package protocol_v3

import (
	"github.com/mindstand/go-bolt/constants"
	"github.com/mindstand/go-bolt/encoding"
	"github.com/mindstand/go-bolt/encoding/encoding_v2"
	"github.com/mindstand/go-bolt/structures"
	"github.com/mindstand/go-bolt/structures/messages"
	"io"
)

type BoltProtocolV3 struct {}

func (b *BoltProtocolV3) SupportsMultiDatabase() bool {
	return false
}

func (b *BoltProtocolV3) GetInitMessage(client string, authToken map[string]interface{}) structures.Structure {
	if authToken == nil {
		authToken = map[string]interface{}{}
	}

	authToken["user_agent"] = client

	return messages.NewHelloMessage(authToken)
}

func (b *BoltProtocolV3) GetCloseMessage() (structures.Structure, bool) {
	return messages.NewGoodbyeMessage(), true
}

func (b *BoltProtocolV3) GetTxBeginMessage(database string, accessMode constants.AccessMode) structures.Structure {
	return messages.NewBeginMessage(messages.BuildTxMetadataWithDatabase(nil, nil, database, accessMode, nil))
}

func (b *BoltProtocolV3) GetTxCommitMessage() structures.Structure {
	return messages.NewCommitMessage()
}

func (b *BoltProtocolV3) GetTxRollbackMessage() structures.Structure {
	return messages.NewRollbackMessage()
}

func (b *BoltProtocolV3) GetRunMessage(query string, params map[string]interface{}, dbName string, mode constants.AccessMode, autoCommit bool) structures.Structure {
	if autoCommit {
		return messages.NewAutoCommitTxRunMessage(query, params, 0, nil, dbName, mode)
	} else {
		return messages.NewUnmanagedTxRunMessage(query, params)
	}
}

func (b *BoltProtocolV3) Marshal(v interface{}) ([]byte, error) {
	return encoding_v2.Marshal(v)
}

func (b *BoltProtocolV3) Unmarshal(bytes []byte) (interface{}, error) {
	return encoding_v2.Unmarshal(bytes)
}

func (b *BoltProtocolV3) NewEncoder(w io.Writer, chunkSize uint16) encoding.IEncoder {
	return encoding_v2.NewEncoder(w, chunkSize)
}

func (b *BoltProtocolV3) NewDecoder(r io.Reader) encoding.IDecoder {
	return encoding_v2.NewDecoder(r)
}

