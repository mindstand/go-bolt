package messages

const (
	// PullMessageSignature is the signature byte for the PULL message
	// PULL
	// binary [0011 1111]
	//PullMessageSignature = 0x3F
	PullMessageSignature = 0x2F
	StreamUnlimited int64 = -1
	AbsentQueryId int64 = -1
)

// PullMessage Represents an PULL_ALL message
type PullMessage struct{
	metadata map[string]interface{}
}

// NewPullMessage Gets a new PullMessage struct
func NewPullMessage(n, id int64) PullMessage {
	mdata := map[string]interface{}{
		"n": n,
	}

	if id != AbsentQueryId {
		mdata["qid"] = id
	}

	return PullMessage{
		metadata: mdata,
	}
}

func NewPull_PullAllMessage() PullMessage {
	return NewPullMessage(StreamUnlimited, AbsentQueryId)
}

// Signature gets the signature byte for the struct
func (i PullMessage) Signature() int {
	return PullMessageSignature
}

// AllFields gets the fields to encode for the struct
func (i PullMessage) AllFields() []interface{} {
	return []interface{}{i.metadata}
}
