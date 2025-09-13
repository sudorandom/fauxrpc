package registry

import (
	"bytes"
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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"buf.build/gen/go/bufbuild/registry/connectrpc/go/buf/registry/module/v1/modulev1connect"
)

// errCacheMiss removed

func looksLikeBSR(path string) bool {
	return strings.HasPrefix(path, "buf.build/")
}

// AddServicesFromBSR resolves, downloads, and registers Protobuf services from the Buf Schema Registry (BSR).
// It uses a local file-based cache to avoid re-downloading modules.
func AddServicesFromBSR(registry LoaderTarget, module string) error {
	module, ref, _ := strings.Cut(module, ":")
	if ref == "" {
		ref = "main"
	}

	httpClient := newBufHttpClient()

	// 1. Resolve the module's 'ref' (e.g., "main") to a specific commit ID.
	rootCommitID, err := resolveBSRCommitID(httpClient, module, ref)
	if err != nil {
		return err
	}

	var fds *descriptorpb.FileDescriptorSet

	// 2. Try to load the compiled file descriptors from the local cache.
	// If any error occurs (including cache miss), proceed to load from BSR.
	cachedFds, err := loadFromCache(module, rootCommitID)
	if err == nil {
		slog.Debug("cache hit", slog.String("module", module), slog.String("commit", rootCommitID))
		fds = cachedFds
	} else {
		slog.Debug("cache load failed or miss, loading from BSR", slog.String("module", module), slog.String("commit", rootCommitID), slog.String("error", err.Error()))
		// 3. Load the module from the BSR.
		bsrFds, bsrErr := loadFromBSR(httpClient, module, rootCommitID)
		if bsrErr != nil {
			return fmt.Errorf("failed to load from BSR: %w", bsrErr)
		}

		fds = bsrFds

		// Asynchronously save the newly downloaded files to the cache.
		go func() {
			if err := saveToCache(module, rootCommitID, bsrFds); err != nil {
				slog.Warn("failed to save to cache", slog.String("error", err.Error()))
			}
		}()
	}

	// 4. Register the file descriptors with the target registry.
	// This step is now common for both cached and non-cached paths.
	slog.Info("registering files for module", slog.String("module", module), slog.Int("count", len(fds.GetFile())))
	files, err := protodesc.FileOptions{AllowUnresolvable: true}.NewFiles(fds)
	if err != nil {
		return err
	}
	var registerErr error
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if err := registry.RegisterFile(fd); err != nil {
			registerErr = fmt.Errorf("RegisterFile for %s: %w", fd.Path(), err)
			return false
		}
		return true
	})
	if registerErr != nil {
		return registerErr
	}
	return nil
}

func resolveBSRCommitID(httpClient *http.Client, module, ref string) (string, error) {
	parts := strings.Split(module, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid module format: %s", module)
	}
	remote, owner, repoAndRef := parts[0], parts[1], parts[2]
	repo, _, _ := strings.Cut(repoAndRef, ":")
	apiURL := "https://" + remote
	labelClient := modulev1connect.NewLabelServiceClient(httpClient, apiURL)

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

// handleCacheLoadError logs a warning and attempts to delete the corrupted cache file.
func handleCacheLoadError(cachePath string, originalErr error) error {
	slog.Warn("failed to load from cache, a new version will be downloaded", slog.String("path", cachePath), slog.String("error", originalErr.Error()))
	if rmErr := os.Remove(cachePath); rmErr != nil {
		slog.Warn("failed to delete cache file", slog.String("path", cachePath), slog.String("error", rmErr.Error()))
	}
	return fmt.Errorf("failed to load from cache: %w", originalErr)
}

// loadFromCache reads a .binpb.gz file from the local cache, unmarshals it into a
// FileDescriptorSet, and then converts it to a slice of FileDescriptors.
func loadFromCache(module, commitID string) (*descriptorpb.FileDescriptorSet, error) {
	cachePath, err := bsrCachePath(module, commitID)
	if err != nil {
		return nil, err // This error is not related to cache file corruption, so don't delete.
	}

	gzipBytes, err := os.ReadFile(cachePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fs.ErrNotExist // This is a cache miss, not an error to warn about or delete.
		}
		return nil, handleCacheLoadError(cachePath, fmt.Errorf("failed to read cache file %q: %w", cachePath, err))
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(gzipBytes))
	if err != nil {
		return nil, handleCacheLoadError(cachePath, fmt.Errorf("failed to create gzip reader: %w", err))
	}
	defer gzipReader.Close()

	bytes, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, handleCacheLoadError(cachePath, fmt.Errorf("failed to read from gzip reader: %w", err))
	}

	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(bytes, fds); err != nil {
		return nil, handleCacheLoadError(cachePath, fmt.Errorf("failed to unmarshal file descriptor set: %w", err))
	}

	return fds, nil
}

// saveToCache writes the given protobuf file contents into a gzipped fds file
// at the appropriate cache location.
func saveToCache(module, commitID string, fds *descriptorpb.FileDescriptorSet) error {
	cachePath, err := bsrCachePath(module, commitID)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	rawBytes, err := proto.Marshal(fds)
	if err != nil {
		return fmt.Errorf("failed to marshal file descriptor set: %w", err)
	}

	var gzipBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipBuffer)
	if _, err := gzipWriter.Write(rawBytes); err != nil {
		return fmt.Errorf("failed to write to gzip writer: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	if err := os.WriteFile(cachePath, gzipBuffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	slog.Debug("wrote to cache", slog.String("path", cachePath))
	return nil
}

func bsrCachePath(module, commitID string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache dir: %w", err)
	}
	return filepath.Join(cacheDir, "fauxrpc", "bsr", module, fmt.Sprintf("%s.binpb.gz", commitID)), nil
}

func loadFromBSR(httpClient *http.Client, module, rootCommitID string) (*descriptorpb.FileDescriptorSet, error) {
	slog.Info("loading from BSR", slog.String("module", module), slog.String("commit", rootCommitID))

	parts := strings.Split(module, "/")
	remote, owner, repoAndRef := parts[0], parts[1], parts[2]
	repo, _, _ := strings.Cut(repoAndRef, ":")
	apiURL := "https://" + remote

	fdsClient := modulev1connect.NewFileDescriptorSetServiceClient(httpClient, apiURL)
	fdsRes, err := fdsClient.GetFileDescriptorSet(context.Background(), connect.NewRequest(&modulev1.GetFileDescriptorSetRequest{
		ResourceRef: modulev1.ResourceRef_builder{
			Name: modulev1.ResourceRef_Name_builder{
				Owner:  owner,
				Module: repo,
				Ref:    &rootCommitID,
			}.Build(),
		}.Build(),
		ExcludeImports:                false,
		IncludeSourceCodeInfo:         false,
		IncludeSourceRetentionOptions: true,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to get file descriptor set for commit %s: %w", rootCommitID, err)
	}

	return fdsRes.Msg.GetFileDescriptorSet(), nil
}

func newBufHttpClient() *http.Client {
	return &http.Client{
		Transport: &bufAuthInterceptor{
			transport: http.DefaultTransport,
		},
	}
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
