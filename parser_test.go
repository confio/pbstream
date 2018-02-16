package pbstream

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type check struct {
	path      []int32
	isMissing bool
	eval      checker
}

func (c check) extractPath(bz []byte) ([]byte, int, error) {
	return ExtractPath(bz, c.path[0], c.path[1:]...)
}

type checker func(int, []byte) error

func checkString(expect string) checker {
	return func(wire int, field []byte) error {
		if wire != WireLengthPrefix {
			return fmt.Errorf("Invalid wire type: %d", wire)
		}
		str, err := ParseString(field)
		if err != nil {
			return err
		}
		if str != expect {
			return fmt.Errorf("Got %s, expected %s", str, expect)
		}
		return nil
	}
}

func checkInt32(expect int32) checker {
	return func(wire int, field []byte) error {
		i, _, err := ParseAnyInt(wire, field)
		if err != nil {
			return err
		}
		if int32(i) != expect {
			return fmt.Errorf("Got %d, expected %d", int32(i), expect)
		}
		return nil
	}
}

func checkInt64(expect int64) checker {
	return func(wire int, field []byte) error {
		i, _, err := ParseAnyInt(wire, field)
		if err != nil {
			return err
		}
		if int64(i) != expect {
			return fmt.Errorf("Got %d, expected %d", int64(i), expect)
		}
		return nil
	}
}

func checkUint32(expect uint32) checker {
	return func(wire int, field []byte) error {
		i, _, err := ParseAnyInt(wire, field)
		if err != nil {
			return err
		}
		if uint32(i) != expect {
			return fmt.Errorf("Got %d, expected %d", uint32(i), expect)
		}
		return nil
	}
}

func checkUint64(expect uint64) checker {
	return func(wire int, field []byte) error {
		i, _, err := ParseAnyInt(wire, field)
		if err != nil {
			return err
		}
		if i != expect {
			return fmt.Errorf("Got %d, expected %d", i, expect)
		}
		return nil
	}
}

func TestExtractField(t *testing.T) {
	// See: _gen/cmd/gen.go to see where data comes from
	cases := []struct {
		pbfile string
		checks []check
	}{
		{
			"testdata/person_john.bin",
			[]check{
				{[]int32{1}, false, checkString("John")},
				{[]int32{2}, false, checkInt32(123)},
				{[]int32{3}, false, checkString("john@doe.com")},
				{[]int32{4}, true, nil},
			},
		},
		{
			"testdata/employee_marmot.bin",
			[]check{
				// Title from top level
				{[]int32{1}, false, checkString("COO")},
				// Embeded Person struct
				{[]int32{2, 1}, false, checkString("Mr. Marmot")},
				{[]int32{2, 2}, false, checkInt32(-37)},
				{[]int32{2, 3}, true, nil},
			},
		},
		{
			"testdata/mixed.bin",
			[]check{
				// {[]int32{1}, false, checkDouble("1.234")},
				// {[]int32{2}, false, checkFloat(56.78)},
				{[]int32{3}, false, checkInt32(654321)},
				{[]int32{4}, false, checkInt64(-8877665544332211)},
				{[]int32{5}, false, checkUint32(87654)},
				{[]int32{6}, false, checkUint64(1122334455667788)},
				// {[]int32{7}, false, checkInt32(162)},
				// {[]int32{8}, false, checkInt(-835)},
				{[]int32{9}, false, checkUint32(19734562)},
				{[]int32{10}, false, checkUint64(2926733)},
				{[]int32{11}, false, checkInt32(-38919)},
				{[]int32{12}, false, checkInt64(20472732987)},
				// {[]int32{13}, false, checkBool(true)},
				{[]int32{14}, false, checkString("Hello")},
				// {[]int32{15}, false, checkBytes([]byte{17, 32, 16, 0, 4})},
				{[]int32{16}, false, checkInt32(3)},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			bz, err := ioutil.ReadFile(tc.pbfile)
			require.NoError(t, err)
			for j, check := range tc.checks {
				field, wire, err := check.extractPath(bz)
				if check.isMissing {
					assert.Error(t, err, "%d", j)
				} else if assert.NoError(t, err, "%d", j) {
					err = check.eval(wire, field)
					assert.NoError(t, err, "%d", j)
				}
			}
		})
	}
}
