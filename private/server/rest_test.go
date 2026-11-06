package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	connectv2 "connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/vanguard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// TestHandler_REST serves a method over REST via vanguard.Mount, using
// WithRules because the eliza schema carries no google.api.http annotations.
func TestHandler_REST(t *testing.T) {
	reg := mustNewRegistry()
	require.NoError(t, reg.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto))
	srv, err := NewServer(ServerOpts{Addr: "127.0.0.1:0"})
	require.NoError(t, err)
	srv.ServiceRegistry = reg

	sd := reg.Get("connectrpc.eliza.v1.ElizaService")
	require.NotNil(t, sd)

	rpcServer := connectv2.NewServer()
	rpcServer.Register(buildMethods(sd, fauxrpc.NewFauxFaker(), nil, srv, srv.logger, 5)...)

	resolver := reg.Resolver()
	mux := http.NewServeMux()
	connecthttp.Mount(mux, rpcServer,
		connecthttp.WithCodecs(newBinaryCodec(resolver), newJSONCodec(resolver)))
	require.NoError(t, vanguard.Mount(mux, rpcServer,
		vanguard.WithTypeResolver(resolver),
		vanguard.WithRules(&annotations.HttpRule{
			Selector: "connectrpc.eliza.v1.ElizaService.Say",
			Pattern:  &annotations.HttpRule_Post{Post: "/v1/say"},
			Body:     "*",
		})))

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/v1/say", "application/json", strings.NewReader(`{"sentence": "hello"}`))
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "fake", resp.Header.Get("X-Fauxrpc-Source"))

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	sentence, ok := body["sentence"].(string)
	require.True(t, ok, "response missing sentence field: %v", body)
	assert.NotEmpty(t, sentence)
}

const edgeCaseProtoSchema = `syntax = "proto3";
package test.rest.v1;

import "google/api/annotations.proto";
import "google/api/httpbody.proto";

service EdgeCaseService {
  rpc RootInfo(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      get: "/"
    };
  }

  rpc GetItem(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      get: "/v1/items/{id}"
    };
  }

  rpc ArchiveItem(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      post: "/v1/items/{id}:archive"
      body: "*"
    };
  }

  rpc GetBook(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      get: "/v1/{name=shelves/*/books/*}"
      additional_bindings {
        get: "/v1/books/{name=*}"
      }
      additional_bindings {
        get: "/v1/archives/{shelf}/books/{id}"
      }
      // Duplicate trie pattern with primary binding; should be skipped without failing server mount.
      additional_bindings {
        get: "/v1/shelves/{shelf}/books/{id}"
      }
    };
  }

  rpc CheckoutBook(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      post: "/v1/{name=shelves/*/books/*}:checkout"
    };
  }

  rpc UpdateBook(UpdateBookRequest) returns (UpdateBookResponse) {
    option (google.api.http) = {
      patch: "/v1/{book.name=shelves/*/books/*}"
      body: "book"
      response_body: "book"
    };
  }

  rpc GetFile(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      get: "/v1/files/{path=**}"
    };
  }

  rpc TrailingWildcardVerb(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      get: "/trailing/{path=**}:slash"
    };
  }

  rpc TrailingSlashLiteral(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      get: "/v1/trailing/"
      additional_bindings {
        post: "/v1/trailing/:action"
      }
    };
  }

  rpc NestedCapture(NestedCaptureRequest) returns (NestedCaptureResponse) {
    option (google.api.http) = {
      get: "/foo/bob/{book_id={author}/{isbn}/*}/details"
    };
  }

  rpc PercentNestedCapture(PercentNestedRequest) returns (PercentNestedResponse) {
    option (google.api.http) = {
      get: "/percent_encoding/{outer={inner}/suffix}"
    };
  }

  rpc EscapedLiterals(EscapedLiteralsRequest) returns (EscapedLiteralsResponse) {
    option (google.api.http) = {
      get: "/foo%2Fbar/%2A/%2A%2a/{starstar=%2A%2a/**}:%2c"
    };
  }

  rpc CustomLock(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      custom: {
        kind: "LOCK"
        path: "/v1/{name=locks/*}"
      }
    };
  }

  rpc CustomWildcard(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      custom: {
        kind: "*"
        path: "/v1/any/{id}"
      }
    };
  }

  rpc DownloadBlob(EchoRequest) returns (google.api.HttpBody) {
    option (google.api.http) = {
      get: "/v1/blobs/{id}:download"
    };
  }

  rpc UploadBlob(stream google.api.HttpBody) returns (EchoResponse) {
    option (google.api.http) = {
      post: "/v1/blobs:upload"
      body: "*"
    };
  }

  // Non-HttpBody server-streaming RPC with google.api.http annotation:
  // must not prevent the server from mounting other REST routes.
  rpc WatchItems(EchoRequest) returns (stream EchoResponse) {
    option (google.api.http) = {
      get: "/v1/items:watch"
    };
  }

  // Malformed google.api.http rule referencing a non-existent field:
  // must be skipped gracefully without breaking other routes.
  rpc BrokenRule(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      get: "/v1/broken/{nonexistent_field}"
    };
  }
}

message EchoRequest {
  string id = 1;
  string name = 2;
  string shelf = 3;
  string path = 4;
  string query = 5;
}

message EchoResponse {
  string id = 1;
  string name = 2;
  string shelf = 3;
  string path = 4;
  string query = 5;
}

message Book {
  string name = 1;
  string title = 2;
}

message UpdateBookRequest {
  Book book = 1;
  string update_mask = 2;
}

message UpdateBookResponse {
  Book book = 1;
}

message NestedCaptureRequest {
  string book_id = 1;
  string author = 2;
  string isbn = 3;
}

message NestedCaptureResponse {
  string book_id = 1;
  string author = 2;
  string isbn = 3;
}

message PercentNestedRequest {
  string outer = 1;
  string inner = 2;
}

message PercentNestedResponse {
  string outer = 1;
  string inner = 2;
}

message EscapedLiteralsRequest {
  string starstar = 1;
}

message EscapedLiteralsResponse {
  string starstar = 1;
}
`

func setupEdgeCaseServer(t *testing.T) (*server, *httptest.Server) {
	t.Helper()
	protoFile := filepath.Join(t.TempDir(), "edge_cases.proto")
	require.NoError(t, os.WriteFile(protoFile, []byte(edgeCaseProtoSchema), 0o600))

	srv, err := NewServer(ServerOpts{Addr: "127.0.0.1:0", OnlyStubs: true})
	require.NoError(t, err)
	require.NoError(t, registry.AddServicesFromSingleFile(srv.ServiceRegistry, protoFile))

	sd := srv.Get("test.rest.v1.EdgeCaseService")
	require.NotNil(t, sd)

	registerCELStub := func(methodName, expr string) {
		md := sd.Methods().ByName(protoreflect.Name(methodName))
		require.NotNil(t, md)
		celMsg, err := protocel.New(srv.Files(), md.Output(), expr)
		require.NoError(t, err)
		srv.AddStub(stubs.StubEntry{
			Key:        stubs.StubKey{Name: md.FullName(), ID: methodName + "-echo"},
			CELMessage: celMsg,
		})
	}

	echoExpr := `{"id": req.id, "name": req.name, "shelf": req.shelf, "path": req.path, "query": req.query}`
	for _, m := range []string{
		"RootInfo",
		"GetItem",
		"ArchiveItem",
		"GetBook",
		"CheckoutBook",
		"GetFile",
		"TrailingWildcardVerb",
		"TrailingSlashLiteral",
		"CustomLock",
		"CustomWildcard",
	} {
		registerCELStub(m, echoExpr)
	}
	registerCELStub("UpdateBook", `{"book": {"name": req.book.name, "title": req.book.title + " (" + req.update_mask + ")"}}`)
	registerCELStub("NestedCapture", `{"book_id": req.book_id, "author": req.author, "isbn": req.isbn}`)
	registerCELStub("PercentNestedCapture", `{"outer": req.outer, "inner": req.inner}`)
	registerCELStub("EscapedLiterals", `{"starstar": req.starstar}`)
	registerCELStub("DownloadBlob", `{"content_type": "text/plain", "data": b"blob:" + bytes(req.id)}`)
	registerCELStub("UploadBlob", `{"id": "uploaded"}`)

	handler, err := srv.Handler()
	require.NoError(t, err)
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return srv, ts
}

func doJSONRequest(t *testing.T, client *http.Client, method, url, body string) (int, map[string]any, http.Header) {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	require.NoError(t, err)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var out map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return resp.StatusCode, out, resp.Header
}

func TestServer_REST_GoogleAPIHTTPPathEdgeCases(t *testing.T) {
	srv, ts := setupEdgeCaseServer(t)
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	t.Run("root path rule", func(t *testing.T) {
		status, body, _ := doJSONRequest(t, client, http.MethodGet, ts.URL+"/?query=root-ok", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "root-ok", body["query"])
	})

	t.Run("trailing slash literal rule and trailing slash before verb", func(t *testing.T) {
		status, body, _ := doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/trailing/?query=slash-ok", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "slash-ok", body["query"])

		status, body, _ = doJSONRequest(t, client, http.MethodPost, ts.URL+"/v1/trailing/:action?query=verb-slash-ok", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "verb-slash-ok", body["query"])
	})

	t.Run("trailing slash matched by double wildcard before verb", func(t *testing.T) {
		status, body, _ := doJSONRequest(t, client, http.MethodGet, ts.URL+"/trailing/:slash?query=wildcard-slash", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "wildcard-slash", body["query"])
		assert.Empty(t, body["path"])

		status, body, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/trailing/seg1/seg2:slash", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "seg1/seg2", body["path"])

		status, _, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/trailing:slash", "")
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("multi-segment resource pattern and additional_bindings", func(t *testing.T) {
		status, body, _ := doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/shelves/sci-fi/books/dune", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "shelves/sci-fi/books/dune", body["name"])

		status, body, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/books/dune", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "dune", body["name"])

		status, body, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/archives/sci-fi/books/dune-id", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "sci-fi", body["shelf"])
		assert.Equal(t, "dune-id", body["id"])
	})

	t.Run("custom verb suffix on multi-segment and single-segment variables", func(t *testing.T) {
		status, body, _ := doJSONRequest(t, client, http.MethodPost, ts.URL+"/v1/shelves/s1/books/b2:checkout", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "shelves/s1/books/b2", body["name"])

		status, body, _ = doJSONRequest(t, client, http.MethodPost, ts.URL+"/v1/items/item-42:archive", `{"query":"archived"}`)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "item-42", body["id"])
		assert.Equal(t, "archived", body["query"])
	})

	t.Run("colons in path variables: unescaped fallback and %3A disambiguation from verb", func(t *testing.T) {
		// Unescaped colons where suffix is not a registered verb on /v1/items/{id}
		status, body, _ := doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/items/my:weird:id", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "my:weird:id", body["id"])

		// Percent-encoded %3A where "archive" IS a verb on /v1/items/{id}:archive:
		// %3A keeps it in the {id} variable instead of matching :archive.
		status, body, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/items/item%3Aarchive", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "item:archive", body["id"])
	})

	t.Run("percent-encoded characters in single-segment vs multi-segment path variables", func(t *testing.T) {
		// %2F and %2f in single-segment variable decode to '/'
		status, body, _ := doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/items/a%2Fb%2fc", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "a/b/c", body["id"])

		// %25 (literal percent) and %252F (double-encoded slash) in single-segment variable
		status, body, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/items/100%25-off", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "100%-off", body["id"])

		status, body, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/items/a%252Fb", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "a%2Fb", body["id"])

		// Spaces (%20) and UTF-8 in single-segment variable
		status, body, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/items/hello%20world%20%F0%9F%9A%80", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "hello world 🚀", body["id"])

		// %2F and %2f in bounded multi-segment variable stay percent-encoded as %2F
		status, body, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/shelves/s%2F1/books/b%2f2", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "shelves/s%2F1/books/b%2F2", body["name"])

		// %2F in unbounded ** multi-segment variable stays percent-encoded as %2F
		status, body, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/files/dir%2Fsub/file%20name.txt", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "dir%2Fsub/file name.txt", body["path"])

		// Nested single-segment variable inside bounded multi-segment variable
		status, body, _ = doJSONRequest(t, client, http.MethodGet, ts.URL+"/percent_encoding/A%2Fb/suffix", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "A/b", body["inner"])
		assert.Equal(t, "A%2Fb/suffix", body["outer"])
	})

	t.Run("overlapping nested variable captures", func(t *testing.T) {
		status, body, _ := doJSONRequest(t, client, http.MethodGet, ts.URL+"/foo/bob/tolkien/9780261102385/first-ed/details", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "tolkien/9780261102385/first-ed", body["bookId"])
		assert.Equal(t, "tolkien", body["author"])
		assert.Equal(t, "9780261102385", body["isbn"])
	})

	t.Run("escaped literals in path segments and verb with uppercase, lowercase, and unescaped sub-delims", func(t *testing.T) {
		for _, suffix := range []string{":%2C", ":%2c", ":,"} {
			status, body, _ := doJSONRequest(t, client, http.MethodGet, ts.URL+"/foo%2Fbar/%2A/%2A%2a/%2A%2a/tail/seg"+suffix, "")
			require.Equal(t, http.StatusOK, status, "suffix %s", suffix)
			assert.Equal(t, "**"+"/tail/seg", body["starstar"], "suffix %s", suffix)
		}
	})

	t.Run("nested message field path with body and response_body mapping", func(t *testing.T) {
		status, body, _ := doJSONRequest(
			t,
			client,
			http.MethodPatch,
			ts.URL+"/v1/shelves/s1/books/b1?updateMask=title",
			`{"title":"Foundation"}`,
		)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "shelves/s1/books/b1", body["name"])
		assert.Equal(t, "Foundation (title)", body["title"])
	})

	t.Run("custom HTTP method and wildcard custom method", func(t *testing.T) {
		status, body, _ := doJSONRequest(t, client, "LOCK", ts.URL+"/v1/locks/my-lock", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "locks/my-lock", body["name"])

		status, body, _ = doJSONRequest(t, client, "PURGE", ts.URL+"/v1/any/item-99", "")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "item-99", body["id"])
	})

	t.Run("method not allowed sets Allow header", func(t *testing.T) {
		status, _, headers := doJSONRequest(t, client, http.MethodDelete, ts.URL+"/v1/items/123", "")
		assert.Equal(t, http.StatusMethodNotAllowed, status)
		assert.Contains(t, headers.Get("Allow"), "GET")
	})

	t.Run("double slash in REST path returns 404 JSON status without 301 redirect", func(t *testing.T) {
		status, body, _ := doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1//items/123", "")
		assert.Equal(t, http.StatusNotFound, status)
		assert.NotEmpty(t, body["message"])
	})

	t.Run("google.api.HttpBody response and client-streaming upload", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/v1/blobs/abc-123:download")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
		raw, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "blob:abc-123", string(raw))

		status, body, _ := doJSONRequest(t, client, http.MethodPost, ts.URL+"/v1/blobs:upload", "raw-binary-payload")
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, "uploaded", body["id"])
	})

	t.Run("unsupported non-HttpBody streaming REST route returns 415 without breaking server", func(t *testing.T) {
		status, _, _ := doJSONRequest(t, client, http.MethodGet, ts.URL+"/v1/items:watch", "")
		assert.Equal(t, http.StatusUnsupportedMediaType, status)
	})

	t.Run("end-to-end round-trip with vanguard.NewTransport and fauxrpc.NewMessage", func(t *testing.T) {
		sd := srv.Get("test.rest.v1.EdgeCaseService")
		require.NotNil(t, sd)

		transport, err := vanguard.NewTransport(client, ts.URL, vanguard.WithTypeResolver(srv.Resolver()))
		require.NoError(t, err)
		connectClient := connectv2.NewClient(transport)

		for _, methodName := range []string{
			"GetItem",
			"GetBook",
			"CheckoutBook",
			"UpdateBook",
			"GetFile",
			"NestedCapture",
			"CustomLock",
		} {
			t.Run(methodName, func(t *testing.T) {
				md := sd.Methods().ByName(protoreflect.Name(methodName))
				require.NotNil(t, md)

				ctx := fauxrpc.WithMethodDescriptor(context.Background(), md)
				reqMsg, err := fauxrpc.NewMessage(md.Input(), fauxrpc.GenOptions{Context: ctx})
				require.NoError(t, err)

				respMsg := dynamicpb.NewMessage(md.Output())
				spec := connectv2.Spec{
					Procedure:  "/" + string(sd.FullName()) + "/" + string(md.Name()),
					StreamType: connectv2.StreamTypeUnary,
					Schema:     md,
				}
				err = connectClient.CallUnary(ctx, spec, reqMsg, respMsg)
				require.NoError(t, err)
			})
		}
	})
}
