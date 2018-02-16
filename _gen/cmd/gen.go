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
	{
		&data.Employee{
			Title: "COO",
			Person: &data.Person{
				Name: "Mr. Marmot",
				Age:  -37,
			},
		},
		"employee_marmot.bin",
	},
	{
		&data.Mixed{
			Flt:  1.234,
			Dbl:  -56.78,
			I32:  654321,
			I64:  -8877665544332211,
			U32:  87654,
			U64:  1122334455667788,
			S32:  162,
			S64:  -835,
			F32:  19734562,
			F64:  2926733,
			Sf32: -38919,
			Sf64: 20472732987,
			B:    true,
			S:    "Hello",
			Bz:   []byte{17, 32, 16, 0, 4},
			En:   data.Mixed_LOCAL,
		},
		"mixed.bin",
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
