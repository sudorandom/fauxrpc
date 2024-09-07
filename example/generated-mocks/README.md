# Status: EXTREMELY Alpha

I'm still actively working on this, but the hope is to have a protoc plugin that generates a handler that sits alongside generated ConnectRPC code. This handler implementation is instrumented with [stretchr/testify's mock package](https://github.com/stretchr/testify?tab=readme-ov-file#mock-package).

- Unary Support: Working
- Server Streaming support: None
- Client Streaming support: None
- Bidirectional Streaming support: None
