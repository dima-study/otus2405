package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

const str = "Hello, OTUS!"

func main() {
	fmt.Print(reverse.String(str))
}
