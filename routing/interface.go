package routing

import "github.com/mindstand/go-bolt/connection"

type IRoutingPool interface {
	Start() error
	Stop() error

	BorrowRConnection() (connection.IConnection, error)
	BorrowRWConnection() (connection.IConnection, error)

	Reclaim(conn connection.IConnection) error
}
