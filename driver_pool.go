package goBolt

import (
	"context"
	pool "github.com/jolestar/go-commons-pool"
	"github.com/mindstand/go-bolt/errors"
	"sync"
)

type driverPool struct {
	connStr  string
	isV4 bool
	maxConns int
	pool     *pool.ObjectPool
	refLock  sync.Mutex
	closed   bool
}

func newDriverPool(connStr string, maxConns int) (*driverPool, error) {
	dPool := pool.NewObjectPool(context.Background(), pool.NewPooledObjectFactorySimple(getPoolFunc([]string{connStr}, false)), &pool.ObjectPoolConfig{
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

func (d *driverPool) open(db string) (IConnection, error) {
	d.refLock.Lock()
	defer d.refLock.Unlock()
	if !d.closed {
		// verify config
		if d.isV4 && db == "" {
			return nil, errors.Wrap(errors.ErrConfiguration, "must specify db when using v4 connection")
		}

		connObj, err := d.pool.BorrowObject(context.Background())
		if err != nil {
			return nil, err
		}

		conn, ok := connObj.(IConnection)
		if !ok {
			return nil, errors.Wrap(errors.ErrInternal, "cannot cast from [%T] to [IConnection]", connObj)
		}

		//check to make sure the connection is open
		if connectionNilOrClosed(conn) {
			//if it isn't, reset it
			err := conn.initialize()
			if err != nil {
				return nil, err
			}

			conn.setClosed(false)
			conn.setConnErr(err)
			conn.setStatement(nil)
			conn.setTx(nil)
		}

		// handle v4 connection

		return conn, nil
	}
	return nil, errors.New("Driver pool has been closed")
}

func (d *driverPool) close() error {
	d.refLock.Lock()
	defer d.refLock.Unlock()

	if d.closed {
		return errors.Wrap(errors.ErrClosed, "driver pool is already closed")
	}

	if d.pool == nil {
		return errors.Wrap(errors.ErrPool, "connection pool is nil")
	}

	d.pool.Close(context.Background())

	d.closed = true

	return nil
}

func (d *driverPool) reclaim(conn IConnection) error {
	if conn == nil {
		return errors.New("cannot reclaim nil connection")
	}
	if connectionNilOrClosed(conn) {
		err := conn.initialize()
		if err != nil {
			return err
		}

		conn.setClosed(false)
		conn.setConnErr(nil)
		conn.setStatement(nil)
		conn.setTx(nil)
	} else {
		if conn.getStatement() != nil {
			if !conn.getStatement().closed {
				if conn.getStatement().rows != nil && !conn.getStatement().rows.closed {
					err := conn.getStatement().rows.Close()
					if err != nil {
						return err
					}
				}
			}

			conn.setStatement(nil)
		}

		if conn.getTx() != nil {
			if !conn.getTx().IsClosed() {
				err := conn.getTx().Rollback()
				if err != nil {
					return err
				}
			}

			conn.setTx(nil)
		}
	}
	return d.pool.ReturnObject(context.Background(), conn)
}

// ----------------------

type DriverPool struct {
	internalPool *driverPool
}

func (d *DriverPool) Open(mode DriverMode) (IConnection, error) {
	return d.internalPool.open("")
}

func (d *DriverPool) Reclaim(conn IConnection) error {
	return d.internalPool.reclaim(conn)
}

func (d *DriverPool) Close() error {
	return d.internalPool.close()
}

type DriverPoolV4 struct {
	internalPool *driverPool
}

func (d *DriverPoolV4) Open(db string, mode DriverMode) (IConnection, error) {
	return d.internalPool.open(db)
}

func (d *DriverPoolV4) Reclaim(conn IConnection) error {
	return d.internalPool.reclaim(conn)
}

func (d *DriverPoolV4) Close() error {
	return d.internalPool.close()
}

