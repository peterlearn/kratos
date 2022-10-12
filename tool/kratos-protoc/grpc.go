package main

import (
	"os/exec"
)

const (
	_getGRPCGen = "go get -u github.com/gogo/protobuf/protoc-gen-gogofast"
	_grpcProtoc = `protoc --proto_path=%s --proto_path=%s --proto_path=%s --proto_path=%s --gogofast_out=plugins=grpc,Mprotobuf/google/protobuf/any.proto=github.com/gogo/protobuf/types,Mprotobuf/google/protobuf/api.proto=github.com/gogo/protobuf/types,Mprotobuf/google/protobuf/descriptor.proto=github.com/gogo/protobuf/types,Mprotobuf/google/protobuf/duration.proto=github.com/gogo/protobuf/types,Mprotobuf/google/protobuf/empty.proto=github.com/gogo/protobuf/types,Mprotobuf/google/protobuf/field_mask.proto=github.com/gogo/protobuf/types,Mprotobuf/google/protobuf/source_context.proto=github.com/gogo/protobuf/types,Mprotobuf/google/protobuf/struct.proto=github.com/gogo/protobuf/types,Mprotobuf/google/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mprotobuf/google/protobuf/type.proto=github.com/gogo/protobuf/types,Mprotobuf/google/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api:.`
)

func installGRPCGen() error {
	if _, err := exec.LookPath("protoc-gen-gofast"); err != nil {
		if err := goget(_getGRPCGen); err != nil {
			return err
		}
	}
	return nil
}

func genGRPC(files []string) error {
	return generate(_grpcProtoc, files)
}
