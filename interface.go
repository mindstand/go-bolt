package goBolt

import (
	"net"
	"time"
)

type IBoltConnectionFactory interface {
	CreateBoltConnection() (IConnection, error)
}

// bolt+routing will not work for non pooled connections
type IDriver interface {
	// OpenNeo opens a Neo-specific connection.
	Open(mode DriverMode) (IConnection, error)
}

type IDriverPool interface {
	// Open opens a Neo-specific connection.
	Open(mode DriverMode) (IConnection, error)
	Reclaim(IConnection) error
	Close() error
}

// bolt+routing will not work for non pooled connections
type IDriverV4 interface {
	// OpenNeo opens a Neo-specific connection.
	Open(db string, mode DriverMode) (IConnection, error)
}

type IDriverPoolV4 interface {
	// allows user of new neo4j multiple db feature
	Open(db string, mode DriverMode) (IConnection, error)
	Reclaim(IConnection) error
	Close() error
}

// Conn represents a connection to Neo4J
//
// Implements a neo-friendly interface.
// Some of the features of this interface implement neo-specific features
// unavailable in the sql/driver compatible interface
//
// Conn objects, and any prepared statements/transactions within ARE NOT
// THREAD SAFE.  If you want to use multipe go routines with these objects,
// you should use a driver to create a new conn for each routine.
type IConnection interface {
	// PrepareNeo prepares a neo4j specific statement
	PrepareNeo(query string) (Stmt, error)
	// PreparePipeline prepares a neo4j specific pipeline statement
	// Useful for running multiple queries at the same time
	PreparePipeline(query ...string) (PipelineStmt, error)
	// QueryNeo queries using the neo4j-specific interface
	QueryNeo(query string, params QueryParams) (Rows, error)
	// QueryNeoAll queries using the neo4j-specific interface and returns all row data and output metadata
	QueryNeoAll(query string, params QueryParams) (NeoRows, map[string]interface{}, map[string]interface{}, error)
	// QueryPipeline queries using the neo4j-specific interface
	// pipelining multiple statements
	QueryPipeline(query []string, params ...QueryParams) (PipelineRows, error)
	// ExecNeo executes a query using the neo4j-specific interface
	ExecNeo(query string, params QueryParams) (Result, error)
	// ExecPipeline executes a query using the neo4j-specific interface
	// pipelining multiple statements
	ExecPipeline(query []string, params ...QueryParams) ([]Result, error)
	// Close closes the connection
	Close() error
	// Begin starts a new transaction
	Begin() (Tx, error)
	// SetChunkSize is used to set the max chunk size of the
	// bytes to send to Neo4j at once
	SetChunkSize(uint16)
	// SetTimeout sets the read/write timeouts for the
	// connection to Neo4j
	SetTimeout(time.Duration)

	// private package level stuff
	getStatement() *boltStmt
	setStatement(*boltStmt)

	getTx() Tx
	setTx(tx Tx)

	getConnection() net.Conn
	setConnection(conn net.Conn)

	getConnErr() error
	setConnErr(err error)

	getClosed() bool
	setClosed(closed bool)

	getReadOnly() bool
	setReadOnly(readOnly bool)

	initialize() error

	sendRunPullAllConsumeSingle(string, map[string]interface{}) (interface{}, interface{}, error)
}