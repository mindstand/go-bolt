package protocol

import (
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/encoding"
	"github.com/mindstand/go-bolt/structures"
	"io"
)

// todo:
// 1) update connection to have exec/execwdb and query and querywdb
// 2) update tx to do that shit too
// 3) go on to protocol 4

//IBoltProtocol describes different versions of bolt protocol
type IBoltProtocol interface {
	// creates correct init message for that impl of the protocol
	GetInitMessage(client string, authToken map[string]interface{}) structures.Structure
	// creates begin message for the tx
	// different versions of the protocol use either BeginMessage or RunMessage with the command BEGIN
	GetTxBeginMessage(database string, accessMode bolt_mode.AccessMode) structures.Structure
	// creates commit message for the tx
	// different versions of the protocol use either CommitMessage or RunMessage with the command COMMIT
	GetTxCommitMessage() structures.Structure
	// creates rollback message for the tx
	// different versions of the protocol use either RollbackMessage or RunMessage with the command ROLLBACK
	GetTxRollbackMessage() structures.Structure
	// creates close message for the tx if needed
	// returns true if explicit close message need, returns false if not
	GetCloseMessage() (structures.Structure, bool)
	// creates run message
	// newer versions of bolt protocol require additional information in run message for database specification, tx, and r/w modes
	GetRunMessage(query string, params map[string]interface{}, dbName string, mode bolt_mode.AccessMode, autoCommit bool) structures.Structure
	// creates pull all message
	GetPullAllMessage() structures.Structure
	// gets discard message
	GetDiscardMessage(qid int64) structures.Structure
	GetDiscardAllMessage() structures.Structure
	// newer versions of bolt protocol allow for multi database support
	SupportsMultiDatabase() bool

	// marshall and unmarshal via the protocol
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(b []byte) (interface{}, error)

	// get encoders/decoders
	NewEncoder(w io.Writer, chunkSize uint16) encoding.IEncoder
	NewDecoder(r io.Reader) encoding.IDecoder
}
