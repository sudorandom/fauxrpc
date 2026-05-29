package fauxrpc

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

func generateBytesSimple(opts GenOptions) []byte {
	return []byte(opts.fake().HipsterSentence())
}

// Bytes returns a fake []byte value given a field descriptor.
func Bytes(fd protoreflect.FieldDescriptor, opts GenOptions) []byte {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return generateBytesSimple(opts)
	}
	rules := constraints.GetBytes()
	if rules == nil {
		return generateBytesSimple(opts)
	}

	if rules.Const != nil {
		return rules.Const
	}
	if len(rules.Example) > 0 {
		return rules.Example[opts.fake().IntRange(0, len(rules.Example)-1)]
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
		return []byte(opts.fake().Regex(*rules.Pattern))
	}

	if len(rules.In) > 0 {
		return rules.In[opts.fake().IntRange(0, len(rules.In)-1)]
	}

	return []byte(opts.fake().Sentence(int(maxLen / uint64(4)))[minLen:maxLen])
}
