package pbstream

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type check struct {
	path []int32
	eval checker
}

func (c check) extractPath(bz []byte) ([]byte, error) {
	return ExtractPath(bz, c.path[0], c.path[1:]...)
}

type checker func([]byte) error

func checkString(expect string) checker {
	return func(field []byte) error {
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
	return func(field []byte) error {
		i, _, err := ParseInt32(field)
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
	cases := []struct {
		pbfile string
		checks []check
	}{
		{
			"testdata/person_john.bin",
			[]check{
				{[]int32{1}, checkString("John")},
				{[]int32{2}, checkInt32(123)},
				{[]int32{3}, checkString("john@doe.com")},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			bz, err := ioutil.ReadFile(tc.pbfile)
			require.NoError(t, err)
			for j, check := range tc.checks {
				field, err := check.extractPath(bz)
				if assert.NoError(t, err, "%d", j) {
					err = check.eval(field)
					assert.NoError(t, err, "%d", j)
				}
			}
		})
	}
}
