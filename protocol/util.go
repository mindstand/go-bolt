package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mindstand/go-bolt/protocol/protocol_v1"
)

func GetProtocol(version []byte) (IBoltProtocol, int, error) {
	if version == nil || len(version) == 0{
		return nil, -1, errors.New("can not get protocol for nil or empty version")
	}

	// todo write this part out for all protocols
	if bytes.Equal(version, protocol_v1.ProtocolVersionBytes) {
		return &protocol_v1.BoltProtocolV1{}, protocol_v1.ProtocolVersion, nil
	} else {
		return nil, -1, fmt.Errorf("protocol with bytes [%v] not supported", version)
	}
}

