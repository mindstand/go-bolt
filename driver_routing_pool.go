package goBolt

import (
	"context"
	"fmt"
	pool "github.com/jolestar/go-commons-pool"
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/constants"
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/protocol/protocol_v4"
	"net/url"
	"strings"
	"sync"
)

type routingDriverPool struct {
	maxConns int
	refLock  sync.Mutex
	closed   bool

	client *Client

	//compose configuration
	config *clusterConnectionConfig

	//write resources
	writeConns int
	writePool  *pool.ObjectPool

	//read resources
	readConns int
	readPool  *pool.ObjectPool
}

func newRoutingDriverPool(connStr string, max int, poolFunc func(context.Context) (interface{}, error)) (*routingDriverPool, error) {

}

func (b *routingDriverPool) refreshConnectionPool() error {

}

func (b *routingDriverPool) close() error {

}

func (b *routingDriverPool) open(db string, mode constants.AccessMode) (connection.IConnection, error) {

}

func (b *routingDriverPool) reclaim(conn connection.IConnection) error {

}

// ------------------------------
// heres the interface impls, the core impl is above

type RoutingDriverPool struct {
	internalPool *routingDriverPool
}

func (r *RoutingDriverPool) Open(mode constants.AccessMode) (connection.IConnection, error) {
	return r.internalPool.open("", mode)
}

func (r *RoutingDriverPool) Reclaim(conn connection.IConnection) error {
	return r.internalPool.reclaim(conn)
}

func (r *RoutingDriverPool) Close() error {
	return r.internalPool.close()
}