package messages

const (
	// HelloMessageSignature is the signature byte for the HELLO message
	// HELLO <metadata>
	// binary [0000 0001]
	HelloMessageSignature = 0x01
)

type HelloMessage struct {
	metadata map[string]interface{}
}

func NewHelloMessage(metadata map[string]interface{}) HelloMessage {
	return HelloMessage{
		metadata: metadata,
	}
}

func (h HelloMessage) Signature() int {
	return HelloMessageSignature
}

func (h HelloMessage) AllFields() []interface{} {
	return []interface{}{h.metadata}
}
