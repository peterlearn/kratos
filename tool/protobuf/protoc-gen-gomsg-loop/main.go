package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/peterlearn/kratos/tool/protobuf/pkg/gen"
	"github.com/peterlearn/kratos/tool/protobuf/pkg/generator"
	gomsggen "github.com/peterlearn/kratos/tool/protobuf/protoc-gen-gomsg/generator"
)

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Println(generator.Version)
		os.Exit(0)
	}

	g := gomsggen.GoMsgGenerator()
	gen.Main(g)
}
