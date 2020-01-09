package goBolt

import (
	"github.com/mindstand/go-bolt/errors"
	"time"
)

type Opt func(*Client) error

// WithConnectionString provides client option with connection string
func WithConnectionString(connString string) Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		if client.host != "" || client.port != 0 {
			return errors.Wrap(errors.ErrConfiguration, "can not call WithConnectionString and WithHostPort")
		}

		client.connStr = connString
		return nil
	}
}

// allows setting the host and port of neo4j
func WithHostPort(host string, port int) Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		if client.connStr != "" {
			return errors.Wrap(errors.ErrConfiguration, "can not call WithHostPort and WithConnectionString")
		}

		client.host = host
		client.port = port
		return nil
	}
}

// allows setting protocol to bolt+routing
func WithRouting() Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		if client.connStr != "" {
			return errors.Wrap(errors.ErrConfiguration, "can not call WithRouting and WithConnectionString")
		}

		client.routing = true
		return nil
	}
}

// allows setting chunk size
func WithChunkSize(size uint16) Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		client.chunkSize = size
		return nil
	}
}

// allows authentication with basic auth
func WithBasicAuth(username, password string) Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		client.user = username
		client.password = password

		return nil
	}
}

// allows authentication with tls
func WithTLS(cacertPath, certPath, keyPath string, tlsNoVerify bool) Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		client.caCertFile = cacertPath
		client.certFile = certPath
		client.keyFile = keyPath
		client.useTLS = true
		client.tlsNoVerify = tlsNoVerify
		return nil
	}
}

// tells client to negotiate version
func WithAPIVersionNegotiation() Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		client.negotiateVersion = true
		return nil
	}
}

// tells client what timeout it should use
func WithTimeout(timeout time.Duration) Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		client.timeout = timeout
		return nil
	}
}

// tells client whether it should be in readonly mode
func WithReadonly() Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		client.readOnly = true

		return nil
	}
}

// tells client which bolt version to use
func WithVersion(version string) Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		if client.negotiateVersion {
			return errors.Wrap(errors.ErrConfiguration, "can not set client version and negotiate version")
		}

		client.serverVersion = []byte(version)

		return nil
	}
}

func WithPool(max int) Opt {
	return func(client *Client) error {
		if client == nil {
			return errors.Wrap(errors.ErrConfiguration, "client can not be nil")
		}

		if max < 1 {
			return errors.Wrap(errors.ErrConfiguration, "client pool has to have at least 1 connection")
		}

		client.pooled = true
		client.maxConnections = max
		return nil
	}
}
