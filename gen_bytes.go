package fauxrpc

import (
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func generateBytesSimple() []byte {
	return []byte(gofakeit.HipsterSentence(3))
}

func GenerateBytes(fd protoreflect.FieldDescriptor) []byte {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return generateBytesSimple()
	}
	rules := constraints.GetBytes()
	if rules == nil {
		return generateBytesSimple()
	}

	if rules.Const != nil {
		return rules.Const
	}
	minLen, maxLen := uint64(0), uint64(20)
	if rules.Len != nil {
		minLen = *rules.Len
		maxLen = *rules.Len
	}
	if rules.MinLen != nil {
		minLen = *rules.MinLen
	}
	if rules.MaxLen != nil {
		maxLen = *rules.MaxLen
	}
	if rules.Pattern != nil {
		return []byte(gofakeit.Regex(*rules.Pattern))
	}

	if len(rules.In) > 0 {
		return rules.In[gofakeit.IntRange(0, len(rules.In)-1)]
	}

	return []byte(gofakeit.Sentence(int(maxLen / uint64(4)))[minLen:maxLen])
}
