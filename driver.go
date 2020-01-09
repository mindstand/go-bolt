package goBolt

type driver struct {
	recorder *recorder
	createIfNotExists bool
	connectionFactory IBoltConnectionFactory
}

type Driver struct {
	internalDriver *driver
}

func (d *Driver) Open(mode DriverMode) (IConnection, error) {
	return d.internalDriver.connectionFactory.CreateBoltConnection()
}

type DriverV4 struct {
	internalDriver *driver
}

func (d DriverV4) Open(db string, mode DriverMode) (IConnection, error) {
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




