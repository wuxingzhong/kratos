package bm

import (
	"fmt"
	"net/http"
	"unicode"

	"github.com/bilibili/kratos/tool/protobuf/pkg/extensions/google/api/google.golang.org/genproto/googleapis/api/annotations"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	contextPkgPath = "context"
	bmPkgPath      = "github.com/bilibili/kratos/pkg/net/http/blademaster"
	bindPkgPath    = "github.com/bilibili/kratos/pkg/net/http/blademaster/binding"
)

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	contextPkg string
	bmPkg      string
	bindPkg    string
)

func init() {
	generator.RegisterPlugin(newPlugin())
}

func lcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

type plugin struct {
	*generator.Generator
	generator.PluginImports
}

func newPlugin() *plugin {
	return &plugin{}
}

func (p *plugin) Name() string {
	return "bm"
}

func (p *plugin) Init(g *generator.Generator) {
	p.Generator = g
}

// GenerateImports generates the import declaration for this file.
func (p *plugin) GenerateImports(file *generator.FileDescriptor) {}

func (p *plugin) Generate(file *generator.FileDescriptor) {
	contextPkg = string(p.Generator.AddImport(contextPkgPath))
	bmPkg = string(p.Generator.AddImport(bmPkgPath))
	bindPkg = string(p.Generator.AddImport(bindPkgPath))
	p.P("// Reference imports to suppress errors if they are not otherwise used.")
	p.P("var _ ", contextPkg, ".Context")
	p.P("var _ ", bmPkg, ".Engine")
	p.P("var _ ", bindPkg, ".Binding")
	p.P()

	for i, service := range file.FileDescriptorProto.Service {
		p.generateService(file, service, i)
	}
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (p *plugin) objectNamed(name string) generator.Object {
	p.Generator.RecordTypeUse(name)
	return p.Generator.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (p *plugin) typeName(str string) string {
	return p.Generator.TypeName(p.objectNamed(str))
}

func (p *plugin) shouldGenForMethod(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) bool {
	_, err := proto.GetExtension(method.GetOptions(), annotations.E_Http)
	if err != nil {
		return false
	}
	return true
}

// generateService generates all the code for the named service.
func (p *plugin) generateService(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto, index int) {
	p.generateInterface(file, service)
	p.generateHandle(file, service)
	p.generateRoute(file, service)
}

func (p *plugin) generateInterface(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto) {
	servName := generator.CamelCase(service.GetName())
	p.P(`// `, servName, `BMServer is the server API for `+servName+` service.`)
	p.P(`type `, servName, `BMServer interface {`)
	for _, method := range service.Method {
		if !p.shouldGenForMethod(file, service, method) {
			continue
		}
		methName := generator.CamelCase(method.GetName())
		inType := p.typeName(method.GetInputType())
		outType := p.typeName(method.GetOutputType())
		p.P(`	`, methName, `(ctx `, contextPkg, `.Context, req *`, inType, `) (resp *`, outType, `, err error)`)
	}
	p.P(`}`)
}

func (p *plugin) generateHandle(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto) {
	servName := generator.CamelCase(service.GetName())
	varName := lcFirst(servName) + `BMServer `
	p.P(`var `, varName, servName, `BMServer`)
	for _, method := range service.Method {
		if !p.shouldGenForMethod(file, service, method) {
			continue
		}
		fnName := lcFirst(servName) + method.GetName()
		inType := p.typeName(method.GetInputType())
		p.P(`func `, fnName, " (c *", bmPkg, ".Context) {")
		p.P(`	p := new(`, inType, `)`)
		p.P(`	if err := c.BindWith(p, `, bindPkg, `.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {`)
		p.P(`		return`)
		p.P(`	}`)
		p.P(`	resp, err := `, varName, `.`, method.GetName(), `(c, p)`)
		p.P(`	c.JSON(resp, err)`)
		p.P(`}`)
		p.P(``)
	}
}

func (p *plugin) generateRoute(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto) {
	servName := generator.CamelCase(service.GetName())
	varName := lcFirst(servName) + `BMServer `
	fnName := fmt.Sprintf("Register%sBMServer", servName)
	p.P(`// `, fnName, ` Register the blademaster route`)
	p.P(`func `, fnName, `(e *`, bmPkg, `.Engine, server `, servName, `BMServer) {`)
	p.P(varName, ` = server`)
	for _, method := range service.Method {
		if !p.shouldGenForMethod(file, service, method) {
			continue
		}
		var (
			httpMethod  string
			pathPattern string
			httpHandle  = lcFirst(servName) + method.GetName()
		)
		ext, err := proto.GetExtension(method.GetOptions(), annotations.E_Http)
		if err != nil {
			continue
		}
		rule := ext.(*annotations.HttpRule)
		switch pattern := rule.Pattern.(type) {
		case *annotations.HttpRule_Get:
			pathPattern = pattern.Get
			httpMethod = http.MethodGet
		case *annotations.HttpRule_Put:
			pathPattern = pattern.Put
			httpMethod = http.MethodPut
		case *annotations.HttpRule_Post:
			pathPattern = pattern.Post
			httpMethod = http.MethodPost
		case *annotations.HttpRule_Patch:
			pathPattern = pattern.Patch
			httpMethod = http.MethodPatch
		case *annotations.HttpRule_Delete:
			pathPattern = pattern.Delete
			httpMethod = http.MethodDelete
		default:
			continue
		}
		p.P(`e.`, httpMethod, `("`, pathPattern, `", `, rule.Selector, `,`, httpHandle, `)`)
	}
	p.P(`}`)
	p.P(``)
}
