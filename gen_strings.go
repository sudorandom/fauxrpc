package fauxrpc

import (
	"math"
	"strconv"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const maxPatternRetryAttempts = 5

func randInt64GeometricDist(p float64, opts GenOptions) int64 {
	return int64(math.Floor(math.Log(opts.fake().Float64()) / math.Log(1.0-p)))
}

func stringByHeuristics(fd protoreflect.FieldDescriptor, opts GenOptions) (string, bool) {
	lowerName := strings.ToLower(string(fd.Name()))
	switch {
	case strings.Contains(lowerName, "firstname"):
		return opts.fake().FirstName(), true
	case strings.Contains(lowerName, "lastname"):
		return opts.fake().LastName(), true
	case strings.Contains(lowerName, "name"):
		return opts.fake().FirstName(), true
	case strings.Contains(lowerName, "fullname"):
		return opts.fake().Name(), true
	case strings.Contains(lowerName, "id"):
		return opts.fake().UUID(), true
	case strings.Contains(lowerName, "token"):
		return opts.fake().UUID(), true
	case strings.Contains(lowerName, "photo") && strings.Contains(lowerName, "url"):
		return "https://picsum.photos/400", true
	case strings.Contains(lowerName, "url"):
		return opts.fake().URL(), true
	case strings.Contains(lowerName, "version"):
		return opts.fake().AppVersion(), true
	case strings.Contains(lowerName, "status"):
		return opts.fake().RandomString([]string{"active", "inactive", "hidden", "archived", "deleted", "pending"}), true
	}
	return "", false
}

func stringSimple(fd protoreflect.FieldDescriptor, opts GenOptions) string {
	if s, ok := stringByHeuristics(fd, opts); ok {
		return s
	}

	return opts.fake().Sentence(int(randInt64GeometricDist(0.5, opts) + 1))
}

// String returns a fake string value given a field descriptor.
func String(fd protoreflect.FieldDescriptor, opts GenOptions) string {
	constraints := getFieldConstraints(fd, opts)
	if constraints == nil {
		return stringSimple(fd, opts)
	}
	var rules *validate.StringRules

	// Check if the constraints are for a repeated field's items
	if repeatedRules := constraints.GetRepeated(); repeatedRules != nil && repeatedRules.GetItems() != nil {
		// If it's a repeated field's items, get the string rules from the items
		rules = repeatedRules.GetItems().GetString()
	} else {
		// Otherwise, get the string rules from the top-level constraints
		rules = constraints.GetString()
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
	if len(rules.In) > 0 {
		return opts.fake().RandomString(rules.In)
	}

	minLen, maxLen := uint64(0), uint64(255)
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
		minLen = *rules.MinBytes
	}
	if rules.MaxBytes != nil {
		maxLen = *rules.MaxBytes
	}

	var generatedString string

	if rules.Pattern != nil {
		for range maxPatternRetryAttempts {
			generatedString = opts.fake().Regex(*rules.Pattern)
			if uint64(len(generatedString)) >= minLen && uint64(len(generatedString)) <= maxLen {
				break
			}
		}
	} else if rules.WellKnown != nil {
		switch rules.WellKnown.(type) {
		case *validate.StringRules_Email:
			generatedString = opts.fake().Email()
		case *validate.StringRules_Hostname:
			generatedString = strings.ToLower(opts.fake().JobDescriptor())
		case *validate.StringRules_Ip:
			generatedString = opts.fake().IPv4Address()
		case *validate.StringRules_Ipv4:
			generatedString = opts.fake().IPv4Address()
		case *validate.StringRules_Ipv6:
			generatedString = opts.fake().IPv6Address()
		case *validate.StringRules_Uri:
			generatedString = opts.fake().URL()
		case *validate.StringRules_Address:
			generatedString = opts.fake().DomainName()
		case *validate.StringRules_Uuid:
			return opts.fake().UUID()
		case *validate.StringRules_Tuuid:
			return strings.ReplaceAll(opts.fake().UUID(), "-", "")
		case *validate.StringRules_IpWithPrefixlen:
			generatedString = opts.fake().IPv4Address() + "/30"
		case *validate.StringRules_Ipv4WithPrefixlen:
			generatedString = opts.fake().IPv4Address() + "/30"
		case *validate.StringRules_Ipv6Prefix:
			generatedString = opts.fake().IPv6Address() + "/64"
		case *validate.StringRules_HostAndPort:
			generatedString = strings.ToLower(opts.fake().JobDescriptor()) + ":" + strconv.FormatInt(int64(opts.fake().IntRange(443, 9000)), 10)
		default:
			// If no specific well-known type is matched, or if WellKnownRegex is encountered,
			// fall back to heuristics or hipster text.
			if s, ok := stringByHeuristics(fd, opts); ok {
				generatedString = s
			} else {
				return generateHipsterText(minLen, maxLen, opts)
			}
		}
	} else if s, ok := stringByHeuristics(fd, opts); ok {
		generatedString = s
	} else {
		return generateHipsterText(minLen, maxLen, opts)
	}

	if uint64(len(generatedString)) < minLen {
		needed := int(minLen - uint64(len(generatedString)))
		padding := opts.fake().LetterN(uint(needed))
		if rules.Pattern != nil {
			if strings.ContainsAny(*rules.Pattern, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
				padding = strings.ToUpper(padding)
			}
			if strings.ContainsAny(*rules.Pattern, "abcdefghijklmnopqrstuvwxyz") {
				padding = strings.ToLower(padding)
			}
		}
		// Insert padding instead of appending to preserve start/end characters
		if len(generatedString) >= 2 && rules.Pattern != nil {
			middleIndex := len(generatedString) - 1
			generatedString = generatedString[:middleIndex] + padding + generatedString[middleIndex:]
		} else {
			generatedString += padding
		}
	}
	if uint64(len(generatedString)) > maxLen {
		generatedString = generatedString[:maxLen]
	}
	return generatedString
}

func generateHipsterText(minLen, maxLen uint64, opts GenOptions) string {
	b := &strings.Builder{}
	addMoreText := func() {
		b.WriteString(opts.fake().Sentence(int(randInt64GeometricDist(0.5, opts) + 1)))
	}
	addMoreText()
	for uint64(b.Len()) < minLen {
		addMoreText()
	}
	if uint64(b.Len()) > maxLen {
		return b.String()[:maxLen]
	}
	return b.String()
}
