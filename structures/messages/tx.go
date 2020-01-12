package messages

const (
	// CommitMessageSignature is the signature byte for COMMIT message
	// COMMIT
	// binary [0001 0010]
	CommitMessageSignature = 0x12

	// BeginMessageSignature is the signature byte for the BEGIN message
	// BEGIN <metadata>
	// binary [0001 0001]
	BeginMessageSignature = 0x11

	// RollbackMessageSignature is the signature byte for the ROLLBACK message
	// ROLLBACK
	// binary [0001 0011]
	RollbackMessageSignature = 0x13
)

type CommitMessage struct{}

func NewCommitMessage() CommitMessage {
	return CommitMessage{}
}

func (c CommitMessage) Signature() int {
	return CommitMessageSignature
}

func (c CommitMessage) AllFields() []interface{} {
	return []interface{}{}
}

// BEGIN <metadata>
type BeginMessage struct {
	metadata map[string]interface{}
}

func NewBeginMessage(metadata map[string]interface{}) BeginMessage {
	return BeginMessage{
		metadata: metadata,
	}
}

func (b BeginMessage) Signature() int {
	return BeginMessageSignature
}

func (b BeginMessage) AllFields() []interface{} {
	return []interface{}{b.metadata}
}

type RollbackMessage struct{}

func NewRollbackMessage() RollbackMessage {
	return RollbackMessage{}
}

func (r RollbackMessage) Signature() int {
	return RollbackMessageSignature
}

func (r RollbackMessage) AllFields() []interface{} {
	return []interface{}{}
}
