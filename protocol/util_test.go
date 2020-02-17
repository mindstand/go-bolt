package protocol

import (
	"github.com/mindstand/go-bolt/protocol/protocol_v1"
	"github.com/mindstand/go-bolt/protocol/protocol_v2"
	"github.com/mindstand/go-bolt/protocol/protocol_v3"
	"github.com/mindstand/go-bolt/protocol/protocol_v4"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestGetProtocol tests grabbing the protocol version for the bytes provided
func TestGetProtocol(t *testing.T) {
	req := require.New(t)
	boltV1 := []byte{0x00, 0x00, 0x00, 0x01}
	boltV2 := []byte{0x00, 0x00, 0x00, 0x02}
	boltV3 := []byte{0x00, 0x00, 0x00, 0x03}
	boltV4 := []byte{0x00, 0x00, 0x00, 0x04}

	// test v1
	protocol, version, err := GetProtocol(boltV1)
	req.Nil(err)
	req.IsType(&protocol_v1.BoltProtocolV1{}, protocol)
	req.Equal(1, version)

	// test v2
	protocol, version, err = GetProtocol(boltV2)
	req.Nil(err)
	req.IsType(&protocol_v2.BoltProtocolV2{}, protocol)
	req.Equal(2, version)

	// test v3
	protocol, version, err = GetProtocol(boltV3)
	req.Nil(err)
	req.IsType(&protocol_v3.BoltProtocolV3{}, protocol)
	req.Equal(3, version)

	// test v4
	protocol, version, err = GetProtocol(boltV4)
	req.Nil(err)
	req.IsType(&protocol_v4.BoltProtocolV4{}, protocol)
	req.Equal(4, version)

	// test nil
	protocol, version, err = GetProtocol(nil)
	req.NotNil(err)
	req.Equal(-1, version)
	req.Nil(protocol)

	// test empty
	protocol, version, err = GetProtocol([]byte{})
	req.NotNil(err)
	req.Equal(-1, version)
	req.Nil(protocol)

	// test invalid version
	protocol, version, err = GetProtocol([]byte{0x00, 0x00, 0x00, 0x09})
	req.NotNil(err)
	req.Equal(-1, version)
	req.Nil(protocol)
}
