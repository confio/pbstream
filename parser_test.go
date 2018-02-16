package pbstream

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnpackSint(t *testing.T) {
	cases := []struct {
		in  uint64
		out int64
	}{
		{0, 0},
		{1, -1},
		{2, 1},
		{3, -2},
		{4294967294, 2147483647},
		{4294967295, -2147483648},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			assert.Equal(t, tc.out, UnpackSint(tc.in))
		})
	}
}

func TestExtractField(t *testing.T) {
	// See: _gen/cmd/gen.go to see where data comes from
	cases := []struct {
		pbfile string
		checks []check
	}{
		// test simplest struct
		0: {
			"testdata/person_john.bin",
			[]check{
				{[]int32{1}, false, assertString("John")},
				{[]int32{2}, false, assertInt32(123)},
				{[]int32{3}, false, assertString("john@doe.com")},
				{[]int32{4}, true, nil},
			},
		},
		// test embedded struct
		1: {
			"testdata/employee_marmot.bin",
			[]check{
				// Title from top level
				{[]int32{1}, false, assertString("COO")},
				// Embeded Person struct
				{[]int32{2, 1}, false, assertString("Mr. Marmot")},
				{[]int32{2, 2}, false, assertInt32(-37)},
				{[]int32{2, 3}, true, nil},
			},
		},
		// test all primative types
		2: {
			"testdata/mixed.bin",
			[]check{
				{[]int32{1}, false, assertFloat32(1.234)},
				{[]int32{2}, false, assertFloat64(-56.78)},
				{[]int32{3}, false, assertInt32(654321)},
				{[]int32{4}, false, assertInt64(-8877665544332211)},
				{[]int32{5}, false, assertUint32(87654)},
				{[]int32{6}, false, assertUint64(1122334455667788)},
				{[]int32{7}, false, assertSint32(162)},
				{[]int32{8}, false, assertSint64(-835)},
				{[]int32{9}, false, assertUint32(19734562)},
				{[]int32{10}, false, assertUint64(2926733)},
				{[]int32{11}, false, assertInt32(-38919)},
				{[]int32{12}, false, assertInt64(20472732987)},
				{[]int32{13}, false, assertInt32(1)},
				{[]int32{14}, false, assertString("Hello")},
				{[]int32{15}, false, assertBytes([]byte{17, 32, 16, 0, 4})},
				{[]int32{16}, false, assertInt32(3)},
			},
		},
		// this tests use of repeated fields
		// TODO: handle repeated structs
		3: {
			"testdata/phonebook.bin",
			[]check{
				{[]int32{1}, false, assertString("Friends")},

				// by defaulit only gets first field....
				// TODO: get them all
				// {[]int32{2, 1}, false, assertString("John")},
				// {[]int32{2, 2}, false, assertString("123-4567")},
				// handle packed repeated fields for varint

				{[]int32{3}, false, assertRepeatedInt(
					WireVarint,
					[]int64{532, -344, 3454230, 543, -234})},
				{[]int32{4}, false, assertRepeatedUint(
					WireFixed32,
					[]uint64{123, 4567, 846273})},
				{[]int32{5}, false, assertInt32(34)},
			},
		},
		// oneof Send
		4: {
			"testdata/send_msg.bin",
			[]check{
				// fee
				{[]int32{1, 1}, false, assertInt64(500)},
				{[]int32{1, 2}, false, assertString("PHO")},
				// check sendtx: recpt and amount
				{[]int32{2, 2}, false, assertBytes([]byte{0x74, 0x23, 0x12, 0x63, 0x82})},
				{[]int32{2, 3, 1}, false, assertInt64(18500)},
				{[]int32{2, 3, 2}, false, assertString("ATOM")},
				// missing
				{[]int32{3}, true, nil},
				{[]int32{3, 1}, true, nil},
			},
		},
		// oneof Issue
		5: {
			"testdata/issue_msg.bin",
			[]check{
				// fee
				{[]int32{1, 1}, false, assertInt64(600)},
				{[]int32{1, 2}, false, assertString("SPC")},
				// missing
				{[]int32{2}, true, nil},
				{[]int32{2, 3}, true, nil},
				// check issuetx: recpt and amount
				{[]int32{3, 1}, false, assertBytes([]byte("G7nh98y7un&YBiuBHO*0"))},
				{[]int32{3, 2, 1}, false, assertInt64(777888)},
				{[]int32{3, 2, 2}, false, assertString("WIN")},
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
					check.eval(t, wire, field)
				}
			}
		})
	}
}

// ExampleExtractPath documents parsing a tx
//
// Code to create it in _gen/cmd/gen.go
//
//   &data.Tx{
//       Fee: &data.Coin{
//           Amount: 500,
//           Denom:  "PHO",
//       },
//       Msg: &data.Tx_Send{
//           &data.SendMsg{
//               Sender:    []byte("12345678901234567890"),
//               Recipient: []byte{0x74, 0x23, 0x12, 0x63, 0x82},
//               Amount: &data.Coin{
//                   Amount: 18500,
//                   Denom:  "ATOM",
//               },
//           },
//       },
//   },
//   "send_msg.bin",
//
// Spec in https://github.com/confio/pbstream/blob/master/_gen/sendtx.proto
func ExampleExtractPath() {
	// In real code, you would check those errors....
	bz, _ := ioutil.ReadFile("testdata/send_msg.bin")

	// get the fee bytes
	raw, wireType, _ := ExtractPath(bz, 1, 1)
	feeAmt, _, _ := ParseAnyInt(wireType, raw)
	raw, _, _ = ExtractPath(bz, 1, 2)
	feeDenom, _ := ParseString(raw)
	fmt.Printf("Fee: %d %s\n", feeAmt, feeDenom)

	// check if it is sendTx
	raw, _, err := ExtractField(bz, 2)
	if err != nil {
		fmt.Println("Not send msg")
	} else {
		sendMsg, _ := ParseBytesField(raw)
		raw, _, _ = ExtractField(sendMsg, 2)
		rcpt, _ := ParseBytesField(raw)
		raw, wireType, _ = ExtractPath(sendMsg, 3, 1)
		sendAmt, _, _ := ParseAnyInt(wireType, raw)
		raw, _, _ = ExtractPath(sendMsg, 3, 2)
		sendDenom, _ := ParseString(raw)
		fmt.Printf("SendTx: %d %s to %X\n", sendAmt, sendDenom, rcpt)
	}

	_, _, err = ExtractField(bz, 3)
	if err != nil {
		fmt.Println("Not issue msg")
	}

	// Output: Fee: 500 PHO
	// SendTx: 18500 ATOM to 7423126382
	// Not issue msg
}

type check struct {
	path      []int32
	isMissing bool
	eval      assertion
}

func (c check) extractPath(bz []byte) ([]byte, int, error) {
	return ExtractPath(bz, c.path[0], c.path[1:]...)
}

type assertion func(*testing.T, int, []byte)

func assertString(expect string) assertion {
	return func(t *testing.T, wire int, field []byte) {
		if !assert.Equal(t, wire, WireLengthPrefix) {
			return
		}
		str, err := ParseString(field)
		assert.NoError(t, err)
		assert.Equal(t, expect, str)
	}
}

func assertBytes(expect []byte) assertion {
	return func(t *testing.T, wire int, field []byte) {
		if !assert.Equal(t, wire, WireLengthPrefix) {
			return
		}
		bz, err := ParseBytesField(field)
		assert.NoError(t, err)
		assert.Equal(t, expect, bz)
	}
}

func assertInt32(expect int32) assertion {
	return func(t *testing.T, wire int, field []byte) {
		i, _, err := ParseAnyInt(wire, field)
		assert.NoError(t, err)
		assert.Equal(t, expect, int32(i))
	}
}

func assertInt64(expect int64) assertion {
	return func(t *testing.T, wire int, field []byte) {
		i, _, err := ParseAnyInt(wire, field)
		assert.NoError(t, err)
		assert.Equal(t, expect, int64(i))
	}
}

func assertSint32(expect int32) assertion {
	return func(t *testing.T, wire int, field []byte) {
		raw, _, err := ParseAnyInt(wire, field)
		assert.NoError(t, err)
		assert.Equal(t, expect, int32(UnpackSint(raw)))
	}
}

func assertSint64(expect int64) assertion {
	return func(t *testing.T, wire int, field []byte) {
		raw, _, err := ParseAnyInt(wire, field)
		assert.NoError(t, err)
		assert.Equal(t, expect, int64(UnpackSint(raw)))
	}
}

func assertUint32(expect uint32) assertion {
	return func(t *testing.T, wire int, field []byte) {
		i, _, err := ParseAnyInt(wire, field)
		assert.NoError(t, err)
		assert.Equal(t, expect, uint32(i))
	}
}

func assertUint64(expect uint64) assertion {
	return func(t *testing.T, wire int, field []byte) {
		i, _, err := ParseAnyInt(wire, field)
		assert.NoError(t, err)
		assert.Equal(t, expect, i)
	}
}

func assertFloat64(expect float64) assertion {
	return func(t *testing.T, wire int, field []byte) {
		f, err := ParseFloat64(wire, field)
		assert.NoError(t, err)
		assert.InEpsilon(t, expect, f, 0.0001)
	}
}

func assertFloat32(expect float32) assertion {
	return func(t *testing.T, wire int, field []byte) {
		f, err := ParseFloat32(wire, field)
		assert.NoError(t, err)
		assert.InEpsilon(t, expect, f, 0.0001)
	}
}

func assertRepeatedUint(wire int, expect []uint64) assertion {
	return func(t *testing.T, _ int, field []byte) {
		vals, err := ParsePackedRepeated(wire, field)
		assert.NoError(t, err)
		assert.Equal(t, expect, vals)
	}
}

func assertRepeatedInt(wire int, expect []int64) assertion {
	return func(t *testing.T, _ int, field []byte) {
		raws, err := ParsePackedRepeated(wire, field)
		vals := make([]int64, len(raws))
		for i := range raws {
			vals[i] = int64(raws[i])
		}
		assert.NoError(t, err)
		assert.Equal(t, expect, vals)
	}
}
