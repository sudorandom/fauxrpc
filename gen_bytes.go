package fauxrpc

import (
	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
)

type BytesHints struct {
	Rules     *validate.BytesRules
	FirstName bool
	LastName  bool
	Name      bool
	UUID      bool
	URL       bool
	Version   bool
}

func GenerateBytes(faker *gofakeit.Faker, hints BytesHints) []byte {
	if hints.Rules == nil {
		return []byte(faker.HipsterSentence(3))
	}

	if hints.Rules.Const != nil {
		return hints.Rules.Const
	}
	minLen, maxLen := uint64(0), uint64(20)
	if hints.Rules.Len != nil {
		minLen = *hints.Rules.Len
		maxLen = *hints.Rules.Len
	}
	if hints.Rules.MinLen != nil {
		minLen = *hints.Rules.MinLen
	}
	if hints.Rules.MaxLen != nil {
		maxLen = *hints.Rules.MaxLen
	}
	if hints.Rules.Pattern != nil {
		return []byte(faker.Regex(*hints.Rules.Pattern))
	}

	if len(hints.Rules.In) > 0 {
		return hints.Rules.In[faker.IntRange(0, len(hints.Rules.In)-1)]
	}

	return []byte(faker.Sentence(int(maxLen / uint64(4)))[minLen:maxLen])
}
