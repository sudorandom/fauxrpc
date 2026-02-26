package registry

import (
	"fmt"
	"testing"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func BenchmarkSortFilesByDependency(b *testing.B) {
	files := new(protoregistry.Files)
	numFiles := 1000
	for i := 0; i < numFiles; i++ {
		name := fmt.Sprintf("file%d.proto", i)
		fdp := &descriptorpb.FileDescriptorProto{
			Name: &name,
		}
		if i > 0 {
			dep := fmt.Sprintf("file%d.proto", i-1)
			fdp.Dependency = []string{dep}
		}

		fd, err := protodesc.NewFile(fdp, files)
		if err != nil {
			b.Fatalf("failed to create file descriptor for %s: %v", name, err)
		}
		if err := files.RegisterFile(fd); err != nil {
			b.Fatalf("failed to register file %s: %v", name, err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sortFilesByDependency(files)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestSortFilesByDependency(t *testing.T) {
	files := new(protoregistry.Files)
	numFiles := 10
	for i := 0; i < numFiles; i++ {
		name := fmt.Sprintf("file%d.proto", i)
		fdp := &descriptorpb.FileDescriptorProto{
			Name: &name,
		}
		if i > 0 {
			dep := fmt.Sprintf("file%d.proto", i-1)
			fdp.Dependency = []string{dep}
		}

		fd, err := protodesc.NewFile(fdp, files)
		if err != nil {
			t.Fatalf("failed to create file descriptor for %s: %v", name, err)
		}
		if err := files.RegisterFile(fd); err != nil {
			t.Fatalf("failed to register file %s: %v", name, err)
		}
	}

	sorted, err := sortFilesByDependency(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != numFiles {
		t.Fatalf("expected %d files, got %d", numFiles, len(sorted))
	}

	for i, fd := range sorted {
		expectedName := fmt.Sprintf("file%d.proto", i)
		if fd.Path() != expectedName {
			t.Errorf("expected file at index %d to be %s, got %s", i, expectedName, fd.Path())
		}
	}
}
