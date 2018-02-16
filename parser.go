package pbstream

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

var (
	ErrInvalidLengthSample = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSample   = fmt.Errorf("proto: integer overflow")
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

func ParseInt64(bz []byte) (wire int64, offset int, err error) {
	var uwire uint64
	uwire, offset, err = parseVarUint(bz)
	wire = int64(uwire)
	return
}

func ParseInt32(bz []byte) (wire int32, offset int, err error) {
	var uwire uint64
	uwire, offset, err = parseVarUint(bz)
	wire = int32(uwire)
	return
}

func ParseInt(bz []byte) (wire int, offset int, err error) {
	var uwire uint64
	uwire, offset, err = parseVarUint(bz)
	wire = int(uwire)
	return
}

func ParseUint64(bz []byte) (wire uint64, offset int, err error) {
	return parseVarUint(bz)
}

func ParseUint32(bz []byte) (wire uint32, offset int, err error) {
	var uwire uint64
	uwire, offset, err = parseVarUint(bz)
	wire = uint32(uwire)
	return
}

func ParseBytesField(bz []byte) ([]byte, error) {
	size, offset, err := ParseInt(bz)
	if err != nil {
		return nil, err
	}
	return bz[offset : offset+size], nil
}

func ParseString(bz []byte) (string, error) {
	field, err := ParseBytesField(bz)
	return string(field), err
}

// ParseAnyInt parses any int field,
// uses wire type to determine how to parse the bytes
func ParseAnyInt(wireType int, bz []byte) (val uint64, offset int, err error) {
	switch wireType {
	case 0: // varint
		val, offset, err := parseVarUint(bz)
		return val, offset, err
	case 1: // fixed64
		if len(bz) != 8 {
			return 0, 0, errors.Errorf("Fixed64 but %d bytes", len(bz))
		}
		val := binary.LittleEndian.Uint64(bz)
		return val, 8, nil
	case 5: // fixed32
		if len(bz) != 4 {
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
	wire, offset, err = ParseUint64(bz)
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
	case 0: // varint
		_, offset, err = ParseUint64(bz[i:])
		if err != nil {
			return 0, err
		}
		i += offset
		return i, nil
	case 1: // fixed 64 byte
		i += 8
		return i, nil
	case 2: // length-delimited
		size, offset, err := ParseInt(bz[i:])
		if err != nil {
			return 0, err
		}
		if size < 0 {
			return 0, ErrInvalidLengthSample
		}
		i += offset + size
		return i, nil
	case 3: // begin group (deprecated)
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
	case 4: // end group
		return i, nil
	case 5: // fixed 32 bit field (fixed32, float)
		i += 4
		return i, nil
	default:
		return 0, errors.Errorf("proto: illegal wireType %d", wireType)
	}
}
