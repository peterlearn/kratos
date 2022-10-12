package main

import "os/exec"

const (
	_getTcpGen = "go get -u github.com/peterlearn/kratos/v1/tool/protobuf/protoc-gen-tcp"
	_tcpProtoc = "protoc --proto_path=%s --proto_path=%s --proto_path=%s --proto_path=%s --tcp_out=:."
)

func genTcp(files []string) error {
	if _, err := exec.LookPath("protoc-gen-tcp"); err != nil {
		if err := goget(_getTcpGen); err != nil {
			return err
		}
	}
	return generate(_tcpProtoc, files)
}
