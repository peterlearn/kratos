package main

import (
	"os/exec"
)

const (
	_getWSGen = "go get -u github.com/peterlearn/kratos/v1/tool/protobuf/protoc-gen-ws"
	_WSProtoc = "protoc --proto_path=%s --proto_path=%s --proto_path=%s --proto_path=%s --ws_out=:."
)

func genWS(files []string) error {
	if _, err := exec.LookPath("protoc-gen-ws"); err != nil {
		if err := goget(_getWSGen); err != nil {
			return err
		}
	}
	return generate(_WSProtoc, files)
}
