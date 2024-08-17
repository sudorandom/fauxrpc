package fauxrpc

import (
	"math"
	"strconv"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func randInt64GeometricDist(p float64) int64 {
	return int64(math.Floor(math.Log(gofakeit.Float64()) / math.Log(1.0-p)))
}

func generateStringSimple(fd protoreflect.FieldDescriptor) string {
	lowerName := strings.ToLower(string(fd.Name()))
	switch {
	case strings.Contains(lowerName, "firstname"):
		return gofakeit.FirstName()
	case strings.Contains(lowerName, "lastname"):
		return gofakeit.LastName()
	case strings.Contains(lowerName, "name"):
		return gofakeit.Name()
	case strings.Contains(lowerName, "id"):
		return gofakeit.UUID()
	case strings.Contains(lowerName, "token"):
		return gofakeit.UUID()
	case strings.Contains(lowerName, "url"):
		return gofakeit.URL()
	case strings.Contains(lowerName, "version"):
		return gofakeit.AppVersion()
	}

	return gofakeit.HipsterSentence(int(randInt64GeometricDist(0.5) + 1))
}

func GenerateString(fd protoreflect.FieldDescriptor) string {
	constraints := getResolver().ResolveFieldConstraints(fd)
	if constraints == nil {
		return generateStringSimple(fd)
	}
	rules := constraints.GetString_()
	if rules == nil {
		return generateStringSimple(fd)
	}

	if rules == nil {
		return generateStringSimple(fd)
	}

	if rules.Const != nil {
		return *rules.Const
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
	if rules.MinBytes != nil {
		maxLen = *rules.MinBytes
	}
	if rules.MaxBytes != nil {
		maxLen = *rules.MaxBytes
	}
	if rules.Pattern != nil {
		return gofakeit.Regex(*rules.Pattern)
	}

	if len(rules.In) > 0 {
		return gofakeit.RandomString(rules.In)
	}

	if rules.WellKnown != nil {
		switch rules.WellKnown.(type) {
		case *validate.StringRules_Email:
			return gofakeit.Email()
		case *validate.StringRules_Hostname:
			return strings.ToLower(gofakeit.JobDescriptor())
		case *validate.StringRules_Ip:
			return gofakeit.IPv4Address()
		case *validate.StringRules_Ipv4:
			return gofakeit.IPv4Address()
		case *validate.StringRules_Ipv6:
			return gofakeit.IPv6Address()
		case *validate.StringRules_Uri:
			return gofakeit.URL()
		case *validate.StringRules_Address:
			return gofakeit.DomainName()
		case *validate.StringRules_Uuid:
			return gofakeit.UUID()
		case *validate.StringRules_Tuuid:
			return strings.ReplaceAll(gofakeit.UUID(), "-", "")
		case *validate.StringRules_IpWithPrefixlen:
			return gofakeit.IPv4Address() + "/30"
		case *validate.StringRules_Ipv4WithPrefixlen:
			return gofakeit.IPv4Address() + "/30"
		case *validate.StringRules_Ipv6Prefix:
			return gofakeit.IPv6Address() + "/64"
		case *validate.StringRules_HostAndPort:
			return strings.ToLower(gofakeit.JobDescriptor()) + ":" + strconv.FormatInt(int64(gofakeit.IntRange(443, 9000)), 10)
		case *validate.StringRules_WellKnownRegex:
		}
	}

	return generateHipsterText(minLen, maxLen)
}

func generateHipsterText(minLen, maxLen uint64) string {
	b := &strings.Builder{}
	addMoreText := func() {
		b.WriteString(gofakeit.HipsterSentence(int(randInt64GeometricDist(0.5) + 1)))
	}
	addMoreText()
	for uint64(b.Len()) < minLen {
		addMoreText()
	}
	if uint64(b.Len()) > maxLen {
		return b.String()[:maxLen-1]
	}
	return b.String()
}
