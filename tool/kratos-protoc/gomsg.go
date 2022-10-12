package main

import (
	"os/exec"
)

const (
	_getGomsgGen = "go get -u github.com/pll/kratos/tool/protobuf/protoc-gen-gomsg"
	_gomsgProtoc = "protoc --proto_path=%s --proto_path=%s --proto_path=%s --proto_path=%s --gomsg_out=:."
)

func genGoMsg(files []string) error {
	if _, err := exec.LookPath("protoc-gen-gomsg"); err != nil {
		if err := goget(_getGomsgGen); err != nil {
			return err
		}
	}
	return generate(_gomsgProtoc, files)
}
