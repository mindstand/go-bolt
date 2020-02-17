package protocol_v1

import (
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/encoding"
	"github.com/mindstand/go-bolt/encoding/encoding_v1"
	"github.com/mindstand/go-bolt/structures"
	"github.com/mindstand/go-bolt/structures/messages"
	"io"
)

type BoltProtocolV1 struct {
}

func (b *BoltProtocolV1) GetResultAvailableAfterKey() string {
	return "result_available_after"
}

func (b *BoltProtocolV1) GetResultConsumedAfterKey() string {
	return "result_consumed_after"
}

func (b *BoltProtocolV1) GetDiscardMessage(qid int64) structures.Structure {
	return messages.NewDiscardAllMessage()
}

func (b *BoltProtocolV1) GetDiscardAllMessage() structures.Structure {
	return messages.NewDiscardAllMessage()
}

func (b *BoltProtocolV1) SupportsMultiDatabase() bool {
	return false
}

func (b *BoltProtocolV1) GetCloseMessage() (structures.Structure, bool) {
	return nil, false
}

func (b *BoltProtocolV1) GetTxBeginMessage(database string, accessMode bolt_mode.AccessMode) structures.Structure {
	return messages.NewRunMessage("BEGIN", nil)
}

func (b *BoltProtocolV1) GetTxCommitMessage() structures.Structure {
	return messages.NewRunMessage("COMMIT", nil)
}

func (b *BoltProtocolV1) GetTxRollbackMessage() structures.Structure {
	return messages.NewRunMessage("ROLLBACK", nil)
}

func (b *BoltProtocolV1) GetInitMessage(client string, authToken map[string]interface{}) structures.Structure {
	return messages.NewInitMessage(client, authToken)
}

func (b *BoltProtocolV1) GetRunMessage(query string, params map[string]interface{}, dbName string, mode bolt_mode.AccessMode, autoCommit bool) structures.Structure {
	return messages.NewRunMessage(query, params)
}

func (b *BoltProtocolV1) GetPullAllMessage() structures.Structure {
	return messages.NewPullAllMessage()
}

func (b *BoltProtocolV1) Marshal(v interface{}) ([]byte, error) {
	return encoding_v1.Marshal(v)
}

func (b *BoltProtocolV1) Unmarshal(bytes []byte) (interface{}, error) {
	return encoding_v1.Unmarshal(bytes)
}

func (b *BoltProtocolV1) NewEncoder(w io.Writer, chunkSize uint16) encoding.IEncoder {
	return encoding_v1.NewEncoder(w, chunkSize)
}

func (b *BoltProtocolV1) NewDecoder(r io.Reader) encoding.IDecoder {
	return encoding_v1.NewDecoder(r)
}
