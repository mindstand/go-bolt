package protocol_v2

import (
	"github.com/mindstand/go-bolt/constants"
	"github.com/mindstand/go-bolt/encoding"
	"github.com/mindstand/go-bolt/encoding/encoding_v2"
	"github.com/mindstand/go-bolt/structures"
	"github.com/mindstand/go-bolt/structures/messages"
	"io"
)

type BoltProtocolV2 struct {

}

func (b *BoltProtocolV2) GetDiscardMessage(qid int64) structures.Structure {
	return messages.NewDiscardAllMessage()
}

func (b *BoltProtocolV2) GetDiscardAllMessage() structures.Structure {
	return messages.NewDiscardAllMessage()
}

func (b *BoltProtocolV2) SupportsMultiDatabase() bool {
	return false
}

func (b *BoltProtocolV2) GetCloseMessage() (structures.Structure, bool) {
	return nil, false
}

func (b *BoltProtocolV2) GetTxBeginMessage(database string, accessMode constants.AccessMode) structures.Structure {
	return messages.NewRunMessage("BEGIN", nil)
}

func (b *BoltProtocolV2) GetTxCommitMessage() structures.Structure {
	return messages.NewRunMessage("COMMIT", nil)
}

func (b *BoltProtocolV2) GetTxRollbackMessage() structures.Structure {
	return messages.NewRunMessage("ROLLBACK", nil)
}

func (b *BoltProtocolV2) GetInitMessage(client string, authToken map[string]interface{}) structures.Structure {
	return messages.NewInitMessage(client, authToken)
}

func (b *BoltProtocolV2) GetRunMessage(query string, params map[string]interface{}, dbName string, mode constants.AccessMode, autoCommit bool) structures.Structure {
	return messages.NewRunMessage(query, params)
}

func (b *BoltProtocolV2) GetPullAllMessage() structures.Structure {
	return messages.NewPullAllMessage()
}

func (b *BoltProtocolV2) Marshal(v interface{}) ([]byte, error) {
	return encoding_v2.Marshal(v)
}

func (b *BoltProtocolV2) Unmarshal(bytes []byte) (interface{}, error) {
	return encoding_v2.Unmarshal(bytes)
}

func (b *BoltProtocolV2) NewEncoder(w io.Writer, chunkSize uint16) encoding.IEncoder {
	return encoding_v2.NewEncoder(w, chunkSize)
}

func (b *BoltProtocolV2) NewDecoder(r io.Reader) encoding.IDecoder {
	return encoding_v2.NewDecoder(r)
}
