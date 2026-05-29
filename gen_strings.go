package fauxrpc

import (
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const maxPatternRetryAttempts = 5

func randInt64GeometricDist(p float64, opts GenOptions) int64 {
	return int64(math.Floor(math.Log(opts.fake().Float64()) / math.Log(1.0-p)))
}

type heuristicFunc func(GenOptions) string

var (
	heuristicCache sync.Map
	statusValues   = []string{"active", "inactive", "hidden", "archived", "deleted", "pending"}
)

func stringByHeuristics(fd protoreflect.FieldDescriptor, opts GenOptions) (string, bool) {
	if v, ok := heuristicCache.Load(fd); ok {
		if v == nil {
			return "", false
		}
		f := v.(heuristicFunc)
		if f == nil {
			return "", false
		}
		return f(opts), true
	}

	lowerName := strings.ToLower(string(fd.Name()))
	var f heuristicFunc
	switch {
	case isTagOrAttributesKey(fd):
		f = func(opts GenOptions) string { return strings.ToLower(opts.fake().Word()) }
	case strings.Contains(lowerName, "id"):
		f = func(opts GenOptions) string { return opts.fake().UUID() }
	case strings.Contains(lowerName, "token"):
		f = func(opts GenOptions) string { return opts.fake().UUID() }
	case strings.Contains(lowerName, "email") || strings.Contains(lowerName, "mail"):
		f = func(opts GenOptions) string { return strings.ToLower(opts.fake().Email()) }
	case isPhoneField(lowerName):
		f = func(opts GenOptions) string { return opts.fake().Phone() }
	case isIPField(lowerName):
		f = func(opts GenOptions) string { return opts.fake().IPv4Address() }
	case isMacField(lowerName):
		f = func(opts GenOptions) string { return opts.fake().MacAddress() }
	case strings.Contains(lowerName, "user_agent") || lowerName == "ua":
		f = func(opts GenOptions) string { return opts.fake().UserAgent() }
	case strings.Contains(lowerName, "color") || strings.Contains(lowerName, "hex"):
		f = func(opts GenOptions) string { return opts.fake().HexColor() }
	case isTimeField(lowerName):
		f = func(opts GenOptions) string { return opts.fake().Date().Format(time.RFC3339) }
	case lowerName == "company" || strings.Contains(lowerName, "company_name") || lowerName == "organization" || lowerName == "org":
		f = func(opts GenOptions) string { return opts.fake().Company() }
	case lowerName == "job_title" || lowerName == "title" || strings.Contains(lowerName, "job_title"):
		f = func(opts GenOptions) string { return opts.fake().JobTitle() }
	case lowerName == "currency" || strings.Contains(lowerName, "currency_code"):
		f = func(opts GenOptions) string { return opts.fake().CurrencyShort() }
	case lowerName == "language" || lowerName == "lang" || strings.HasSuffix(lowerName, "_lang"):
		f = func(opts GenOptions) string { return opts.fake().LanguageAbbreviation() }
	case lowerName == "locale" || lowerName == "bcp47" || strings.HasSuffix(lowerName, "_locale"):
		f = func(opts GenOptions) string { return opts.fake().LanguageBCP() }
	case strings.Contains(lowerName, "address"):
		f = func(opts GenOptions) string { return opts.fake().Address().Address }
	case strings.Contains(lowerName, "street"):
		f = func(opts GenOptions) string { return opts.fake().Address().Street }
	case isCityField(lowerName):
		f = func(opts GenOptions) string { return opts.fake().City() }
	case strings.Contains(lowerName, "country"):
		f = func(opts GenOptions) string { return opts.fake().Country() }
	case strings.Contains(lowerName, "zip") || strings.Contains(lowerName, "postcode") || strings.Contains(lowerName, "postal"):
		f = func(opts GenOptions) string { return opts.fake().Zip() }
	case strings.Contains(lowerName, "description") || strings.Contains(lowerName, "bio") || strings.Contains(lowerName, "summary") || strings.Contains(lowerName, "comment") || strings.Contains(lowerName, "body") || strings.Contains(lowerName, "content"):
		f = func(opts GenOptions) string { return opts.fake().Paragraph() }
	case strings.Contains(lowerName, "name"):
		if strings.Contains(lowerName, "firstname") {
			f = func(opts GenOptions) string { return opts.fake().FirstName() }
		} else if strings.Contains(lowerName, "lastname") {
			f = func(opts GenOptions) string { return opts.fake().LastName() }
		} else if strings.Contains(lowerName, "fullname") {
			f = func(opts GenOptions) string { return opts.fake().Name() }
		} else {
			f = func(opts GenOptions) string { return opts.fake().FirstName() }
		}
	case strings.Contains(lowerName, "url"):
		if strings.Contains(lowerName, "photo") {
			f = func(opts GenOptions) string { return "https://picsum.photos/400" }
		} else {
			f = func(opts GenOptions) string { return opts.fake().URL() }
		}
	case strings.Contains(lowerName, "version"):
		f = func(opts GenOptions) string { return opts.fake().AppVersion() }
	case strings.Contains(lowerName, "status"):
		f = func(opts GenOptions) string {
			return opts.fake().RandomString(statusValues)
		}
	}

	heuristicCache.Store(fd, f)
	if f == nil {
		return "", false
	}
	return f(opts), true
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

	minLen, maxLen := uint64(0), uint64(20)
	isMaxLenSet := false
	if rules.Len != nil {
		minLen = *rules.Len
		maxLen = *rules.Len
		isMaxLenSet = true
	}
	if rules.MinLen != nil {
		minLen = *rules.MinLen
	}
	if rules.MaxLen != nil {
		maxLen = *rules.MaxLen
		isMaxLenSet = true
	}
	if rules.MinBytes != nil {
		minLen = *rules.MinBytes
	}
	if rules.MaxBytes != nil {
		maxLen = *rules.MaxBytes
		isMaxLenSet = true
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
			generatedString = strings.ToLower(opts.fake().FirstName() + "@" + opts.fake().DomainName())
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
		if !isMaxLenSet && rules.WellKnown != nil {
			// Don't truncate well-known types unless specifically requested
		} else {
			generatedString = generatedString[:maxLen]
		}
	}
	if isTagOrAttributesKey(fd) {
		if rules == nil || rules.Pattern == nil {
			generatedString = strings.ToLower(generatedString)
		}
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

func isTagOrAttributesKey(fd protoreflect.FieldDescriptor) bool {
	lowerName := strings.ToLower(string(fd.Name()))
	if strings.Contains(lowerName, "tag") {
		return true
	}
	if fd.Name() != "key" || fd.Kind() != protoreflect.StringKind {
		return false
	}
	md := fd.ContainingMessage()
	if md == nil || !md.IsMapEntry() {
		return false
	}
	parent := md.Parent()
	if parent == nil {
		return false
	}
	outerMsg, ok := parent.(protoreflect.MessageDescriptor)
	if !ok {
		return false
	}
	for i := range outerMsg.Fields().Len() {
		field := outerMsg.Fields().Get(i)
		if field.IsMap() && field.Message() == md {
			lowerMapName := strings.ToLower(string(field.Name()))
			return strings.Contains(lowerMapName, "attribute") || strings.Contains(lowerMapName, "attr")
		}
	}
	return false
}

func isIPField(lowerName string) bool {
	return lowerName == "ip" ||
		strings.Contains(lowerName, "ip_address") ||
		strings.HasPrefix(lowerName, "ip_") ||
		strings.HasPrefix(lowerName, "ip-") ||
		strings.HasSuffix(lowerName, "_ip") ||
		strings.HasSuffix(lowerName, "-ip") ||
		strings.Contains(lowerName, "_ip_") ||
		strings.Contains(lowerName, "-ip-")
}

func isMacField(lowerName string) bool {
	return lowerName == "mac" ||
		strings.Contains(lowerName, "mac_address") ||
		strings.HasPrefix(lowerName, "mac_") ||
		strings.HasPrefix(lowerName, "mac-") ||
		strings.HasSuffix(lowerName, "_mac") ||
		strings.HasSuffix(lowerName, "-mac") ||
		strings.Contains(lowerName, "_mac_") ||
		strings.Contains(lowerName, "-mac-")
}

func isCityField(lowerName string) bool {
	return lowerName == "city" ||
		strings.HasPrefix(lowerName, "city_") ||
		strings.HasPrefix(lowerName, "city-") ||
		strings.HasSuffix(lowerName, "_city") ||
		strings.HasSuffix(lowerName, "-city")
}

func isPhoneField(lowerName string) bool {
	return strings.Contains(lowerName, "phone") ||
		strings.Contains(lowerName, "mobile") ||
		strings.Contains(lowerName, "fax") ||
		lowerName == "tel" ||
		strings.HasPrefix(lowerName, "tel_") ||
		strings.HasPrefix(lowerName, "tel-") ||
		strings.HasSuffix(lowerName, "_tel") ||
		strings.HasSuffix(lowerName, "-tel")
}

func isTimeField(lowerName string) bool {
	return strings.HasSuffix(lowerName, "_at") ||
		lowerName == "date" ||
		strings.HasSuffix(lowerName, "_date") ||
		lowerName == "timestamp" ||
		strings.HasSuffix(lowerName, "_timestamp")
}
