package fauxrpc

import (
	"math"
	"strconv"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func randInt64GeometricDist(p float64, opts GenOptions) int64 {
	return int64(math.Floor(math.Log(opts.fake().Float64()) / math.Log(1.0-p)))
}

func stringSimple(fd protoreflect.FieldDescriptor, opts GenOptions) string {
	lowerName := strings.ToLower(string(fd.Name()))
	switch {
	case strings.Contains(lowerName, "firstname"):
		return opts.fake().FirstName()
	case strings.Contains(lowerName, "lastname"):
		return opts.fake().LastName()
	case strings.Contains(lowerName, "name"):
		return opts.fake().FirstName()
	case strings.Contains(lowerName, "fullname"):
		return opts.fake().Name()
	case strings.Contains(lowerName, "id"):
		return opts.fake().UUID()
	case strings.Contains(lowerName, "token"):
		return opts.fake().UUID()
	case strings.Contains(lowerName, "photo") && strings.Contains(lowerName, "url"):
		return "https://picsum.photos/400"
	case strings.Contains(lowerName, "url"):
		return opts.fake().URL()
	case strings.Contains(lowerName, "version"):
		return opts.fake().AppVersion()
	case strings.Contains(lowerName, "status"):
		return opts.fake().RandomString([]string{"active", "inactive", "hidden", "archived", "deleted", "pending"})
	}

	return opts.fake().HipsterSentence(int(randInt64GeometricDist(0.5, opts) + 1))
}

// String returns a fake string value given a field descriptor.
func String(fd protoreflect.FieldDescriptor, opts GenOptions) string {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return stringSimple(fd, opts)
	}
	rules := constraints.GetString_()
	if rules == nil {
		return stringSimple(fd, opts)
	}

	if rules == nil {
		return stringSimple(fd, opts)
	}

	if rules.Const != nil {
		return *rules.Const
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
	if rules.MinBytes != nil {
		maxLen = *rules.MinBytes
	}
	if rules.MaxBytes != nil {
		maxLen = *rules.MaxBytes
	}
	if rules.Pattern != nil {
		return opts.fake().Regex(*rules.Pattern)
	}

	if len(rules.In) > 0 {
		return opts.fake().RandomString(rules.In)
	}

	if rules.WellKnown != nil {
		switch rules.WellKnown.(type) {
		case *validate.StringRules_Email:
			return opts.fake().Email()
		case *validate.StringRules_Hostname:
			return strings.ToLower(opts.fake().JobDescriptor())
		case *validate.StringRules_Ip:
			return opts.fake().IPv4Address()
		case *validate.StringRules_Ipv4:
			return opts.fake().IPv4Address()
		case *validate.StringRules_Ipv6:
			return opts.fake().IPv6Address()
		case *validate.StringRules_Uri:
			return opts.fake().URL()
		case *validate.StringRules_Address:
			return opts.fake().DomainName()
		case *validate.StringRules_Uuid:
			return opts.fake().UUID()
		case *validate.StringRules_Tuuid:
			return strings.ReplaceAll(opts.fake().UUID(), "-", "")
		case *validate.StringRules_IpWithPrefixlen:
			return opts.fake().IPv4Address() + "/30"
		case *validate.StringRules_Ipv4WithPrefixlen:
			return opts.fake().IPv4Address() + "/30"
		case *validate.StringRules_Ipv6Prefix:
			return opts.fake().IPv6Address() + "/64"
		case *validate.StringRules_HostAndPort:
			return strings.ToLower(opts.fake().JobDescriptor()) + ":" + strconv.FormatInt(int64(opts.fake().IntRange(443, 9000)), 10)
		case *validate.StringRules_WellKnownRegex:
		}
	}

	return generateHipsterText(minLen, maxLen, opts)
}

func generateHipsterText(minLen, maxLen uint64, opts GenOptions) string {
	b := &strings.Builder{}
	addMoreText := func() {
		b.WriteString(opts.fake().HipsterSentence(int(randInt64GeometricDist(0.5, opts)) + 1))
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
