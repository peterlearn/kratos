package generator

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pll/kratos/tool/protobuf/pkg/generator"
	"github.com/pll/kratos/tool/protobuf/pkg/naming"
	"github.com/pll/kratos/tool/protobuf/pkg/tag"
	"github.com/pll/kratos/tool/protobuf/pkg/typemap"
)

type ws struct {
	generator.Base
	filesHandled int
}

// wsGenerator ws generator.
func WSGenerator() *ws {
	t := &ws{}
	return t
}

// Generate ...
func (t *ws) Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
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

func (t *ws) generateForFile(file *descriptor.FileDescriptorProto) *plugin.CodeGeneratorResponse_File {
	resp := new(plugin.CodeGeneratorResponse_File)

	t.generateFileHeader(file, t.GenPkgName)
	t.generateImports(file)
	count := 0
	for i, service := range file.Service {
		count += t.generateWSInterface(file, service)
		t.generateService(file, service, i)
	}

	resp.Name = proto.String(naming.GenFileName(file, ".ws.go"))
	resp.Content = proto.String(t.FormattedOutput())
	t.Output.Reset()

	t.filesHandled++
	return resp
}

func (t *ws) generateFileHeader(file *descriptor.FileDescriptorProto, pkgName string) {
	t.P("// Code generated by protoc-gen-ws ", generator.Version, ", DO NOT EDIT.")
	t.P("// source: ", file.GetName())
	t.P()
	if t.filesHandled == 0 {
		comment, err := t.Reg.FileComments(file)
		if err == nil && comment.Leading != "" {
			// doc for the first file
			t.P("/*")
			t.P("Package ", t.GenPkgName, " is a generated websocket stub package.")
			t.P("This code was generated with kratos/tool/protobuf/protoc-gen-ws ", generator.Version, ".")
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

func (t *ws) generateImports(file *descriptor.FileDescriptorProto) {
	//if len(file.Service) == 0 {
	//	return
	//}
	t.P(`import (`)
	//t.P(`	`,t.pkgs["context"], ` "context"`)
	t.P(`	"context"`)
	t.P(`	"fmt"`)
	t.P()
	t.P(`	"git.huoys.com/middle-business/gomsg/pkg"`)
	t.P(`	"git.huoys.com/middle-business/gomsg/pkg/ws/server"`)
	t.P(`	"github.com/gogo/protobuf/proto"`)
	t.P(`	"github.com/pll/kratos/pkg/ecode"`)

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
	t.P(`// to suppressed 'imported but not used warning'`)
	t.P(`var _ context.Context`)

}

func (t *ws) generateWSInterface(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) int {
	count := 0
	servName := naming.ServiceName(service)
	t.P("// " + servName + "WsServer is the server API for " + servName + " service.")

	comments, err := t.Reg.ServiceComments(file, service)
	if err == nil {
		t.PrintComments(comments)
	}
	t.P(`type `, servName, `WSServer interface {`)
	t.P("\tStartNotify(ws *server.Server)")
	for _, method := range service.Method {
		if !t.ShouldGenForMethod(file, service, method) {
			continue
		}
		count++
		t.generateInterfaceMethod(file, service, method, comments)
		t.P()
	}
	t.P(`}`)
	return count
}

func (t *ws) generateInterfaceMethod(file *descriptor.FileDescriptorProto,
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

func (t *ws) generateService(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, index int) {
	servName := naming.ServiceName(service)
	servAlias := servName + "Server"

	t.P("var srv ", servAlias)
	t.P("type shandler struct {")
	t.P("closeChan chan string")
	t.P("}")
	t.P()
	t.P("func (h *shandler) OnOpen(p pkg.Session) {}")
	t.P("func (h *shandler) OnClose(p pkg.Session, b bool) {")
	t.P("h.closeChan <- fmt.Sprint(p.ID())")
	t.P("}")
	t.P()
	t.P("func (h *shandler) OnReq(pk pkg.Session, data []byte) pkg.IRet {")
	t.P("var (\n\t\tpb  Message\n\t\tctx = context.WithValue(context.Background(), \"id\", pk.ID())\n\t)\n\terr := proto.Unmarshal(data, &pb)\n\tif err != nil {\n\t\treturn nil\n\t}\n\tswitch pb.Ops {")
	for _, method := range service.Method {
		t.generateWsHandler(method)
	}
	t.P("default:\n\t\treturn pkg.Error(int16(pkg.NoHandler), \"not implemented\")\n\t}")
	t.P("}")
	t.P()
	t.P("func (h *shandler) OnPush(pk pkg.Session, data []byte) pkg.IRet {")
	t.P("return pkg.Ok(nil)")
	t.P("}")
	t.P()
	t.P("func Register", servName, "WsServer(e *server.Server, addr string, closeChan chan string, s ", servAlias, ") error {")
	t.P("srv = s")
	t.P("return e.ListenAndServe(addr, &shandler{closeChan: closeChan})")
	t.P("}")
	t.P()
}

func (t *ws) generateWsHandler(method *descriptor.MethodDescriptorProto) {
	methName := naming.MethodName(method)
	reqArg := t.GoTypeName(method.GetInputType())
	t.P("case int32(GameCommand_", methName, "):")
	t.P("var req = &", reqArg, "{}")
	t.P("if err = proto.Unmarshal(pb.Data, req); err != nil {\n\t\t\treturn pkg.Error(int16(pkg.ReadErrorNo), err.Error())\n\t\t}")
	t.P("resp, err := srv.", methName, "(ctx, req)")
	t.P("if err != nil {return pkg.Error(int16(ecode.Cause(err).Code()), err.Error())}")
	t.P("res, _ := proto.Marshal(resp)")
	t.P("return pkg.Ok(res)")
}
