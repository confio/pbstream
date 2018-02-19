package pbstream

import (
	"fmt"
	"io/ioutil"
)

func ExampleStruct() {
	bz, err := ioutil.ReadFile("testdata/send_msg.bin")
	if err != nil {
		fmt.Println("Cannot read file")
		return
	}

	str := Parse(bz)
	fee := str.Struct(1)
	feeAmt := fee.Number(1).Int64()
	feeDenom := fee.String(2)
	if err := fee.Close(); err != nil {
		fmt.Printf("Fee parse error: %+v\n", err)
	}
	fmt.Printf("Fee: %d %s\n", feeAmt, feeDenom)

	msg, idx := str.OneOf(2, 3, 4)
	switch idx {
	case 2:
		rcpt := msg.Bytes(2)
		coin := msg.Struct(3)
		sendAmt := coin.Number(1).Int64()
		sendDenom := coin.String(2)
		if err := msg.Close(); err != nil {
			fmt.Printf("Send parse error: %+v\n", err)
		}
		fmt.Printf("SendTx: %d %s to %X\n", sendAmt, sendDenom, rcpt)
	case 3:
		fmt.Println("Is issue msg")
	case 4:
		fmt.Println("Is other msg...")
	default:
		fmt.Printf("Unknown oneof %d\n", idx)
	}
	if err := msg.Close(); err != nil {
		fmt.Printf("Msg error: %+v\n", err)
		return
	}

	// this can check for duplicates, bad-fields
	if err := str.Close(); err != nil {
		fmt.Printf("Buffer parse error: %+v\n", err)
	}

	// Output: Fee: 500 PHO
	// SendTx: 18500 ATOM to 7423126382
}
