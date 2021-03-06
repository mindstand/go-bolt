package connection

import (
	"github.com/mindstand/go-bolt/structures"
	"time"
)

// Result represents a result from a runQuery that returns no data
type IResult interface {
	GetStats() (map[string]interface{}, bool)
	GetNodesCreated() (int64, bool)
	GetRelationshipsCreated() (int64, bool)
	GetNodesDeleted() (int64, bool)
	GetRelationshipsDeleted() (int64, bool)

	// Metadata returns the metadata response from neo4j
	Metadata() map[string]interface{}
}

type IQuery interface {
	// ExecNeo executes a runQuery that returns no rows.
	Exec(query string, params QueryParams) (IResult, error)

	ExecWithDb(query string, params QueryParams, db string) (IResult, error)

	// QueryNeo executes a runQuery that returns data.
	Query(query string, params QueryParams) ([][]interface{}, IResult, error)

	QueryWithDb(query string, params QueryParams, db string) ([][]interface{}, IResult, error)
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

	// ackFailure(failure messages.FailureMessage) error
	reset() error

	// closes connection
	Close() error

	GetProtocolVersionNumber() int
	GetProtocolVersionBytes() []byte

	// returns true if open, returns false if not
	ValidateOpen() bool

	MakeIdle() error

	Begin() (ITransaction, error)
	BeginWithDatabase(db string) (ITransaction, error)
	// SetTimeout sets the read/write timeouts for the
	// connection to Neo4j
	SetTimeout(time.Duration)
	SetChunkSize(uint16)

	// connection id's are for the routing driver to keep track of connections
	GetConnectionId() string
	SetConnectionId(id string)
}
