package goBolt

import "github.com/mindstand/go-bolt/connection"

type internalDriver struct {
	recorder          *recorder
	createIfNotExists bool
	connectionFactory IBoltConnectionFactory
}

type Driver struct {
	internalDriver *internalDriver
}

func (d *Driver) Open(mode connection.DriverMode) (IConnection, error) {
	return d.internalDriver.connectionFactory.CreateBoltConnection()
}

type DriverV4 struct {
	internalDriver *internalDriver
}

func (d DriverV4) Open(db string, mode connection.DriverMode) (IConnection, error) {
	conn, err := d.internalDriver.connectionFactory.CreateBoltConnection()
	if err != nil {
		return nil, err
	}

	err = handleV4OpenConnection(conn, db, d.internalDriver.createIfNotExists)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
