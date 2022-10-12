package main

import (
	"os/exec"
)

const (
	_getCometGen = "go get -u github.com/peterlearn/kratos/tool/protobuf/protoc-gen-comet"
	_cometProtoc = "protoc --proto_path=%s --proto_path=%s --proto_path=%s --proto_path=%s --comet_out=:."
)

func genComet(files []string) error {
	if _, err := exec.LookPath("protoc-gen-comet"); err != nil {
		if err := goget(_getCometGen); err != nil {
			return err
		}
	}
	return generate(_cometProtoc, files)
}
