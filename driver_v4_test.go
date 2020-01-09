package goBolt

type DriverV4 struct {

}

func (d *DriverV4) Open(db string, mode DriverMode) (IConnection, error) {
	panic("implement me")
}

func (d *DriverV4) Close() error {
	panic("implement me")
}
