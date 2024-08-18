package fauxrpc_test

import (
	"fmt"
	"strings"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func wrapProto(inner string) string {
	return fmt.Sprintf(`syntax = "proto3";

package test.v1;

import "buf/validate/validate.proto";

%s`, inner)
}

func mustCompileField(fieldType, fieldName, fieldOptions string) protoreflect.FieldDescriptor {
	optionsSection := ""
	if len(fieldOptions) > 0 {
		optionsSection = "[" + fieldOptions + "]"
	}
	protoText := wrapProto(fmt.Sprintf(`
	message Test {
		%s %s = 1 %s;
	}
	`, fieldType, fieldName, optionsSection))
	handler := reporter.NewHandler(nil)
	ast, err := parser.Parse("test.proto", strings.NewReader(protoText), handler)
	if err != nil {
		panic(err)
	}
	res, err := parser.ResultFromAST(ast, true, handler)
	if err != nil {
		panic(fmt.Errorf("convert from AST: %w", err))
	}
	// fdText, _ := protojson.MarshalOptions{Indent: "    "}.Marshal(res.FileDescriptorProto())
	// fmt.Println(string(fdText))
	fd, err := protodesc.NewFile(res.FileDescriptorProto(), protoregistry.GlobalFiles)
	if err != nil {
		panic(fmt.Errorf("protodesc.NewFile: %w", err))
	}

	return fd.Messages().ByName("Test").Fields().ByName(protoreflect.Name(fieldName))
}
