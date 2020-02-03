package goBolt

import (
	"context"
	pool "github.com/jolestar/go-commons-pool"
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/errors"
	"sync"
)

type driverPool struct {
	connStr  string
	maxConns int
	pool     *pool.ObjectPool
	refLock  sync.Mutex
	closed   bool
}

func newDriverPool(connStr string, maxConns int) (*driverPool, error) {
	dPool := pool.NewObjectPool(context.Background(), &ConnectionPooledObjectFactory{connectionString: connStr}, &pool.ObjectPoolConfig{
		LIFO:                     true,
		MaxTotal:                 maxConns,
		MaxIdle:                  maxConns,
		MinIdle:                  0,
		TestOnCreate:             true,
		TestOnBorrow:             true,
		TestOnReturn:             true,
		TestWhileIdle:            true,
		BlockWhenExhausted:       true,
		MinEvictableIdleTime:     pool.DefaultMinEvictableIdleTime,
		SoftMinEvictableIdleTime: pool.DefaultSoftMinEvictableIdleTime,
		NumTestsPerEvictionRun:   3,
		TimeBetweenEvictionRuns:  0,
	})

	return &driverPool{
		connStr:  connStr,
		maxConns: maxConns,
		pool:     dPool,
	}, nil
}

func (d *driverPool) open() (connection.IConnection, error) {
	d.refLock.Lock()
	defer d.refLock.Unlock()
	if !d.closed {
		connObj, err := d.pool.BorrowObject(context.Background())
		if err != nil {
			return nil, err
		}

		conn, ok := connObj.(connection.IConnection)
		if !ok {
			return nil, errors.Wrap(errors.ErrInternal, "cannot cast from [%T] to [IConnection]", connObj)
		}

		if !conn.ValidateOpen() {
			return nil, errors.New("pool returned dead connection")
		}

		return conn, nil
	}
	return nil, errors.New("Driver pool has been closed")
}

func (d *driverPool) close() error {
	d.refLock.Lock()
	defer d.refLock.Unlock()

	if d.closed {
		return errors.Wrap(errors.ErrClosed, "internalDriver pool is already closed")
	}

	if d.pool == nil {
		return errors.Wrap(errors.ErrPool, "connection pool is nil")
	}

	d.pool.Close(context.Background())

	d.closed = true

	return nil
}

func (d *driverPool) reclaim(conn connection.IConnection) error {
	if conn == nil {
		return errors.New("cannot reclaim nil connection")
	}

	return d.pool.ReturnObject(context.Background(), conn)
}

// ----------------------

type DriverPool struct {
	internalPool *driverPool
}

func (d *DriverPool) Open(mode bolt_mode.AccessMode) (connection.IConnection, error) {
	return d.internalPool.open()
}

func (d *DriverPool) Reclaim(conn connection.IConnection) error {
	return d.internalPool.reclaim(conn)
}

func (d *DriverPool) Close() error {
	return d.internalPool.close()
}
