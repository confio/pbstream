package pbstream

import (
	"fmt"
	"math"
	"strings"

	"github.com/pkg/errors"
)

// Parse takes a message and begins the parsing....
func Parse(bz []byte) *Struct {
	return &Struct{
		data: bz,
	}
}

// Struct is a set of bytes being parsed.
// It can store and detect error
type Struct struct {
	data  []byte
	index int // current distance read, advances each call
	err   multierror
	// TODO: bitmask of viewed/Repeated fields
}

// AddError will append one more error to our result
func (s *Struct) AddError(err error) {
	s.err = s.err.WithError(err)
}

func (s *Struct) Bytes(i int) []byte {
	if s == nil {
		return nil
	}
	bz, _, err := ExtractField(s.data, int32(i))
	if err != nil {
		s.AddError(err)
		return nil
	}
	res, err := ParseBytesField(bz)
	if err != nil {
		s.AddError(err)
		return nil
	}
	return res
}

func (s *Struct) String(i int) string {
	// TODO
	return string(s.Bytes(i))
}

func (s *Struct) Number(i int) Number {
	if s == nil {
		return Number(0)
	}
	bz, wireType, err := ExtractField(s.data, int32(i))
	if err != nil {
		s.AddError(err)
		return Number(0)
	}
	val, _, err := ParseAnyInt(wireType, bz)
	if err != nil {
		s.AddError(err)
		return Number(0)
	}
	return Number(val)
}

func (s *Struct) Struct(i int) *Struct {
	if s == nil {
		return nil
	}
	bz := s.Bytes(i)
	return Parse(bz)
}

// OneOf will find the first field that matches any of those choices
func (s *Struct) OneOf(choices ...int32) (*Struct, int) {
	raw, _, field, err := extractAnyField(s.data, choices...)
	if err != nil {
		s.AddError(err)
		return nil, 0
	}
	bz, err := ParseBytesField(raw)
	if err != nil {
		s.AddError(err)
		return nil, 0
	}
	return Parse(bz), int(field)
}

func (s *Struct) Error() error {
	if s == nil {
		// TODO: return error not found?????
		return nil
	}
	return s.err.Resolve()
}

func (s *Struct) Close() error {
	// TODO: skip til end, look for dups
	return s.Error()
}

/*
RepeatedNumber gives us an iterator to see all the numbers
at the field.

  var sum int32
  iter := f.RepeatedNumber(3)
  for ; iter.Valid(); iter.Next() {
    sum += iter.Value().Int32
  }
  if err := iter.Close(); err != nil {
      return err
  }
*/
func (s *Struct) RepeatedNumber(i int) IterNum {
	return nil
}

func (s *Struct) RepeatedStruct(i int) IterStruct {
	return nil
}

// IterNum allows iteration over a series of numbers...
type IterNum interface {
	Valid() bool
	Next()
	Value() Number
	Close() error // (or stored in the parent struct???)
}

// IterStruct allows iteration over a series of structs...
type IterStruct interface {
	Valid() bool
	Next()
	Value() *Struct
	Close() error // (or stored in the parent struct???), needed????
}

// Number is the raw bytes parsed from a numeric struct
// Caller should interpret them as below
type Number uint64

func (n Number) Int64() int64 {
	return int64(n)
}

func (n Number) Int32() int32 {
	return int32(n)
}

func (n Number) Uint64() uint64 {
	return uint64(n)
}

func (n Number) Uint32() uint32 {
	return uint32(n)
}

func (n Number) Bool() bool {
	return n != 0
}

func (n Number) Float64() float64 {
	return math.Float64frombits(uint64(n))
}

func (n Number) Sint64() int64 {
	return UnpackSint(uint64(n))
}

// multierror does nice handling to join errors
type multierror []error

// WithError can concatonate, even for empty me,
// returns new mutlierror
func (me multierror) WithError(err error) multierror {
	err = errors.WithStack(err)
	var base multierror
	if me != nil {
		base = me
	}
	return append(base, err)
}

func (me multierror) Resolve() error {
	if me == nil || len(me) == 0 {
		return nil
	}
	if len(me) == 1 {
		return me[0]
	}
	return me
}

func (me multierror) Error() string {
	res := make([]string, len(me))
	for i := range me {
		res[i] = fmt.Sprintf("%d: %+v\n", i, me[i])
	}
	return strings.Join(res, "\n")
}

// Fmt should work like pkg.Errors, show all sub-errors, concatentated
func (me multierror) Fmt() string {
	return "TODO: combine all"
}

// Bitmask stores info for up to 32 fields,
// Each one is represented by 2 bits:
//
//   * 0 - never seen, single field
//   * 1 - already seen, error upon next seen
//   * 2 - expect field multiple times
//
// Valid transitions are:
//
//   * 0 -> {Add, Close} -> 1
//   * 0 -> {Repeat} -> 2
//   * 2 -> {Add} -> 2
//   * 2 -> {Close} -> 1
//   * All others will return errors
type Bitmask uint64

func (b *Bitmask) Seen(i int) error {
	// TODO
	return nil
}

func (b *Bitmask) Close(i int) error {
	// TODO
	return nil
}

func (b *Bitmask) Repeat(i int) error {
	// TODO
	return nil
}
