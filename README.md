![](<assets/logo-wide.jpg>)

# FauxRPC
[![Go](https://github.com/sudorandom/fauxrpc/actions/workflows/go.yml/badge.svg)](https://github.com/sudorandom/fauxrpc/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/sudorandom/fauxrpc)](https://goreportcard.com/report/github.com/sudorandom/fauxrpc) [![Go Reference](https://pkg.go.dev/badge/github.com/sudorandom/fauxrpc.svg)](https://pkg.go.dev/github.com/sudorandom/fauxrpc)

[FauxRPC](https://fauxrpc.com) accelerates development and testing by generating fake implementations directly from Protobuf and OpenAPI schemas. A single server can expose gRPC, gRPC-Web, Connect, and HTTP/REST services without writing an implementation.

## Why FauxRPC?

* **Faster Development & Testing:** Work independently without relying on fully functional backend services.
* **Isolation & Control:** Test frontend components in isolation with controlled fake data.
* **Multi-Protocol Support:** Supports multiple protocols (gRPC, gRPC-Web, Connect, and REST).
* **OpenAPI Support:** Serve OpenAPI operations, validate requests, generate schema-shaped bodies and headers, and browse interactive API documentation.
* **Prototyping & Demos:** Create prototypes and demos quickly without building the full backend. Fake it till you make it.
* **API Stubs:** Define static or dynamic API responses with powerful stubs featuring [CEL expressions](https://cel.dev/) for precise behavior control. Stubs can be defined using config files or dynamically at runtime.
* **Improved Collaboration:** Bridge the gap between frontend and backend teams.
* **Plays well with others:** Generated data follows OpenAPI schema constraints and tries to follow any [protovalidate](https://github.com/bufbuild/protovalidate) constraints defined in Protobuf schemas.
* **Request Validation:** Validate HTTP requests against OpenAPI operations and RPC messages with [protovalidate](https://github.com/bufbuild/protovalidate).

See the [documentation website](https://fauxrpc.com) for more!

## Get Started

### Install via source

```shell
go install github.com/sudorandom/fauxrpc/cmd/fauxrpc@v0.23.0
```

### Pre-built binaries
Binaries are built for several platforms for each release. See the latest ones on [the releases page](https://github.com/sudorandom/fauxrpc/releases/latest).

---

## Usage

### Running the Server

The core command is `fauxrpc run`, which starts the server based on your Protobuf or OpenAPI schema. You can combine flags to configure the server on startup.

For example, this command starts the server with a specific schema, loads a stub for a method, and enables the dashboard:

```shell
fauxrpc run --schema=buf.build/connectrpc/eliza --stubs=example/stubs.eliza --dashboard
```

### Loading Schemas

You must provide schemas so FauxRPC knows which services to fake. Protobuf descriptors and OpenAPI specifications use the same `--schema` option, and you can mix and match sources.

#### From a local file

```shell
fauxrpc run --schema=service.binpb
```

#### From the Buf Schema Registry (BSR)

```shell
fauxrpc run --schema=buf.build/bufbuild/eliza
```

#### From an OpenAPI specification

```shell
fauxrpc run --schema=openapi.yaml
```

OpenAPI specifications can be YAML or JSON files, URLs, or directories containing specifications. See [OpenAPI Support](#openapi-support) for a complete example.

#### From multiple sources at once

```shell
fauxrpc run --schema=service.binpb --schema=openapi.yaml
```

## OpenAPI Support

FauxRPC can serve OpenAPI operations alongside Protobuf services. It detects OpenAPI YAML and JSON documents passed through the same `--schema` option used for Protobuf descriptors.

For an operation without a matching stub, FauxRPC:

* Matches the HTTP method, server base path, and templated path.
* Validates path parameters, query parameters, headers, and request bodies against the operation.
* Selects a successful response, preferring `200`, then `201`, then the default or first declared response.
* Generates response bodies from schemas, including examples, defaults, enums, constraints, formats, arrays, objects, `allOf`, `oneOf`, `anyOf`, and recursive references.
* Generates response headers declared by the selected OpenAPI response.

Explicit OpenAPI examples and defaults are preserved. Other generated values vary between requests by default. Use `--static-seed` to make unstubbed OpenAPI and Protobuf responses deterministic:

```shell
fauxrpc run --schema=openapi.yaml --static-seed
```

### Swagger Petstore example

The repository includes a locally runnable adaptation of the canonical [Swagger Petstore OpenAPI 3.0 specification](https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml), plus several conditional stubs:

```shell
fauxrpc run \
  --schema=example/swagger-petstore-openapi.yaml \
  --stubs=example/stubs.swagger-petstore.yaml
```

Try the path-parameter and query-parameter stubs:

```shell
curl http://127.0.0.1:6660/api/v3/pet/42
curl "http://127.0.0.1:6660/api/v3/pet/findByStatus?status=available"
```

The interactive OpenAPI documentation is available at [http://127.0.0.1:6660/fauxrpc/openapi-docs/](http://127.0.0.1:6660/fauxrpc/openapi-docs/). The served document points its primary server URL at FauxRPC while preserving the schema's base path.

### OpenAPI stubs

OpenAPI stubs can target an `operationId` or an HTTP path and method. Matches can be narrowed using path parameters, query parameters, headers, or a truthy [GJSON path](https://github.com/tidwall/gjson) into the request body.

```yaml
stubs:
  - name: The answer to pets, life, and everything
    target:
      operationId: getPetById
    match:
      pathParams:
        petId: "42"
    response:
      status: 200
      headers:
        X-Pet-Mood: existential
      body:
        id: 42
        name: Deep Thought
        photoUrls:
          - https://fauxrpc.local/pets/deep-thought.jpg
        status: available
```

Load the file with `--stubs`:

```shell
fauxrpc run --schema=openapi.yaml --stubs=openapi-stubs.yaml
```

## Using Stubs

While FauxRPC generates fake data by default, **stubs** let you define specific, predictable responses for RPC and OpenAPI operations. This is useful for testing particular scenarios.

You can load a single stub file or an entire directory of them.

Add `--only-stubs` to disable generated fallback responses. An OpenAPI operation without a matching stub returns HTTP `501 Not Implemented`; Protobuf RPCs return an empty response message.

#### Load a single stub file

```shell
fauxrpc run --schema=eliza.binpb --stubs=example/stubs.eliza/say.json
```

#### Load all stubs from a directory

```shell
fauxrpc run --schema=eliza.binpb --stubs=example/stubs.eliza/
```

## Proxying and Ingesting Real Traffic

FauxRPC can act as an intercepting proxy to ingest real gRPC/Connect traffic and automatically generate stateful, predictable mock profiles (stubs) on disk.

When in proxy mode, FauxRPC forwards incoming requests to an upstream server, captures the request and response payloads, translates them into the stub format, and writes/appends them to files under a structured directory by service/method (e.g., `<record-dir>/<service>/<method>.json`).

To run FauxRPC in proxy mode:

```shell
fauxrpc run --proxy-to=127.0.0.1:8080 --record-dir=stubs/
```

* `--proxy-to`: The address of the upstream gRPC or Connect server to forward requests to.
* `--record-dir`: The directory path where the recorded stubs should be saved, structured by service and method.

### Unimplemented Fallback

If the upstream server returns an `UNIMPLEMENTED` status code (indicating that the endpoint is not yet implemented), FauxRPC will automatically catch the error and fall back to serving a mock response (from stubs or random fake generation). This allows frontend and backend teams to co-develop APIs incrementally.

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
*   📁 **Schema Browser:** Explore Protobuf schemas loaded into the server.
*   🔌 **Stubs:** Manage and view details of registered stubs.
*   📚 **API Documentation:** Access generated Protobuf documentation and interactive OpenAPI documentation.

![](<assets/dashboard-event-log.gif>)

Go to [the documentation website](https://fauxrpc.com) for more!
