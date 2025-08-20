package server

// contextKey is a private type used for context keys.
type contextKey string

const (
	clientProtocolKey contextKey = "clientProtocol"
	requestHeadersKey contextKey = "requestHeaders"
)