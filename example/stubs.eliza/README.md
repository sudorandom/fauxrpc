# Eliza Stub Example

For this example we're using the [Eliza demo service ](https://buf.build/connectrpc/eliza), which is a simple service that emulates interactions between a patient and an obtuse psychotherapist.

```protobuf
service ElizaService {
  rpc Say(SayRequest) returns (SayResponse) {
    option idempotency_level = NO_SIDE_EFFECTS;
  }
  rpc Converse(stream ConverseRequest) returns (stream ConverseResponse) {}
  rpc Introduce(IntroduceRequest) returns (stream IntroduceResponse) {}
}
```
