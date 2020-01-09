package goBolt

import (
	"fmt"
	"github.com/mindstand/go-bolt/errors"
	"time"
)

type IClient interface {
	// opens a new driver to neo4j
	NewDriver() (IDriver, error)

	// opens a driver pool to neo4j
	NewDriverPool(size int) (IDriverPool, error)

	// opens a v4 driver
	NewDriverV4() (IDriverV4, error)

	// opens a v4 driver pool
	NewDriverPoolV4(size int) (IDriverPoolV4, error)
}

type Client struct {
	// config stuff
	connStr          string
	host             string
	port             int
	routing          bool
	pooled           bool
	maxConnections   int
	negotiateVersion bool
	user             string
	password         string
	serverVersion    []byte
	timeout          time.Duration
	chunkSize        uint16
	useTLS           bool
	certFile         string
	caCertFile       string
	keyFile          string
	tlsNoVerify      bool
	readOnly         bool
	supportsV4 bool
	createDbIfNotExists bool
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

		// todo check if neo4j allows passwordless users
		if client.password == "" {
			return nil, errors.Wrap(errors.ErrConfiguration, "password can not be empty")
		}

		client.connStr = fmt.Sprintf("%s://%s:%s@%s:%v", protocol, client.user, client.password, client.host, client.port)

		// append tls portion if needed
		if client.useTLS {
			tlsPortion := fmt.Sprintf("?tls_cert_file=%s&tls_key_file=%s&tls_ca_cert_file=%s&tls_no_verify=%t",
				client.certFile, client.keyFile, client.caCertFile, client.tlsNoVerify)
			client.connStr += tlsPortion
		}
	}

	return client, nil
}

func (c *Client) NewDriver() (IDriver, error) {
	driver := &driver{
		createIfNotExists: c.createDbIfNotExists,
		connectionFactory: &boltConnectionFactory{
			timeout: c.timeout,
			chunkSize: c.chunkSize,
			serverVersion: c.serverVersion,
			connStr: c.connStr,
		},
	}

	return &Driver{internalDriver: driver}, nil
}

func (c *Client) NewDriverPool(size int) (IDriverPool, error) {
	driverPool, err := newDriverPool(c.connStr, size)
	if err != nil {
		return nil, err
	}

	return &DriverPool{
		internalPool: driverPool,
	}, nil
}

func (c *Client) NewDriverV4() (IDriverV4, error) {
	if !c.supportsV4 {
		return nil, errors.Wrap(errors.ErrInvalidVersion, "attempting to use v4 driver when actual version is [%s]", string(c.serverVersion))
	}

	driver := &driver{
		createIfNotExists: c.createDbIfNotExists,
		connectionFactory: &boltConnectionFactory{
			timeout: c.timeout,
			chunkSize: c.chunkSize,
			serverVersion: c.serverVersion,
			connStr: c.connStr,
		},
	}

	return &DriverV4{
		internalDriver: driver,
	}, nil
}

func (c *Client) NewDriverPoolV4(size int) (IDriverPoolV4, error) {
	if !c.supportsV4 {
		return nil, errors.Wrap(errors.ErrInvalidVersion, "attempting to use v4 driver when actual version is [%s]", string(c.serverVersion))
	}
	panic("implement me")
}




