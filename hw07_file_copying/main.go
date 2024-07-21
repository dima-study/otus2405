package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	from, to      string
	limit, offset int64
)

func init() {
	flag.StringVar(&from, "from", "", "file to read from *required*")
	flag.StringVar(&to, "to", "", "file to write to *required*")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy *>=0*")
	flag.Int64Var(&offset, "offset", 0, "offset in input file *>=0*")
}

func main() {
	flag.Parse()

	switch {
	case from == "":
		flag.Usage()
		fmt.Println("\nInvalid arguments: 'from' is required")
		os.Exit(1)
	case to == "":
		flag.Usage()
		fmt.Println("\nInvalid arguments: 'to' is required")
		os.Exit(1)
	case offset < 0:
		flag.Usage()
		fmt.Println("\nInvalid arguments: 'offset' must be >=0")
		os.Exit(1)
	case limit < 0:
		flag.Usage()
		fmt.Println("\nInvalid arguments: 'limit' must be >=0")
		os.Exit(1)
	}

	err := Copy(from, to, offset, limit)
	if err != nil {
		fmt.Printf("Error occurred: %s\n", err)
		os.Exit(2)
	}

	fmt.Println("Success!")
}
