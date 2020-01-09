package goBolt

import (
	pool "github.com/jolestar/go-commons-pool"
	"github.com/mindstand/go-bolt/errors"
	"time"
)

type IClient interface {
	// opens a new driver to neo4j
	NewDriver() (IDriver, error)

	// opens a driver pool to neo4j
	NewDriverPool() (IDriverPool, error)

	// opens a v4 driver
	NewDriverV4() (IDriverV4, error)

	// opens a v4 driver pool
	NewDriverPoolV4() (IDriverPoolV4, error)
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
	// pool stuff
	connectionPool pool.ObjectPool
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

	// todo calculate some stuff

	return client, nil
}

func (c *Client) NewDriver() (IDriver, error) {
	panic("implement me")
}

func (c *Client) NewDriverPool() (IDriverPool, error) {
	panic("implement me")
}

func (c *Client) NewDriverV4() (IDriverV4, error) {
	if !c.supportsV4 {
		return nil, errors.Wrap(errors.ErrInvalidVersion, "attempting to use v4 driver when actual version is [%s]", string(c.serverVersion))
	}
	panic("implement me")
}

func (c *Client) NewDriverPoolV4() (IDriverPoolV4, error) {
	if !c.supportsV4 {
		return nil, errors.Wrap(errors.ErrInvalidVersion, "attempting to use v4 driver when actual version is [%s]", string(c.serverVersion))
	}
	panic("implement me")
}




