package fauxrpc

import (
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *dataGenerator) genGoogleDuration() *protoreflect.Value {
	duration := time.Duration(g.faker.Rand.Uint64() % uint64(30*time.Hour*24))
	v := protoreflect.ValueOf(durationpb.New(duration).ProtoReflect())
	return &v
}

func (g *dataGenerator) genGoogleTimestamp() *protoreflect.Value {
	v := protoreflect.ValueOf(timestamppb.New(g.faker.Date()).ProtoReflect())
	return &v
}

func (g *dataGenerator) genGoogleValue() *protoreflect.Value {
	scalarOptions := []func() *structpb.Value{
		func() *structpb.Value { return structpb.NewNullValue() },
		func() *structpb.Value { return structpb.NewBoolValue(g.faker.Bool()) },
		func() *structpb.Value { return structpb.NewNumberValue(g.faker.Float64()) },
		func() *structpb.Value { return structpb.NewStringValue(g.faker.SentenceSimple()) },
	}
	msgOptions := []func() *structpb.Value{
		// TODO: structpb.NewList()
		// TODO: structpb.NewStruct()
	}
	options := append(scalarOptions, msgOptions...)
	fn := options[g.faker.IntRange(0, len(options)-1)]
	v := protoreflect.ValueOf(fn().ProtoReflect())
	return &v
}
