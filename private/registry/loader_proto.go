package registry

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/protobuf/reflect/protodesc"
)

// AddServicesFromProtoFile imports services from a .proto file
func AddServicesFromProtoFile(registry LoaderTarget, filepath string) error {
	slog.Debug("AddServicesFromProtoFile", slog.String("filepath", filepath))
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	handler := reporter.NewHandler(nil)
	ast, err := parser.Parse(filepath, f, handler)
	if err != nil {
		return err
	}
	res, err := parser.ResultFromAST(ast, true, handler)
	if err != nil {
		return fmt.Errorf("convert from AST: %w", err)
	}
	fd, err := protodesc.NewFile(res.FileDescriptorProto(), registry)
	if err != nil {
		return fmt.Errorf("protodesc.NewFile: %w", err)
	}

	return registry.RegisterFile(fd)
}
