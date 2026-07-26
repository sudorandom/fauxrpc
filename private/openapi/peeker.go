package openapi

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

const schemaPeekTimeout = 10 * time.Second

// IsOpenAPISpec checks whether a given file path or URL content resembles an OpenAPI specification (v3 or v2/swagger).
func IsOpenAPISpec(pathOrURL string) bool {
	return isOpenAPISpec(pathOrURL, &http.Client{Timeout: schemaPeekTimeout})
}

func isOpenAPISpec(pathOrURL string, client *http.Client) bool {
	var contents []byte
	var err error

	if isURL(pathOrURL) {
		resp, httpErr := client.Get(pathOrURL)
		if httpErr != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return false
		}
		contents, err = io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		if err != nil {
			return false
		}
	} else {
		stat, statErr := os.Stat(pathOrURL)
		if statErr != nil || stat.IsDir() {
			return false
		}
		file, openErr := os.Open(pathOrURL)
		if openErr != nil {
			return false
		}
		defer func() { _ = file.Close() }()
		contents, err = io.ReadAll(io.LimitReader(file, 64*1024))
		if err != nil {
			return false
		}
	}

	// Peeking check: search for openapi or swagger keys in JSON/YAML
	var doc struct {
		OpenAPI string `yaml:"openapi"`
		Swagger string `yaml:"swagger"`
	}

	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	if err := decoder.Decode(&doc); err == nil {
		if strings.HasPrefix(doc.OpenAPI, "3.") || strings.HasPrefix(doc.Swagger, "2.") || doc.OpenAPI != "" || doc.Swagger != "" {
			return true
		}
	}

	// Substring fallback check
	contentStr := string(contents)
	if strings.Contains(contentStr, `"openapi"`) || strings.Contains(contentStr, `openapi:`) ||
		strings.Contains(contentStr, `"swagger"`) || strings.Contains(contentStr, `swagger:`) {
		return true
	}

	return false
}
