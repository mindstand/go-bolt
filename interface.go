package goBolt

import (
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/connection"
)

// bolt+routing will not work for non pooled connections
type IDriver interface {
	Open(mode bolt_mode.AccessMode) (connection.IConnection, error)
}

type IDriverPool interface {
	// Open opens a Neo-specific connection.
	Open(mode bolt_mode.AccessMode) (connection.IConnection, error)
	Reclaim(connection.IConnection) error
	Close() error
}
