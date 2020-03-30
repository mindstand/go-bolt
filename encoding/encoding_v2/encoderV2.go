package encoding_v2

import (
	"encoding/binary"
	"github.com/mindstand/go-bolt/encoding/encode_consts"
	"io"
	"math"
	"reflect"
	"time"

	"bytes"

	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/structures"
)

// EncoderV2 encodes objects of different types to the given stream.
// Attempts to support all builtin golang types, when it can be confidently
// mapped to a data type from: http://alpha.neohq.net/docs/server-manual/bolt-serialization.html#bolt-packstream-structures
// (version v3.1.0-M02 at the time of writing this.
//
// Maps and Slices are a special case, where only
// map[string]interface{} and []interface{} are supported.
// The interface for maps and slices may be more permissive in the future.
type EncoderV2 struct {
	w         io.Writer
	buf       *bytes.Buffer
	chunkSize uint16
}

// NewEncoder Creates a new EncoderV2 object
func NewEncoder(w io.Writer, chunkSize uint16) EncoderV2 {
	return EncoderV2{
		w:         w,
		buf:       &bytes.Buffer{},
		chunkSize: chunkSize,
	}
}

// Marshal is used to marshal an object to the bolt interface encoded bytes
func Marshal(v interface{}) ([]byte, error) {
	x := &bytes.Buffer{}
	err := NewEncoder(x, math.MaxUint16).Encode(v)
	return x.Bytes(), err
}

// write writes to the writer.  Buffers the writes using chunkSize.
func (e EncoderV2) Write(p []byte) (n int, err error) {
	//log.Trace("in encode write")
	//log.Trace(fmt.Sprintf("len writing is %v", len(p)))
	//log.Trace(fmt.Sprintf("%x", p))

	n, err = e.buf.Write(p)
	if err != nil {
		err = errors.Wrap(err, "An error occurred writing to encoder temp buffer")
		return n, err
	}

	length := e.buf.Len()
	for length >= int(e.chunkSize) {
		if err := binary.Write(e.w, binary.BigEndian, e.chunkSize); err != nil {
			return 0, errors.Wrap(err, "An error occured writing chunksize")
		}

		numWritten, err := e.w.Write(e.buf.Next(int(e.chunkSize)))
		if err != nil {
			err = errors.Wrap(err, "An error occured writing a chunk")
		}

		return numWritten, err
	}

	return n, nil
}

// flush finishes the encoding stream by flushing it to the writer
func (e EncoderV2) flush() error {
	length := e.buf.Len()
	if length > 0 {
		if err := binary.Write(e.w, binary.BigEndian, uint16(length)); err != nil {
			return errors.Wrap(err, "An error occured writing length bytes during flush")
		}

		if _, err := e.buf.WriteTo(e.w); err != nil {
			return errors.Wrap(err, "An error occured writing message bytes during flush")
		}
	}

	_, err := e.w.Write(encode_consts.EndMessage)
	if err != nil {
		return errors.Wrap(err, "An error occurred ending encoding message")
	}
	e.buf.Reset()

	return nil
}

// Encode encodes an object to the stream
func (e EncoderV2) Encode(iVal interface{}) error {

	err := e.encode(iVal)
	if err != nil {
		return err
	}

	// Whatever is left in the buffer for the chunk at the end, write it out
	return e.flush()
}

// Encode encodes an object to the stream
func (e EncoderV2) encode(iVal interface{}) error {

	var err error
	switch val := iVal.(type) {
	case nil:
		err = e.encodeNil()
	case bool:
		err = e.encodeBool(val)
	case int:
		err = e.encodeInt(int64(val))
	case int8:
		err = e.encodeInt(int64(val))
	case int16:
		err = e.encodeInt(int64(val))
	case int32:
		err = e.encodeInt(int64(val))
	case int64:
		err = e.encodeInt(val)
	case uint:
		err = e.encodeInt(int64(val))
	case uint8:
		err = e.encodeInt(int64(val))
	case uint16:
		err = e.encodeInt(int64(val))
	case uint32:
		err = e.encodeInt(int64(val))
	case uint64:
		if val > math.MaxInt64 {
			return errors.New("Integer too big: %d. Max integer supported: %d", val, int64(math.MaxInt64))
		}
		err = e.encodeInt(int64(val))
	case float32:
		err = e.encodeFloat(float64(val))
	case float64:
		err = e.encodeFloat(val)
	case string:
		err = e.encodeString(val)
	case []interface{}:
		err = e.encodeSlice(val)
	case map[string]interface{}:
		err = e.encodeMap(val)
	case time.Time:
		err = e.encodeTime(val)
	case time.Duration:
		err = e.encodeDuration(val)
	case structures.Structure:
		err = e.encodeStructure(val)
	default:
		// arbitrary slice types
		if reflect.TypeOf(iVal).Kind() == reflect.Slice {
			s := reflect.ValueOf(iVal)
			newSlice := make([]interface{}, s.Len())
			for i := 0; i < s.Len(); i++ {
				newSlice[i] = s.Index(i).Interface()
			}
			return e.encodeSlice(newSlice)
		}

		return errors.New("Unrecognized type when encoding data for Bolt transport: %T %+v", val, val)
	}

	return err
}

func roundTime(input float64) int {
	var result float64

	if input < 0 {
		result = math.Ceil(input - 0.5)
	} else {
		result = math.Floor(input + 0.5)
	}

	// only interested in integer, ignore fractional
	i, _ := math.Modf(result)

	return int(i)
}

// encodeTime encodes native golang time.Time object as neo DateTime
func (e EncoderV2) encodeTime(t time.Time) error {
	_, offset := t.Zone()

	t = t.Add(time.Duration(offset) * time.Second) // add offset before converting to unix local
	epochSeconds := t.Unix()
	nano := t.Nanosecond()

	_, err := e.Write([]byte{byte(encode_consts.TinyStructMarker | encode_consts.DateTimeStructSize)})
	if err != nil {
		return err
	}

	_, err = e.Write([]byte{encode_consts.DateTimeWithZoneOffsetSignature})
	if err != nil {
		return err
	}

	err = e.encode(epochSeconds)
	if err != nil {
		return err
	}

	err = e.encode(nano)
	if err != nil {
		return err
	}

	return e.encode(offset)
}

func (e EncoderV2) encodeDuration(d time.Duration) error {
	months := d.Seconds() / 2600640
	days := d.Seconds() / 86400
	seconds := d.Seconds()
	nano := d.Nanoseconds()

	_, err := e.Write([]byte{byte(encode_consts.TinyStructMarker | encode_consts.DurationTimeStructSize)})
	if err != nil {
		return err
	}

	_, err = e.Write([]byte{encode_consts.DurationSignature})
	if err != nil {
		return err
	}

	err = e.encode(months)
	if err != nil {
		return err
	}

	err = e.encode(days)
	if err != nil {
		return err
	}

	err = e.encode(seconds)
	if err != nil {
		return err
	}

	return e.encode(nano)
}

func (e EncoderV2) encodeNil() error {
	_, err := e.Write([]byte{encode_consts.NilMarker})
	return err
}

func (e EncoderV2) encodeBool(val bool) error {
	var err error
	if val {
		_, err = e.Write([]byte{encode_consts.TrueMarker})
	} else {
		_, err = e.Write([]byte{encode_consts.FalseMarker})
	}
	return err
}

func (e EncoderV2) encodeInt(val int64) error {
	var err error
	switch {
	case val >= math.MinInt64 && val < math.MinInt32:
		// Write as INT_64
		if _, err = e.Write([]byte{encode_consts.Int64Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, val)
	case val >= math.MinInt32 && val < math.MinInt16:
		// Write as INT_32
		if _, err = e.Write([]byte{encode_consts.Int32Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, int32(val))
	case val >= math.MinInt16 && val < math.MinInt8:
		// Write as INT_16
		if _, err = e.Write([]byte{encode_consts.Int16Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, int16(val))
	case val >= math.MinInt8 && val < -16:
		// Write as INT_8
		if _, err = e.Write([]byte{encode_consts.Int8Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, int8(val))
	case val >= -16 && val <= math.MaxInt8:
		// Write as TINY_INT
		err = binary.Write(e, binary.BigEndian, int8(val))
	case val > math.MaxInt8 && val <= math.MaxInt16:
		// Write as INT_16
		if _, err = e.Write([]byte{encode_consts.Int16Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, int16(val))
	case val > math.MaxInt16 && val <= math.MaxInt32:
		// Write as INT_32
		if _, err = e.Write([]byte{encode_consts.Int32Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, int32(val))
	case val > math.MaxInt32 && val <= math.MaxInt64:
		// Write as INT_64
		if _, err = e.Write([]byte{encode_consts.Int64Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, val)
	default:
		return errors.New("Int too long to write: %d", val)
	}
	if err != nil {
		return errors.Wrap(err, "An error occurred writing an int to bolt")
	}
	return err
}

func (e EncoderV2) encodeFloat(val float64) error {
	if _, err := e.Write([]byte{encode_consts.FloatMarker}); err != nil {
		return err
	}

	err := binary.Write(e, binary.BigEndian, val)
	if err != nil {
		return errors.Wrap(err, "An error occured writing a float to bolt")
	}

	return err
}

func (e EncoderV2) encodeString(val string) error {
	var err error
	bytes := []byte(val)

	length := len(bytes)
	switch {
	case length <= 15:
		if _, err = e.Write([]byte{byte(encode_consts.TinyStringMarker | length)}); err != nil {
			return err
		}
		_, err = e.Write(bytes)
	case length > 15 && length <= math.MaxUint8:
		if _, err = e.Write([]byte{encode_consts.String8Marker}); err != nil {
			return err
		}
		if err = binary.Write(e, binary.BigEndian, int8(length)); err != nil {
			return err
		}
		_, err = e.Write(bytes)
	case length > math.MaxUint8 && length <= math.MaxUint16:
		if _, err = e.Write([]byte{encode_consts.String16Marker}); err != nil {
			return err
		}
		if err = binary.Write(e, binary.BigEndian, int16(length)); err != nil {
			return err
		}
		_, err = e.Write(bytes)
	case length > math.MaxUint16 && int64(length) <= math.MaxUint32:
		if _, err = e.Write([]byte{encode_consts.String32Marker}); err != nil {
			return err
		}
		if err = binary.Write(e, binary.BigEndian, int32(length)); err != nil {
			return err
		}
		_, err = e.Write(bytes)
	default:
		return errors.New("String too long to write: %s", val)
	}
	return err
}

func (e EncoderV2) encodeSlice(val []interface{}) error {
	length := len(val)
	switch {
	case length <= 15:
		if _, err := e.Write([]byte{byte(encode_consts.TinySliceMarker | length)}); err != nil {
			return err
		}
	case length > 15 && length <= math.MaxUint8:
		if _, err := e.Write([]byte{encode_consts.Slice8Marker}); err != nil {
			return err
		}
		if err := binary.Write(e, binary.BigEndian, int8(length)); err != nil {
			return err
		}
	case length > math.MaxUint8 && length <= math.MaxUint16:
		if _, err := e.Write([]byte{encode_consts.Slice16Marker}); err != nil {
			return err
		}
		if err := binary.Write(e, binary.BigEndian, int16(length)); err != nil {
			return err
		}
	case length >= math.MaxUint16 && int64(length) <= math.MaxUint32:
		if _, err := e.Write([]byte{encode_consts.Slice32Marker}); err != nil {
			return err
		}
		if err := binary.Write(e, binary.BigEndian, int32(length)); err != nil {
			return err
		}
	default:
		return errors.New("Slice too long to write: %+v", val)
	}

	// Encode Slice values
	for _, item := range val {
		if err := e.encode(item); err != nil {
			return err
		}
	}

	return nil
}

func (e EncoderV2) encodeMap(val map[string]interface{}) error {
	length := len(val)
	switch {
	case length <= 15:
		if _, err := e.Write([]byte{byte(encode_consts.TinyMapMarker | length)}); err != nil {
			return err
		}
	case length > 15 && length <= math.MaxUint8:
		if _, err := e.Write([]byte{encode_consts.Map8Marker}); err != nil {
			return err
		}
		if err := binary.Write(e, binary.BigEndian, int8(length)); err != nil {
			return err
		}
	case length > math.MaxUint8 && length <= math.MaxUint16:
		if _, err := e.Write([]byte{encode_consts.Map16Marker}); err != nil {
			return err
		}
		if err := binary.Write(e, binary.BigEndian, int16(length)); err != nil {
			return err
		}
	case length >= math.MaxUint16 && int64(length) <= math.MaxUint32:
		if _, err := e.Write([]byte{encode_consts.Map32Marker}); err != nil {
			return err
		}
		if err := binary.Write(e, binary.BigEndian, int32(length)); err != nil {
			return err
		}
	default:
		return errors.New("Map too long to write: %+v", val)
	}

	// Encode Map values
	for k, v := range val {
		if err := e.encode(k); err != nil {
			return err
		}
		if err := e.encode(v); err != nil {
			return err
		}
	}

	return nil
}

func (e EncoderV2) encodeStructure(val structures.Structure) error {

	fields := val.AllFields()
	length := len(fields)
	switch {
	case length <= 15:
		if _, err := e.Write([]byte{byte(encode_consts.TinyStructMarker | length)}); err != nil {
			return err
		}
	case length > 15 && length <= math.MaxUint8:
		if _, err := e.Write([]byte{encode_consts.Struct8Marker}); err != nil {
			return err
		}
		if err := binary.Write(e, binary.BigEndian, int8(length)); err != nil {
			return err
		}
	case length > math.MaxUint8 && length <= math.MaxUint16:
		if _, err := e.Write([]byte{encode_consts.Struct16Marker}); err != nil {
			return err
		}
		if err := binary.Write(e, binary.BigEndian, int16(length)); err != nil {
			return err
		}
	default:
		return errors.New("Structure too long to write: %+v", val)
	}

	_, err := e.Write([]byte{byte(val.Signature())})
	if err != nil {
		return errors.Wrap(err, "An error occurred writing to encoder a struct field")
	}

	for _, field := range fields {
		if err := e.encode(field); err != nil {
			return errors.Wrap(err, "An error occurred encoding a struct field")
		}
	}

	return nil
}
