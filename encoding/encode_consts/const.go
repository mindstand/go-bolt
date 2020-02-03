package encode_consts

const (
	// NilMarker represents the encoding marker byte for a nil object
	NilMarker = 0xC0

	// TrueMarker represents the encoding marker byte for a true boolean object
	TrueMarker = 0xC3
	// FalseMarker represents the encoding marker byte for a false boolean object
	FalseMarker = 0xC2

	// Int8Marker represents the encoding marker byte for a int8 object
	Int8Marker = 0xC8
	// Int16Marker represents the encoding marker byte for a int16 object
	Int16Marker = 0xC9
	// Int32Marker represents the encoding marker byte for a int32 object
	Int32Marker = 0xCA
	// Int64Marker represents the encoding marker byte for a int64 object
	Int64Marker = 0xCB

	// FloatMarker represents the encoding marker byte for a float32/64 object
	FloatMarker = 0xC1

	// TinyStringMarker represents the encoding marker byte for a string object
	TinyStringMarker = 0x80
	// String8Marker represents the encoding marker byte for a string object
	String8Marker = 0xD0
	// String16Marker represents the encoding marker byte for a string object
	String16Marker = 0xD1
	// String32Marker represents the encoding marker byte for a string object
	String32Marker = 0xD2

	// TinySliceMarker represents the encoding marker byte for a slice object
	TinySliceMarker = 0x90
	// Slice8Marker represents the encoding marker byte for a slice object
	Slice8Marker = 0xD4
	// Slice16Marker represents the encoding marker byte for a slice object
	Slice16Marker = 0xD5
	// Slice32Marker represents the encoding marker byte for a slice object
	Slice32Marker = 0xD6

	// TinyMapMarker represents the encoding marker byte for a map object
	TinyMapMarker = 0xA0
	// Map8Marker represents the encoding marker byte for a map object
	Map8Marker = 0xD8
	// Map16Marker represents the encoding marker byte for a map object
	Map16Marker = 0xD9
	// Map32Marker represents the encoding marker byte for a map object
	Map32Marker = 0xDA

	// TinyStructMarker represents the encoding marker byte for a struct object
	TinyStructMarker = 0xB0
	// Struct8Marker represents the encoding marker byte for a struct object
	Struct8Marker = 0xDC
	// Struct16Marker represents the encoding marker byte for a struct object
	Struct16Marker = 0xDD

	DateSignature  byte = 'D'
	DateStructSize int  = 1

	TimeSignature  byte = 'T'
	TimeStructSize int  = 2

	LocalTimeSignature  byte = 't'
	LocalTimeStructSize int  = 2

	LocalDateTimeSignature  byte = 'd'
	LocalDateTimeStructSize int  = 2

	DateTimeWithZoneOffsetSignature byte = 'F'
	DateTimeWithZoneIdSignature     byte = 'f'
	DateTimeStructSize              int  = 3

	DurationSignature      byte = 'E'
	DurationTimeStructSize int  = 4
)

var (
	// EndMessage is the data to send to end a message
	EndMessage = []byte{byte(0x00), byte(0x00)}
)
