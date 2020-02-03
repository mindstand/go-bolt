package goBolt

import (
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/routing"
	"time"
)

const refreshInterval = time.Minute * 5

type RoutingDriverPool struct {
	internalPool routing.IRoutingPool
}

func newRoutingPool(client *Client, size int) (*RoutingDriverPool, error) {
	if client == nil {
		return nil, errors.New("client can not be nil")
	}

	internalPool, err := routing.NewRoutingPool(client.connStr, size, refreshInterval)
	if err != nil {
		return nil, err
	}

	err = internalPool.Start()
	if err != nil {
		return nil, err
	}

	return &RoutingDriverPool{internalPool: internalPool}, nil
}

func (r *RoutingDriverPool) Open(mode bolt_mode.AccessMode) (connection.IConnection, error) {
	if mode == bolt_mode.ReadMode {
		return r.internalPool.BorrowRConnection()
	} else {
		return r.internalPool.BorrowRWConnection()
	}
}

func (r *RoutingDriverPool) Reclaim(conn connection.IConnection) error {
	return r.internalPool.Reclaim(conn)
}

func (r *RoutingDriverPool) Close() error {
	return r.internalPool.Stop()
}
