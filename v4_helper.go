package goBolt

// queries
var (
	// crd on dbs
	createQuery = `create database %s`
	deleteDatabase = `drop database %s`
	startDatabase = `start database %s`
	stopDatabase = `stop database %s`
	showDatabase = `show database %s`

	// more read queries
	showAllDatabases = `show databases`
	showDefaultDatabase = `show default database`
)

func handleV4OpenConnection(conn IConnection, db string, createIfNotExists bool) error {
	panic("needs to be implemented")
}