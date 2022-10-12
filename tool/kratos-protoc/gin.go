package main

import (
	"os/exec"
)

const (
	_getGinGen = "go get -u github.com/pll/kratos/tool/protobuf/protoc-gen-gin"
	_GinProtoc = "protoc --proto_path=%s --proto_path=%s --proto_path=%s --proto_path=%s --gin_out=:."
)

func installGinGen() error {
	if _, err := exec.LookPath("protoc-gen-gin"); err != nil {
		if err := goget(_getGinGen); err != nil {
			return err
		}
	}
	return nil
}

func genGin(files []string) error {
	return generate(_GinProtoc, files)
}
