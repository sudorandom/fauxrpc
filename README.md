![](<assets/logo-wide.jpg>)

# FauxRPC
[![Go](https://github.com/sudorandom/fauxrpc/actions/workflows/go.yml/badge.svg)](https://github.com/sudorandom/fauxrpc/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/sudorandom/fauxrpc)](https://goreportcard.com/report/github.com/sudorandom/fauxrpc) [![Go Reference](https://pkg.go.dev/badge/github.com/sudorandom/fauxrpc.svg)](https://pkg.go.dev/github.com/sudorandom/fauxrpc)

[FauxRPC](https://fauxrpc.com) is a powerful tool that empowers you to accelerate development and testing by effortlessly generating fake implementations of gRPC, gRPC-Web, Connect, and REST services. If you have a protobuf-based workflow, this tool could help.

## Why FauxRPC?
* **Faster Development & Testing:** Work independently without relying on fully functional backend services.
* **Isolation & Control:** Test frontend components in isolation with controlled fake data.
* **Multi-Protocol Support:** Supports multiple protocols (gRPC, gRPC-Web, Connect, and REST).
* **Prototyping & Demos:** Create prototypes and demos quickly without building the full backend. Fake it till you make it.
* **Improved Collaboration:** Bridge the gap between frontend and backend teams.
* **Plays well with others:** Test data from FauxRPC will try to automatically follow any [protovalidate](https://github.com/bufbuild/protovalidate) constraints that are defined.
* **Request Validation:** Ensure data integrity with automatic request validation using [protovalidate](https://github.com/bufbuild/protovalidate). Catch errors early and prevent invalid data from reaching your application logic.

See the [the documentation website](https://fauxrpc.com) for more!

## Get Started

### Install via source
```
go install github.com/sudorandom/fauxrpc/cmd/fauxrpc@latest
```

### Pre-built binaries
Binaries are built for several platforms for each release. See the latest ones on [the releases page](https://github.com/sudorandom/fauxrpc/releases/latest).

## Quick Start

Pass [protobuf descriptors](https://buf.build/docs/reference/descriptors) to FauxRPC and a test server will be created, returning random fake data!

```shell
$ fauxrpc run --schema=service.binpb
```

That's... it. Now you can call it with your gRPC/gRPC-Web/Connect clients:

```shell
$ buf curl --http2-prior-knowledge http://127.0.0.1:6660/my.own.v1.service/HelloWorld
{
  "text": "Thundercats."
}
```

Go to [the documentation website](https://fauxrpc.com) for more!