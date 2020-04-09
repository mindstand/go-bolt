package encoding_v2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/mindstand/go-bolt/encoding"
	"github.com/mindstand/go-bolt/encoding/encode_consts"
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/structures/graph"
	"github.com/mindstand/go-bolt/structures/messages"
	"github.com/mindstand/go-bolt/structures/types"
	"github.com/mindstand/gotime"
	"io"
	"time"
)

type DecoderV2 struct {
	r   io.Reader
	buf *bytes.Buffer
}

// NewDecoder Creates a new DecoderV2 object
func NewDecoder(r io.Reader) DecoderV2 {
	return DecoderV2{
		r:   r,
		buf: &bytes.Buffer{},
	}
}

// Unmarshal is used to marshal an object to the bolt interface encoded bytes
func Unmarshal(b []byte) (interface{}, error) {
	return NewDecoder(bytes.NewBuffer(b)).Decode()
}

// Read out the object bytes to decode
func (d DecoderV2) read() (*bytes.Buffer, error) {
	output := &bytes.Buffer{}
	for {
		lengthBytes := make([]byte, 2)
		if numRead, err := io.ReadFull(d.r, lengthBytes); numRead != 2 {
			return nil, errors.Wrap(err, "Couldn't read expected bytes for message length. Read: %d Expected: 2.", numRead)
		}

		// Chunk header contains length of current message
		messageLen := binary.BigEndian.Uint16(lengthBytes)
		if messageLen == 0 {
			// If the length is 0, the chunk is done.
			return output, nil
		}

		data, err := d.readData(messageLen)
		if err != nil {
			return output, errors.Wrap(err, "An error occurred reading message data")
		}

		numWritten, err := output.Write(data)
		if numWritten < len(data) {
			return output, errors.New("Didn't write full data on output. Expected: %d Wrote: %d", len(data), numWritten)
		} else if err != nil {
			return output, errors.Wrap(err, "Error writing data to output")
		}
	}
}

func (d DecoderV2) readData(messageLen uint16) ([]byte, error) {
	output := make([]byte, messageLen)
	var totalRead uint16
	for totalRead < messageLen {
		data := make([]byte, messageLen-totalRead)
		numRead, err := d.r.Read(data)
		if err != nil {
			return nil, errors.Wrap(err, "An error occurred reading from stream")
		} else if numRead == 0 {
			return nil, errors.Wrap(err, "Couldn't read expected bytes for message. Read: %d Expected: %d.", totalRead, messageLen)
		}

		for idx, b := range data {
			output[uint16(idx)+totalRead] = b
		}

		totalRead += uint16(numRead)
	}

	return output, nil
}

// Decode decodes the stream to an object
func (d DecoderV2) Decode() (interface{}, error) {
	data, err := d.read()
	if err != nil {
		return nil, err
	}

	return d.decode(data)
}

func (d DecoderV2) decode(buffer *bytes.Buffer) (interface{}, error) {

	marker, err := buffer.ReadByte()
	if err != nil {
		return nil, errors.Wrap(err, "Error reading marker")
	}

	// Here we have to get the marker as an int to check and see
	// if it's a TINYINT
	var markerInt int8
	err = binary.Read(bytes.NewBuffer([]byte{marker}), binary.BigEndian, &markerInt)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading marker as int8 from bolt message")
	}

	switch {

	// NIL
	case marker == encode_consts.NilMarker:
		return nil, nil

	// BOOL
	case marker == encode_consts.TrueMarker:
		return true, nil
	case marker == encode_consts.FalseMarker:
		return false, nil

	// INT
	case markerInt >= -16 && markerInt <= 127:
		return int64(int8(marker)), nil
	case marker == encode_consts.Int8Marker:
		var out int8
		err := binary.Read(buffer, binary.BigEndian, &out)
		return int64(out), err
	case marker == encode_consts.Int16Marker:
		var out int16
		err := binary.Read(buffer, binary.BigEndian, &out)
		return int64(out), err
	case marker == encode_consts.Int32Marker:
		var out int32
		err := binary.Read(buffer, binary.BigEndian, &out)
		return int64(out), err
	case marker == encode_consts.Int64Marker:
		var out int64
		err := binary.Read(buffer, binary.BigEndian, &out)
		return int64(out), err

	// FLOAT
	case marker == encode_consts.FloatMarker:
		var out float64
		err := binary.Read(buffer, binary.BigEndian, &out)
		return out, err

	// STRING
	case marker >= encode_consts.TinyStringMarker && marker <= encode_consts.TinyStringMarker+0x0F:
		size := int(marker) - int(encode_consts.TinyStringMarker)
		if size == 0 {
			return "", nil
		}
		return string(buffer.Next(size)), nil
	case marker == encode_consts.String8Marker:
		var size int8
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading string size")
		}
		return string(buffer.Next(int(size))), nil
	case marker == encode_consts.String16Marker:
		var size int16
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading string size")
		}
		return string(buffer.Next(int(size))), nil
	case marker == encode_consts.String32Marker:
		var size int32
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading string size")
		}
		return string(buffer.Next(int(size))), nil

	// SLICE
	case marker >= encode_consts.TinySliceMarker && marker <= encode_consts.TinySliceMarker+0x0F:
		size := int(marker) - int(encode_consts.TinySliceMarker)
		return d.decodeSlice(buffer, size)
	case marker == encode_consts.Slice8Marker:
		var size int8
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading slice size")
		}
		return d.decodeSlice(buffer, int(size))
	case marker == encode_consts.Slice16Marker:
		var size int16
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading slice size")
		}
		return d.decodeSlice(buffer, int(size))
	case marker == encode_consts.Slice32Marker:
		var size int32
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading slice size")
		}
		return d.decodeSlice(buffer, int(size))

	// MAP
	case marker >= encode_consts.TinyMapMarker && marker <= encode_consts.TinyMapMarker+0x0F:
		size := int(marker) - int(encode_consts.TinyMapMarker)
		return d.decodeMap(buffer, size)
	case marker == encode_consts.Map8Marker:
		var size int8
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading map size")
		}
		return d.decodeMap(buffer, int(size))
	case marker == encode_consts.Map16Marker:
		var size int16
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading map size")
		}
		return d.decodeMap(buffer, int(size))
	case marker == encode_consts.Map32Marker:
		var size int32
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading map size")
		}
		return d.decodeMap(buffer, int(size))

	// STRUCTURES
	case marker >= encode_consts.TinyStructMarker && marker <= encode_consts.TinyStructMarker+0x0F:
		size := int(marker) - int(encode_consts.TinyStructMarker)
		return d.decodeStruct(buffer, size)
	case marker == encode_consts.Struct8Marker:
		var size int8
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading struct size")
		}
		return d.decodeStruct(buffer, int(size))
	case marker == encode_consts.Struct16Marker:
		var size int16
		if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
			return nil, errors.Wrap(err, "An error occurred reading struct size")
		}
		return d.decodeStruct(buffer, int(size))

	default:
		return nil, errors.New("Unrecognized marker byte!: %x", marker)
	}

}

func (d DecoderV2) decodeSlice(buffer *bytes.Buffer, size int) ([]interface{}, error) {
	slice := make([]interface{}, size)
	for i := 0; i < size; i++ {
		item, err := d.decode(buffer)
		if err != nil {
			return nil, err
		}
		slice[i] = item
	}

	return slice, nil
}

func (d DecoderV2) decodeMap(buffer *bytes.Buffer, size int) (map[string]interface{}, error) {
	mapp := make(map[string]interface{}, size)
	for i := 0; i < size; i++ {
		keyInt, err := d.decode(buffer)
		if err != nil {
			return nil, err
		}
		val, err := d.decode(buffer)
		if err != nil {
			return nil, err
		}

		key, ok := keyInt.(string)
		if !ok {
			return nil, errors.New("Unexpected key type: %T with value %+v", keyInt, keyInt)
		}
		mapp[key] = val
	}

	return mapp, nil
}

func (d DecoderV2) decodeStruct(buffer *bytes.Buffer, size int) (interface{}, error) {

	signature, err := buffer.ReadByte()
	if err != nil {
		return nil, errors.Wrap(err, "An error occurred reading struct signature byte")
	}

	switch signature {
	case graph.NodeSignature:
		return d.decodeNode(buffer)
	case graph.RelationshipSignature:
		return d.decodeRelationship(buffer)
	case graph.PathSignature:
		return d.decodePath(buffer)
	case graph.UnboundRelationshipSignature:
		return d.decodeUnboundRelationship(buffer)
	case types.Point2DStructSignature:
		return d.decodePoint2D(buffer)
	case types.Point3DStructSignature:
		return d.decodePoint3D(buffer)
	case messages.RecordMessageSignature:
		return d.decodeRecordMessage(buffer)
	case messages.FailureMessageSignature:
		return d.decodeFailureMessage(buffer)
	case messages.IgnoredMessageSignature:
		return d.decodeIgnoredMessage(buffer)
	case messages.SuccessMessageSignature:
		return d.decodeSuccessMessage(buffer)
	case messages.DiscardAllMessageSignature:
		return d.decodeDiscardAllMessage(buffer)
	case messages.PullAllMessageSignature:
		return d.decodePullAllMessage(buffer)
	case messages.ResetMessageSignature:
		return d.decodeResetMessage(buffer)
	case encode_consts.DateSignature:
		return d.decodeDate(buffer)
	case encode_consts.TimeSignature:
		return d.decodeTime(buffer)
	case encode_consts.LocalTimeSignature:
		return d.decodeLocalTime(buffer)
	case encode_consts.LocalDateTimeSignature:
		return d.decodeLocalDateTime(buffer)
	case encode_consts.DateTimeWithZoneOffsetSignature:
		return d.decodeDateTimeWithZoneOffset(buffer)
	case encode_consts.DateTimeWithZoneIdSignature:
		return d.decodeTimeWithZoneId(buffer)
	case encode_consts.DurationSignature:
		return d.decodeDuration(buffer)
	default:
		return nil, errors.New("Unrecognized type decoding struct with signature %x", signature)
	}
}

func (d DecoderV2) decodeDate(buffer *bytes.Buffer) (gotime.Date, error) {
	epochDayI, err := d.decode(buffer)
	if err != nil {
		return gotime.Date{}, err
	}

	// days since unix epoch
	epochDay, ok := epochDayI.(int64)
	if !ok {
		return gotime.Date{}, errors.New("unable to cast to int64")
	}
	// add epochDays to unix epoch
	return gotime.NewDateOfEpochDays(int(epochDay)), nil
}

func (d DecoderV2) decodeTime(buffer *bytes.Buffer) (gotime.Clock, error) {
	nanoOfDayLocalI, err := d.decode(buffer)
	if err != nil {
		return gotime.Clock{}, err
	}

	nanoOfDayLocal, ok := nanoOfDayLocalI.(int64)
	if !ok {
		return gotime.Clock{}, fmt.Errorf("unable to cast [%T] to [int64]", nanoOfDayLocalI)
	}

	offsetI, err := d.decode(buffer)
	if err != nil {
		return gotime.Clock{}, err
	}

	offset, ok := offsetI.(int64)
	if !ok {
		return gotime.Clock{}, fmt.Errorf("unable to cast [%T] to [int64]", offsetI)
	}

	tzName := fmt.Sprintf("UTC%+d", int(offset)/(60*60)) // offset seconds to offset hours
	tz := time.FixedZone(tzName, int(offset))            // create named time zone UTC+(offset in hours)

	return gotime.NewClockOfDayNano(nanoOfDayLocal, tz), nil
}

func (d DecoderV2) decodeLocalTime(buffer *bytes.Buffer) (gotime.LocalClock, error) {
	nanoOfDayLocalI, err := d.decode(buffer)
	if err != nil {
		return gotime.LocalClock{}, err
	}

	nanoOfDayLocal, ok := nanoOfDayLocalI.(int64)
	fmt.Println(nanoOfDayLocal)
	if !ok {
		return gotime.LocalClock{}, fmt.Errorf("unable to cast [%T] to [int64]", nanoOfDayLocalI)
	}

	return gotime.NewLocalClockOfDayNano(nanoOfDayLocal), nil
}

func (d DecoderV2) decodeLocalDateTime(buffer *bytes.Buffer) (gotime.LocalTime, error) {
	epochSecondsI, err := d.decode(buffer)
	if err != nil {
		return gotime.LocalTime{}, err
	}

	epochSeconds, ok := epochSecondsI.(int64)
	if !ok {
		return gotime.LocalTime{}, fmt.Errorf("unable to cast [%T] to [int64]", epochSecondsI)
	}

	nanoOfDayI, err := d.decode(buffer)
	if err != nil {
		return gotime.LocalTime{}, err
	}

	nanoOfDay, ok := nanoOfDayI.(int64)
	if !ok {
		return gotime.LocalTime{}, fmt.Errorf("unable to cast [%T] to [int64]", nanoOfDayI)
	}

	return gotime.NewLocalTimeFromUnix(epochSeconds, nanoOfDay), nil
}

func (d DecoderV2) decodeDateTimeWithZoneOffset(buffer *bytes.Buffer) (time.Time, error) {
	epochSecondsLocalI, err := d.decode(buffer)
	if err != nil {
		return time.Time{}, err
	}

	epochSecondsLocal, ok := epochSecondsLocalI.(int64)
	if !ok {
		return time.Time{}, fmt.Errorf("unable to cat [%T] to [int64]", epochSecondsLocalI)
	}

	nanoOfDayLocalI, err := d.decode(buffer)
	if err != nil {
		return time.Time{}, err
	}

	nanoOfDayLocal, ok := nanoOfDayLocalI.(int64)
	if !ok {
		return time.Time{}, fmt.Errorf("unable to cast [%T] to [int64]", nanoOfDayLocalI)
	}

	offsetI, err := d.decode(buffer)
	if err != nil {
		return time.Time{}, err
	}

	offset, ok := offsetI.(int64)
	if !ok {
		return time.Time{}, fmt.Errorf("unable to cast [%T] to [int64]", offsetI)
	}

	tzName := fmt.Sprintf("UTC%+d", int(offset)/(60*60)) // offset seconds to offset hours
	tz := time.FixedZone(tzName, int(offset))            // create named time zone UTC+(offset in hours)

	return time.Unix(epochSecondsLocal, nanoOfDayLocal).
			In(tz).                                    // change timezone (sets offset)
			Add(time.Duration(-offset) * time.Second), // add back reverse of offset
		nil
}

// todo improve time functions
func (d DecoderV2) decodeTimeWithZoneId(buffer *bytes.Buffer) (time.Time, error) {
	epochSecondLocalI, err := d.decode(buffer)
	if err != nil {
		return time.Time{}, err
	}

	epochSecondLocal, ok := epochSecondLocalI.(int64)
	if !ok {
		return time.Time{}, fmt.Errorf("unable to cast [%T] to [int64]", epochSecondLocalI)
	}

	nanoOfDayLocalI, err := d.decode(buffer)
	if err != nil {
		return time.Time{}, err
	}

	nanoOfDayLocal, ok := nanoOfDayLocalI.(int64)
	if !ok {
		return time.Time{}, fmt.Errorf("unable to cast [%T] to [int64]", nanoOfDayLocalI)
	}

	zoneIdI, err := d.decode(buffer)
	if err != nil {
		return time.Time{}, err
	}

	zoneId, ok := zoneIdI.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("unable to cast [%T] to [string]", zoneIdI)
	}

	loc, err := time.LoadLocation(zoneId)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to load zone info for [%v]", loc)
	}

	timeWithoutOffsetCorrection := time.Unix(epochSecondLocal, nanoOfDayLocal).In(loc)

	_, offset := timeWithoutOffsetCorrection.Zone() // get second offset from zone

	return timeWithoutOffsetCorrection.
			Add(time.Duration(-offset) * time.Second), // correct for zone offset
		nil
}

// todo improve time functions
func (d DecoderV2) decodeDuration(buffer *bytes.Buffer) (time.Duration, error) {
	monthsI, err := d.decode(buffer)
	if err != nil {
		return -1, err
	}

	month, ok := monthsI.(int64)
	if !ok {
		return -1, fmt.Errorf("unable to cast [%T] to [int64]", monthsI)
	}

	daysI, err := d.decode(buffer)
	if err != nil {
		return -1, err
	}

	days, ok := daysI.(int64)
	if !ok {
		return -1, fmt.Errorf("unable to cast [%T] to [int64]", daysI)
	}

	secondsI, err := d.decode(buffer)
	if err != nil {
		return -1, err
	}

	seconds, ok := secondsI.(int64)
	if !ok {
		return -1, fmt.Errorf("unable to cast [%T] to [int64]", secondsI)
	}

	nanoSecondsI, err := d.decode(buffer)
	if err != nil {
		return -1, err
	}

	nanoSeconds, ok := nanoSecondsI.(int64)
	if !ok {
		return -1, fmt.Errorf("unable to cast [%T] to [int64]", nanoSecondsI)
	}

	return time.Duration(month * days * seconds * nanoSeconds), nil
}

func (d DecoderV2) decodePoint2D(buffer *bytes.Buffer) (types.Point2D, error) {
	sridI, err := d.decode(buffer)
	if err != nil {
		return types.Point2D{}, err
	}

	srid, ok := sridI.(int64)
	if !ok {
		return types.Point2D{}, fmt.Errorf("unable to cast [%T] to [int64]", sridI)
	}

	xI, err := d.decode(buffer)
	if err != nil {
		return types.Point2D{}, err
	}

	x, ok := xI.(float64)
	if !ok {
		return types.Point2D{}, fmt.Errorf("unable to cast [%T] to [int64]", xI)
	}

	yI, err := d.decode(buffer)
	if err != nil {
		return types.Point2D{}, err
	}

	y, ok := yI.(float64)
	if !ok {
		return types.Point2D{}, fmt.Errorf("unable to cast [%T] to [int64]", yI)
	}

	return types.Point2D{
		SRID: int(srid),
		X:    x,
		Y:    y,
	}, nil
}

func (d DecoderV2) decodePoint3D(buffer *bytes.Buffer) (types.Point3D, error) {
	sridI, err := d.decode(buffer)
	if err != nil {
		return types.Point3D{}, err
	}

	srid, ok := sridI.(int64)
	if !ok {
		return types.Point3D{}, fmt.Errorf("unable to cast [%T] to [int64]", sridI)
	}

	xI, err := d.decode(buffer)
	if err != nil {
		return types.Point3D{}, err
	}

	x, ok := xI.(float64)
	if !ok {
		return types.Point3D{}, fmt.Errorf("unable to cast [%T] to [int64]", xI)
	}

	yI, err := d.decode(buffer)
	if err != nil {
		return types.Point3D{}, err
	}

	y, ok := yI.(float64)
	if !ok {
		return types.Point3D{}, fmt.Errorf("unable to cast [%T] to [int64]", yI)
	}

	zI, err := d.decode(buffer)
	if err != nil {
		return types.Point3D{}, err
	}

	z, ok := zI.(float64)
	if !ok {
		return types.Point3D{}, fmt.Errorf("unable to cast [%T] to [int64]", zI)
	}

	return types.Point3D{
		SRID: int(srid),
		X:    x,
		Y:    y,
		Z:    z,
	}, nil
}

func (d DecoderV2) decodeNode(buffer *bytes.Buffer) (graph.Node, error) {
	node := graph.Node{}

	nodeIdentityInt, err := d.decode(buffer)
	if err != nil {
		return node, err
	}
	node.NodeIdentity = nodeIdentityInt.(int64)

	labelInt, err := d.decode(buffer)
	if err != nil {
		return node, err
	}
	labelIntSlice, ok := labelInt.([]interface{})
	if !ok {
		return node, errors.New("Expected: Labels []string, but got %T %+v", labelInt, labelInt)
	}
	node.Labels, err = encoding.SliceInterfaceToString(labelIntSlice)
	if err != nil {
		return node, err
	}

	propertiesInt, err := d.decode(buffer)
	if err != nil {
		return node, err
	}
	node.Properties, ok = propertiesInt.(map[string]interface{})
	if !ok {
		return node, errors.New("Expected: Properties map[string]interface{}, but got %T %+v", propertiesInt, propertiesInt)
	}

	return node, nil

}

func (d DecoderV2) decodeRelationship(buffer *bytes.Buffer) (graph.Relationship, error) {
	rel := graph.Relationship{}

	relIdentityInt, err := d.decode(buffer)
	if err != nil {
		return rel, err
	}
	rel.RelIdentity = relIdentityInt.(int64)

	startNodeIdentityInt, err := d.decode(buffer)
	if err != nil {
		return rel, err
	}
	rel.StartNodeIdentity = startNodeIdentityInt.(int64)

	endNodeIdentityInt, err := d.decode(buffer)
	if err != nil {
		return rel, err
	}
	rel.EndNodeIdentity = endNodeIdentityInt.(int64)

	var ok bool
	typeInt, err := d.decode(buffer)
	if err != nil {
		return rel, err
	}
	rel.Type, ok = typeInt.(string)
	if !ok {
		return rel, errors.New("Expected: Type string, but got %T %+v", typeInt, typeInt)
	}

	propertiesInt, err := d.decode(buffer)
	if err != nil {
		return rel, err
	}
	rel.Properties, ok = propertiesInt.(map[string]interface{})
	if !ok {
		return rel, errors.New("Expected: Properties map[string]interface{}, but got %T %+v", propertiesInt, propertiesInt)
	}

	return rel, nil
}

func (d DecoderV2) decodePath(buffer *bytes.Buffer) (graph.Path, error) {
	path := graph.Path{}

	nodesInt, err := d.decode(buffer)
	if err != nil {
		return path, err
	}
	nodesIntSlice, ok := nodesInt.([]interface{})
	if !ok {
		return path, errors.New("Expected: Nodes []Node, but got %T %+v", nodesInt, nodesInt)
	}
	path.Nodes, err = encoding.SliceInterfaceToNode(nodesIntSlice)
	if err != nil {
		return path, err
	}

	relsInt, err := d.decode(buffer)
	if err != nil {
		return path, err
	}
	relsIntSlice, ok := relsInt.([]interface{})
	if !ok {
		return path, errors.New("Expected: Relationships []Relationship, but got %T %+v", relsInt, relsInt)
	}
	path.Relationships, err = encoding.SliceInterfaceToUnboundRelationship(relsIntSlice)
	if err != nil {
		return path, err
	}

	seqInt, err := d.decode(buffer)
	if err != nil {
		return path, err
	}
	seqIntSlice, ok := seqInt.([]interface{})
	if !ok {
		return path, errors.New("Expected: Sequence []int, but got %T %+v", seqInt, seqInt)
	}
	path.Sequence, err = encoding.SliceInterfaceToInt(seqIntSlice)

	return path, err
}

func (d DecoderV2) decodeUnboundRelationship(buffer *bytes.Buffer) (graph.UnboundRelationship, error) {
	rel := graph.UnboundRelationship{}

	relIdentityInt, err := d.decode(buffer)
	if err != nil {
		return rel, err
	}
	rel.RelIdentity = relIdentityInt.(int64)

	var ok bool
	typeInt, err := d.decode(buffer)
	if err != nil {
		return rel, err
	}
	rel.Type, ok = typeInt.(string)
	if !ok {
		return rel, errors.New("Expected: Type string, but got %T %+v", typeInt, typeInt)
	}

	propertiesInt, err := d.decode(buffer)
	if err != nil {
		return rel, err
	}
	rel.Properties, ok = propertiesInt.(map[string]interface{})
	if !ok {
		return rel, errors.New("Expected: Properties map[string]interface{}, but got %T %+v", propertiesInt, propertiesInt)
	}

	return rel, nil
}

func (d DecoderV2) decodeRecordMessage(buffer *bytes.Buffer) (messages.RecordMessage, error) {
	fieldsInt, err := d.decode(buffer)
	if err != nil {
		return messages.RecordMessage{}, err
	}
	fields, ok := fieldsInt.([]interface{})
	if !ok {
		return messages.RecordMessage{}, errors.New("Expected: Fields []interface{}, but got %T %+v", fieldsInt, fieldsInt)
	}

	return messages.NewRecordMessage(fields), nil
}

func (d DecoderV2) decodeFailureMessage(buffer *bytes.Buffer) (messages.FailureMessage, error) {
	metadataInt, err := d.decode(buffer)
	if err != nil {
		return messages.FailureMessage{}, err
	}
	metadata, ok := metadataInt.(map[string]interface{})
	if !ok {
		return messages.FailureMessage{}, errors.New("Expected: Metadata map[string]interface{}, but got %T %+v", metadataInt, metadataInt)
	}

	return messages.NewFailureMessage(metadata), nil
}

func (d DecoderV2) decodeIgnoredMessage(buffer *bytes.Buffer) (messages.IgnoredMessage, error) {
	return messages.NewIgnoredMessage(), nil
}

func (d DecoderV2) decodeSuccessMessage(buffer *bytes.Buffer) (messages.SuccessMessage, error) {
	metadataInt, err := d.decode(buffer)
	if err != nil {
		return messages.SuccessMessage{}, err
	}
	metadata, ok := metadataInt.(map[string]interface{})
	if !ok {
		return messages.SuccessMessage{}, errors.New("Expected: Metadata map[string]interface{}, but got %T %+v", metadataInt, metadataInt)
	}

	return messages.NewSuccessMessage(metadata), nil
}

func (d DecoderV2) decodeDiscardAllMessage(buffer *bytes.Buffer) (messages.DiscardAllMessage, error) {
	return messages.NewDiscardAllMessage(), nil
}

func (d DecoderV2) decodePullAllMessage(buffer *bytes.Buffer) (messages.PullAllMessage, error) {
	return messages.NewPullAllMessage(), nil
}

func (d DecoderV2) decodeResetMessage(buffer *bytes.Buffer) (messages.ResetMessage, error) {
	return messages.NewResetMessage(), nil
}
