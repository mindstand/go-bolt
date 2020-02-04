package encoding_v2

import (
	"bytes"
	"encoding/binary"
	"github.com/mindstand/go-bolt/encoding/encode_consts"
	"github.com/mindstand/go-bolt/log"
	"github.com/stretchr/testify/require"
	"io"
	"math"
	"testing"
	"testing/quick"

	"github.com/mindstand/go-bolt/errors"
)

const (
	maxBufSize = math.MaxUint16
	maxKeySize = 10
)

func createNewTestEncoder() (EncoderV2, io.Reader) {
	buf := bytes.NewBuffer([]byte{})
	return NewEncoder(buf, maxBufSize), buf
}

func TestEncodeNil(t *testing.T) {
	req := require.New(t)
	encoder, buf := createNewTestEncoder()

	req.Nil(encoder.Encode(nil))

	output := make([]byte, maxBufSize)
	outputCount, err := buf.Read(output)
	req.Nil(err)

	expectedBuf := bytes.NewBuffer([]byte{})
	expected := make([]byte, maxBufSize)

	req.Nil(binary.Write(expectedBuf, binary.BigEndian, uint16(1)))
	expectedBuf.Write([]byte{encode_consts.NilMarker})
	expectedBuf.Write(encode_consts.EndMessage)

	expectedCount, err := expectedBuf.Read(expected)
	req.Nil(err)
	req.EqualValues(expected[:expectedCount], output[:outputCount])
}

func TestEncodeBool(t *testing.T) {
	req := require.New(t)

	expected := func(val bool) []byte {
		expectedBuf := bytes.NewBuffer([]byte{})
		expected := make([]byte, maxBufSize)

		req.Nil(binary.Write(expectedBuf, binary.BigEndian, uint16(1)))

		var marker byte

		if val == true {
			marker = encode_consts.TrueMarker
		} else {
			marker = encode_consts.FalseMarker
		}

		expectedBuf.Write([]byte{marker})
		expectedBuf.Write(encode_consts.EndMessage)

		expectedCount, err := expectedBuf.Read(expected)
		req.Nil(err)

		return expected[:expectedCount]
	}

	result := func(val bool) []byte {
		encoder, buf := createNewTestEncoder()
		req.NotNil(encoder)
		req.NotNil(buf)

		req.Nil(encoder.Encode(val))

		output := make([]byte, maxBufSize)
		outputCount, err := buf.Read(output)
		req.Nil(err)
		return output[:outputCount]
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func generateIntExpectedBuf(val int64) ([]byte, error) {
	expectedBuf := bytes.NewBuffer([]byte{})
	expected := make([]byte, maxBufSize)

	switch {
	case val >= math.MinInt64 && val < math.MinInt32:
		binary.Write(expectedBuf, binary.BigEndian, uint16(9))
		expectedBuf.Write([]byte{encode_consts.Int64Marker})
		binary.Write(expectedBuf, binary.BigEndian, int64(val))
	case val >= math.MinInt32 && val < math.MinInt16:
		binary.Write(expectedBuf, binary.BigEndian, uint16(5))
		expectedBuf.Write([]byte{encode_consts.Int32Marker})
		binary.Write(expectedBuf, binary.BigEndian, int32(val))
	case val >= math.MinInt16 && val < math.MinInt8:
		binary.Write(expectedBuf, binary.BigEndian, uint16(3))
		expectedBuf.Write([]byte{encode_consts.Int16Marker})
		binary.Write(expectedBuf, binary.BigEndian, int16(val))
	case val >= math.MinInt8 && val < -16:
		binary.Write(expectedBuf, binary.BigEndian, uint16(2))
		expectedBuf.Write([]byte{encode_consts.Int8Marker})
		binary.Write(expectedBuf, binary.BigEndian, int8(val))
	case val >= -16 && val <= math.MaxInt8:
		binary.Write(expectedBuf, binary.BigEndian, uint16(1))
		binary.Write(expectedBuf, binary.BigEndian, int8(val))
	case val > math.MaxInt8 && val <= math.MaxInt16:
		binary.Write(expectedBuf, binary.BigEndian, uint16(3))
		expectedBuf.Write([]byte{encode_consts.Int16Marker})
		binary.Write(expectedBuf, binary.BigEndian, int16(val))
	case val > math.MaxInt16 && val <= math.MaxInt32:
		binary.Write(expectedBuf, binary.BigEndian, uint16(5))
		expectedBuf.Write([]byte{encode_consts.Int32Marker})
		binary.Write(expectedBuf, binary.BigEndian, int32(val))
	case val > math.MaxInt32 && val <= math.MaxInt64:
		binary.Write(expectedBuf, binary.BigEndian, uint16(9))
		expectedBuf.Write([]byte{encode_consts.Int64Marker})
		binary.Write(expectedBuf, binary.BigEndian, int64(val))
	default:
		return nil, errors.New("Int too long to write: %d", val)
	}
	expectedBuf.Write(encode_consts.EndMessage)

	expectedCount, _ := expectedBuf.Read(expected)

	return expected[:expectedCount], nil
}

func generateIntResultBuf(val interface{}) ([]byte, error) {
	encoder, buf := createNewTestEncoder()

	err := encoder.Encode(val)
	if err != nil {
		return nil, errors.New("Error while encoding: %v", err)
	}

	output := make([]byte, maxBufSize)

	outputCount, err := buf.Read(output)
	if err != nil {
		return nil, errors.New("Error while reading output: %v", err)
	}

	return output[:outputCount], nil
}

func TestEncodeInt(t *testing.T) {
	req := require.New(t)
	expected := func(val int) []byte {
		output, err := generateIntExpectedBuf(int64(val))
		req.Nil(err)

		return output
	}
	result := func(val int) []byte {
		output, err := generateIntResultBuf(val)
		req.Nil(err)

		return output
	}

	req.Nil(quick.CheckEqual(expected, result, nil))

}

func TestEncodeUint(t *testing.T) {
	req := require.New(t)
	expected := func(val uint) []byte {
		output, err := generateIntExpectedBuf(int64(val))
		req.Nil(err)

		return output
	}
	result := func(val uint) []byte {
		output, err := generateIntResultBuf(val)
		req.Nil(err)

		return output
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeInt8(t *testing.T) {
	req := require.New(t)
	expected := func(val int8) []byte {
		output, err := generateIntExpectedBuf(int64(val))
		req.Nil(err)

		return output
	}
	result := func(val int8) []byte {
		output, err := generateIntResultBuf(val)
		req.Nil(err)

		return output
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeUint8(t *testing.T) {
	req := require.New(t)
	expected := func(val uint8) []byte {
		output, err := generateIntExpectedBuf(int64(val))
		req.Nil(err)

		return output
	}
	result := func(val uint8) []byte {
		output, err := generateIntResultBuf(val)
		req.Nil(err)

		return output
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeInt16(t *testing.T) {
	req := require.New(t)
	expected := func(val int16) []byte {
		output, err := generateIntExpectedBuf(int64(val))
		req.Nil(err)

		return output
	}
	result := func(val int16) []byte {
		output, err := generateIntResultBuf(val)
		req.Nil(err)

		return output
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeUint16(t *testing.T) {
	req := require.New(t)
	expected := func(val uint16) []byte {
		output, err := generateIntExpectedBuf(int64(val))
		req.Nil(err)

		return output
	}
	result := func(val uint16) []byte {
		output, err := generateIntResultBuf(val)
		req.Nil(err)

		return output
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeInt32(t *testing.T) {
	req := require.New(t)
	expected := func(val int32) []byte {
		output, err := generateIntExpectedBuf(int64(val))
		req.Nil(err)

		return output
	}
	result := func(val int32) []byte {
		output, err := generateIntResultBuf(val)
		req.Nil(err)

		return output
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeUint32(t *testing.T) {
	req := require.New(t)
	expected := func(val uint32) []byte {
		output, err := generateIntExpectedBuf(int64(val))
		req.Nil(err)

		return output
	}
	result := func(val uint32) []byte {
		output, err := generateIntResultBuf(val)
		req.Nil(err)

		return output
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeInt64(t *testing.T) {
	req := require.New(t)
	expected := func(val int64) []byte {
		output, err := generateIntExpectedBuf(int64(val))
		req.Nil(err)

		return output
	}
	result := func(val int64) []byte {
		output, err := generateIntResultBuf(val)
		req.Nil(err)

		return output
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeFloat32(t *testing.T) {
	req := require.New(t)
	expected := func(val float32) []byte {
		expectedBuf := bytes.NewBuffer([]byte{})
		expected := make([]byte, maxBufSize)

		req.Nil(binary.Write(expectedBuf, binary.BigEndian, uint16(9)))
		expectedBuf.Write([]byte{encode_consts.FloatMarker})
		req.Nil(binary.Write(expectedBuf, binary.BigEndian, float64(val)))
		expectedBuf.Write(encode_consts.EndMessage)

		expectedCount, err := expectedBuf.Read(expected)
		req.Nil(err)

		return expected[:expectedCount]
	}
	result := func(val float32) []byte {
		encoder, buf := createNewTestEncoder()
		req.Nil(encoder.Encode(val))

		output := make([]byte, maxBufSize)
		outputCount, err := buf.Read(output)
		req.Nil(err)

		return output[:outputCount]
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeFloat64(t *testing.T) {
	req := require.New(t)
	expected := func(val float64) []byte {
		expectedBuf := bytes.NewBuffer([]byte{})
		expected := make([]byte, maxBufSize)

		req.Nil(binary.Write(expectedBuf, binary.BigEndian, uint16(9)))
		expectedBuf.Write([]byte{encode_consts.FloatMarker})
		req.Nil(binary.Write(expectedBuf, binary.BigEndian, float64(val)))
		expectedBuf.Write(encode_consts.EndMessage)

		expectedCount, err := expectedBuf.Read(expected)
		req.Nil(err)

		return expected[:expectedCount]
	}
	result := func(val float64) []byte {
		encoder, buf := createNewTestEncoder()
		req.Nil(encoder.Encode(val))

		output := make([]byte, maxBufSize)
		outputCount, err := buf.Read(output)
		req.Nil(err)

		return output[:outputCount]
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeString(t *testing.T) {
	log.SetLevel("trace")
	req := require.New(t)
	expected := func(val string) []byte {
		expectedBuf := bytes.NewBuffer([]byte{})
		resultExpectedBuf := bytes.NewBuffer([]byte{})
		expected := make([]byte, maxBufSize)

		bytes := []byte(val)

		length := len(bytes)

		switch {
		case length <= 15:
			expectedBuf.Write([]byte{byte(encode_consts.TinyStringMarker | length)})
			expectedBuf.Write(bytes)
		case length > 15 && length <= math.MaxUint8:
			expectedBuf.Write([]byte{encode_consts.String8Marker})
			req.Nil(binary.Write(expectedBuf, binary.BigEndian, int8(length)))
			expectedBuf.Write(bytes)
		case length > math.MaxUint8 && length <= math.MaxUint16:
			expectedBuf.Write([]byte{encode_consts.String16Marker})
			req.Nil(binary.Write(expectedBuf, binary.BigEndian, int16(length)))
			expectedBuf.Write(bytes)
		case length > math.MaxUint16 && int64(length) <= math.MaxUint32:
			expectedBuf.Write([]byte{encode_consts.String32Marker})
			req.Nil(binary.Write(expectedBuf, binary.BigEndian, int32(length)))
			expectedBuf.Write(bytes)
		default:
			t.Fatalf("String too long to write: %s", val)
		}

		req.Nil(binary.Write(resultExpectedBuf, binary.BigEndian, uint16(expectedBuf.Len())))
		_, err := resultExpectedBuf.ReadFrom(expectedBuf)
		req.Nil(err)
		resultExpectedBuf.Write(encode_consts.EndMessage)

		expectedCount, err := resultExpectedBuf.Read(expected)
		req.Nil(err)

		return expected[:expectedCount]
	}

	result := func(val string) []byte {
		encoder, buf := createNewTestEncoder()
		req.Nil(encoder.Encode(val))

		output := make([]byte, maxBufSize)
		outputCount, err := buf.Read(output)
		req.Nil(err)

		return output[:outputCount]
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeInterfaceSlice(t *testing.T) {
	req := require.New(t)
	expected := func(val []bool) []byte {
		expectedBuf := bytes.NewBuffer([]byte{})
		resultExpectedBuf := bytes.NewBuffer([]byte{})
		expected := make([]byte, maxBufSize)
		length := len(val)

		switch {
		case length <= 15:
			expectedBuf.Write([]byte{byte(encode_consts.TinySliceMarker | length)})
		case length > 15 && length <= math.MaxUint8:
			expectedBuf.Write([]byte{encode_consts.Slice8Marker})
			req.Nil(binary.Write(expectedBuf, binary.BigEndian, int8(length)))
		case length > math.MaxUint8 && length <= math.MaxUint16:
			expectedBuf.Write([]byte{encode_consts.Slice16Marker})
			req.Nil(binary.Write(expectedBuf, binary.BigEndian, int16(length)))
		case length >= math.MaxUint16 && int64(length) <= math.MaxUint32:
			expectedBuf.Write([]byte{encode_consts.Slice32Marker})
			req.Nil(binary.Write(expectedBuf, binary.BigEndian, int32(length)))
		default:
			t.Fatalf("Slice too long to write: %+v", val)
		}

		var marker byte

		for _, item := range val {
			if item {
				marker = encode_consts.TrueMarker
			} else {
				marker = encode_consts.FalseMarker
			}

			expectedBuf.Write([]byte{marker})
		}

		req.Nil(binary.Write(resultExpectedBuf, binary.BigEndian, uint16(expectedBuf.Len())))
		_, err := resultExpectedBuf.ReadFrom(expectedBuf)
		req.Nil(err)
		resultExpectedBuf.Write(encode_consts.EndMessage)

		expectedCount, err := resultExpectedBuf.Read(expected)
		req.Nil(err)

		return expected[:expectedCount]
	}

	result := func(val []bool) []byte {
		encoder, buf := createNewTestEncoder()
		req.Nil(encoder.Encode(val))

		output := make([]byte, maxBufSize)
		outputCount, err := buf.Read(output)
		req.Nil(err)

		return output[:outputCount]
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

func TestEncodeStringSlice(t *testing.T) {
	req := require.New(t)
	expected := func(val []string) []byte {
		expectedBuf := bytes.NewBuffer([]byte{})
		resultExpectedBuf := bytes.NewBuffer([]byte{})
		expected := make([]byte, maxBufSize)
		length := len(val)

		switch {
		case length <= 15:
			expectedBuf.Write([]byte{byte(encode_consts.TinySliceMarker | length)})
		case length > 15 && length <= math.MaxUint8:
			expectedBuf.Write([]byte{encode_consts.Slice8Marker})
			req.Nil(binary.Write(expectedBuf, binary.BigEndian, int8(length)))
		case length > math.MaxUint8 && length <= math.MaxUint16:
			expectedBuf.Write([]byte{encode_consts.Slice16Marker})
			req.Nil(binary.Write(expectedBuf, binary.BigEndian, int16(length)))
		case length >= math.MaxUint16 && int64(length) <= math.MaxUint32:
			expectedBuf.Write([]byte{encode_consts.Slice32Marker})
			req.Nil(binary.Write(expectedBuf, binary.BigEndian, int32(length)))
		default:
			t.Fatalf("Slice too long to write: %+v", val)
		}

		for _, item := range val {
			bytes := []byte(item)

			length := len(bytes)

			switch {
			case length <= 15:
				expectedBuf.Write([]byte{byte(encode_consts.TinyStringMarker + length)})
				expectedBuf.Write(bytes)
			case length > 15 && length <= math.MaxUint8:
				expectedBuf.Write([]byte{encode_consts.String8Marker})
				req.Nil(binary.Write(expectedBuf, binary.BigEndian, int8(length)))
				expectedBuf.Write(bytes)
			case length > math.MaxUint8 && length <= math.MaxUint16:
				expectedBuf.Write([]byte{encode_consts.String16Marker})
				req.Nil(binary.Write(expectedBuf, binary.BigEndian, int16(length)))
				expectedBuf.Write(bytes)
			case length > math.MaxUint16 && int64(length) <= math.MaxUint32:
				expectedBuf.Write([]byte{encode_consts.String32Marker})
				req.Nil(binary.Write(expectedBuf, binary.BigEndian, int32(length)))
				expectedBuf.Write(bytes)
			default:
				t.Fatalf("String too long to write: %s", val)
			}
		}

		req.Nil(binary.Write(resultExpectedBuf, binary.BigEndian, uint16(expectedBuf.Len())))
		_, err := resultExpectedBuf.ReadFrom(expectedBuf)
		req.Nil(err)
		resultExpectedBuf.Write(encode_consts.EndMessage)

		expectedCount, err := resultExpectedBuf.Read(expected)
		req.Nil(err)

		return expected[:expectedCount]
	}

	result := func(val []string) []byte {
		encoder, buf := createNewTestEncoder()
		req.Nil(encoder.Encode(val))

		output := make([]byte, maxBufSize)
		outputCount, err := buf.Read(output)
		req.Nil(err)

		return output[:outputCount]
	}

	req.Nil(quick.CheckEqual(expected, result, nil))
}

// todo implement this test
func TestEncodeTime(t *testing.T) {
	t.Log("not implemented")
	t.FailNow()
}

// todo implement this test
func TestEncodeDuration(t *testing.T) {
	t.Log("not implemented")
	t.FailNow()
}