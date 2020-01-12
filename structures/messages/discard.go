package messages

const (
	// DiscardMessageSignature is the signature byte for the DISCARD message
	// DISCARD
	// binary [0010 1111]
	DiscardMessageSignature = 0x2f
)

// DiscardMessage Represents an DISCARD_ALL message
type DiscardMessage struct{}

// NewDiscardMessage Gets a new DiscardMessage struct
func NewDiscardMessage() DiscardMessage {
	return DiscardMessage{}
}

// Signature gets the signature byte for the struct
func (i DiscardMessage) Signature() int {
	return DiscardMessageSignature
}

// AllFields gets the fields to encode for the struct
func (i DiscardMessage) AllFields() []interface{} {
	return []interface{}{}
}
