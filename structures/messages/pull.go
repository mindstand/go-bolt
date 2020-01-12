package messages

const (
	// PullMessageSignature is the signature byte for the PULL message
	// PULL
	// binary [0011 1111]
	PullMessageSignature = 0x3F
)

// PullMessage Represents an PULL_ALL message
type PullMessage struct{}

// NewPullMessage Gets a new PullMessage struct
func NewPullMessage() PullMessage {
	return PullMessage{}
}

// Signature gets the signature byte for the struct
func (i PullMessage) Signature() int {
	return PullMessageSignature
}

// AllFields gets the fields to encode for the struct
func (i PullMessage) AllFields() []interface{} {
	return []interface{}{}
}
