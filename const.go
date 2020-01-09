package goBolt

var (
	magicPreamble     = []byte{0x60, 0x60, 0xb0, 0x17}
	supportedVersions = []byte{
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	handShake          = append(magicPreamble, supportedVersions...)
	noVersionSupported = []byte{0x00, 0x00, 0x00, 0x00}
	// Version is the current version of this driver
	Version = "1.0"
	// ClientID is the id of this client
	ClientID = "GolangNeo4jBolt/" + Version
)

type DriverMode int

const (
	ReadOnlyMode  DriverMode = 0
	ReadWriteMode DriverMode = 1
)

type QueryParams map[string]interface{}

func (q *QueryParams) GetMap() map[string]interface{} {
	return *q
}

func (q *QueryParams) GetKeys() []string {
	keys := make([]string, len(*q))
	i := 0
	for k := range *q {
		keys[i] = k
		i++
	}

	return keys
}

type NeoRows [][]interface{}

func (n *NeoRows) Get2DSlice() [][]interface{} {
	return *n
}