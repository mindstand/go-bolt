package goBolt

import (
	"github.com/mindstand/go-bolt/errors"
	"time"
)

type Opt func(*Driver) error

// WithConnectionString provides driver option with connection string
func WithConnectionString(connString string) Opt {
	return func(driver *Driver) error {
		if driver == nil {
			return errors.Wrap(errors.ErrConfiguration, "driver can not be nil")
		}

		if driver.host != "" || driver.port != 0 {
			return errors.Wrap(errors.ErrConfiguration, "can not call WithConnectionString and WithHostPort")
		}

		driver.connStr = connString
		return nil
	}
}

// allows setting the host and port of neo4j
func WithHostPort(host string, port int) Opt {
	return func(driver *Driver) error {
		if driver == nil {
			return errors.Wrap(errors.ErrConfiguration, "driver can not be nil")
		}

		if driver.connStr != "" {
			return errors.Wrap(errors.ErrConfiguration, "can not call WithHostPort and WithConnectionString")
		}

		driver.host = host
		driver.port = port
		return nil
	}
}

// allows setting protocol to bolt+routing
func WithRouting() Opt {
	return func(driver *Driver) error {
		if driver == nil {
			return errors.Wrap(errors.ErrConfiguration, "driver can not be nil")
		}

		if driver.connStr != "" {
			return errors.Wrap(errors.ErrConfiguration, "can not call WithRouting and WithConnectionString")
		}

		driver.routing = true
		return nil
	}
}

// allows setting chunk size
func WithChunkSize(size uint16) Opt {
	return func(driver *Driver) error {
		if driver == nil {
			return errors.Wrap(errors.ErrConfiguration, "driver can not be nil")
		}

		driver.chunkSize = size
		return nil
	}
}

// allows authentication with basic auth
func WithBasicAuth(username, password string) Opt {
	return func(driver *Driver) error {
		if driver == nil {
			return errors.Wrap(errors.ErrConfiguration, "driver can not be nil")
		}

		driver.user = username
		driver.password = password

		return nil
	}
}

// allows authentication with tls
func WithTLS(cacertPath, certPath, keyPath string, tlsNoVerify bool) Opt {
	return func(driver *Driver) error {
		if driver == nil {
			return errors.Wrap(errors.ErrConfiguration, "driver can not be nil")
		}

		driver.caCertFile = cacertPath
		driver.certFile = certPath
		driver.keyFile = keyPath
		driver.useTLS = true
		driver.tlsNoVerify = tlsNoVerify
		return nil
	}
}

// tells driver to negotiate version
func WithAPIVersionNegotiation() Opt {
	return func(driver *Driver) error {
		if driver == nil {
			return errors.Wrap(errors.ErrConfiguration, "driver can not be nil")
		}

		driver.negotiateVersion = true
		return nil
	}
}

// tells driver what timeout it should use
func WithTimeout(timeout time.Duration) Opt {
	return func(driver *Driver) error {
		if driver == nil {
			return errors.Wrap(errors.ErrConfiguration, "driver can not be nil")
		}

		driver.timeout = timeout
		return nil
	}
}

// tells driver whether it should be in readonly mode
func WithReadonly() Opt {
	return func(driver *Driver) error {
		if driver == nil {
			return errors.Wrap(errors.ErrConfiguration, "driver can not be nil")
		}

		driver.readOnly = true

		return nil
	}
}

// tells driver which bolt version to use
func WithVersion(version string) Opt {
	return func(driver *Driver) error {
		if driver == nil {
			return errors.Wrap(errors.ErrConfiguration, "driver can not be nil")
		}

		if driver.negotiateVersion {
			return errors.Wrap(errors.ErrConfiguration, "can not set driver version and negotiate version")
		}

		driver.serverVersion = []byte(version)

		return nil
	}
}
