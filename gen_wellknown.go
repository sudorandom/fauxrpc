package fauxrpc

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func genGoogleDuration() *protoreflect.Value {
	duration := time.Duration(gofakeit.Uint64() % uint64(30*time.Hour*24))
	v := protoreflect.ValueOf(durationpb.New(duration).ProtoReflect())
	return &v
}

func genGoogleTimestamp() *protoreflect.Value {
	v := protoreflect.ValueOf(timestamppb.New(gofakeit.Date()).ProtoReflect())
	return &v
}

func genGoogleValue() *protoreflect.Value {
	scalarOptions := []func() *structpb.Value{
		func() *structpb.Value { return structpb.NewNullValue() },
		func() *structpb.Value { return structpb.NewBoolValue(gofakeit.Bool()) },
		func() *structpb.Value { return structpb.NewNumberValue(gofakeit.Float64()) },
		func() *structpb.Value { return structpb.NewStringValue(gofakeit.SentenceSimple()) },
	}
	msgOptions := []func() *structpb.Value{
		// TODO: structpb.NewList()
		// TODO: structpb.NewStruct()
	}
	options := append(scalarOptions, msgOptions...)
	fn := options[gofakeit.IntRange(0, len(options)-1)]
	v := protoreflect.ValueOf(fn().ProtoReflect())
	return &v
}
