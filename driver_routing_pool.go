package goBolt

import (
	"context"
	"fmt"
	pool "github.com/jolestar/go-commons-pool"
	"github.com/mindstand/go-bolt/errors"
	"net/url"
	"strings"
	"sync"
)

type routingDriverPool struct {
	connStr  string
	maxConns int
	refLock  sync.Mutex
	closed   bool

	username string
	password string

	client *Client

	//compose configuration
	config *clusterConnectionConfig

	poolFunc func(context.Context) (interface{}, error)

	//write resources
	writeConns int
	writePool  *pool.ObjectPool

	//read resources
	readConns int
	readPool  *pool.ObjectPool
}

func newRoutingDriverPool(connStr string, max int, poolFunc func(context.Context) (interface{}, error)) (*routingDriverPool, error) {
	if max < 2 {
		return nil, errors.Wrap(errors.ErrConfiguration, "max must be at least 2")
	}

	var writeConns int
	var readConns int

	//max conns is even
	if max%2 == 0 {
		writeConns = max / 2
		readConns = max / 2
	} else {
		//if given odd, make more write conns
		c := (max - 1) / 2
		writeConns = c + 1
		readConns = c
	}

	u, err := url.Parse(connStr)
	if err != nil {
		return nil, err
	}

	pwd, ok := u.User.Password()
	if !ok {
		pwd = ""
	}

	d := &routingDriverPool{
		//main stuff
		maxConns: max,
		connStr:  connStr,
		username: u.User.Username(),
		password: pwd,
		poolFunc: poolFunc,
		//write
		writeConns: writeConns,
		//read
		readConns: readConns,
	}

	err = d.refreshConnectionPool()
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (b *routingDriverPool) refreshConnectionPool() error {
	//get the connection info
	var clusterInfoConn IConnection

	if b.client.supportsV4 {
		clusterInfoDriver, err := b.client.NewDriverV4()
		if err != nil {
			return err
		}

		clusterInfoConn, err = clusterInfoDriver.Open("system", ReadWriteMode)
		if err != nil {
			return err
		}
	} else {
		clusterInfoDriver, err := b.client.NewDriver()
		if err != nil {
			return err
		}

		clusterInfoConn, err = clusterInfoDriver.Open(ReadWriteMode)
		if err != nil {
			return err
		}
	}

	clusterInfo, err := getClusterInfo(clusterInfoConn)
	if err != nil {
		return err
	}

	b.config = clusterInfo

	//close original internalDriver
	err = clusterInfoConn.Close()
	if err != nil {
		return err
	}

	writeConnStr := ""

	for _, l := range b.config.Leaders {
		for _, a := range l.Addresses {
			if strings.Contains(a, "bolt") {
				//retain login info
				if b.username != "" {
					u, err := url.Parse(a)
					if err != nil {
						return err
					}

					pwdStr := ""
					if b.password != "" {
						pwdStr = fmt.Sprintf(":%s@", b.password)
					}

					writeConnStr = fmt.Sprintf("bolt://%s%s%s:%s", b.username, pwdStr, u.Hostname(), u.Port())
					break
				}

				writeConnStr = a
				break
			}
		}
	}
	//b.writePool
	writeFactory := pool.NewPooledObjectFactorySimple(getPoolFunc([]string{writeConnStr}, false))

	b.writePool = pool.NewObjectPool(context.Background(), writeFactory, &pool.ObjectPoolConfig{
		LIFO:                     true,
		MaxTotal:                 b.writeConns,
		MaxIdle:                  b.writeConns,
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

	var readUris []string
	//parse followers
	for _, follow := range b.config.Followers {
		for _, a := range follow.Addresses {
			if strings.Contains(a, "bolt") {
				if b.username != "" {
					u, err := url.Parse(a)
					if err != nil {
						return err
					}

					pwdStr := ""
					if b.password != "" {
						pwdStr = fmt.Sprintf(":%s@", b.password)
					}

					readUris = append(readUris, fmt.Sprintf("bolt://%s%s%s:%s", b.username, pwdStr, u.Hostname(), u.Port()))
					break
				}
				readUris = append(readUris, a)
				break
			}
		}
	}

	//parse replicas
	for _, replica := range b.config.ReadReplicas {
		for _, a := range replica.Addresses {
			if strings.Contains(a, "bolt") {
				if b.username != "" {
					u, err := url.Parse(a)
					if err != nil {
						return err
					}

					pwdStr := ""
					if b.password != "" {
						pwdStr = fmt.Sprintf(":%s@", b.password)
					}

					readUris = append(readUris, fmt.Sprintf("bolt://%s%s%s:%s", b.username, pwdStr, u.Hostname(), u.Port()))
					break
				}
				readUris = append(readUris, a)
				break
			}
		}
	}

	if readUris == nil || len(readUris) == 0 {
		return errors.New("no read nodes to connect to")
	}

	readFactory := pool.NewPooledObjectFactorySimple(getPoolFunc(readUris, true))
	b.readPool = pool.NewObjectPool(context.Background(), readFactory, &pool.ObjectPoolConfig{
		LIFO:                     true,
		MaxTotal:                 b.readConns,
		MaxIdle:                  b.readConns,
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

	return nil
}

func (b *routingDriverPool) close() error {
	b.refLock.Lock()
	defer b.refLock.Unlock()

	b.writePool.Close(context.Background())
	b.readPool.Close(context.Background())

	b.closed = true
	return nil
}

func (b *routingDriverPool) open(db string, mode DriverMode) (IConnection, error) {
	// For each connection request we need to block in case the Close function is called. This gives us a guarantee
	// when closing the pool no new connections are made.
	b.refLock.Lock()
	defer b.refLock.Unlock()
	ctx := context.Background()
	if !b.closed {
		var conn IConnection
		var ok bool

		switch mode {
		case ReadOnlyMode:
			connObj, err := b.readPool.BorrowObject(ctx)
			if err != nil {
				return nil, err
			}

			conn, ok = connObj.(IConnection)
			if !ok {
				return nil, errors.New("unable to cast to *BoltConn")
			}
			break
		case ReadWriteMode:
			connObj, err := b.writePool.BorrowObject(ctx)
			if err != nil {
				return nil, err
			}

			conn, ok = connObj.(IConnection)
			if !ok {
				return nil, errors.New("unable to cast to *BoltConn")
			}
			break
		default:
			return nil, errors.New("invalid internalDriver mode")
		}

		//check to make sure the connection is open
		if connectionNilOrClosed(conn) {
			//if it isn't, reset it
			err := conn.initialize()
			if err != nil {
				return nil, err
			}

			conn.setClosed(false)
			conn.setConnErr(nil)
			conn.setStatement(nil)
			conn.setTx(nil)
		}

		return conn, nil
	}
	return nil, errors.New("Driver pool has been closed")
}

func (b *routingDriverPool) reclaim(conn IConnection) error {
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

	if conn.getReadOnly() {
		return b.readPool.ReturnObject(context.Background(), conn)
	} else {
		return b.writePool.ReturnObject(context.Background(), conn)
	}
}

// ------------------------------
// heres the interface impls, the core impl is above

type RoutingDriverPool struct {
	internalPool *routingDriverPool
}

func (r *RoutingDriverPool) Open(mode DriverMode) (IConnection, error) {
	return r.internalPool.open("", mode)
}

func (r *RoutingDriverPool) Reclaim(conn IConnection) error {
	return r.internalPool.reclaim(conn)
}

func (r *RoutingDriverPool) Close() error {
	return r.internalPool.close()
}

type RoutingDriverPoolV4 struct {
	internalPool *routingDriverPool
}

func (r *RoutingDriverPoolV4) Open(db string, mode DriverMode) (IConnection, error) {
	return r.internalPool.open(db, mode)
}

func (r *RoutingDriverPoolV4) Reclaim(conn IConnection) error {
	return r.internalPool.reclaim(conn)
}

func (r *RoutingDriverPoolV4) Close() error {
	return r.internalPool.close()
}
