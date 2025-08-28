package registry

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func sortFilesByDependency(files *protoregistry.Files) ([]protoreflect.FileDescriptor, error) {
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Build the dependency graph.
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		inDegree[fd.Path()] = 0
		imports := fd.Imports()
		for i := 0; i < imports.Len(); i++ {
			imp := imports.Get(i)
			graph[imp.Path()] = append(graph[imp.Path()], fd.Path())
			inDegree[fd.Path()]++
		}
		return true
	})

	// Topological sort using Kahn's algorithm.
	var queue []string
	for fileName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, fileName)
		}
	}

	var sortedFiles []protoreflect.FileDescriptor
	for len(queue) > 0 {
		currentFile := queue[0]
		queue = queue[1:]

		fd, err := files.FindFileByPath(currentFile)
		if err != nil {
			return nil, fmt.Errorf("failed to find file %q: %v", currentFile, err)
		}
		sortedFiles = append(sortedFiles, fd)

		for _, neighbor := range graph[currentFile] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return sortedFiles, nil
}
