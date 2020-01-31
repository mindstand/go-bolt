package goBolt

import (
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/constants"
)

// bolt+routing will not work for non pooled connections
type IDriver interface {
	Open(mode constants.AccessMode) (connection.IConnection, error)
}

type IDriverPool interface {
	// Open opens a Neo-specific connection.
	Open(mode constants.AccessMode) (connection.IConnection, error)
	Reclaim(connection.IConnection) error
	Close() error
}

