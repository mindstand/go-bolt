package messages

import (
	"github.com/mindstand/go-bolt/constants"
	"time"
)

const (
	// RunMessageWithMetadataSignature is the signature byte for the RUN message
	// RUN <query> <params> <metadata>
	// binary [0001 0000]
	RunMessageWithMetadataSignature = 0x10
)

type RunWithMetadataMessage struct {
	statement  string
	parameters map[string]interface{}
	metadata   map[string]interface{}
}

// todo update to support bookmarks at some point
// this would be used outside an explicit tx
func NewAutoCommitTxRunMessage(query string, params map[string]interface{}, timeout time.Duration, txConfig map[string]interface{}, dbName string, mode constants.AccessMode) RunWithMetadataMessage {
	return newRunMessageWithMetadata(query, params, BuildTxMetadataWithDatabase(&timeout, txConfig, dbName, mode, nil))
}

// this would be used in an explicit transaction
func NewUnmanagedTxRunMessage(query string, params map[string]interface{}) RunWithMetadataMessage {
	return newRunMessageWithMetadata(query, params, map[string]interface{}{})
}

func newRunMessageWithMetadata(query string, params, metadata map[string]interface{}) RunWithMetadataMessage {
	return RunWithMetadataMessage{
		statement:  query,
		parameters: params,
		metadata:   metadata,
	}
}

func (r RunWithMetadataMessage) Signature() int {
	return RunMessageWithMetadataSignature
}

func (r RunWithMetadataMessage) AllFields() []interface{} {
	return []interface{}{r.statement, r.parameters, r.metadata}
}
