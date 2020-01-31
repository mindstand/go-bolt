package goBolt

import (
	"context"
	"fmt"
	pool "github.com/jolestar/go-commons-pool"
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/errors"
)

type ConnectionPooledObjectFactory struct {
	connectionString string
}

func (c *ConnectionPooledObjectFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	conn, err := connection.CreateBoltConn(c.connectionString)
	if err != nil {
		return nil, err
	}

	return pool.NewPooledObject(conn), nil
}

func (c *ConnectionPooledObjectFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	if object == nil {
		return errors.New("pooled object wrapper can not be nil")
	}

	if object.Object == nil {
		return errors.New("pooled object can not be nil")
	}

	conn, ok := object.Object.(connection.IConnection)
	if !ok {
		return fmt.Errorf("unable to cast [%T] to [connection.IConnection]", object.Object)
	}

	if conn.ValidateOpen() {
		return conn.Close()
	} else {
		return nil
	}
}

func (c *ConnectionPooledObjectFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	if object == nil {
		return false
	}

	if object.Object == nil {
		return false
	}

	conn, ok := object.Object.(connection.IConnection)
	if !ok {
		return false
	}

	return conn.ValidateOpen()
}

func (c *ConnectionPooledObjectFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	if object == nil {
		return errors.New("pooled object wrapper can not be nil")
	}

	if object.Object == nil {
		return errors.New("pooled object can not be nil")
	}

	conn, ok := object.Object.(connection.IConnection)
	if !ok {
		return fmt.Errorf("unable to cast [%T] to [connection.IConnection]", object.Object)
	}

	var err error

	if !conn.ValidateOpen() {
		conn, err = connection.CreateBoltConn(c.connectionString)
		if err != nil {
			return err
		}

		object.Object = conn
	}

	return nil
}

func (c *ConnectionPooledObjectFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	if object == nil {
		return errors.New("pooled object wrapper can not be nil")
	}

	if object.Object == nil {
		return errors.New("pooled object can not be nil")
	}

	conn, ok := object.Object.(connection.IConnection)
	if !ok {
		return fmt.Errorf("unable to cast [%T] to [connection.IConnection]", object.Object)
	}

	return conn.MakeIdle()
}

