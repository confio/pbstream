package pbstream

import (
	"math"

	"github.com/pkg/errors"
)

// Parse takes a message and begins the parsing....
func Parse(bz []byte) *Field {
	return &Field{
		data: bz,
	}
}

// Field is a set of bytes being parsed.
// It can store and detect error
type Field struct {
	data  []byte
	index int // current distance read, advances each call
	err   *multierror
	// TODO: bitmask of viewed/Repeated fields
}

func (f *Field) Bytes(i int) []byte {
	if f == nil {
		return nil
	}
	// TODO
	return nil
}

func (f *Field) String(i int) string {
	// TODO
	return string(f.Bytes(i))
}

func (f *Field) Number(i int) Number {
	if f == nil {
		return Number(0)
	}
	// TODO
	return Number(0)
}

func (f *Field) Field(i int) *Field {
	if f == nil {
		return nil
	}
	// TODO
	return nil
}

func (f *Field) Error() error {
	if f == nil {
		// TODO: return error not found?????
		return nil
	}
	return f.err.Resolve()
}

func (f *Field) Close() error {
	// TODO: skip til end, look for dups
	return f.Error()
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
func (f *Field) RepeatedNumber(i int) IterNum {
	return nil
}

func (f *Field) RepeatedField(i int) IterField {
	return nil
}

// IterNum allows iteration over a series of numbers...
type IterNum interface {
	Valid() bool
	Next()
	Value() Number
	Close() error // (or stored in the parent field???)
}

// IterField allows iteration over a series of fields...
type IterField interface {
	Valid() bool
	Next()
	Value() *Field
	Close() error // (or stored in the parent field???), needed????
}

// Number is the raw bytes parsed from a numeric field
// Caller should interpret them as below
type Number uint64

func (n Number) Int64() int64 {
	return int64(n)
}

func (n Number) Int32() int32 {
	return int32(n)
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

// Add can concatonate, even for empty me
func (me *multierror) Add(err error) *multierror {
	err = errors.WithStack(err)
	var base []error
	if me != nil {
		base = *me
	}
	*me = append(base, err)
	return me
}

func (me *multierror) Resolve() error {
	if me == nil || len(*me) == 0 {
		return nil
	}
	if len(*me) == 1 {
		return (*me)[0]
	}
	return me
}

func (me *multierror) Error() string {
	return "TODO: combine all"
}

// Fmt should work like pkg.Errors, show all sub-errors, concatentated
func (me *multierror) Fmt() string {
	return "TODO: combine all"
}
