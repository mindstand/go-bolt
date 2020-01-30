package protocol

import (
	"github.com/mindstand/go-bolt/encoding"
	"github.com/mindstand/go-bolt/structures"
	"io"
)

type IBoltProtocol interface {
	GetInitMessage(client string, authToken map[string]interface{}) structures.Structure

	// marshall and unmarshal via the protocol
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(b []byte) (interface{}, error)

	// get encoders/decoders
	NewEncoder(w io.Writer, chunkSize uint16) encoding.IEncoder
	NewDecoder(r io.Reader) encoding.IDecoder
}
