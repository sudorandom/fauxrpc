![](<assets/logo-wide.jpg>)

# FauxRPC
[![Go](https://github.com/sudorandom/fauxrpc/actions/workflows/go.yml/badge.svg)](https://github.com/sudorandom/fauxrpc/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/sudorandom/fauxrpc)](https://goreportcard.com/report/github.com/sudorandom/fauxrpc) [![Go Reference](https://pkg.go.dev/badge/github.com/sudorandom/fauxrpc.svg)](https://pkg.go.dev/github.com/sudorandom/fauxrpc)

[FauxRPC](https://fauxrpc.com) is a powerful tool that empowers you to accelerate development and testing by effortlessly generating fake implementations of gRPC, gRPC-Web, Connect, and REST services. If you have a protobuf-based workflow, this tool could help.

## Why FauxRPC?
* **Faster Development & Testing:** Work independently without relying on fully functional backend services.
* **Isolation & Control:** Test frontend components in isolation with controlled fake data.
* **Multi-Protocol Support:** Supports multiple protocols (gRPC, gRPC-Web, Connect, and REST).
* **Prototyping & Demos:** Create prototypes and demos quickly without building the full backend. Fake it till you make it.
* **API Stubs:** Define static or dynamic API responses with powerful stubs featuring [CEL expressions](https://cel.dev/) for precise behavior control. Stubs can be defined using config files or dynamically at runtime.
* **Improved Collaboration:** Bridge the gap between frontend and backend teams.
* **Plays well with others:** Test data from FauxRPC will try to automatically follow any [protovalidate](https://github.com/bufbuild/protovalidate) constraints that are defined.
* **Request Validation:** Ensure data integrity with automatic request validation using [protovalidate](https://github.com/bufbuild/protovalidate). Catch errors early and prevent invalid data from reaching your application logic.

See the [the documentation website](https://fauxrpc.com) for more!

## Get Started

### Install via source
```
go install github.com/sudorandom/fauxrpc/cmd/fauxrpc@v0.16.1
```

### Pre-built binaries
Binaries are built for several platforms for each release. See the latest ones on [the releases page](https://github.com/sudorandom/fauxrpc/releases/latest).

--------------

## Usage

### Running the Server

The core command is `fauxrpc run`, which starts the server based on your Protobuf schema. You can combine flags to configure the server on startup.

For example, this command starts the server with a specific schema, loads a stub for a method, and enables the dashboard:

```shell
fauxrpc run --schema=eliza.binpb --stubs=example/stubs.eliza/say.json --dashboard
```

### Loading Schemas

You must provide Protobuf descriptors so FauxRPC knows which services to fake. Schemas can be loaded from multiple sources, and you can mix and match them.

#### From a local file:

```shell
fauxrpc run --schema=service.binpb
```

#### From the Buf Schema Registry (BSR)

```shell
fauxrpc run --schema=buf.build/bufbuild/eliza
```

#### From multiple sources at once
```shell
fauxrpc run --schema=service.binpb --schema=buf.build/bufbuild/eliza
```

## Using Stubs

While FauxRPC generates random fake data by default, **stubs** let you define specific, predictable responses for your RPCs. This is great for testing specific scenarios.

You can load a single stub file or an entire directory of them.

#### Load a single stub file

```shell
fauxrpc run --schema=eliza.binpb --stubs=example/stubs.eliza/say.json
```

#### Load all stubs from a directory
```shell
fauxrpc run --schema=eliza.binpb --stubs=example/stubs.eliza/
```

## Making Requests with `fauxrpc curl`

FauxRPC includes a handy built-in client, `fauxrpc curl`, for making requests to your services without needing external tools. It automatically sources the schema to provide a seamless testing experience.

### Hit all RPCs in a service with default data

```shell
fauxrpc curl --http2-prior-knowledge --schema=buf.build/bufbuild/registry
```

#### Hit a specific RPC

```shell
fauxrpc curl --http2-prior-knowledge --schema=buf.build/bufbuild/registry buf.registry.plugin.v1beta1.LabelService/ListLabels
```

#### Using server reflection

If no `--schema` option is provided, server reflection will be used to figure out the type and service information.

```shell
fauxrpc curl --http2-prior-knowledge buf.registry.plugin.v1beta1.LabelService/ListLabels
```

## Dashboard
Enhance your FauxRPC experience with the interactive dashboard, providing real-time insights into your server's operations.

To enable the dashboard, simply start FauxRPC with the `--dashboard` option:

```
fauxrpc run --schema=service.binpb --dashboard
```

Access the dashboard in your browser at [http://127.0.0.1:6660/fauxrpc](http://127.0.0.1:6660/fauxrpc).

![](<assets/dashboard.png>)

The dashboard provides:
*   📊 **Summary:** View overall server statistics.
*   📜 **Request Log:** Live stream of all incoming requests.
*   📁 **Schema Browser:** Explore all Protobuf schemas loaded into the server.
*   🔌 **Stubs:** Manage and view details of registered stubs.
*   📚 **API Documentation:** Access auto-generated API documentation.

![](<assets/dashboard-event-log.gif>)

Go to [the documentation website](https://fauxrpc.com) for more!
