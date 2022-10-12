package generator

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pll/kratos/tool/protobuf/pkg/generator"
	"github.com/pll/kratos/tool/protobuf/pkg/naming"
	"github.com/pll/kratos/tool/protobuf/pkg/tag"
	"github.com/pll/kratos/tool/protobuf/pkg/typemap"
	"github.com/pll/kratos/tool/protobuf/pkg/utils"
)

type kgin struct {
	generator.Base
	filesHandled int
}

// KginGenerator Kgin generator.
func KginGenerator() *kgin {
	t := &kgin{}
	return t
}

// Generate ...
func (t *kgin) Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
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

func (t *kgin) generateForFile(file *descriptor.FileDescriptorProto) *plugin.CodeGeneratorResponse_File {
	resp := new(plugin.CodeGeneratorResponse_File)

	t.generateFileHeader(file, t.GenPkgName)
	t.generateImports(file)
	t.generatePathConstants(file)
	count := 0
	for i, service := range file.Service {
		count += t.generateGinInterface(file, service)
		t.generateGinRoute(file, service, i)
	}

	resp.Name = proto.String(naming.GenFileName(file, ".gin.go"))
	resp.Content = proto.String(t.FormattedOutput())
	t.Output.Reset()

	t.filesHandled++
	return resp
}

func (t *kgin) generatePathConstants(file *descriptor.FileDescriptorProto) {
	t.P()
	for _, service := range file.Service {
		name := naming.ServiceName(service)
		for _, method := range service.Method {
			if !t.ShouldGenForMethod(file, service, method) {
				continue
			}
			apiInfo := t.GetHttpInfoCached(file, service, method)
			t.P(`var Path`, name, naming.MethodName(method), ` = "`, apiInfo.Path, `"`)
		}
		t.P()
	}
}

func (t *kgin) generateFileHeader(file *descriptor.FileDescriptorProto, pkgName string) {
	t.P("// Code generated by protoc-gen-gin ", generator.Version, ", DO NOT EDIT.")
	t.P("// source: ", file.GetName())
	t.P()
	if t.filesHandled == 0 {
		comment, err := t.Reg.FileComments(file)
		if err == nil && comment.Leading != "" {
			// doc for the first file
			t.P("/*")
			t.P("Package ", t.GenPkgName, " is a generated gin stub package.")
			t.P("This code was generated with kratos/tool/protobuf/protoc-gen-gin ", generator.Version, ".")
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

func (t *kgin) generateImports(file *descriptor.FileDescriptorProto) {
	//if len(file.Service) == 0 {
	//	return
	//}
	t.P(`import (`)
	//t.P(`	`,t.pkgs["context"], ` "context"`)
	t.P(`	"context"`)
	t.P()
	t.P(`	kgin "github.com/pll/kratos/pkg/net/http/gin"`)
	t.P(`	"github.com/gin-gonic/gin"`)
	t.P(`	"github.com/gogo/protobuf/proto"`)
	t.P(`	"net/http"`)

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
	t.P(`var _ *gin.Context`)
	t.P(`var _ context.Context`)
	t.P(`var _ proto.Message`)

}

// Big header comments to makes it easier to visually parse a generated file.
func (t *kgin) sectionComment(sectionTitle string) {
	t.P()
	t.P(`// `, strings.Repeat("=", len(sectionTitle)))
	t.P(`// `, sectionTitle)
	t.P(`// `, strings.Repeat("=", len(sectionTitle)))
	t.P()
}

func (t *kgin) generateGinRoute(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	index int) {
	// old mode is generate xx.route.go in the http pkg
	// new mode is generate route code in the same .gin.go
	// route rule /x{department}/{project-name}/{path_prefix}/method_name
	// generate each route method
	servName := naming.ServiceName(service)
	versionPrefix := naming.GetVersionPrefix(t.GenPkgName)
	svcName := utils.LcFirst(utils.CamelCase(versionPrefix)) + servName + "Svc"
	t.P(`var `, svcName, ` `, servName, `GinServer`)

	type methodInfo struct {
		midwares      []string
		routeFuncName string
		apiInfo       *generator.HTTPInfo
		methodName    string
	}
	var methList []methodInfo
	var allMidwareMap = make(map[string]bool)
	var isLegacyPkg = false
	for _, method := range service.Method {
		if !t.ShouldGenForMethod(file, service, method) {
			continue
		}
		var midwares []string
		comments, _ := t.Reg.MethodComments(file, service, method)
		tags := tag.GetTagsInComment(comments.Leading)
		if tag.GetTagValue("dynamic", tags) == "true" {
			continue
		}
		apiInfo := t.GetHttpInfoCached(file, service, method)
		isLegacyPkg = apiInfo.IsLegacyPath
		//httpMethod, legacyPath, path := getHttpInfo(file, service, method, t.reg)
		//if legacyPath != "" {
		//	isLegacyPkg = true
		//}

		midStr := tag.GetTagValue("midware", tags)
		if midStr != "" {
			midwares = strings.Split(midStr, ",")
			for _, m := range midwares {
				allMidwareMap[m] = true
			}
		}

		methName := naming.MethodName(method)
		inputType := t.GoTypeName(method.GetInputType())

		routeName := utils.LcFirst(utils.CamelCase(servName) +
			utils.CamelCase(methName))

		methList = append(methList, methodInfo{
			apiInfo:       apiInfo,
			midwares:      midwares,
			routeFuncName: routeName,
			methodName:    method.GetName(),
		})

		if apiInfo.ContextType == "X-PROTOBUF" {
			t.P(fmt.Sprintf("func %s (c *gin.Context) {", routeName))
			t.P(`	p := new(`, inputType, `)`)
			t.P(`	if err := c.Bind(p); err != nil {`)
			t.P(`		return`)
			t.P(`	}`)
			t.P(`	var buf []byte`)
			t.P(`	resp, err := `, svcName, `.`, methName, `(c, p)`)
			t.P(`	if err == nil && resp != nil {`)
			t.P(`   	var marshalErr error`)
			t.P(`		buf, marshalErr = proto.Marshal(resp)`)
			t.P(`		if marshalErr != nil {`)
			t.P(`			c.ProtoBuf(http.StatusInternalServerError, nil)`)
			t.P(`			return`)
			t.P(`		}`)
			t.P(`	}`)
			t.P(`	c.ProtoBuf(http.StatusOK, kgin.TOPROTO(buf, err))`)
			t.P(`}`)
			t.P(``)
		} else {
			t.P(fmt.Sprintf("func %s (c *gin.Context) {", routeName))
			t.P(`	p := new(`, inputType, `)`)
			t.P(`	if err := c.Bind(p); err != nil {`)
			t.P(`		return`)
			t.P(`	}`)
			t.P(`	resp, err := `, svcName, `.`, methName, `(c, p)`)
			//t.P(`	res := kgin.TOJSON(resp, err)`)
			//t.P(`	c.JSON(http.StatusOK, res)`)
			//t.P(`	c = context.WithValue(c, "Path", Path`, servName, methName, `)`)
			t.P(`	res, t := kgin.ToResponse(c, resp, err)`)
			t.P(`	if t == kgin.RespProtobuf {`)
			t.P(`		c.ProtoBuf(http.StatusOK, res)`)
			t.P(`	} else {`)
			t.P(`		c.JSON(http.StatusOK, res)`)
			t.P(`	}`)
			t.P(`}`)
			t.P(``)
		}
	}

	// generate route group
	var midList []string
	for m := range allMidwareMap {
		midList = append(midList, m+" gin.HandlerFunc")
	}

	sort.Strings(midList)

	// 注册老的路由的方法
	if isLegacyPkg {
		funcName := `Register` + utils.CamelCase(versionPrefix) + servName + `Service`
		t.P(`// `, funcName, ` Register the gin route with middleware map`)
		t.P(`// midMap is the middleware map, the key is defined in proto`)
		t.P(`func `, funcName, `(e *gin.Engine, svc `, servName, "GinServer, midMap map[string]gin.HandlerFunc)", ` {`)
		var keys []string
		for m := range allMidwareMap {
			keys = append(keys, m)
		}
		// to keep generated code consistent
		sort.Strings(keys)
		for _, m := range keys {
			t.P(m, ` := midMap["`, m, `"]`)
		}

		t.P(svcName, ` = svc`)
		for _, methInfo := range methList {
			var midArgStr string
			if len(methInfo.midwares) == 0 {
				midArgStr = ""
			} else {
				midArgStr = strings.Join(methInfo.midwares, ", ") + ", "
			}
			t.P(`e.`, methInfo.apiInfo.HttpMethod, `("`, methInfo.apiInfo.LegacyPath, `", `, midArgStr, methInfo.routeFuncName, `)`)
		}
		t.P(`	}`)
	} else {
		// 新的注册路由的方法
		var ginFuncName = fmt.Sprintf("Register%sGinServer", servName)
		t.P(`// `, ginFuncName, ` Register the gin route`)
		t.P(`func `, ginFuncName, `(e *gin.Engine, server `, servName, `GinServer) {`)
		t.P(svcName, ` = server`)
		for _, methInfo := range methList {
			t.P(`e.`, methInfo.apiInfo.HttpMethod, `("`, methInfo.apiInfo.NewPath, `",`, methInfo.routeFuncName, ` )`)
		}
		t.P(`	}`)
	}
}

func (t *kgin) hasHeaderTag(md *typemap.MessageDefinition) bool {
	if md.Descriptor.Field == nil {
		return false
	}
	for _, f := range md.Descriptor.Field {
		t := tag.GetMoreTags(f)
		if t != nil {
			st := reflect.StructTag(*t)
			if st.Get("request") != "" {
				return true
			}
			if st.Get("header") != "" {
				return true
			}
		}
	}
	return false
}

func (t *kgin) generateGinInterface(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) int {
	count := 0
	servName := naming.ServiceName(service)
	t.P("// " + servName + "GinServer is the server API for " + servName + " service.")

	comments, err := t.Reg.ServiceComments(file, service)
	if err == nil {
		t.PrintComments(comments)
	}
	t.P(`type `, servName, `GinServer interface {`)
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

func (t *kgin) generateInterfaceMethod(file *descriptor.FileDescriptorProto,
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
