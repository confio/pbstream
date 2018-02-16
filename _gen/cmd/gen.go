package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	data "github.com/confio/pbstream/_gen"
	"github.com/gogo/protobuf/proto"
)

// we will generate a file for every element in outputs
var outputs = []struct {
	obj  proto.Message
	file string
}{
	{
		&data.Person{Name: "John", Age: 123, Email: "john@doe.com"},
		"person_john.bin",
	},
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gen [output_dir]")
		os.Exit(1)
	}
	outdir := os.Args[1]

	for i, out := range outputs {
		bz, err := proto.Marshal(out.obj)
		if err != nil {
			fmt.Printf("Error generating %d: %v\n", i, err)
		}
		path := filepath.Join(outdir, out.file)
		err = ioutil.WriteFile(path, bz, 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", path, err)
		}
	}
}
