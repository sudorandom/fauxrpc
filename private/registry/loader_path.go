package registry

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AddServicesFromPath imports services from a given 'path' which can be a local file path, directory,
// BSR repo, server address for server reflection.
func AddServicesFromPath(ctx context.Context, registry LoaderTarget, path string) error {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return AddServicesFromReflection(registry, http.DefaultClient, path)
	}
	stat, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) && looksLikeBSR(path) {
		return AddServicesFromBSR(ctx, registry, path)
	} else if err != nil {
		return err
	}

	dirPath := path
	if !stat.IsDir() {
		dirPath = filepath.Dir(path)
	}

	bufYamlPath := filepath.Join(dirPath, "buf.yaml")
	slog.Debug("checking for buf.yaml", "dirPath", dirPath, "bufYamlPath", bufYamlPath)
	if _, err := os.Stat(bufYamlPath); err == nil {
		// buf.yaml exists, so we use buf to build the descriptors
		return addServicesWithBuf(registry, dirPath)
	} else {
		slog.Info("buf.yaml not found", "error", err)
	}

	if stat.IsDir() {
		if err := fs.WalkDir(os.DirFS(path), ".", func(childpath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if err := AddServicesFromSingleFile(registry, filepath.Join(path, childpath)); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	return AddServicesFromSingleFile(registry, path)
}

func addServicesWithBuf(registry LoaderTarget, dirPath string) error {
	slog.Info("found buf.yaml, building with buf", "path", dirPath)
	tmpFile, err := os.CreateTemp("", "bufbuild-*.binpb")
	if err != nil {
		return fmt.Errorf("failed to create temp file for buf build: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			slog.Error("failed to remove temp file", "path", tmpFile.Name(), "error", err)
		}
	}()
	// Close the file so that buf can write to it
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	cmd := exec.Command("buf", "build", "-o", tmpFile.Name())
	cmd.Dir = dirPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("buf build failed: %w\n%s", err, string(output))
	}

	return AddServicesFromSingleFile(registry, tmpFile.Name())
}
