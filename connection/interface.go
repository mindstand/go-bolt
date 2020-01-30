package connection

import (
	"github.com/mindstand/go-bolt/structures"
	"github.com/mindstand/go-bolt/structures/messages"
	"time"
)

// Result represents a result from a query that returns no data
type IResult interface {
	GetStats() (map[string]interface{}, bool)
	GetNodesCreated() (int64, bool)
	GetRelationshipsCreated() (int64, bool)
	GetNodesDeleted() (int64, bool)
	GetRelationshipsDeleted() (int64, bool)

	// Metadata returns the metadata response from neo4j
	Metadata() map[string]interface{}
}

// IRows represents results of rows from the DB
//
// Row objects ARE NOT THREAD SAFE.
// If you want to use multiple go routines with these objects,
// you should use a internalDriver to create a new conn for each routine.
type IRows interface {
	// Columns Gets the names of the columns in the returned dataset
	Columns() []string
	// Metadata Gets all of the metadata returned from Neo on query start
	Metadata() map[string]interface{}
	// Close the rows, flushing any existing datastream
	Close() error
	// NextNeo gets the next row result
	// When the rows are completed, returns the success metadata
	// and io.EOF
	Next() ([]interface{}, map[string]interface{}, error)
	// All gets all of the results from the row set. It's recommended to use NextNeo when
	// there are a lot of rows
	All() ([][]interface{}, map[string]interface{}, error)
}

// IStmt represents a statement to run against the database
//
// IStmt objects, and any rows prepared within ARE NOT
// THREAD SAFE.  If you want to use multiple go routines with these objects,
// you should use a internalDriver to create a new conn for each routine.
//type IStmt interface {
//	// Close Closes the statement. See sql/internalDriver.IStmt.
//	Close() error
//
//	// query stuff
//	IQuery
//}

type IQuery interface {
	// ExecNeo executes a query that returns no rows. Implements a Neo-friendly alternative to sql/internalDriver.
	Exec(query string, params QueryParams) (IResult, error)
	// QueryNeo executes a query that returns data. Implements a Neo-friendly alternative to sql/internalDriver.
	Query(query string, params QueryParams) (IRows, error)
}

// ITransaction controls a transaction
type ITransaction interface {
	// Query
	IQuery
	// Commit commits the transaction
	Commit() error
	// Rollback rolls back the transaction
	Rollback() error
	// IsClosed determines if the transaction has been closed
	IsClosed() bool
}

// IConnection
type IConnection interface {
	// Query functionality
	IQuery

	sendMessage(message structures.Structure) error
	sendMessageConsume(message structures.Structure) (interface{}, error)
	consume() (interface{}, error)

	ackFailure(failure messages.FailureMessage) error
	reset() error

	// closes connection
	Close() error

	Begin() (ITransaction, error)

	// SetTimeout sets the read/write timeouts for the
	// connection to Neo4j
	SetTimeout(time.Duration)
	SetChunkSize(uint16)
}