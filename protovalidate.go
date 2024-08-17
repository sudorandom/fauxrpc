package fauxrpc

import (
	"sync"

	"github.com/bufbuild/protovalidate-go/resolver"
)

var makeResolverOnce sync.Once
var globalResolver resolver.DefaultResolver

func getResolver() resolver.DefaultResolver {
	makeResolverOnce.Do(func() {
		globalResolver = resolver.DefaultResolver{}
	})
	return globalResolver
}
