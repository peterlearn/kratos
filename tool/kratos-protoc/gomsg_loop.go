package main

import (
	"os/exec"
)

const (
	_getGomsgLoopGen = "go get -u github.com/pll/kratos/tool/protobuf/protoc-gen-gomsg-loop"
	_gomsgLoopProtoc = "protoc --proto_path=%s --proto_path=%s --proto_path=%s --proto_path=%s --gomsg_out=:."
)

func genGoMsgLoop(files []string) error {
	if _, err := exec.LookPath("protoc-gen-gomsg-loop"); err != nil {
		if err := goget(_getGomsgGen); err != nil {
			return err
		}
	}
	return generate(_gomsgProtoc, files)
}
