package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/peterlearn/kratos/v1/tool/protobuf/pkg/gen"
	"github.com/peterlearn/kratos/v1/tool/protobuf/pkg/generator"
	tcpgen "github.com/peterlearn/kratos/v1/tool/protobuf/protoc-gen-tcp/generator"
)

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Println(generator.Version)
		os.Exit(0)
	}

	g := tcpgen.TcpGenerator()
	gen.Main(g)
}
