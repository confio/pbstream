package pbstream

import (
	"fmt"
	"io/ioutil"
)

func ExampleField() {
	bz, err := ioutil.ReadFile("testdata/send_msg.bin")
	if err != nil {
		fmt.Println("Cannot read file")
		return
	}

	str := Parse(bz)
	fee := str.Field(1)
	feeAmt := fee.Number(1).Int64()
	feeDenom := fee.String(2)
	if err := fee.Close(); err != nil {
		fmt.Printf("Fee parse error: %+v\n", err)
	}
	fmt.Printf("Fee: %d %s\n", feeAmt, feeDenom)

	send := str.Field(2)
	if send == nil {
		fmt.Println("Not send msg")
	} else {
		rcpt := send.Bytes(2)
		coin := send.Field(3)
		sendAmt := coin.Number(1).Int64()
		sendDenom := coin.String(2)
		if err := send.Close(); err != nil {
			fmt.Printf("Send parse error: %+v\n", err)
		}
		fmt.Printf("SendTx: %d %s to %X\n", sendAmt, sendDenom, rcpt)
	}

	if str.Field(3) == nil {
		fmt.Println("Not issue msg")
	}
	// this can check for duplicates, bad-fields
	if err := str.Close(); err != nil {
		fmt.Printf("Buffer parse error: %+v\n", err)
	}

	// Output: Fee: 500 PHO
	// SendTx: 18500 ATOM to 7423126382
	// Not issue msg
}
