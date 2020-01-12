package messages

const (
	// InitMessageSignature is the signature byte for the INIT message
	// INIT <user_agent> <authentication_token>
	// binary [0000 0001]
	InitMessageSignature = 0x01
)

// InitMessage Represents an INIT message
type InitMessage struct {
	data map[string]interface{}
}

// NewInitMessage Gets a new InitMessage struct
func NewInitMessage(clientName string, user string, password string) InitMessage {
	var data map[string]interface{}
	if user == "" {
		data = map[string]interface{}{
			"scheme": "none",
		}
	} else {
		data = map[string]interface{}{
			"scheme":      "basic",
			"principal":   user,
			"credentials": password,
		}
	}

	data["user_agent"] = clientName

	return InitMessage{
		data: data,
	}
}

// Signature gets the signature byte for the struct
func (i InitMessage) Signature() int {
	return InitMessageSignature
}

// AllFields gets the fields to encode for the struct
func (i InitMessage) AllFields() []interface{} {
	return []interface{}{i.data}
}
