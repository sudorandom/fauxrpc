![](<assets/logo-wide.jpg>)

# FauxRPC
[![Go](https://github.com/sudorandom/fauxrpc/actions/workflows/go.yml/badge.svg)](https://github.com/sudorandom/fauxrpc/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/sudorandom/fauxrpc)](https://goreportcard.com/report/github.com/sudorandom/fauxrpc) [![Go Reference](https://pkg.go.dev/badge/github.com/sudorandom/fauxrpc.svg)](https://pkg.go.dev/github.com/sudorandom/fauxrpc)

FauxRPC is a powerful tool that empowers you to accelerate development and testing by effortlessly generating fake implementations of gRPC, gRPC-Web, Connect, and REST services. If you have a protobuf-based workflow, this tool could help.

## Why FauxRPC?
* **Faster Development & Testing:** Work independently without relying on fully functional backend services.
* **Isolation & Control:** Test frontend components in isolation with controlled fake data.
* **Multi-Protocol Support:** Supports multiple protocols (gRPC, gRPC-Web, Connect, and REST).
* **Prototyping & Demos:** Create prototypes and demos quickly without building the full backend. Fake it till you make it.
* **Improved Collaboration:** Bridge the gap between frontend and backend teams.
* **Plays well with others:** Test data from FauxRPC will try to automatically follow any [protovalidate](https://github.com/bufbuild/protovalidate) constraints that are defined.

## How it Works
FauxRPC leverages your Protobuf definitions to generate fake services that mimic the behavior of real ones. You can easily configure the fake data returned, allowing you to simulate various scenarios and edge cases. It takes in `*.proto` files or protobuf descriptors (in binpb, json, txtpb, yaml formats), then it automatically starts up a server that can speak gRPC/gRPC-Web/Connect and REST (as long as there are `google.api.http` annotations defined). Descriptors contain all of the information found in a set of `.proto` files. You can generate them with `protoc` or the `buf build` command.

![](<assets/diagram.svg>)

## Get Started

### Install via source
```
go install github.com/sudorandom/fauxrpc/cmd/fauxrpc@latest
```

### Pre-built binaries
Binaries are built for several platforms for each release. See the latest ones on [the releases page](https://github.com/sudorandom/fauxrpc/releases/latest).

### Using Descriptors
Make an `example.proto` file (or use a file that already exists):
```protobuf
syntax = "proto3";

package greet.v1;

message GreetRequest {
  string name = 1;
}

message GreetResponse {
  string greeting = 1;
}

service GreetService {
  rpc Greet(GreetRequest) returns (GreetResponse) {}
}
```

Create a descriptors file and use it to start the FauxRPC server:
```shell
$ buf build ./example.proto -o ./example.binpb
$ fauxrpc run --schema=./example.binpb
2024/08/17 08:01:19 INFO Listening on http://127.0.0.1:6660
2024/08/17 08:01:19 INFO See available methods: buf curl --http2-prior-knowledge http://127.0.0.1:6660 --list-methods
```
Done! It's that easy. Now you can call the service with any tooling that supports gRPC, gRPC-Web, or connect. So [buf curl](https://buf.build/docs/reference/cli/buf/curl), [grpcurl](https://github.com/fullstorydev/grpcurl), [Postman](https://www.postman.com/), [Insomnia](https://insomnia.rest/) all work fine!

```shell
$ buf curl --http2-prior-knowledge http://127.0.0.1:6660/greet.v1.GreetService/Greet
{
  "greeting":  "3 wolf moon fashion axe."
}
```

### Using Server Reflection
If there's an existing gRPC service running that you want to emulate, you can use server reflection to start the FauxRPC service:
```shell
$ fauxrpc run --schema=https://demo.connectrpc.com
```

### From BSR (Buf Schema Registry)
Buf has a [schema registry](https://buf.build/product/bsr) where many schemas are hosted. Here's how to use FauxRPC using images from the registry.

```shell
$ buf build buf.build/bufbuild/registry -o bufbuild.registry.json
$ fauxrpc run --schema=./bufbuild.registry.json
```

This will start a fake version of the BSR API by downloading descriptors for [bufbuild/registry](https://buf.build/bufbuild/registry) from the BSR and using them with FauxRPC. Very meta.

### Multiple Sources
You can define this `--schema` option as many times as you want. That means you can add services from multiple descriptors and even mix and match from descriptors and from server reflection:
```shell
$ fauxrpc run --schema=https://demo.connectrpc.com --schema=./example.binpb
```

## Multi-protocol Support
The multi-protocol support [is based on ConnectRPC](https://connectrpc.com/docs/multi-protocol/). So with FauxRPC, you get **gRPC, gRPC-Web and Connect** out of the box. However, FauxRPC does one thing more. It allows you to use [`google.api.http` annotations](https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/adding_annotations/) to present a JSON/HTTP API, so you can gRPC and REST together! This is normally done with [an additional service](https://github.com/grpc-ecosystem/grpc-gateway) that runs in-between the outside world and your actual gRPC service but with FauxRPC you get the so-called transcoding from HTTP/JSON to gRPC all in the same package. Here's a concrete example:

```protobuf
syntax = "proto3";

package http.service;

import "google/api/annotations.proto";

service HTTPService {
  rpc GetMessage(GetMessageRequest) returns (Message) {
    option (google.api.http) = {get: "/v1/{name=messages/*}"};
  }
}
message GetMessageRequest {
  string name = 1; // Mapped to URL path.
}
message Message {
  string text = 1; // The resource content.
}
```

Again, we start the service by building the descriptors and using
```
$ buf build ./httpservice.proto -o ./httpservice.binpb
$ fauxrpc run --schema=httpservice.binpb
```

Now that we have the server running we can test this with the "normal" curl:
```shell
$ curl http://127.0.0.1:6660/v1/messages/123456
{"text":"Retro."}⏎
```
Sweet. You can now easily support REST alongside gRPC. If you are wondering how to do this with "real" services, look into [vangaurd-go](https://github.com/connectrpc/vanguard-go). This library is doing the real heavy lifting.

## What does the fake data look like?
You might be wondering what actual responses look like. FauxRPC's fake data generation is continually improving so these details might change as time goes on. It uses a library called [fakeit](https://github.com/brianvoe/gofakeit) to generate fake data. Because protobufs have pretty well-defined types, we can easily generate data that technically matches the types. This works well for most use cases, but FauxRPC tries to be a little bit better. If you annotate your protobuf files with [protovalidate](https://github.com/bufbuild/protovalidate) constraints, FauxRPC will try its best to generate data that matches these constraints. Let's look at some examples!

```protobuf
syntax = "proto3";

package greet.v1;

message GreetRequest {
  string name = 1;
}

message GreetResponse {
  string greeting = 1;
}

service GreetService {
  rpc Greet(GreetRequest) returns (GreetResponse) {}
}
```

With FauxRPC, you will get any kind of word, so it might look like this:
```json
{
  "greeting": "Poutine."
}
```
This is fine, but for the RPC, we know a bit more about the type being returned. We know that it sends a greeting back that looks like "Hello, [name]". So here's what the same protobuf file might look like with protovalidate constraints:


Now let's see what this looks like with protovalidate constraints:
```protobuf
syntax = "proto3";

import "buf/validate/validate.proto";

package greet.v1;

message GreetRequest {
  string name = 1 [(buf.validate.field).string = {min_len: 3, max_len: 100}];
}

message GreetResponse {
  string greeting = 1 [(buf.validate.field).string.pattern = "^Hello, [a-zA-Z]+$"];
}

service GreetService {
  rpc Greet(GreetRequest) returns (GreetResponse) {}
}
```

With this new protobuf file, this is what FauxRPC might output now:

```json
{
  "greeting": "Hello, TWXxF"
}
```
In essence, protovalidate constraints enable FauxRPC to generate more realistic and contextually relevant fake data, aligning it closer to the expected behavior of your actual services.

## Status: Alpha
This project is just starting out. I plan to add a lot of things that make this tool actually usable in more situations.

- Service for adding/updating/removing stub responses with a CLI to add/remove/replace these stubs
- Configuration file
- BSR Support (this is a 'maybe' because using `buf build` to emit descriptors works well enough IMO)
- Better streaming support. FauxRPC does work with streaming calls but it only returns a single response
