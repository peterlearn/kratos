package naming

import (
	"path"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// GoFileName returns the output name for the generated Go file.
func GoFileName(f *descriptor.FileDescriptorProto, suffix string) string {
	name := *f.Name
	if ext := path.Ext(name); ext == ".pb" || ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}
	name += suffix

	// Does the file have a "go_package" option? If it does, it may override the
	// filename.
	impPath, pkg, ok := goPackageOption(f)
	if ok && impPath != "" {
		// Replace the existing dirname with the declared import path.
		_, name = path.Split(name)
		name = path.Join(impPath, name)
		return name
	}
	if ok && pkg != "" {
		protopkg := *f.Package
		if protopkg == "google.protobuf" && pkg == "types" {
			impPath = "github.com/gogo/protobuf/types"
			_, name = path.Split(name)
			name = path.Join(impPath, name)
			return name
		}
	}

	return name
}
