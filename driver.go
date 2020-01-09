package goBolt

import (
	"time"
)

type Driver struct {
	connectionFactory IBoltConnectionFactory
	connStr string
	timeout time.Duration
	chunkSize uint16
	serverVersion []byte
}

func (d *Driver) Open(mode DriverMode) (IConnection, error) {
	return d.connectionFactory.CreateBoltConnection(d.connStr, d.timeout, d.chunkSize, mode == ReadOnlyMode, d.serverVersion)
}


