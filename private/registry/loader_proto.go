package registry

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// AddServicesFromProtoFile imports services from a .proto file
func AddServicesFromProtoFile(registry LoaderTarget, filePath string) error {
	slog.Debug("AddServicesFromProtoFile", slog.String("filepath", filePath))
	if _, err := os.Stat(filePath); err != nil {
		return err
	}

	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(protocompile.CompositeResolver{
			protocompile.ResolverFunc(func(path string) (protocompile.SearchResult, error) {
				if path == filePath {
					f, err := os.Open(filePath)
					if err != nil {
						return protocompile.SearchResult{}, err
					}
					return protocompile.SearchResult{Source: f}, nil
				}
				if fd, err := registry.FindFileByPath(path); err == nil {
					return protocompile.SearchResult{Desc: fd}, nil
				}
				if fd, err := protoregistry.GlobalFiles.FindFileByPath(path); err == nil {
					return protocompile.SearchResult{Desc: fd}, nil
				}
				return protocompile.SearchResult{}, protoregistry.NotFound
			}),
			&protocompile.SourceResolver{
				ImportPaths: []string{filepath.Dir(filePath), "."},
			},
		}),
	}

	files, err := compiler.Compile(context.Background(), filePath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := registerFileWithDeps(registry, f); err != nil {
			return err
		}
	}
	return nil
}

func registerFileWithDeps(registry LoaderTarget, fd protoreflect.FileDescriptor) error {
	if _, err := registry.FindFileByPath(fd.Path()); err == nil {
		return nil
	}
	imports := fd.Imports()
	for i := 0; i < imports.Len(); i++ {
		imp := imports.Get(i).FileDescriptor
		if imp == nil {
			continue
		}
		if err := registerFileWithDeps(registry, imp); err != nil {
			return err
		}
	}
	if res, ok := fd.(linker.Result); ok {
		b, err := proto.Marshal(res.FileDescriptorProto())
		if err == nil {
			fdp := new(descriptorpb.FileDescriptorProto)
			unmarshalOpts := proto.UnmarshalOptions{}
			if r, ok := registry.(interface{ Resolver() *Resolver }); ok {
				unmarshalOpts.Resolver = r.Resolver()
			}
			if err := unmarshalOpts.Unmarshal(b, fdp); err == nil {
				if converted, err := protodesc.NewFile(fdp, registry); err == nil {
					fd = converted
				}
			}
		}
	}
	return registry.RegisterFile(fd)
}
