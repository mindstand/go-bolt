package encoding

type IDecoder interface {
	Decode() (interface{}, error)
}

type IEncoder interface {
	Write(p []byte) (n int, err error)
	Encode(iVal interface{}) error
}