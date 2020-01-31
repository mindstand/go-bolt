package protocol_v4

import (
	"github.com/mindstand/go-bolt/constants"
	"github.com/mindstand/go-bolt/encoding"
	"github.com/mindstand/go-bolt/encoding/encoding_v2"
	"github.com/mindstand/go-bolt/structures"
	"github.com/mindstand/go-bolt/structures/messages"
	"io"
)

type BoltProtocolV4 struct {}

func (b *BoltProtocolV4) GetDiscardMessage(qid int64) structures.Structure {
	return messages.NewDiscardMessage(messages.StreamUnlimited, qid)
}

func (b *BoltProtocolV4) GetDiscardAllMessage() structures.Structure {
	return messages.NewDiscardMessage(messages.StreamUnlimited, messages.AbsentQueryId)
}

func (b *BoltProtocolV4) SupportsMultiDatabase() bool {
	return true
}

func (b *BoltProtocolV4) GetInitMessage(client string, authToken map[string]interface{}) structures.Structure {
	if authToken == nil {
		authToken = map[string]interface{}{}
	}

	authToken["user_agent"] = client

	return messages.NewHelloMessage(authToken)
}

func (b *BoltProtocolV4) GetCloseMessage() (structures.Structure, bool) {
	return messages.NewGoodbyeMessage(), true
}

func (b *BoltProtocolV4) GetTxBeginMessage(database string, accessMode constants.AccessMode) structures.Structure {
	return messages.NewBeginMessage(messages.BuildTxMetadataWithDatabase(nil, nil, database, accessMode, nil))
}

func (b *BoltProtocolV4) GetTxCommitMessage() structures.Structure {
	return messages.NewCommitMessage()
}

func (b *BoltProtocolV4) GetTxRollbackMessage() structures.Structure {
	return messages.NewRollbackMessage()
}

func (b *BoltProtocolV4) GetRunMessage(query string, params map[string]interface{}, dbName string, mode constants.AccessMode, autoCommit bool) structures.Structure {
	if autoCommit {
		return messages.NewAutoCommitTxRunMessage(query, params, 0, nil, dbName, mode)
	} else {
		return messages.NewUnmanagedTxRunMessage(query, params)
	}
}

func (b *BoltProtocolV4) GetPullAllMessage() structures.Structure {
	return messages.NewPull_PullAllMessage()
}

func (b *BoltProtocolV4) Marshal(v interface{}) ([]byte, error) {
	return encoding_v2.Marshal(v)
}

func (b *BoltProtocolV4) Unmarshal(bytes []byte) (interface{}, error) {
	return encoding_v2.Unmarshal(bytes)
}

func (b *BoltProtocolV4) NewEncoder(w io.Writer, chunkSize uint16) encoding.IEncoder {
	return encoding_v2.NewEncoder(w, chunkSize)
}

func (b *BoltProtocolV4) NewDecoder(r io.Reader) encoding.IDecoder {
	return encoding_v2.NewDecoder(r)
}
