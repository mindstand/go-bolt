package messages

import (
	"fmt"
)

const (
	// SuccessMessageSignature is the signature byte for the SUCCESS message
	// SUCCESS <metadata>
	// binary [0111 0000]
	SuccessMessageSignature = 0x70

	// RecordMessageSignature is the signature byte for the RECORD message
	// RECORD <value>
	// binary [0111 0001]
	RecordMessageSignature = 0x71

	// IgnoredMessageSignature is the signature byte for the IGNORED message
	// IGNORED <metadata>
	// binary [0111 1110]
	IgnoredMessageSignature = 0x7e

	// FailureMessageSignature is the signature byte for the FAILURE message
	// FAILURE <metadata>
	// binary [0111 1111]
	FailureMessageSignature = 0x7f
)

// SuccessMessage Represents an SUCCESS message
type SuccessMessage struct {
	Metadata map[string]interface{}
}

// NewSuccessMessage Gets a new SuccessMessage struct
func NewSuccessMessage(metadata map[string]interface{}) SuccessMessage {
	return SuccessMessage{
		Metadata: metadata,
	}
}

func (i SuccessMessage) GetAvailableAfter() int64 {
	if i.Metadata == nil || len(i.Metadata) == 0 {
		return -1
	}

	if after, ok := i.Metadata["result_available_after"]; ok {
		afterAcual, ok := after.(int64)
		if !ok {
			return -1
		}

		return afterAcual
	}

	return -1
}

// Signature gets the signature byte for the struct
func (i SuccessMessage) Signature() int {
	return SuccessMessageSignature
}

// AllFields gets the fields to encode for the struct
func (i SuccessMessage) AllFields() []interface{} {
	return []interface{}{i.Metadata}
}

// RecordMessage Represents an RECORD message
type RecordMessage struct {
	Fields []interface{}
}

// NewRecordMessage Gets a new RecordMessage struct
func NewRecordMessage(fields []interface{}) RecordMessage {
	return RecordMessage{
		Fields: fields,
	}
}

// Signature gets the signature byte for the struct
func (i RecordMessage) Signature() int {
	return RecordMessageSignature
}

// AllFields gets the fields to encode for the struct
func (i RecordMessage) AllFields() []interface{} {
	return []interface{}{i.Fields}
}

// IgnoredMessage Represents an IGNORED message
type IgnoredMessage struct{}

// NewIgnoredMessage Gets a new IgnoredMessage struct
func NewIgnoredMessage() IgnoredMessage {
	return IgnoredMessage{}
}

// Signature gets the signature byte for the struct
func (i IgnoredMessage) Signature() int {
	return IgnoredMessageSignature
}

// AllFields gets the fields to encode for the struct
func (i IgnoredMessage) AllFields() []interface{} {
	return []interface{}{}
}

// FailureMessage Represents an FAILURE message
type FailureMessage struct {
	Metadata map[string]interface{}
}

// NewFailureMessage Gets a new FailureMessage struct
func NewFailureMessage(metadata map[string]interface{}) FailureMessage {
	return FailureMessage{
		Metadata: metadata,
	}
}

// Signature gets the signature byte for the struct
func (i FailureMessage) Signature() int {
	return FailureMessageSignature
}

// AllFields gets the fields to encode for the struct
func (i FailureMessage) AllFields() []interface{} {
	return []interface{}{i.Metadata}
}

// GetCode returns the internal error code from neo4j
func (i FailureMessage) GetCode() string {
	code, ok := i.Metadata["code"]
	if !ok {
		return "none"
	}

	codeStr, ok := code.(string)
	if !ok {
		return "failed to parse code string"
	}

	return codeStr
}

// GetMessage extracts the internal failure message returned by neo4j
func (i FailureMessage) GetMessage() string {
	msg, ok := i.Metadata["message"]
	if !ok {
		return "none"
	}

	msgStr, ok := msg.(string)
	if !ok {
		return "failed to parse message to string"
	}

	return msgStr
}

// Error is the implementation of the Golang error interface so a failure message
// can be treated like a normal error
func (i FailureMessage) Error() string {
	return fmt.Sprintf("code: [%s]; message: [%s]", i.GetCode(), i.GetMessage())
}
