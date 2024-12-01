## Petstore Example
This example uses the famous "pet store" example that was original created as an example for OpenAPI.

# Build the protobuf descriptors for petstore.proto.
Thanks to buf for [this wonderful example](https://github.com/connectrpc/vanguard-go/blob/main/internal/examples/pets/internal/proto/io/swagger/petstore/v2/pets.proto)!

```shell
buf build ssh://git@github.com/connectrpc/vanguard-go.git -o petstore.binpb --path internal/examples/pets/internal/proto/io/swagger/petstore/v2
```

# Launch FauxRPC with the descriptors and stubs!
```shell
fauxrpc run --schema petstore.binpb --only-stubs --stubs=example/stubs.petstore/
```

# Use your fake service!
```shell
$ buf curl --http2-prior-knowledge -d '{"pet_id": "1"}' http://127.0.0.1:6660/io.swagger.petstore.v2.PetService/GetPetByID
```

Using the `google.api.http` annotations:
```shell
curl http://127.0.0.1:6660/pet/1
curl http://127.0.0.1:6660/pet/2
```
