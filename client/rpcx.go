package client

import (
	"fmt"
	"io/ioutil"
	"strings"

	pb "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

const (
	rpcxServerPkgPath   = "github.com/smallnest/rpcx/server"
	rpcxClientPkgPath   = "github.com/smallnest/rpcx/client"
	rpcxProtocolPkgPath = "github.com/smallnest/rpcx/protocol"
)

func init() {
	generator.RegisterPlugin(new(rpcx))
}

type rpcx struct {
	gen *generator.Generator
}

// Name returns the name of this plugin
func (p *rpcx) Name() string {
	return "rpcx"
}

// Init initializes the plugin.
func (p *rpcx) Init(gen *generator.Generator) {
	p.gen = gen
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (p *rpcx) objectNamed(name string) generator.Object {
	p.gen.RecordTypeUse(name)
	return p.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (p *rpcx) typeName(str string) string {
	return p.gen.TypeName(p.objectNamed(str))
}

// GenerateImports generates the import declaration for this file.
func (p *rpcx) GenerateImports(file *generator.FileDescriptor) {
}

// P forwards to g.gen.P.
func (p *rpcx) P(args ...interface{}) { p.gen.P(args...) }

// Generate generates code for the services in the given file.
func (p *rpcx) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}

	// generate all services
	for i, service := range file.FileDescriptorProto.Service {
		p.gen.Reset()
		p.generateService(file, service, i)
		serviceName := strings.ToLower(service.GetName())
		//b := append([]byte(fmt.Sprintf("package %s\n", serviceName)), p.gen.Bytes()...)
		err := ioutil.WriteFile(fmt.Sprintf("%sclient.go", serviceName), p.gen.Bytes(), 0644)
		if err != nil {
			panic(err)
		}
		p.gen.Reset()
	}
}

// generateService generates all the code for the named service
func (p *rpcx) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	originServiceName := service.GetName()
	serviceName := upperFirstLatter(originServiceName)
	p.P("// This following code was generated by rpcx")
	p.P(fmt.Sprintf("// Gernerated from %s", file.GetName()))

	// xclient
	p.P(fmt.Sprintf("package %s", strings.ToLower(service.GetName())))
	p.P()
	p.P(`import (
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"context"
)`)
	p.P()

	p.P()
	p.P(fmt.Sprintf(`type %[1]sClient struct{
	xclient client.XClient
}

func New%[1]sClient(xclient client.XClient) *%[1]sClient {
	return &%[1]sClient{xclient: xclient}
}

func NewXClientFor%[1]s(addr string) (client.XClient, error) {
	d, err := client.NewPeer2PeerDiscovery("tcp@"+addr, "")
	if err != nil {
		return nil, err
	}
	
	opt := client.DefaultOption
	opt.SerializeType = protocol.ProtoBuffer

	xclient := client.NewXClient("%[1]s", client.Failtry, client.RoundRobin, d, opt)

	return xclient,nil
}`, serviceName))
	for _, method := range service.Method {
		p.generateClientCode(service, method)
	}

	// one client
	p.P()
	p.P(fmt.Sprintf(`type %[1]sOneClient struct{
	serviceName string
	oneclient client.OneClient
}

func New%[1]sOneClient(oneclient client.OneClient) *%[1]sOneClient {
	return &%[1]sOneClient{
		serviceName: "%[1]s",
		oneclient: oneclient,
	}
}`, serviceName))
	for _, method := range service.Method {
		p.generateOneClientCode(service, method)
	}
}

func (p *rpcx) generateAbleCode(method *pb.MethodDescriptorProto) {
	methodName := upperFirstLatter(method.GetName())
	inType := p.typeName(method.GetInputType())
	outType := p.typeName(method.GetOutputType())
	p.P(fmt.Sprintf(`// %[1]s is server rpc method as defined
		%[1]s(ctx context.Context, args *%[2]s, reply *%[3]s) (err error)
	`, methodName, inType, outType))
}

func (p *rpcx) generateServerCode(service *pb.ServiceDescriptorProto, method *pb.MethodDescriptorProto) {
	methodName := upperFirstLatter(method.GetName())
	serviceName := upperFirstLatter(service.GetName())
	inType := p.typeName(method.GetInputType())
	outType := p.typeName(method.GetOutputType())
	p.P(fmt.Sprintf(`func (s *%sImpl) %s(ctx context.Context, args *%s, reply *%s) (err error){
	// TODO: add business logics

	// TODO: setting return values
	*reply = %s{}

	return nil
}`, serviceName, methodName, inType, outType, outType))
}

func (p *rpcx) generateClientCode(service *pb.ServiceDescriptorProto, method *pb.MethodDescriptorProto) {
	methodName := upperFirstLatter(method.GetName())
	serviceName := upperFirstLatter(service.GetName())
	inType := p.typeName(method.GetInputType())
	outType := p.typeName(method.GetOutputType())
	p.P(fmt.Sprintf(`func (c *%sClient) %s(ctx context.Context, args *%s)(reply *%s, err error){
	reply = &%s{}
	err = c.xclient.Call(ctx,"%s",args, reply)
	return reply, err
}`, serviceName, methodName, inType, outType, outType, method.GetName()))
}

func (p *rpcx) generateOneClientCode(service *pb.ServiceDescriptorProto, method *pb.MethodDescriptorProto) {
	methodName := upperFirstLatter(method.GetName())
	serviceName := upperFirstLatter(service.GetName())
	inType := p.typeName(method.GetInputType())
	outType := p.typeName(method.GetOutputType())
	p.P(fmt.Sprintf(`func (c *%sOneClient) %s(ctx context.Context, args *%s)(reply *%s, err error){
	reply = &%s{}
	err = c.oneclient.Call(ctx,c.serviceName,"%s",args, reply)
	return reply, err
}`, serviceName, methodName, inType, outType, outType, method.GetName()))
}

// upperFirstLatter make the fisrt charater of given string  upper class
func upperFirstLatter(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return strings.ToUpper(string(s[0]))
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}
