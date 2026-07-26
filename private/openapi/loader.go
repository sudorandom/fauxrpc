package openapi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

type Document struct {
	Doc *openapi3.T
}

func LoadSchema(ctx context.Context, pathOrURL string) (*Document, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	var doc *openapi3.T
	var err error

	if isURL(pathOrURL) {
		u, parseErr := url.Parse(pathOrURL)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse url %s: %w", pathOrURL, parseErr)
		}
		doc, err = loader.LoadFromURI(u)
	} else {
		doc, err = loader.LoadFromFile(pathOrURL)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load openapi schema from %s: %w", pathOrURL, err)
	}

	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid openapi schema in %s: %w", pathOrURL, err)
	}

	return &Document{Doc: doc}, nil
}

func isURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || s[:8] == "https://")
}
