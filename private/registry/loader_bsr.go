package registry

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	modulev1 "buf.build/gen/go/bufbuild/registry/protocolbuffers/go/buf/registry/module/v1"
	"connectrpc.com/connect"
	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protoreflect"

	"buf.build/gen/go/bufbuild/registry/connectrpc/go/buf/registry/module/v1/modulev1connect"
)

var errCacheMiss = errors.New("cache miss")

func looksLikeBSR(path string) bool {
	return strings.HasPrefix(path, "buf.build/")
}

// AddServicesFromBSR resolves, downloads, and registers Protobuf services from the Buf Schema Registry (BSR).
// It uses a local file-based cache to avoid re-downloading modules.
func AddServicesFromBSR(registry LoaderTarget, module string) error {
	slog.Info("resolving BSR module", slog.String("module", module))

	module, ref, _ := strings.Cut(module, ":")
	if ref == "" {
		ref = "main"
	}

	// 1. Resolve the module's 'ref' (e.g., "main") to a specific commit ID.
	rootCommitID, err := resolveBSRCommitID(module, ref)
	if err != nil {
		return err
	}

	var fds []protoreflect.FileDescriptor
	var fileContents map[string]string

	// 2. Try to load the compiled file descriptors from the local cache.
	cachedFds, err := loadFromCache(module, rootCommitID)
	if err == nil {
		slog.Debug("cache hit", slog.String("module", module), slog.String("commit", rootCommitID))
		fds = cachedFds
	} else if errors.Is(err, errCacheMiss) {
		slog.Debug("cache miss", slog.String("module", module), slog.String("commit", rootCommitID))
		// 3. If it's a cache miss, load the module from the BSR.
		bsrFds, contents, bsrErr := loadFromBSR(module, rootCommitID)
		if bsrErr != nil {
			return fmt.Errorf("failed to load from BSR: %w", bsrErr)
		}
		fds = bsrFds
		fileContents = contents

		// Asynchronously save the newly downloaded files to the cache.
		go func() {
			if err := saveToCache(module, rootCommitID, fileContents); err != nil {
				slog.Warn("failed to save to cache", slog.String("error", err.Error()))
			}
		}()
	} else {
		// A non-miss error from the cache is treated as a fatal error.
		return fmt.Errorf("failed to load from cache: %w", err)
	}

	// 4. Register the file descriptors with the target registry.
	// This step is now common for both cached and non-cached paths.
	slog.Info("registering files for module", slog.String("module", module), slog.Int("count", len(fds)))
	for _, fd := range fds {
		if err := registry.RegisterFile(fd); err != nil {
			return fmt.Errorf("RegisterFile for %s: %w", fd.Path(), err)
		}
	}

	return nil
}

func resolveBSRCommitID(module, ref string) (string, error) {
	parts := strings.Split(module, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid module format: %s", module)
	}
	remote, owner, repoAndRef := parts[0], parts[1], parts[2]
	repo, _, _ := strings.Cut(repoAndRef, ":")
	apiURL := "https://" + remote
	labelClient := modulev1connect.NewLabelServiceClient(newBufHttpClient(), apiURL)

	// A 32-character hex string is likely a commit ID already.
	if len(ref) == 32 && !strings.ContainsAny(ref, "ghijklmnopqrstuvwxyz") {
		slog.Debug("ref looks like a commit ID, skipping label lookup", slog.String("ref", ref))
		return ref, nil
	}

	getLabelsResp, err := labelClient.GetLabels(context.Background(), connect.NewRequest(&modulev1.GetLabelsRequest{
		LabelRefs: []*modulev1.LabelRef{
			{
				Value: &modulev1.LabelRef_Name_{
					Name: &modulev1.LabelRef_Name{
						Label:  ref,
						Owner:  owner,
						Module: repo,
					},
				},
			},
		},
	}))
	if err != nil {
		return "", fmt.Errorf("failed to get label for ref %q: %w", ref, err)
	}
	if len(getLabelsResp.Msg.Labels) == 0 {
		return "", fmt.Errorf("no label found for ref %q", ref)
	}
	return getLabelsResp.Msg.Labels[0].CommitId, nil
}

// loadFromCache reads a .tar.gz file from the local cache, extracts the .proto files,
// and compiles them into file descriptors.
func loadFromCache(module, commitID string) ([]protoreflect.FileDescriptor, error) {
	cachePath, err := bsrCachePath(module, commitID)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(cachePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, errCacheMiss
		}
		return nil, fmt.Errorf("failed to open cache file %q: %w", cachePath, err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader for %q: %w", cachePath, err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	fileContents := make(map[string]string)
	var fileNames []string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header from %q: %w", cachePath, err)
		}

		if header.Typeflag == tar.TypeReg {
			contentBytes, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("failed to read content of %s from %q: %w", header.Name, cachePath, err)
			}
			fileContents[header.Name] = string(contentBytes)
			fileNames = append(fileNames, header.Name)
		}
	}

	if len(fileNames) == 0 {
		return nil, fmt.Errorf("cache file %q is empty: %w", cachePath, errCacheMiss)
	}

	return compileAndRegisterFiles(fileContents, fileNames)
}

// saveToCache writes the given protobuf file contents into a .tar.gz archive
// at the appropriate cache location.
func saveToCache(module, commitID string, fileContents map[string]string) error {
	cachePath, err := bsrCachePath(module, commitID)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	file, err := os.Create(cachePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for name, content := range fileContents {
		header := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %w", name, err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			return fmt.Errorf("failed to write tar content for %s: %w", name, err)
		}
	}

	slog.Debug("wrote to cache", slog.String("path", cachePath))
	return nil
}

func bsrCachePath(module, commitID string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache dir: %w", err)
	}
	return filepath.Join(cacheDir, "fauxrpc", "bsr", module, fmt.Sprintf("%s.tar.gz", commitID)), nil
}

func loadFromBSR(module, rootCommitID string) ([]protoreflect.FileDescriptor, map[string]string, error) {
	slog.Info("loading from BSR", slog.String("module", module), slog.String("commit", rootCommitID))

	parts := strings.Split(module, "/")
	remote, owner, repoAndRef := parts[0], parts[1], parts[2]
	repo, _, _ := strings.Cut(repoAndRef, ":")
	apiURL := "https://" + remote

	httpClient := newBufHttpClient()
	graphClient := modulev1connect.NewGraphServiceClient(httpClient, apiURL)
	downloadClient := modulev1connect.NewDownloadServiceClient(httpClient, apiURL)

	getGraphResp, err := graphClient.GetGraph(context.Background(), connect.NewRequest(&modulev1.GetGraphRequest{
		ResourceRefs: []*modulev1.ResourceRef{
			{
				Value: &modulev1.ResourceRef_Name_{
					Name: &modulev1.ResourceRef_Name{
						Owner:  owner,
						Module: repo,
						Child: &modulev1.ResourceRef_Name_Ref{
							Ref: rootCommitID,
						},
					},
				},
			},
		},
	}))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get graph for commit %s: %w", rootCommitID, err)
	}

	fileContents := map[string]string{}
	var fileNames []string
	for _, commit := range getGraphResp.Msg.Graph.Commits {
		downloadResp, err := downloadClient.Download(context.Background(), connect.NewRequest(&modulev1.DownloadRequest{
			Values: []*modulev1.DownloadRequest_Value{
				{
					FileTypes: []modulev1.FileType{modulev1.FileType_FILE_TYPE_PROTO},
					ResourceRef: &modulev1.ResourceRef{
						Value: &modulev1.ResourceRef_Id{
							Id: commit.Id,
						},
					},
				},
			},
		}))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to download module for commit %s: %w", commit.Id, err)
		}
		for _, content := range downloadResp.Msg.Contents {
			for _, file := range content.GetFiles() {
				fileContents[file.Path] = string(file.Content)
				fileNames = append(fileNames, file.Path)
			}
		}
	}

	fds, err := compileAndRegisterFiles(fileContents, fileNames)
	if err != nil {
		return nil, nil, err
	}

	return []protoreflect.FileDescriptor(fds), fileContents, nil
}

func compileAndRegisterFiles(fileContents map[string]string, fileNames []string) ([]protoreflect.FileDescriptor, error) {
	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(protocompile.ResolverFunc(func(path string) (protocompile.SearchResult, error) {
			if content, ok := fileContents[path]; ok {
				return protocompile.SearchResult{Source: io.NopCloser(strings.NewReader(content))}, nil
			}
			return protocompile.SearchResult{}, fs.ErrNotExist
		})),
	}
	compiledFiles, err := compiler.Compile(context.Background(), fileNames...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile protos: %w", err)
	}
	// The result of Compile is already sorted by dependency, so we can just return them.
	fds := make([]protoreflect.FileDescriptor, len(compiledFiles))
	for i, fd := range compiledFiles {
		fds[i] = fd
	}
	return fds, nil
}

type bufAuthInterceptor struct {
	transport http.RoundTripper
}

func (b *bufAuthInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	if token := os.Getenv("BUF_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return b.transport.RoundTrip(req)
}

func newBufHttpClient() *http.Client {
	return &http.Client{
		Transport: &bufAuthInterceptor{
			transport: http.DefaultTransport,
		},
	}
}
