package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	connectPackage = protogen.GoImportPath("connectrpc.com/connect")
	contextPackage = protogen.GoImportPath("context")
	errorsPackage  = protogen.GoImportPath("errors")
	mockPackage    = protogen.GoImportPath("github.com/stretchr/testify/mock")
	fauxrpcPackage = protogen.GoImportPath("github.com/sudorandom/fauxrpc")

	generatedPackageSuffix = "connect"

	usage = `See https://fauxrpc.com to learn how to use this plugin.

Flags:
	-h, --help	Print this help and exit.
	--version	Print the version and exit.`
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Fprintln(os.Stdout, connect.Version)
		os.Exit(0)
	}
	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Fprintln(os.Stdout, usage)
		os.Exit(0)
	}
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}
	protogen.Options{}.Run(
		func(plugin *protogen.Plugin) error {
			plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) | uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
			plugin.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
			plugin.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2023
			for _, file := range plugin.Files {
				if file.Generate {
					generate(plugin, file)
				}
			}
			return nil
		},
	)
}

func generate(plugin *protogen.Plugin, file *protogen.File) {
	if len(file.Services) == 0 {
		return
	}
	file.GoPackageName += generatedPackageSuffix

	generatedFilenamePrefixToSlash := filepath.ToSlash(file.GeneratedFilenamePrefix)
	file.GeneratedFilenamePrefix = path.Join(
		path.Dir(generatedFilenamePrefixToSlash),
		string(file.GoPackageName),
		path.Base(generatedFilenamePrefixToSlash),
	)
	generatedFile := plugin.NewGeneratedFile(
		file.GeneratedFilenamePrefix+".faux.go",
		protogen.GoImportPath(path.Join(
			string(file.GoImportPath),
			string(file.GoPackageName),
		)),
	)

	generatedFile.Import(file.GoImportPath)
	generatePreamble(generatedFile, file)
	for _, service := range file.Services {
		generateService(generatedFile, service)
	}
}

func generatePreamble(g *protogen.GeneratedFile, file *protogen.File) {
	g.P("package ", file.GoPackageName)
	g.P()
}

func generateService(g *protogen.GeneratedFile, service *protogen.Service) {
	names := newNames(service)

	g.P("type ", names.FauxHandler, " struct {")
	g.P(mockPackage.Ident("Mock"))
	g.P("}")
	g.P()
	g.P("func ", names.NewFauxHandler, "() *", names.FauxHandler, " {")
	g.P("h := new(", names.FauxHandler, ")")
	// for _, method := range service.Methods {
	// 	g.P("{")
	// 	g.P("resp := &", g.QualifiedGoIdent(method.Output.GoIdent), "{}")
	// 	g.P(fauxrpcPackage.Ident("SetDataOnMessage"), "(resp, opts)")
	// 	g.P(`h.On("`,
	// 		method.GoName, `", `,
	// 		mockPackage.Ident("Anything"), ", ",
	// 		mockPackage.Ident("Anything"), ")",
	// 		".Return(connect.NewResponse(resp), nil)",
	// 	)
	// 	g.P("}")
	// }
	g.P("return h")
	g.P("}")
	g.P()
	for _, method := range service.Methods {
		g.P("func (h *", names.FauxHandler, ") ", serverSignature(g, method), "{")
		if method.Desc.IsStreamingServer() {
			// TODO: Support the three variants of streaming calls.
			// The idea is to maybe make a pseudo-method called "MethodName:stream"
			// that will call a user-provided function to get the connect.BidiStream,
			// ClientStream or ServerStream instance, so callers can can mock out the
			// behavior more precisely.
			// For cases where you don't need a lot of control, an "MethodName:onMessage"
			// pseudo-method will be called each time a message is received.
			g.P("h.Called(ctx, stream)")
			g.P("return ", connectPackage.Ident("NewError"), "(",
				connectPackage.Ident("CodeUnimplemented"), ", ", errorsPackage.Ident("New"),
				`("`, method.Desc.FullName(), ` is not implemented"))`)
		} else {
			g.P("args := h.Called(ctx, req)")
			g.P("resp, _ = args.Get(0).(*",
				g.QualifiedGoIdent(connectPackage.Ident("Response")),
				"["+g.QualifiedGoIdent(method.Output.GoIdent)+"])")
			g.P("err, _ = args.Get(1).(error)")
			g.P("return resp, err")
		}
		g.P("}")
		g.P()
	}
	g.P()
}

func serverSignature(g *protogen.GeneratedFile, method *protogen.Method) string {
	return method.GoName + serverSignatureParams(g, method)
}

func serverSignatureParams(g *protogen.GeneratedFile, method *protogen.Method) string {
	if method.Desc.IsStreamingClient() && method.Desc.IsStreamingServer() {
		// bidi streaming
		return "(ctx " + g.QualifiedGoIdent(contextPackage.Ident("Context")) + ", " +
			"stream *" + g.QualifiedGoIdent(connectPackage.Ident("BidiStream")) +
			"[" + g.QualifiedGoIdent(method.Input.GoIdent) + ", " + g.QualifiedGoIdent(method.Output.GoIdent) + "]" +
			") error"
	}
	if method.Desc.IsStreamingClient() {
		// client streaming
		return "(ctx " + g.QualifiedGoIdent(contextPackage.Ident("Context")) + ", " +
			"stream *" + g.QualifiedGoIdent(connectPackage.Ident("ClientStream")) +
			"[" + g.QualifiedGoIdent(method.Input.GoIdent) + "]" +
			") (*" + g.QualifiedGoIdent(connectPackage.Ident("Response")) + "[" + g.QualifiedGoIdent(method.Output.GoIdent) + "] ,error)"
	}
	if method.Desc.IsStreamingServer() {
		// server streaming
		return "(ctx " + g.QualifiedGoIdent(contextPackage.Ident("Context")) +
			", req *" + g.QualifiedGoIdent(connectPackage.Ident("Request")) + "[" +
			g.QualifiedGoIdent(method.Input.GoIdent) + "], " +
			"stream *" + g.QualifiedGoIdent(connectPackage.Ident("ServerStream")) +
			"[" + g.QualifiedGoIdent(method.Output.GoIdent) + "]" +
			") error"
	}
	// unary
	return "(ctx " + g.QualifiedGoIdent(contextPackage.Ident("Context")) +
		", req *" + g.QualifiedGoIdent(connectPackage.Ident("Request")) + "[" +
		g.QualifiedGoIdent(method.Input.GoIdent) + "]) " +
		"(resp *" + g.QualifiedGoIdent(connectPackage.Ident("Response")) + "[" +
		g.QualifiedGoIdent(method.Output.GoIdent) + "], err error)"
}

type names struct {
	Handler        string
	FauxHandler    string
	NewFauxHandler string
}

func newNames(service *protogen.Service) names {
	base := service.GoName
	return names{
		Handler:        fmt.Sprintf("%sHandler", base),
		FauxHandler:    fmt.Sprintf("mocked%sHandler", base),
		NewFauxHandler: fmt.Sprintf("NewMocked%sHandler", base),
	}
}
