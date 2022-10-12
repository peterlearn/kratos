package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/peterlearn/kratos/v1/tool/protobuf/pkg/generator"
	"github.com/peterlearn/kratos/v1/tool/protobuf/pkg/naming"
	"github.com/peterlearn/kratos/v1/tool/protobuf/pkg/tag"
	"github.com/peterlearn/kratos/v1/tool/protobuf/pkg/typemap"
)

type comet struct {
	generator.Base
	filesHandled int
}

// CometGenerator Comet generator
func CometGenerator() *comet {
	t := &comet{}
	return t
}

// Generate ...
func (t *comet) Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	t.Setup(in)

	// Showtime! Generate the response.
	resp := new(plugin.CodeGeneratorResponse)
	for _, f := range t.GenFiles {
		respFile := t.generateForFile(f)
		if respFile != nil {
			resp.File = append(resp.File, respFile)
		}
	}
	return resp
}

func (t *comet) generateForFile(file *descriptor.FileDescriptorProto) *plugin.CodeGeneratorResponse_File {
	resp := new(plugin.CodeGeneratorResponse_File)
	t.generateFileHeader(file, t.GenPkgName)
	t.generateImports(file)
	count := 0
	for i, service := range file.Service {
		count += t.generateCometInterface(file, service)
		t.generateCometService(file, service, i)
	}
	resp.Name = proto.String(naming.GenFileName(file, ".comet.go"))
	resp.Content = proto.String(t.FormattedOutput())
	t.Output.Reset()

	t.filesHandled++
	return resp
}

func (t *comet) generateFileHeader(file *descriptor.FileDescriptorProto, pkgName string) {
	t.P("// Code generated by protoc-gen-comet ", generator.Version, ", DO NOT EDIT.")
	t.P("// source: ", file.GetName())
	t.P()
	if t.filesHandled == 0 {
		comment, err := t.Reg.FileComments(file)
		if err == nil && comment.Leading != "" {
			// doc for the first file
			t.P("/*")
			t.P("Package ", t.GenPkgName, " is a generated comet stub package.")
			t.P("This code was generated with kratos/tool/protobuf/protoc-gen-comet ", generator.Version, ".")
			t.P()
			for _, line := range strings.Split(comment.Leading, "\n") {
				line = strings.TrimPrefix(line, " ")
				// ensure we don't escape from the block comment
				line = strings.Replace(line, "*/", "* /", -1)
				t.P(line)
			}
			t.P()
			t.P("It is generated from these files:")
			for _, f := range t.GenFiles {
				t.P("\t", f.GetName())
			}
			t.P("*/")
		}
	}
	t.P(`package `, pkgName)
	t.P()
}

func (t *comet) generateImports(file *descriptor.FileDescriptorProto) {
	//if len(file.Service) == 0 {
	//	return
	//}
	t.P(`import (`)
	//t.P(`	`,t.pkgs["context"], ` "context"`)
	t.P(`	"context"`)
	t.P()
	t.P(`	"git.huoys.com/middle-end/library/pkg/net/comet"`)
	t.P(`	"github.com/gogo/protobuf/proto"`)

	t.P(`)`)
	// It's legal to import a message and use it as an input or output for a
	// method. Make sure to import the package of any such message. First, dedupe
	// them.
	deps := make(map[string]string) // Map of package name to quoted import path.
	deps = t.DeduceDeps(file)
	for pkg, importPath := range deps {
		t.P(`import `, pkg, ` `, importPath)
	}
	t.P()
}

func (t *comet) generateCometInterface(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) int {
	count := 0
	servName := naming.ServiceName(service)
	t.P("// " + servName + "CometServer is the server API for " + servName + " service.")

	comments, err := t.Reg.ServiceComments(file, service)
	if err == nil {
		t.PrintComments(comments)
	}
	t.P(`type `, servName, `CometServer interface {`)
	t.P(`SetCometChan(cl *comet.ChanList, cs *comet.Server)`)
	for _, method := range service.Method {
		if !t.ShouldGenForMethod(file, service, method) {
			continue
		}
		count++
		t.generateInterfaceMethod(file, service, method, comments)
		t.P()
	}
	t.P(`}`)
	t.P()
	return count
}

func (t *comet) generateInterfaceMethod(file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	method *descriptor.MethodDescriptorProto,
	comments typemap.DefinitionComments) {
	comments, err := t.Reg.MethodComments(file, service, method)

	methName := naming.MethodName(method)
	outputType := t.GoTypeName(method.GetOutputType())
	inputType := t.GoTypeName(method.GetInputType())
	tags := tag.GetTagsInComment(comments.Leading)
	if tag.GetTagValue("dynamic", tags) == "true" {
		return
	}

	if err == nil {
		t.PrintComments(comments)
	}

	respDynamic := tag.GetTagValue("dynamic_resp", tags) == "true"
	if respDynamic {
		t.P(fmt.Sprintf(`	%s(ctx context.Context, req *%s) (resp interface{}, err error)`,
			methName, inputType))
	} else {
		t.P(fmt.Sprintf(`	%s(ctx context.Context, req *%s) (resp *%s, err error)`,
			methName, inputType, outputType))
	}
}

func (t *comet) generateCometService(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	index int) {
	servName := naming.ServiceName(service)
	origServName := service.GetName()
	fullServName := origServName
	if pkg := file.GetPackage(); pkg != "" {
		fullServName = pkg + "." + fullServName
	}
	serviceDescVar := "_Comet_" + servName + "_serviceDesc"

	// Server registration.
	var cometFuncName = fmt.Sprintf("Register%sCometServer", servName)
	t.P(`func `, cometFuncName, `(s *comet.Server, srv `, servName, `CometServer) {`)
	t.P(`chanList := s.RegisterService(&`, serviceDescVar, `, srv)`)
	t.P(`srv.SetCometChan(chanList, s)`)
	t.P("}")
	t.P()

	// Server handler implementations.
	var handlerNames []string
	for _, method := range service.Method {
		hname := t.generateServerMethod(servName, fullServName, method)
		handlerNames = append(handlerNames, hname)
	}

	// get GameCommand
	gameCommand := make(map[string]string)
	for _, enumType := range file.EnumType {
		if enumType.GetName() != "GameCommand" {
			break
		}
		for _, values := range enumType.Value {
			gameCommand[values.GetName()] = fmt.Sprintf("%d", values.GetNumber())
		}
	}
	//Service descriptor.
	t.P(`var `, serviceDescVar, ` = comet.ServiceDesc{`)
	t.P(`	ServiceName: `, strconv.Quote(fullServName), `,`)
	t.P(`	HandlerType: (*`, servName, `CometServer)(nil),`)
	t.P(`	Methods: []comet.MethodDesc{`)
	for i, method := range service.Method {
		t.P(`		{`)
		t.P(`			MethodName: `, strconv.Quote(naming.MethodName(method)), `,`)
		t.P(`			Handler: `, handlerNames[i], `,`)
		t.P(`			Ops: `, gameCommand[naming.MethodName(method)], `,`)
		t.P(`		},`)
	}
	t.P(`	},`)
	t.P(`}`)
}

func (t *comet) generateServerMethod(servName, fullServName string, method *descriptor.MethodDescriptorProto) string {
	methName := naming.MethodName(method)
	hname := fmt.Sprintf("_Comet_%s_%s_Handler", servName, methName)
	inputType := t.GoTypeName(method.GetInputType())
	t.P(`func `, hname, `(srv interface{}, ctx context.Context, data []byte, interceptor comet.UnaryServerInterceptor) ([]byte, error) {`)
	t.P(`	in := new(`, inputType, `)`)
	t.P(`	err := proto.Unmarshal(data, in)`)
	t.P(`	if err != nil {`)
	t.P(`		return nil, err`)
	t.P(`	}`)
	t.P(`	if interceptor == nil {`)
	t.P(`		out, err := srv.(`, servName, `CometServer).`, methName, `(ctx, in)`)
	t.P(`		data, _ := proto.Marshal(out)`)
	t.P(`			return data, err`)
	t.P(`	}`)
	t.P(`	info := &comet.UnaryServerInfo{`)
	t.P(`		Server:     srv,`)
	t.P(`		FullMethod: `, strconv.Quote(fmt.Sprintf("/%s/%s", fullServName, methName)), `,`)
	t.P(`	}`)
	t.P(`	handler := func(ctx context.Context, req interface{}) ([]byte, error) {`)
	t.P(`		out, err := srv.(`, servName, `CometServer).`, methName, `(ctx, req.(*`, inputType, `))`)
	t.P(`		if out != nil {`)
	t.P(`			data, _ := proto.Marshal(out)`)
	t.P(`			return data, err`)
	t.P(`		}`)
	t.P(`		return nil, err`)
	t.P(`	}`)
	t.P(`	return interceptor(ctx, in, info, handler)`)
	t.P(`}`)
	return hname
}
