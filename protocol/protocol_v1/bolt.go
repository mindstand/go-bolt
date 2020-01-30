package protocol_v1

import (
	"github.com/mindstand/go-bolt/encoding"
	"github.com/mindstand/go-bolt/encoding/encoding_v1"
	"github.com/mindstand/go-bolt/structures"
	"github.com/mindstand/go-bolt/structures/messages"
	"io"
)

type BoltProtocolV1 struct {

}

func (b *BoltProtocolV1) GetInitMessage(client string, authToken map[string]interface{}) structures.Structure {
	return messages.NewInitMessage(client, authToken)
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
