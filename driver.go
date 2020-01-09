package goBolt

import (
	"github.com/mindstand/go-bolt/errors"
	"time"
)

type Driver struct {
	connStr       string
	host string
	port int
	routing bool
	negotiateVersion bool
	user          string
	password      string
	serverVersion []byte
	timeout       time.Duration
	chunkSize     uint16
	useTLS        bool
	certFile      string
	caCertFile    string
	keyFile       string
	tlsNoVerify   bool
	readOnly      bool
}

func NewDriver(opts ...Opt) (*Driver, error) {
	if len(opts) == 0 {
		return nil, errors.Wrap(errors.ErrConfiguration, "no options for driver")
	}

	driver := new(Driver)

	for _, opt := range opts {
		if opt == nil {
			return nil, errors.Wrap(errors.ErrConfiguration, "found nil option function in new driver")
		}

		err := opt(driver)
		if err != nil {
			return nil, errors.Wrap(errors.ErrConfiguration, err.Error())
		}
	}

	return driver, nil
}