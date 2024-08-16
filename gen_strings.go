package fauxrpc

import (
	"strconv"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
)

type StringHints struct {
	Rules     *validate.StringRules
	FirstName bool
	LastName  bool
	Name      bool
	UUID      bool
	URL       bool
}

func GenerateString(faker *gofakeit.Faker, hints StringHints) string {
	if hints.Rules == nil {
		switch {
		case hints.FirstName:
			return faker.FirstName()
		case hints.LastName:
			return faker.LastName()
		case hints.Name:
			return faker.Name()
		case hints.UUID:
			return faker.UUID()
		case hints.URL:
			return faker.URL()
		}
		return faker.Word()
	}

	if hints.Rules.Const != nil {
		return *hints.Rules.Const
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
	if hints.Rules.MinBytes != nil {
		maxLen = *hints.Rules.MinBytes
	}
	if hints.Rules.MaxBytes != nil {
		maxLen = *hints.Rules.MaxBytes
	}
	if hints.Rules.Pattern != nil {
		return faker.Regex(*hints.Rules.Pattern)
	}

	if len(hints.Rules.In) > 0 {
		return faker.RandomString(hints.Rules.In)
	}

	if hints.Rules.WellKnown != nil {
		switch hints.Rules.WellKnown.(type) {
		case *validate.StringRules_Email:
			return faker.Email()
		case *validate.StringRules_Hostname:
			return strings.ToLower(faker.JobDescriptor())
		case *validate.StringRules_Ip:
			return faker.IPv4Address()
		case *validate.StringRules_Ipv4:
			return faker.IPv4Address()
		case *validate.StringRules_Ipv6:
			return faker.IPv6Address()
		case *validate.StringRules_Uri:
			return faker.URL()
		case *validate.StringRules_Address:
			return faker.DomainName()
		case *validate.StringRules_Uuid:
			return faker.UUID()
		case *validate.StringRules_Tuuid:
			return strings.ReplaceAll(faker.UUID(), "-", "")
		case *validate.StringRules_IpWithPrefixlen:
			return faker.IPv4Address() + "/30"
		case *validate.StringRules_Ipv4WithPrefixlen:
			return faker.IPv4Address() + "/30"
		case *validate.StringRules_Ipv6Prefix:
			return faker.IPv6Address() + "/64"
		case *validate.StringRules_HostAndPort:
			return strings.ToLower(faker.JobDescriptor()) + ":" + strconv.FormatInt(int64(faker.IntRange(443, 9000)), 10)
		case *validate.StringRules_WellKnownRegex:
		}
	}

	return faker.Sentence(int(maxLen / uint64(4)))[minLen:maxLen]
}
