package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/peterlearn/kratos/v1/tool/protobuf/pkg/generator"
	"github.com/peterlearn/kratos/v1/tool/protobuf/pkg/naming"
	"github.com/peterlearn/kratos/v1/tool/protobuf/pkg/tag"
	"github.com/peterlearn/kratos/v1/tool/protobuf/pkg/typemap"
	"google.golang.org/protobuf/proto"
)

type gomsg struct {
	generator.Base
	filesHandled int
}

// GoMsgGenerator GoMsg generator
func GoMsgGenerator() *gomsg {
	t := &gomsg{}
	return t
}

// Generate ...
func (t *gomsg) Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
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

func (t *gomsg) generateForFile(file *descriptor.FileDescriptorProto) *plugin.CodeGeneratorResponse_File {
	resp := new(plugin.CodeGeneratorResponse_File)
	t.generateFileHeader(file, t.GenPkgName)
	t.generateImports(file)
	count := 0
	for i, service := range file.Service {
		count += t.generateGoMsgInterface(file, service)
		t.generateGoMsgService(file, service, i)
	}
	resp.Name = proto.String(naming.GenFileName(file, ".gomsg.go"))
	resp.Content = proto.String(t.FormattedOutput())
	t.Output.Reset()

	t.filesHandled++
	return resp
}

func (t *gomsg) generateFileHeader(file *descriptor.FileDescriptorProto, pkgName string) {
	t.P("// Code generated by protoc-gen-gomsg ", generator.Version, ", DO NOT EDIT.")
	t.P("// source: ", file.GetName())
	t.P()
	if t.filesHandled == 0 {
		comment, err := t.Reg.FileComments(file)
		if err == nil && comment.Leading != "" {
			// doc for the first file
			t.P("/*")
			t.P("Package ", t.GenPkgName, " is a generated gomsg stub package.")
			t.P("This code was generated with kratos/tool/protobuf/protoc-gen-gomsg ", generator.Version, ".")
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

func (t *gomsg) generateImports(file *descriptor.FileDescriptorProto) {
	//if len(file.Service) == 0 {
	//	return
	//}
	t.P(`import (`)
	//t.P(`	`,t.pkgs["context"], ` "context"`)
	t.P(`	"context"`)
	t.P(`	"fmt"`)
	t.P()
	t.P(`	gomsg "git.huoys.com/middle-business/gomsg/pkg/ws/server"`)
	t.P(`	"git.huoys.com/middle-business/gomsg/pkg"`)
	t.P(`	"github.com/gogo/protobuf/proto"`)
	t.P(`	pb "git.huoys.com/middle-business/gomsg/pkg/ws/proto"`)
	t.P(`	"git.huoys.com/middle-business/gomsg/pkg/util"`)

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

func (t *gomsg) generateGoMsgInterface(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) int {
	count := 0
	servName := naming.ServiceName(service)
	t.P("// " + servName + "GoMsgServer is the server API for " + servName + " service.")

	comments, err := t.Reg.ServiceComments(file, service)
	if err == nil {
		t.PrintComments(comments)
	}
	t.P(`type `, servName, `GoMsgServer interface {`)
	t.P(`	//连接打开通知`)
	t.P(`	OnSessionOpen(ctx context.Context)`)
	t.P(`	//连接关闭通知`)
	t.P(`	OnSessionClose(ctx context.Context)`)
	t.P(`	//获取桌子的Loop队列`)
	t.P(`	GetTableLoop(sid int32) util.ILoop`)
	t.P()
	for _, method := range service.Method {
		if !t.ShouldGenForMethod(file, service, method) {
			continue
		}
		count++
		t.generateInterfaceMethod(file, service, method, comments)
	}
	t.P(`}`)
	t.P()
	return count
}

func (t *gomsg) generateInterfaceMethod(file *descriptor.FileDescriptorProto,
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

func (t *gomsg) generateGoMsgService(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	index int) {
	servName := naming.ServiceName(service)
	origServName := service.GetName()
	fullServName := origServName
	if pkg := file.GetPackage(); pkg != "" {
		fullServName = pkg + "." + fullServName
	}
	serviceDescVar := "_GoMsg_" + servName + "_serviceDesc"

	// Server registration.
	var gomsgFuncName = fmt.Sprintf("Register%sGoMsgServer", servName)
	t.P(`func `, gomsgFuncName, `(s *gomsg.Server, srv `, servName, `GoMsgServer) gomsg.IHandlerProxy {`)
	t.P(`	return s.RegisterService(&`, serviceDescVar, `, srv)`)
	t.P("}")
	t.P()

	// Server handler implementations.
	var handlerNames []string
	for _, method := range service.Method {
		comments, _ := t.Reg.MethodComments(file, service, method)
		tags := tag.GetTagsInComment(comments.Leading)
		hname := t.generateServerMethod(servName, fullServName, method, tag.GetTagValue("room", tags) == "true")
		handlerNames = append(handlerNames, hname)
	}

	openHandlerName := t.generateServerMustMethod(servName, fullServName, "OnSessionOpen")
	closeHandlerName := t.generateServerMustMethod(servName, fullServName, "OnSessionClose")

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
	t.P(`var `, serviceDescVar, ` = gomsg.ServiceDesc{`)
	t.P(`	ServiceName: `, strconv.Quote(fullServName), `,`)
	t.P(`	HandlerType: (*`, servName, `GoMsgServer)(nil),`)
	t.P(`	Methods: []gomsg.MethodDesc{`)
	for i, method := range service.Method {
		t.P(`		{`)
		t.P(`			MethodName: `, strconv.Quote(naming.MethodName(method)), `,`)
		t.P(`			Handler: `, handlerNames[i], `,`)
		t.P(`			Ops: `, gameCommand[naming.MethodName(method)], `,`)
		t.P(`		},`)
	}
	t.P(`	},`)
	t.P(`	OpenHandler: `, openHandlerName, `,`)
	t.P(`	CloseHandler: `, closeHandlerName, `,`)
	t.P(`}`)
}

func (t *gomsg) generateServerMethod(servName, fullServName string, method *descriptor.MethodDescriptorProto, room bool) string {
	methName := naming.MethodName(method)
	hname := fmt.Sprintf("_GoMsg_%s_%s_Handler", servName, methName)
	inputType := t.GoTypeName(method.GetInputType())
	t.P(`func `, hname, `(srv interface{}, session pkg.Session, msg *pb.Message, interceptor gomsg.UnaryServerInterceptor) ([]byte, error) {`)
	t.P(`	server, ok := srv.(`, servName, `GoMsgServer)`)
	t.P(`	if !ok {`)
	t.P(`		return nil, fmt.Errorf("srv not implement `, servName, `GoMsgServer")`)
	t.P(`	}`)
	t.P(`	ctx := context.WithValue(context.Background(), gomsg.CTXSessionKey, session)`)
	t.P(`	in := new(`, inputType, `)`)
	t.P(`	err := proto.Unmarshal(msg.Data, in)`)
	t.P(`	if err != nil {`)
	t.P(`		return nil, err`)
	t.P(`	}`)
	t.P(`	doReqFunc := func() ([]byte, error) {`)
	t.P(`		out, err := server.`, methName, `(ctx, in)`)
	t.P(`		if out != nil && err == nil {`)
	t.P(`			return proto.Marshal(out)`)
	t.P(`		}`)
	t.P(`		return nil, err`)
	t.P(`	}`)
	if room {
		t.P(`	if interceptor == nil {`)
		t.P(`		return doReqFunc()`)
		t.P(`	}`)
		t.P(`	info := &gomsg.UnaryServerInfo{`)
		t.P(`		Server:     srv,`)
		t.P(`		FullMethod: `, strconv.Quote(fmt.Sprintf("/%s/%s", fullServName, methName)), `,`)
		t.P(`	}`)
		t.P(`	handler := func(ctx context.Context, req interface{}) ([]byte, error) {`)
		t.P(`		in = req.(*`, inputType, `)`)
		t.P(`		return doReqFunc()`)
		t.P(`	}`)
		t.P(`	return interceptor(ctx, in, info, handler)`)
		t.P(`}`)
	} else {
		t.P(`	loop := server.GetTableLoop(session.ID())`)
		t.P(`	if interceptor == nil {`)
		t.P(`		if loop == nil {`)
		t.P(`			return doReqFunc()`)
		t.P(`		}`)
		t.P(`		return loop.PostAndWait(doReqFunc)`)
		t.P(`	}`)
		t.P(`	info := &gomsg.UnaryServerInfo{`)
		t.P(`		Server:     srv,`)
		t.P(`		FullMethod: `, strconv.Quote(fmt.Sprintf("/%s/%s", fullServName, methName)), `,`)
		t.P(`	}`)
		t.P(`	handler := func(ctx context.Context, req interface{}) ([]byte, error) {`)
		t.P(`		in = req.(*`, inputType, `)`)
		t.P(`		if loop == nil {`)
		t.P(`			return doReqFunc()`)
		t.P(`		}`)
		t.P(`		return loop.PostAndWait(doReqFunc)`)
		t.P(`	}`)
		t.P(`	return interceptor(ctx, in, info, handler)`)
		t.P(`}`)
	}

	return hname
}

func (t *gomsg) generateServerMustMethod(servName, fullServName string, methName string) string {
	hname := fmt.Sprintf("_GoMsg_%s_%s_Handler", servName, methName)
	t.P(`func `, hname, `(srv interface{}, session pkg.Session, msg *pb.Message, interceptor gomsg.UnaryServerInterceptor) ([]byte, error) {`)
	t.P(`	server, ok := srv.(`, servName, `GoMsgServer)`)
	t.P(`	if !ok {`)
	t.P(`		return nil, fmt.Errorf("srv not implement `, servName, `GoMsgServer")`)
	t.P(`	}`)
	t.P(`	ctx := context.WithValue(context.Background(), gomsg.CTXSessionKey, session)`)
	t.P(`	if interceptor == nil {`)
	t.P(`		server.`, methName, `(ctx)`)
	t.P(`		return nil, nil`)
	t.P(`	}`)
	t.P(`	info := &gomsg.UnaryServerInfo{`)
	t.P(`		Server:     srv,`)
	t.P(`		FullMethod: `, strconv.Quote(fmt.Sprintf("/%s/%s", fullServName, methName)), `,`)
	t.P(`	}`)
	t.P(`	handler := func(ctx context.Context, req interface{}) ([]byte, error) {`)
	t.P(`		server.`, methName, `(ctx)`)
	t.P(`		return nil, nil`)
	t.P(`	}`)
	t.P(`	return interceptor(ctx, nil, info, handler)`)
	t.P(`}`)
	return hname
}
