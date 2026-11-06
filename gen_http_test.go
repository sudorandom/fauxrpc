package fauxrpc_test

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/bufbuild/protocompile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func compileProtoForHTTPTest(t *testing.T, files map[string]string, mainFile string) protoreflect.FileDescriptor {
	t.Helper()
	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(protocompile.CompositeResolver{
			&protocompile.SourceResolver{
				Accessor: protocompile.SourceAccessorFromMap(files),
			},
			protocompile.ResolverFunc(func(path string) (protocompile.SearchResult, error) {
				fd, err := protoregistry.GlobalFiles.FindFileByPath(path)
				if err != nil {
					return protocompile.SearchResult{}, err
				}
				return protocompile.SearchResult{Desc: fd}, nil
			}),
		}),
	}
	compiled, err := compiler.Compile(context.Background(), mainFile)
	require.NoError(t, err)
	return compiled[0]
}

func TestGenHTTPPathTemplates(t *testing.T) {
	protoSrc := `syntax = "proto3";
package test.http.v1;

import "google/api/annotations.proto";

service BookService {
  rpc GetBook(GetBookRequest) returns (Book) {
    option (google.api.http) = {
      get: "/v1/{name=shelves/*/books/*}"
    };
  }

  rpc CheckoutBook(CheckoutBookRequest) returns (Book) {
    option (google.api.http) = {
      post: "/v1/{name=shelves/*/books/*}:checkout"
    };
  }

  rpc UpdateBook(UpdateBookRequest) returns (Book) {
    option (google.api.http) = {
      patch: "/v1/{book.name=shelves/*/books/*}"
      body: "book"
    };
  }

  rpc GetBookDetails(GetBookDetailsRequest) returns (Book) {
    option (google.api.http) = {
      get: "/foo/bob/{book_id={author}/{isbn}/*}/details"
    };
  }

  rpc GetFile(GetFileRequest) returns (GetFileRequest) {
    option (google.api.http) = {
      get: "/v1/files/{path=**}:download"
    };
  }

  rpc GetEscapedMulti(GetEscapedMultiRequest) returns (GetEscapedMultiRequest) {
    option (google.api.http) = {
      get: "/foo%2Fbar/%2A/{name=hello%2Fworld/*}/test:%2C"
    };
  }

  rpc GetEscapedSingle(GetEscapedSingleRequest) returns (GetEscapedSingleRequest) {
    option (google.api.http) = {
      get: "/v1/items/{name=hello%2Fworld}"
    };
  }

  rpc LockResource(LockResourceRequest) returns (LockResourceRequest) {
    option (google.api.http) = {
      custom: {
        kind: "LOCK"
        path: "/v1/{resource=projects/*/locks/*}"
      }
    };
  }
}

message GetBookRequest {
  string name = 1;
}

message CheckoutBookRequest {
  string name = 1;
}

message UpdateBookRequest {
  Book book = 1;
}

message Book {
  string name = 1;
  string title = 2;
}

message GetBookDetailsRequest {
  string book_id = 1;
  string author = 2;
  int64 isbn = 3;
}

message GetFileRequest {
  string path = 1;
}

message GetEscapedMultiRequest {
  string name = 1;
}

message GetEscapedSingleRequest {
  string name = 1;
}

message LockResourceRequest {
  string resource = 1;
}
`
	fd := compileProtoForHTTPTest(t, map[string]string{"service.proto": protoSrc}, "service.proto")
	shelfBookPattern := regexp.MustCompile(`^shelves/[a-zA-Z0-9._~-]+/books/[a-zA-Z0-9._~-]+$`)

	t.Run("multi-segment resource pattern on request", func(t *testing.T) {
		md := fd.Messages().ByName("GetBookRequest")
		msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{})
		require.NoError(t, err)
		name := msg.ProtoReflect().Get(md.Fields().ByName("name")).String()
		assert.Regexp(t, shelfBookPattern, name)

		// Also via fauxrpc.String directly
		strVal := fauxrpc.String(md.Fields().ByName("name"), fauxrpc.GenOptions{})
		assert.Regexp(t, shelfBookPattern, strVal)
	})

	t.Run("custom verb suffix in path template", func(t *testing.T) {
		md := fd.Messages().ByName("CheckoutBookRequest")
		msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{})
		require.NoError(t, err)
		name := msg.ProtoReflect().Get(md.Fields().ByName("name")).String()
		assert.Regexp(t, shelfBookPattern, name)
	})

	t.Run("nested field path in request and derived resource message", func(t *testing.T) {
		md := fd.Messages().ByName("UpdateBookRequest")
		msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{})
		require.NoError(t, err)
		bookMsg := msg.ProtoReflect().Get(md.Fields().ByName("book")).Message()
		name := bookMsg.Get(bookMsg.Descriptor().Fields().ByName("name")).String()
		assert.Regexp(t, shelfBookPattern, name)

		// Generating Book directly also uses the derived shelves/*/books/* template
		bookDesc := fd.Messages().ByName("Book")
		standaloneBook, err := fauxrpc.NewMessage(bookDesc, fauxrpc.GenOptions{})
		require.NoError(t, err)
		standaloneName := standaloneBook.ProtoReflect().Get(bookDesc.Fields().ByName("name")).String()
		assert.Regexp(t, shelfBookPattern, standaloneName)
	})

	t.Run("overlapping nested variable captures", func(t *testing.T) {
		md := fd.Messages().ByName("GetBookDetailsRequest")
		msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{})
		require.NoError(t, err)
		refl := msg.ProtoReflect()
		bookID := refl.Get(md.Fields().ByName("book_id")).String()
		author := refl.Get(md.Fields().ByName("author")).String()
		isbn := refl.Get(md.Fields().ByName("isbn")).Int()

		parts := strings.Split(bookID, "/")
		require.Len(t, parts, 3)
		assert.Equal(t, author, parts[0])
		assert.False(t, strings.ContainsAny(author, " /:?#"))
		assert.NotEmpty(t, parts[1])
		assert.NotZero(t, isbn)
		assert.NotEmpty(t, parts[2])
	})

	t.Run("double wildcard multi-segment path", func(t *testing.T) {
		md := fd.Messages().ByName("GetFileRequest")
		msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{})
		require.NoError(t, err)
		p := msg.ProtoReflect().Get(md.Fields().ByName("path")).String()
		parts := strings.Split(p, "/")
		assert.GreaterOrEqual(t, len(parts), 2)
		for _, part := range parts {
			assert.NotEmpty(t, part)
		}
	})

	t.Run("escaped literals in multi-segment vs single-segment capture", func(t *testing.T) {
		mdMulti := fd.Messages().ByName("GetEscapedMultiRequest")
		msgMulti, err := fauxrpc.NewMessage(mdMulti, fauxrpc.GenOptions{})
		require.NoError(t, err)
		nameMulti := msgMulti.ProtoReflect().Get(mdMulti.Fields().ByName("name")).String()
		assert.True(t, strings.HasPrefix(nameMulti, "hello%2Fworld/"), "got %q", nameMulti)

		mdSingle := fd.Messages().ByName("GetEscapedSingleRequest")
		msgSingle, err := fauxrpc.NewMessage(mdSingle, fauxrpc.GenOptions{})
		require.NoError(t, err)
		nameSingle := msgSingle.ProtoReflect().Get(mdSingle.Fields().ByName("name")).String()
		assert.Equal(t, "hello/world", nameSingle)
	})

	t.Run("custom HTTP pattern", func(t *testing.T) {
		md := fd.Messages().ByName("LockResourceRequest")
		msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{})
		require.NoError(t, err)
		res := msg.ProtoReflect().Get(md.Fields().ByName("resource")).String()
		assert.Regexp(t, regexp.MustCompile(`^projects/[a-zA-Z0-9._~-]+/locks/[a-zA-Z0-9._~-]+$`), res)
	})
}
