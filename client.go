package goBolt

import (
	"fmt"
	"github.com/mindstand/go-bolt/errors"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type IClient interface {
	// opens a new internalDriver to neo4j
	NewDriver() (IDriver, error)

	// opens a internalDriver pool to neo4j
	NewDriverPool(size int) (IDriverPool, error)
}

type Client struct {
	// config stuff
	connStr             string
	host                string
	port                int
	routing             bool
	pooled              bool
	maxConnections      int
	negotiateVersion    bool
	user                string
	password            string
	serverVersionBytes  []byte
	serverVersion       int
	timeout             time.Duration
	chunkSize           uint16
	useTLS              bool
	certFile            string
	caCertFile          string
	keyFile             string
	tlsNoVerify         bool
	protocolGreaterThan int
	protocolLessThan    int
}

func NewClient(opts ...Opt) (IClient, error) {
	if len(opts) == 0 {
		return nil, errors.Wrap(errors.ErrConfiguration, "no options for client")
	}

	client := new(Client)

	for _, opt := range opts {
		if opt == nil {
			return nil, errors.Wrap(errors.ErrConfiguration, "found nil option function in new client")
		}

		err := opt(client)
		if err != nil {
			return nil, errors.Wrap(errors.ErrConfiguration, err.Error())
		}
	}

	// timeout not set
	if client.timeout == 0 {
		client.timeout = time.Second * time.Duration(60)
	}

	// check chunk size
	if client.chunkSize == 0 {
		// set default chunk size
		client.chunkSize = math.MaxUint16
	}

	// figure out the connection string
	if client.connStr == "" {
		var protocol string

		// figure out schema
		if client.routing {
			protocol = "bolt+routing"
		} else {
			protocol = "bolt"
		}

		// validate other stuff
		if client.host == "" {
			return nil, errors.Wrap(errors.ErrConfiguration, "host can not be empty")
		}

		if client.port <= 0 {
			return nil, errors.Wrap(errors.ErrConfiguration, "invalid port [%v]", client.port)
		}

		if client.user == "" {
			return nil, errors.Wrap(errors.ErrConfiguration, "user can not be empty")
		}

		client.connStr = fmt.Sprintf("%s://%s:%s@%s:%v", protocol, client.user, client.password, client.host, client.port)

		// append tls portion if needed
		if client.useTLS {
			tlsPortion := fmt.Sprintf("?tls_cert_file=%s&tls_key_file=%s&tls_ca_cert_file=%s&tls_no_verify=%t",
				client.certFile, client.keyFile, client.caCertFile, client.tlsNoVerify)
			client.connStr += tlsPortion
		}
	} else {
		client.routing = strings.Contains(client.connStr, "+routing")
		_url, err := url.Parse(client.connStr)
		if err != nil {
			return nil, fmt.Errorf("an error occurred parsing bolt URL, %w", err)
		} else if strings.ToLower(_url.Scheme) != "bolt" && strings.ToLower(_url.Scheme) != "bolt+routing" {
			return nil, fmt.Errorf("unsupported connection string scheme: %s. Driver only supports 'bolt' and 'bolt+routing' scheme", _url.Scheme)
		}

		hostPort := _url.Host
		if strings.Contains(hostPort, ":") {
			parts := strings.Split(hostPort, ":")
			if len(parts) != 2 {
				return nil, errors.New("host/port invalid")
			}
			client.host = parts[0]
			_port, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}

			client.port = int(_port)
		}

		if _url.User != nil {
			client.user = _url.User.Username()
			var isSet bool
			client.password, isSet = _url.User.Password()
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

			client.timeout = time.Duration(timeoutInt) * time.Second
		}

		useTLS := _url.Query().Get("tls")
		client.useTLS = strings.HasPrefix(strings.ToLower(useTLS), "t") || useTLS == "1"

		if client.useTLS {
			client.certFile = _url.Query().Get("tls_cert_file")
			client.keyFile = _url.Query().Get("tls_key_file")
			client.caCertFile = _url.Query().Get("tls_ca_cert_file")
			noVerify := _url.Query().Get("tls_no_verify")
			client.tlsNoVerify = strings.HasPrefix(strings.ToLower(noVerify), "t") || noVerify == "1"
		}
	}

	return client, nil
}

func (c *Client) getTlsPortion() (string, error) {
	if c.certFile != "" {
		return fmt.Sprintf("?tls_cert_file=%s&tls_key_file=%s&tls_ca_cert_file=%s&tls_no_verify=%t",
			c.certFile, c.keyFile, c.caCertFile, c.tlsNoVerify), nil
	} else {
		// parse connection string
		u, err := url.Parse(c.connStr)
		if err != nil {
			return "", nil
		}

		certFile := u.Query().Get("tls_cert_file")
		keyFile := u.Query().Get("tls_key_file")
		caCertFile := u.Query().Get("tls_ca_cert_file")
		tlsNoVerify := u.Query().Get("tls_no_verify")

		if certFile == "" {
			return "", nil
		}

		return fmt.Sprintf("?tls_cert_file=%s&tls_key_file=%s&tls_ca_cert_file=%s&tls_no_verify=%s",
			certFile, keyFile, caCertFile, tlsNoVerify), nil
	}
}

func (c *Client) getUsernamePassword() (string, error) {
	u, err := url.Parse(c.connStr)
	if err != nil {
		return "", nil
	}

	if u.User.Username() == "" {
		return "", nil
	}

	pwd, ok := u.User.Password()
	if !ok {
		pwd = ""
	}

	return fmt.Sprintf("%s:%s", u.User.Username(), pwd), nil
}

func (c *Client) NewDriver() (IDriver, error) {
	if c.routing {
		return nil, errors.New("can not open non pooled driver with routing enabled")
	}

	driver := &internalDriver{
		client: c,
	}

	return &Driver{internalDriver: driver}, nil
}

func (c *Client) NewDriverPool(size int) (IDriverPool, error) {
	driverPool, err := newDriverPool(c.connStr, size)
	if err != nil {
		return nil, err
	}

	if c.routing {
		return newRoutingPool(c, size)
	} else {
		return &DriverPool{
			internalPool: driverPool,
		}, nil
	}
}
