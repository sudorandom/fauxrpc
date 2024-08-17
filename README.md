# FauxRPC
![](<assets/logo-wide.jpg>)

Quickly and easily set up a mock gRPC/gRPC-Web/Connect/Protobuf-powered REST API that returns random test data. No complicated code generation step, just pass the protobuf descriptors and go!

![](<assets/diagram.svg>)

## Get Started
FauxRPC is available as an open-source project.

### Install via source
```
go install github.com/sudorandom/fauxrpc/cmd/fauxrpc@latest
```

### Pre-built binaries
Binaries are built for several platforms for each release. See the latest ones on [the releases page](https://github.com/sudorandom/fauxrpc/releases/latest).

### Use Descriptors
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

### Server Reflection
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
- Testing for REST translations. I have no idea if this actually works yet
- Better streaming support. FauxRPC does work with streaming calls but it only returns a single response
