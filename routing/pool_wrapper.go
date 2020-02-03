package routing

import (
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/connection"
)

type connectionPoolWrapper struct {
	Connection      connection.IConnection
	ConnStr         string
	ConnType        bolt_mode.AccessMode
	borrowed        bool
	numBorrows      int
	markForDeletion bool
}
