/*
Package pbstream will extract arbitrary data from a protobuf
encoded message, using only the field indentifiers as arguments
(not even the original .proto file). It should use minimal
memory and resources and return just the desired field without
parsing the whole structure.

Please see the test cases for usage examples until we improve
the documentation.
*/
package pbstream

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/pkg/errors"
)

var (
	ErrInvalidLengthSample = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSample   = fmt.Errorf("proto: integer overflow")
)

const (
	WireVarint       int = 0
	WireFixed64          = 1
	WireLengthPrefix     = 2
	WireBeginGroup       = 3 // deprecated
	WireEndGroup         = 4 // deprecated
	WireFixed32          = 5
)

// ExtractField goes through this object, field by field until we
// find the field we want.
//
// Second return value is the field type (encoding), which can
// be useful to extract integers
func ExtractField(bz []byte, field int32) ([]byte, int, error) {
	for len(bz) > 0 {
		// parse the header from field type
		offset, fieldNum, wireType, err := parseFieldHeader(bz)
		if err != nil {
			return nil, 0, err
		}

		// we got it!
		if fieldNum == field {
			return bz[offset:], wireType, nil
		}

		// skip field
		skippy, err := skipField(bz)
		if err != nil {
			return nil, 0, err
		}
		if skippy < 0 {
			return nil, 0, errors.WithStack(ErrInvalidLengthSample)
		}
		if (skippy) > len(bz) {
			return nil, 0, errors.WithStack(io.ErrUnexpectedEOF)
		}
		bz = bz[skippy:]
	}
	return nil, 0, errors.Errorf("Desired field %d not found", field)
}

// ExtractPath digs into sub-objects, selecting field #1,
// then field #2 from the bytes that come out, then...
// Returns the final field or an error if anything failed.
func ExtractPath(bz []byte, next int32, rest ...int32) ([]byte, int, error) {
	field, wireType, err := ExtractField(bz, next)
	if err != nil {
		return nil, 0, err
	}
	// recursion guard - we got to the end
	if len(rest) == 0 {
		return field, wireType, nil
	}

	// pop one off the rest list
	next, rest = rest[0], rest[1:]
	// and extract the bytes from the embedded struct in the field
	bz, err = ParseBytesField(field)
	if err != nil {
		return nil, 0, err
	}

	// repeat on sub-structure
	return ExtractPath(bz, next, rest...)
}

// ParseBytesField takes a WireLengthPrefix field, and
// extracts the contents into a byte slice
func ParseBytesField(bz []byte) ([]byte, error) {
	size, offset, err := parseVarUint(bz)
	if err != nil {
		return nil, err
	}
	return bz[offset : offset+int(size)], nil
}

// ParseString takes a WireLengthPrefix field, and
// extracts the contents into a string
func ParseString(bz []byte) (string, error) {
	field, err := ParseBytesField(bz)
	return string(field), err
}

// ParsePackedRepeated handles case when many
// small varints are packed into one field.
//
// Note that the wire type of the field will always be 2 (WireLengthPrefix)
// However, we need to know (out-of-bound info) what kind
// of encoding was used for each number:
// WireVarint, WireFixed64, WireFixed32
//
// See: https://developers.google.com/protocol-buffers/docs/encoding#packed
func ParsePackedRepeated(wire int, bz []byte) ([]uint64, error) {
	data, err := ParseBytesField(bz)
	if err != nil {
		return nil, err
	}

	// pre-allocate a reasonable amount of space
	var bytesPerNum int
	switch wire {
	case WireFixed32:
		bytesPerNum = 4
	case WireFixed64:
		bytesPerNum = 8
	case WireVarint:
		bytesPerNum = 2
	}
	res := make([]uint64, 0, len(data)/bytesPerNum)

	// now, let's keep getting more....
	for len(data) > 0 {
		val, offset, err := ParseAnyInt(wire, data)
		if err != nil {
			return nil, err
		}
		res = append(res, val)
		data = data[offset:]
	}

	return res, nil
}

// UnpackSint takes the raw bytes from ParseAnyInt and
// and unpacks them as needed if they were encoded
// with sint32 or sint64
func UnpackSint(raw uint64) int64 {
	neg := raw&0x1 == 1
	val := int64(raw >> 1)
	if neg {
		val = -val - 1
	}
	return val
}

// ParseFloat64 can decode double fields
func ParseFloat64(wireType int, bz []byte) (float64, error) {
	if wireType != WireFixed64 {
		return 0, errors.Errorf("Unknown wireType for ParseFloat64: %d", wireType)
	}
	if len(bz) < 8 {
		return 0, errors.Errorf("Double but %d bytes", len(bz))
	}
	val := binary.LittleEndian.Uint64(bz)
	return math.Float64frombits(val), nil
}

// ParseFloat32 can decode double fields
func ParseFloat32(wireType int, bz []byte) (float32, error) {
	if wireType != WireFixed32 {
		return 0, errors.Errorf("Unknown wireType for ParseFloat32: %d", wireType)
	}
	if len(bz) < 4 {
		return 0, errors.Errorf("Float but %d bytes", len(bz))
	}
	val := binary.LittleEndian.Uint32(bz)
	return math.Float32frombits(val), nil
}

// ParseAnyInt parses any int field,
// uses wire type to determine how to parse the bytes
func ParseAnyInt(wireType int, bz []byte) (val uint64, offset int, err error) {
	switch wireType {
	case WireVarint:
		val, offset, err := parseVarUint(bz)
		return val, offset, err
	case WireFixed64:
		if len(bz) < 8 {
			return 0, 0, errors.Errorf("Fixed64 but %d bytes", len(bz))
		}
		val := binary.LittleEndian.Uint64(bz)
		return val, 8, nil
	case WireFixed32:
		if len(bz) < 4 {
			return 0, 0, errors.Errorf("Fixed32 but %d bytes", len(bz))
		}
		val := binary.LittleEndian.Uint32(bz)
		return uint64(val), 4, nil
	default:
		return 0, 0, errors.Errorf("Unknown wireType for ParseInt: %d", wireType)
	}
}

// parseVarUint is a helper and returns bytes as uint64
// to be converted by wrapper
func parseVarUint(bz []byte) (wire uint64, offset int, err error) {
	const maxShift uint = 64
	l := len(bz)
	for shift := uint(0); ; shift += 7 {
		if shift >= maxShift {
			err = errors.WithStack(ErrIntOverflowSample)
			return
		}
		if offset >= l {
			err = errors.WithStack(io.ErrUnexpectedEOF)
			return
		}
		b := bz[offset]
		offset++
		wire |= (uint64(b) & 0x7F) << shift
		if b < 0x80 {
			break
		}
	}
	return wire, offset, nil
}

func parseFieldHeader(bz []byte) (offset int, fieldNum int32, wireType int, err error) {
	var wire uint64
	wire, offset, err = parseVarUint(bz)
	if err != nil {
		return
	}
	wireType = int(wire & 0x7)
	fieldNum = int32(wire >> 3)
	if fieldNum <= 0 {
		err = errors.Errorf("proto: Person: illegal tag %d (wire type %d)", fieldNum, wireType)
		return
	}
	return
}

func skipField(bz []byte) (size int, err error) {
	var i int
	offset, _, wireType, err := parseFieldHeader(bz)
	if err != nil {
		return 0, err
	}
	i += offset

	switch wireType {
	case WireVarint:
		_, offset, err = parseVarUint(bz[i:])
		if err != nil {
			return 0, err
		}
		i += offset
		return i, nil
	case WireFixed64:
		i += 8
		return i, nil
	case WireLengthPrefix:
		size, offset, err := parseVarUint(bz[i:])
		if err != nil {
			return 0, err
		}
		if size < 0 {
			return 0, ErrInvalidLengthSample
		}
		i += offset + int(size)
		return i, nil
	case WireBeginGroup: // (deprecated)
		for {
			// we stop if it hits 4, and return up to that point
			_, _, innerWireType, err := parseFieldHeader(bz[i:])
			if innerWireType == 4 {
				return i, nil
			}
			// otherwise, keep skipping the entries in the group
			next, err := skipField(bz[i:])
			if err != nil {
				return 0, err
			}
			i += next
		}
		return i, nil
	case WireEndGroup: // (deprecated)
		return i, nil
	case WireFixed32:
		i += 4
		return i, nil
	default:
		return 0, errors.Errorf("proto: illegal wireType %d", wireType)
	}
}
