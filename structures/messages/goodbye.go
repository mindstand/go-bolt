package messages

const (
	// GoodbyeMessageSignature is the signature byte for the GOODBYE message
	// GOODBYE
	// binary [0000 00010]
	GoodbyeMessageSignature = 0x02
)

type GoodbyeMessage struct{}

func NewGoodbyeMessage() GoodbyeMessage {
	return GoodbyeMessage{}
}

func (g GoodbyeMessage) Signature() int {
	return GoodbyeMessageSignature
}

func (g GoodbyeMessage) AllFields() []interface{} {
	return []interface{}{}
}
