package main

import (
	"os/exec"
)

const (
	_getTcpLoopGen = "go get -u github.com/peterlearn/kratos/v1/tool/protobuf/protoc-gen-tcp-loop"
	_tcpLoopProtoc = "protoc --proto_path=%s --proto_path=%s --proto_path=%s --proto_path=%s --tcp-loop_out=:."
)

func genTcpLoop(files []string) error {
	if _, err := exec.LookPath("protoc-gen-tcp-loop"); err != nil {
		if err := goget(_getTcpLoopGen); err != nil {
			return err
		}
	}
	return generate(_tcpLoopProtoc, files)
}
