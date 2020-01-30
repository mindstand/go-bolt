package messages

const (
	// InitMessageSignature is the signature byte for the INIT message
	// INIT <user_agent> <authentication_token>
	// binary [0000 0001]
	InitMessageSignature = 0x01
)

const (
	SchemeKey = "scheme"
	PrincipalKey = "principal"
	CredentialsKey = "credentials"
	RealmKey = "realm"
	ParametersKey = "parameters"
)

// InitMessage Represents an INIT message
type InitMessage struct {
	loc string
	data map[string]interface{}
}

func BuildAuthTokenBasicWithRealm(username, password, realm string) map[string]interface{} {
	toReturn := map[string]interface{}{
		SchemeKey: "basic",
		PrincipalKey: username,
		CredentialsKey: password,
	}

	if realm == "" {
		toReturn[RealmKey] = realm
	}

	return toReturn
}

func BuildAuthTokenBasic(username, password string) map[string]interface{} {
	return BuildAuthTokenBasicWithRealm(username, password, "")
}

func BuildAuthTokenKerberos(base64EncodedTicket string) map[string]interface{} {
	return map[string]interface{}{
		SchemeKey: "kerberos",
		PrincipalKey: "",
		CredentialsKey: base64EncodedTicket,
	}
}

// NewInitMessage Gets a new InitMessage struct
func NewInitMessage(clientName string, authToken map[string]interface{}) InitMessage {
	//authToken["user_agent"] = clientName

	return InitMessage{
		data: authToken,
		loc: clientName,
	}
}

// Signature gets the signature byte for the struct
func (i InitMessage) Signature() int {
	return InitMessageSignature
}

// AllFields gets the fields to encode for the struct
func (i InitMessage) AllFields() []interface{} {
	return []interface{}{i.loc, i.data}
}
