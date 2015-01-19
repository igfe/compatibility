package main

import (
	"fmt"
	"github.com/gogo/protobuf/parser"
	// descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	newer, err1 := parser.ParseFile("./p.proto", ".")
	check(err1)
	fds := newer.File
	fmt.Println(fds)
}
