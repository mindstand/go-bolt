package goBolt

import (
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/constants"
)

type internalDriver struct {
	client *Client
}

// standard driver is basically a factory
// its not keeping track of connections
// connections are expected to be killed when done
type Driver struct {
	internalDriver *internalDriver
}

// mode doesn't matter since its not a pooled or routing driver
func (d *Driver) Open(mode constants.AccessMode) (connection.IConnection, error) {
	return connection.CreateBoltConn(d.internalDriver.client.connStr)
}
