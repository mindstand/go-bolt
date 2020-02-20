package connection

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"database/sql/driver"
	"fmt"
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/log"
	"github.com/mindstand/go-bolt/protocol"
	"github.com/mindstand/go-bolt/structures"
	"github.com/mindstand/go-bolt/structures/messages"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type readWrite struct {
	connection *Connection
}

func (r *readWrite) Write(p []byte) (n int, err error) {
	if err := r.connection.conn.SetWriteDeadline(time.Now().Add(r.connection.timeout)); err != nil {
		//c.connErr = errors.Wrap(err, "An error occurred setting write deadline")
		return 0, driver.ErrBadConn
	}

	n, err = r.connection.conn.Write(p)

	if err != nil {
		err = driver.ErrBadConn
	}
	return n, err
}

func (r *readWrite) Read(p []byte) (n int, err error) {
	if err := r.connection.conn.SetReadDeadline(time.Now().Add(r.connection.timeout)); err != nil {
		return 0, driver.ErrBadConn
	}

	n, err = r.connection.conn.Read(p)
	if err != nil && err != io.EOF {
		err = driver.ErrBadConn
	}
	return n, err
}

type Connection struct {
	boltProtocol         protocol.IBoltProtocol
	protocolVersion      int
	protocolVersionBytes []byte

	// connection information
	user     string
	password string
	hostPort string

	// tls information
	useTLS      bool
	certFile    string
	caCertFile  string
	keyFile     string
	tlsNoVerify bool

	// connection config
	accessMode bolt_mode.AccessMode

	// handlers
	readWrite   *readWrite
	transaction ITransaction
	openRows    IRows

	// connection stuff
	timeout   time.Duration
	chunkSize uint16
	conn      net.Conn
	closed    bool
	openQuery bool

	// for pool tracking
	id string
}

func CreateBoltConn(connStr string) (IConnection, error) {
	conn, err := newConnectionFromConnectionString(connStr)
	if err != nil {
		return nil, err
	}

	err = conn.initialize()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func newConnectionFromConnectionString(connStr string) (*Connection, error) {
	_url, err := url.Parse(connStr)
	if err != nil {
		return nil, fmt.Errorf("an error occurred parsing bolt URL, %w", err)
	} else if strings.ToLower(_url.Scheme) != "bolt" && strings.ToLower(_url.Scheme) != "bolt+routing" {
		return nil, fmt.Errorf("unsupported connection string scheme: %s. Driver only supports 'bolt' and 'bolt+routing' scheme", _url.Scheme)
	}

	connection := Connection{
		timeout:   time.Second * time.Duration(60),
		chunkSize: math.MaxUint16,
	}

	connection.hostPort = _url.Host

	if _url.User != nil {
		connection.user = _url.User.Username()
		var isSet bool
		connection.password, isSet = _url.User.Password()
		if !isSet {
			return nil, errors.New("must specify password when passing user")
		}
	}

	timeout := _url.Query().Get("timeout")
	if timeout != "" {
		timeoutInt, err := strconv.Atoi(timeout)
		if err != nil {
			return nil, fmt.Errorf("Invalid format for timeout: %s.  Must be integer", timeout)
		}

		connection.timeout = time.Duration(timeoutInt) * time.Second
	}

	useTLS := _url.Query().Get("tls")
	connection.useTLS = strings.HasPrefix(strings.ToLower(useTLS), "t") || useTLS == "1"

	if connection.useTLS {
		connection.certFile = _url.Query().Get("tls_cert_file")
		connection.keyFile = _url.Query().Get("tls_key_file")
		connection.caCertFile = _url.Query().Get("tls_ca_cert_file")
		noVerify := _url.Query().Get("tls_no_verify")
		connection.tlsNoVerify = strings.HasPrefix(strings.ToLower(noVerify), "t") || noVerify == "1"
	}

	return &connection, nil
}

func (c *Connection) GetConnectionId() string {
	return c.id
}

func (c *Connection) SetConnectionId(id string) {
	c.id = id
}

func (c *Connection) GetProtocolVersionNumber() int {
	return c.protocolVersion
}

func (c *Connection) GetProtocolVersionBytes() []byte {
	return c.protocolVersionBytes
}

// todo better errors (wrap stuff)
func (c *Connection) createConnection() error {
	var conn net.Conn
	var err error
	if c.useTLS {
		config, err := c.tlsConfig()
		if err != nil {
			return err
		}

		conn, err = tls.Dial("tcp", c.hostPort, config)
		if err != nil {
			return err
		}
	} else {
		conn, err = net.DialTimeout("tcp", c.hostPort, c.timeout)
		if err != nil {
			return err

		}
	}

	c.conn = conn
	c.closed = false
	return nil
}

func (c *Connection) tlsConfig() (*tls.Config, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS10,
		MaxVersion: tls.VersionTLS12,
	}

	if c.caCertFile != "" {
		// Load CA cert - usually for self-signed certificates
		caCert, err := ioutil.ReadFile(c.caCertFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		config.RootCAs = caCertPool
	}

	if c.certFile != "" {
		if c.keyFile == "" {
			return nil, errors.New("If you're providing a cert file, you must also provide a key file")
		}

		cert, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
		if err != nil {
			return nil, err
		}

		config.Certificates = []tls.Certificate{cert}
	}

	if c.tlsNoVerify {
		config.InsecureSkipVerify = true
	}

	return config, nil
}

func (c *Connection) handshake() ([]byte, error) {
	numWritten, err := c.readWrite.Write(handShake)
	if err != nil {
		return nil, err
	} else if numWritten != 20 {
		return nil, fmt.Errorf("num written bytes [%v] is not equal to expected [20]", numWritten)
	}

	version := make([]byte, 4, 4)
	numRead, err := c.readWrite.Read(version)
	if err != nil {
		return nil, err
	} else if numRead != 4 {
		return nil, fmt.Errorf("expected [4] bytes but got [%v]", numRead)
	} else if bytes.Equal(version, noVersionSupported) {
		return nil, errors.New("server responded with no supported version")
	}

	return version, nil
}

func (c *Connection) initialize() error {
	err := c.createConnection()
	if err != nil {
		return err
	}

	c.readWrite = &readWrite{
		connection: c,
	}

	versionBytes, err := c.handshake()
	if err != nil {
		return fmt.Errorf("handshake failed, %w", err)
	}

	boltProtocol, version, err := protocol.GetProtocol(versionBytes)
	if err != nil {
		return err
	}

	log.Tracef("Using protocol version %v", version)

	c.protocolVersion = version
	c.protocolVersionBytes = versionBytes
	c.boltProtocol = boltProtocol

	return c.sendInit(c.boltProtocol.GetInitMessage(ClientID, messages.BuildAuthTokenBasic(c.user, c.password)))
}

func (c *Connection) sendInit(message structures.Structure) error {
	err := c.sendMessage(message)
	if err != nil {
		return err
	}

	respMsg, err := c.consume()
	if err != nil {
		return err
	}

	switch resp := respMsg.(type) {
	case messages.SuccessMessage:
		log.Infof("Successfully initiated Bolt connection: %+v", resp)
		return nil
	default:
		log.Errorf("Got an unrecognized message when initializing connection :%+v", resp)
		return c.Close()
	}
}

func (c *Connection) ValidateOpen() bool {
	if c.closed {
		return false
	}

	if c.conn == nil {
		return false
	}

	// todo more checks to validate that this connection still works

	return true
}

// Sets the size of the chunks to write to the stream
func (c *Connection) SetChunkSize(chunkSize uint16) {
	c.chunkSize = chunkSize
}

// Sets the timeout for reading and writing to the stream
func (c *Connection) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

func (c *Connection) Exec(query string, params QueryParams) (IResult, error) {
	return c.ExecWithDb(query, params, "")
}

func (c *Connection) ExecWithDb(query string, params QueryParams, db string) (IResult, error) {
	if !c.boltProtocol.SupportsMultiDatabase() && db != "" {
		return nil, fmt.Errorf("bolt protocol version [%v] does not have multi database support", c.protocolVersion)
	}

	success, err := c.runQuery(query, params, db, false, true)
	if err != nil {
		return nil, err
	}

	return newBoltResult(success.Metadata), nil
}

func (c *Connection) Query(query string, params QueryParams) (IRows, error) {
	return c.QueryWithDb(query, params, "")
}

func (c *Connection) QueryWithDb(query string, params QueryParams, db string) (IRows, error) {
	if !c.boltProtocol.SupportsMultiDatabase() && db != "" {
		return nil, fmt.Errorf("bolt protocol version [%v] does not have multi database support", c.protocolVersion)
	}

	success, err := c.runQuery(query, params, db, false, false)
	if err != nil {
		return nil, err
	}

	return newQueryRows(c, success.Metadata, c.boltProtocol.GetResultAvailableAfterKey(), c.boltProtocol.GetResultConsumedAfterKey()), nil
}

func (c *Connection) runQuery(query string, params QueryParams, dbName string, inTx, isExec bool) (*messages.SuccessMessage, error) {
	if c.openQuery {
		return nil, errors.New("runQuery already open")
	}

	if c.closed {
		return nil, errors.New("connection already closed")
	}

	err := c.sendMessage(c.boltProtocol.GetRunMessage(query, params, dbName, c.accessMode, !inTx))
	if err != nil {
		return nil, err
	}

	err = c.sendMessage(c.boltProtocol.GetPullAllMessage())
	if err != nil {
		return nil, err
	}

	resp, err := c.consume()
	if err != nil {
		return nil, err
	}

	success, ok := resp.(messages.SuccessMessage)
	if !ok {
		return nil, fmt.Errorf("unexpected response of type [%T], should be [messages.SuccessMessage]", resp)
	}

	// we dont care about what we're consuming next, so flush until we hit another success (this is identical to how rows are treated)
	if isExec {
		for {
			_resp, err := c.consume()
			if err != nil {
				return nil, err
			}

			switch _resp.(type) {
			case messages.SuccessMessage:
				success, ok := _resp.(messages.SuccessMessage)
				if !ok {
					return nil, fmt.Errorf("unexpected response of type [%T], should be [messages.SuccessMessage]", resp)
				}
				return &success, nil
			default:
				continue
			}
		}
	}

	return &success, nil
}

func (c *Connection) Close() error {
	if c.closed {
		return errors.New("connection is ready closed")
	}

	if c.transaction != nil {
		err := c.transaction.Rollback()
		if err != nil {
			return err
		}
	}

	if c.conn == nil {
		return errors.New("can not close nil transaction")
	}

	if msg, ok := c.boltProtocol.GetCloseMessage(); ok {
		err := c.sendMessage(msg)
		if err != nil {
			return err
		}

		// explicitly not consuming since we're closing the connection
	}

	err := c.conn.Close()
	c.closed = true
	if err != nil {
		return err
	}

	return nil
}

func (c *Connection) MakeIdle() error {
	if c.transaction != nil {
		return c.transaction.Rollback()
	}

	return nil
}

func (c *Connection) BeginWithDatabase(db string) (ITransaction, error) {
	if c.transaction != nil {
		return nil, errors.New("transaction already open")
	}

	if c.closed {
		return nil, errors.New("can not open transaction on closed connection")
	}

	msg := c.boltProtocol.GetTxBeginMessage(db, c.accessMode)

	_, isBeginMsg := msg.(messages.BeginMessage)

	// send BEGIN
	err := c.sendMessage(msg)
	if err != nil {
		return nil, err
	}

	if !isBeginMsg {
		err = c.sendMessage(c.boltProtocol.GetPullAllMessage())
		if err != nil {
			return nil, err
		}
	}

	runSucc, err := c.consume()
	if err != nil {
		return nil, err
	}

	var pullSucc interface{}
	if !isBeginMsg {
		pullSucc, err = c.consume()
		if err != nil {
			return nil, err
		}
	}

	success, ok := runSucc.(messages.SuccessMessage)
	if !ok {
		return nil, errors.New("Unrecognized response type beginning transaction: %#v", success)
	}

	if !isBeginMsg {
		pull, ok := pullSucc.(messages.SuccessMessage)
		if !ok {
			return nil, errors.New("Unrecognized response beginning transaction:  %#v", pull)
		}
	}

	c.transaction = &boltTransaction{
		conn:   c,
		closed: false,
	}

	return c.transaction, nil
}

func (c *Connection) Begin() (ITransaction, error) {
	return c.BeginWithDatabase("")
}

func (c *Connection) sendMessage(message structures.Structure) error {
	if message == nil {
		return errors.New("message can not be nil")
	}
	return c.boltProtocol.NewEncoder(c.readWrite, c.chunkSize).Encode(message)
}

func (c *Connection) sendMessageConsume(message structures.Structure) (interface{}, error) {
	err := c.sendMessage(message)
	if err != nil {
		return nil, err
	}

	return c.consume()
}

func (c *Connection) consume() (interface{}, error) {
	log.Trace("Consuming response from bolt stream")

	respInt, err := c.boltProtocol.NewDecoder(c.readWrite).Decode()
	if err != nil {
		return respInt, err
	}

	if log.GetLevel() >= log.TraceLevel {
		log.Tracef("Consumed Response: %#v", respInt)
	}

	if failure, isFail := respInt.(messages.FailureMessage); isFail {
		log.Errorf("Got failure message: %#v", failure)
		err := c.ackFailure(failure)
		if err != nil {
			return nil, err
		}
		return failure, errors.Wrap(failure, "Neo4J reported a failure for the runQuery")
	}

	return respInt, err
}

func (c *Connection) ackFailure(failure messages.FailureMessage) error {
	log.Tracef("Acknowledging Failure: %#v", failure)

	ack := messages.NewAckFailureMessage()
	err := c.boltProtocol.NewEncoder(c.readWrite, c.chunkSize).Encode(ack)
	if err != nil {
		return errors.Wrap(err, "An error occurred encoding ack failure message")
	}

	for {
		respInt, err := c.boltProtocol.NewDecoder(c.readWrite).Decode()
		if err != nil {
			return errors.Wrap(err, "An error occurred decoding ack failure message response")
		}

		switch resp := respInt.(type) {
		case messages.IgnoredMessage:
			log.Tracef("Got ignored message when acking failure: %#v", resp)
			continue
		case messages.SuccessMessage:
			log.Tracef("Got success message when acking failure: %#v", resp)
			return nil
		case messages.FailureMessage:
			log.Errorf("Got failure message when acking failure: %#v", resp)
			return c.reset()
		default:
			log.Errorf("Got unrecognized response from acking failure: %#v", resp)
			return c.Close()
		}
	}
}

func (c *Connection) reset() error {
	log.Trace("Resetting session")

	reset := messages.NewResetMessage()
	err := c.boltProtocol.NewEncoder(c.readWrite, c.chunkSize).Encode(reset)
	if err != nil {
		return errors.Wrap(err, "An error occurred encoding reset message")
	}

	for {
		respInt, err := c.boltProtocol.NewDecoder(c.readWrite).Decode()
		if err != nil {
			return errors.Wrap(err, "An error occurred decoding reset message response")
		}

		switch resp := respInt.(type) {
		case messages.IgnoredMessage:
			log.Tracef("Got ignored message when resetting session: %#v", resp)
			continue
		case messages.SuccessMessage:
			log.Tracef("Got success message when resetting session: %#v", resp)
			return nil
		case messages.FailureMessage:
			log.Errorf("Got failure message when resetting session: %#v", resp)
			err = c.Close()
			if err != nil {
				log.Errorf("An error occurred closing the session: %s", err)
			}
			return errors.Wrap(resp, "Error resetting session. CLOSING SESSION!")
		default:
			log.Errorf("Got unrecognized response from resetting session: %#v", resp)
			c.Close()
			return driver.ErrBadConn
		}
	}
}
