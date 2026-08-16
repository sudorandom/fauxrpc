package fauxrpc

import (
	"regexp"
	"strings"
	"time"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// violationMarker is appended to or used in place of otherwise valid values to
// build something that no reasonable rule accepts.
const violationMarker = "fauxrpc-invalid"

// maxViolatingItems bounds how many items are generated when violating an
// items/pairs upper bound. Bounds above this are left alone rather than
// allocating an unreasonable amount of data.
const maxViolatingItems = 100

// pickOne returns a random element of candidates, or false when there are none.
func pickOne[T any](candidates []T, opts GenOptions) (T, bool) {
	if len(candidates) == 0 {
		var zero T
		return zero, false
	}
	return candidates[opts.fake().IntRange(0, len(candidates)-1)], true
}

// violatingFieldValue returns a value for fd that fails one of the protovalidate
// rules attached to it, picked at random from every rule a value can break. The
// second return value is false when the field has no such rule, in which case
// the caller falls back to normal generation.
func violatingFieldValue(fd protoreflect.FieldDescriptor, opts GenOptions) (protoreflect.Value, bool) {
	rules := getFieldConstraints(fd, opts)
	if rules == nil || rules.GetIgnore() == validate.Ignore_IGNORE_ALWAYS {
		return protoreflect.Value{}, false
	}

	val, ok := violatingValue(fd, rules, opts)
	if !ok {
		return protoreflect.Value{}, false
	}
	// A zero value would make protovalidate skip the rules entirely, so it is
	// not a violation for these fields.
	if rules.GetIgnore() == validate.Ignore_IGNORE_IF_ZERO_VALUE && isZeroValue(fd, val) {
		return protoreflect.Value{}, false
	}
	return val, true
}

func violatingValue(fd protoreflect.FieldDescriptor, rules *validate.FieldRules, opts GenOptions) (protoreflect.Value, bool) {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		if v, ok := violateBoolRules(rules.GetBool()); ok {
			return protoreflect.ValueOfBool(v), true
		}
	case protoreflect.EnumKind:
		if v, ok := violateEnumRules(fd.Enum(), rules.GetEnum(), opts); ok {
			return protoreflect.ValueOfEnum(v), true
		}
	case protoreflect.StringKind:
		if v, ok := violateStringRules(rules.GetString(), opts); ok {
			return protoreflect.ValueOfString(v), true
		}
	case protoreflect.BytesKind:
		if v, ok := violateBytesRules(rules.GetBytes(), opts); ok {
			return protoreflect.ValueOfBytes(v), true
		}
	case protoreflect.MessageKind, protoreflect.GroupKind:
		switch fd.Message().FullName() {
		case "google.protobuf.Duration":
			if v, ok := violateDurationRules(rules.GetDuration(), opts); ok {
				return protoreflect.ValueOfMessage(v.ProtoReflect()), true
			}
		case "google.protobuf.Timestamp":
			if v, ok := violateTimestampRules(rules.GetTimestamp(), opts); ok {
				return protoreflect.ValueOfMessage(v.ProtoReflect()), true
			}
		}
	default:
		return violatingNumericValue(fd, rules, opts)
	}
	return protoreflect.Value{}, false
}

// violatesByOmission reports whether leaving fd unset fails its rules. Only
// `required` can be violated that way.
func violatesByOmission(fd protoreflect.FieldDescriptor, opts GenOptions) bool {
	if fd.IsList() || fd.IsMap() || fd.ContainingMessage().IsMapEntry() {
		return false
	}
	rules := getFieldConstraints(fd, opts)
	if rules == nil || !rules.GetRequired() {
		return false
	}
	// An unset field is a zero value, which these modes exempt from validation.
	switch rules.GetIgnore() {
	case validate.Ignore_IGNORE_ALWAYS, validate.Ignore_IGNORE_IF_ZERO_VALUE:
		return false
	}
	return true
}

func isZeroValue(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return !val.Bool()
	case protoreflect.StringKind:
		return val.String() == ""
	case protoreflect.BytesKind:
		return len(val.Bytes()) == 0
	case protoreflect.EnumKind:
		return val.Enum() == 0
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return false
	default:
		return isZeroNumber(fd, val)
	}
}

func violateBoolRules(rules *validate.BoolRules) (bool, bool) {
	if rules == nil || rules.Const == nil {
		return false, false
	}
	return !*rules.Const, true
}

func violateEnumRules(ed protoreflect.EnumDescriptor, rules *validate.EnumRules, opts GenOptions) (protoreflect.EnumNumber, bool) {
	if rules == nil {
		return 0, false
	}
	var candidates []int32
	if rules.Const != nil {
		if *rules.Const < maxInt32 {
			candidates = append(candidates, *rules.Const+1)
		} else {
			candidates = append(candidates, *rules.Const-1)
		}
	}
	candidates = append(candidates, rules.NotIn...)
	if len(rules.In) > 0 {
		if v, ok := valueNotIn(rules.In, 0, maxInt32); ok {
			candidates = append(candidates, v)
		}
	}
	if rules.GetDefinedOnly() && ed != nil {
		if v, ok := undefinedEnumNumber(ed); ok {
			candidates = append(candidates, int32(v))
		}
	}
	v, ok := pickOne(candidates, opts)
	return protoreflect.EnumNumber(v), ok
}

// undefinedEnumNumber returns a number that ed does not declare.
func undefinedEnumNumber(ed protoreflect.EnumDescriptor) (protoreflect.EnumNumber, bool) {
	values := ed.Values()
	highest := protoreflect.EnumNumber(0)
	for i := range values.Len() {
		if n := values.Get(i).Number(); n > highest {
			highest = n
		}
	}
	if highest == maxInt32 {
		return 0, false
	}
	return highest + 1, true
}

func violateStringRules(rules *validate.StringRules, opts GenOptions) (string, bool) {
	if rules == nil {
		return "", false
	}
	var candidates []string
	if rules.Const != nil {
		candidates = append(candidates, *rules.Const+violationMarker)
	}
	candidates = append(candidates, rules.NotIn...)
	if len(rules.In) > 0 {
		candidates = append(candidates, stringNotIn(rules.In))
	}
	if rules.WellKnown != nil {
		candidates = append(candidates, violateStringWellKnown(rules.WellKnown))
	}
	if rules.Pattern != nil {
		if v, ok := stringNotMatching(*rules.Pattern); ok {
			candidates = append(candidates, v)
		}
	}
	candidates = append(candidates, violateStringLengths(rules)...)
	if rules.NotContains != nil && *rules.NotContains != "" {
		candidates = append(candidates, violationMarker+*rules.NotContains)
	}
	if rules.Contains != nil && *rules.Contains != "" {
		candidates = append(candidates, stringWithout(*rules.Contains, strings.Contains))
	}
	if rules.Prefix != nil && *rules.Prefix != "" {
		candidates = append(candidates, stringWithout(*rules.Prefix, strings.HasPrefix))
	}
	if rules.Suffix != nil && *rules.Suffix != "" {
		candidates = append(candidates, stringWithout(*rules.Suffix, strings.HasSuffix))
	}
	return pickOne(candidates, opts)
}

// violateStringLengths returns a string per length rule that the string breaks.
// `len`/`min_len`/`max_len` count code points and the `*_bytes` variants count
// bytes; the generated strings are ASCII so the two agree.
func violateStringLengths(rules *validate.StringRules) []string {
	var candidates []string
	tooLong := func(n uint64) {
		if n < maxViolatingLength {
			candidates = append(candidates, strings.Repeat("a", int(n)+1))
		}
	}
	tooShort := func(n uint64) {
		if n > 0 {
			candidates = append(candidates, strings.Repeat("a", int(n)-1))
		}
	}
	for _, exact := range []*uint64{rules.Len, rules.LenBytes} {
		if exact != nil {
			tooLong(*exact)
		}
	}
	for _, maxRule := range []*uint64{rules.MaxLen, rules.MaxBytes} {
		if maxRule != nil {
			tooLong(*maxRule)
		}
	}
	for _, minRule := range []*uint64{rules.MinLen, rules.MinBytes} {
		if minRule != nil {
			tooShort(*minRule)
		}
	}
	return candidates
}

// violateStringWellKnown returns a value that none of the well-known string
// formats accept.
func violateStringWellKnown(wellKnown any) string {
	if _, ok := wellKnown.(*validate.StringRules_WellKnownRegex); ok {
		// HTTP header rules accept most printable text, but never control
		// characters.
		return "\x00\x01"
	}
	return "!!" + violationMarker + "!!"
}

func violateBytesRules(rules *validate.BytesRules, opts GenOptions) ([]byte, bool) {
	if rules == nil {
		return nil, false
	}
	var candidates [][]byte
	if rules.Const != nil {
		candidates = append(candidates, append(append([]byte{}, rules.Const...), violationMarker...))
	}
	candidates = append(candidates, rules.NotIn...)
	if len(rules.In) > 0 {
		candidates = append(candidates, []byte(stringNotIn(bytesToStrings(rules.In))))
	}
	if rules.WellKnown != nil {
		// Every well-known bytes format is a fixed width (4 or 16 bytes); three
		// bytes matches none of them.
		candidates = append(candidates, []byte{0x00, 0x01, 0x02})
	}
	if rules.Pattern != nil {
		if v, ok := stringNotMatching(*rules.Pattern); ok {
			candidates = append(candidates, []byte(v))
		}
	}
	candidates = append(candidates, violateBytesLengths(rules)...)
	if len(rules.Contains) > 0 {
		candidates = append(candidates, []byte(stringWithout(string(rules.Contains), strings.Contains)))
	}
	if len(rules.Prefix) > 0 {
		candidates = append(candidates, []byte(stringWithout(string(rules.Prefix), strings.HasPrefix)))
	}
	if len(rules.Suffix) > 0 {
		candidates = append(candidates, []byte(stringWithout(string(rules.Suffix), strings.HasSuffix)))
	}
	return pickOne(candidates, opts)
}

func violateBytesLengths(rules *validate.BytesRules) [][]byte {
	var candidates [][]byte
	for _, tooLong := range []*uint64{rules.Len, rules.MaxLen} {
		if tooLong != nil && *tooLong < maxViolatingLength {
			candidates = append(candidates, make([]byte, *tooLong+1))
		}
	}
	if rules.MinLen != nil && *rules.MinLen > 0 {
		candidates = append(candidates, make([]byte, *rules.MinLen-1))
	}
	return candidates
}

func violateDurationRules(rules *validate.DurationRules, opts GenOptions) (*durationpb.Duration, bool) {
	if rules == nil {
		return nil, false
	}
	var candidates []*durationpb.Duration
	if rules.Const != nil {
		candidates = append(candidates, durationpb.New(rules.Const.AsDuration()+time.Second))
	}
	candidates = append(candidates, rules.NotIn...)
	if len(rules.In) > 0 {
		candidates = append(candidates, durationpb.New(durationNotIn(rules.In)))
	}
	// A bound is not greater/less than itself, so it is its own counterexample
	// for the exclusive forms.
	switch v := rules.GreaterThan.(type) {
	case *validate.DurationRules_Gt:
		candidates = append(candidates, v.Gt)
	case *validate.DurationRules_Gte:
		candidates = append(candidates, durationpb.New(v.Gte.AsDuration()-1))
	}
	switch v := rules.LessThan.(type) {
	case *validate.DurationRules_Lt:
		candidates = append(candidates, v.Lt)
	case *validate.DurationRules_Lte:
		candidates = append(candidates, durationpb.New(v.Lte.AsDuration()+1))
	}
	return pickOne(candidates, opts)
}

func violateTimestampRules(rules *validate.TimestampRules, opts GenOptions) (*timestamppb.Timestamp, bool) {
	if rules == nil {
		return nil, false
	}
	var candidates []*timestamppb.Timestamp
	if rules.Const != nil {
		candidates = append(candidates, timestamppb.New(rules.Const.AsTime().Add(time.Second)))
	}
	switch v := rules.GreaterThan.(type) {
	case *validate.TimestampRules_Gt:
		candidates = append(candidates, v.Gt)
	case *validate.TimestampRules_Gte:
		candidates = append(candidates, timestamppb.New(v.Gte.AsTime().Add(-time.Second)))
	}
	switch v := rules.LessThan.(type) {
	case *validate.TimestampRules_Lt:
		candidates = append(candidates, v.Lt)
	case *validate.TimestampRules_Lte:
		candidates = append(candidates, timestamppb.New(v.Lte.AsTime().Add(time.Second)))
	}
	if rules.Within != nil {
		// Twice the window, in either direction, is outside of it.
		window := rules.Within.AsDuration()
		candidates = append(candidates,
			timestamppb.New(time.Now().Add(-2*window)),
			timestamppb.New(time.Now().Add(2*window)),
		)
	}
	return pickOne(candidates, opts)
}

// violateRepeated builds a list whose length or uniqueness breaks the rules on
// fd. Rules on the items themselves are violated by FieldValue instead.
func violateRepeated(msg protoreflect.Message, fd protoreflect.FieldDescriptor, rules *validate.FieldRules, opts GenOptions) (*protoreflect.Value, bool) {
	repeated := rules.GetRepeated()
	itemOpts := opts.WithExtraFieldConstraints(repeated.GetItems())

	buildList := func(count int) *protoreflect.Value {
		listVal := msg.NewField(fd)
		for range count {
			if v := FieldValue(fd, itemOpts); v != nil {
				listVal.List().Append(*v)
			}
		}
		return &listVal
	}

	var strategies []func() *protoreflect.Value
	if repeated.GetMaxItems() > 0 && repeated.GetMaxItems() < maxViolatingItems {
		count := int(repeated.GetMaxItems()) + 1
		strategies = append(strategies, func() *protoreflect.Value { return buildList(count) })
	}
	if repeated.GetMinItems() > 0 {
		count := int(repeated.GetMinItems()) - 1
		strategies = append(strategies, func() *protoreflect.Value { return buildList(count) })
	}
	if repeated.GetUnique() {
		strategies = append(strategies, func() *protoreflect.Value {
			listVal := msg.NewField(fd)
			v := FieldValue(fd, itemOpts)
			if v == nil {
				return nil
			}
			listVal.List().Append(*v)
			listVal.List().Append(*v)
			return &listVal
		})
	}
	if rules.GetRequired() {
		strategies = append(strategies, func() *protoreflect.Value {
			listVal := msg.NewField(fd)
			return &listVal
		})
	}

	strategy, ok := pickOne(strategies, opts)
	if !ok {
		return nil, false
	}
	val := strategy()
	return val, val != nil
}

// violateMap builds a map whose size breaks the rules on fd. Rules on the keys
// and values themselves are violated by FieldValue instead.
func violateMap(msg protoreflect.Message, fd protoreflect.FieldDescriptor, rules *validate.FieldRules, opts GenOptions) (*protoreflect.Value, bool) {
	mapRules := rules.GetMap()

	buildMap := func(count int) *protoreflect.Value {
		mapVal := msg.NewField(fd)
		// Keys are generated normally: a violated key is often a constant,
		// which would collapse every entry into one and defeat the size
		// violation.
		keyOpts := opts.WithExtraFieldConstraints(mapRules.GetKeys())
		keyOpts.ViolateRules = 0
		valueOpts := opts.WithExtraFieldConstraints(mapRules.GetValues())
		for attempt := 0; mapVal.Map().Len() < count && attempt < count*10; attempt++ {
			k := FieldValue(fd.MapKey(), keyOpts)
			v := FieldValue(fd.MapValue(), valueOpts)
			if k != nil && v != nil {
				mapVal.Map().Set((*k).MapKey(), *v)
			}
		}
		return &mapVal
	}

	var strategies []func() *protoreflect.Value
	if mapRules.GetMaxPairs() > 0 && mapRules.GetMaxPairs() < maxViolatingItems {
		count := int(mapRules.GetMaxPairs()) + 1
		strategies = append(strategies, func() *protoreflect.Value { return buildMap(count) })
	}
	if mapRules.GetMinPairs() > 0 {
		count := int(mapRules.GetMinPairs()) - 1
		strategies = append(strategies, func() *protoreflect.Value { return buildMap(count) })
	}
	if rules.GetRequired() {
		strategies = append(strategies, func() *protoreflect.Value {
			mapVal := msg.NewField(fd)
			return &mapVal
		})
	}

	strategy, ok := pickOne(strategies, opts)
	if !ok {
		return nil, false
	}
	val := strategy()
	return val, val != nil
}

func stringNotIn(in []string) string {
	allowed := make(map[string]struct{}, len(in))
	for _, v := range in {
		allowed[v] = struct{}{}
	}
	candidate := violationMarker
	for {
		if _, ok := allowed[candidate]; !ok {
			return candidate
		}
		candidate += "!"
	}
}

// stringWithout returns a string for which has(s, substr) is false.
func stringWithout(substr string, has func(s, substr string) bool) string {
	for _, filler := range []string{"z", "a"} {
		s := strings.Repeat(filler, len(substr)+1)
		if !has(s, substr) {
			return s
		}
	}
	return ""
}

// stringNotMatching returns a string that the RE2 pattern does not match. It
// returns false for patterns that match everything it tries.
func stringNotMatching(pattern string) (string, bool) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", false
	}
	for _, candidate := range []string{violationMarker, "!!!", "\x00\x01", "", "0", "zzzzzzzz"} {
		if !re.MatchString(candidate) {
			return candidate, true
		}
	}
	return "", false
}

func durationNotIn(in []*durationpb.Duration) time.Duration {
	allowed := make(map[time.Duration]struct{}, len(in))
	for _, v := range in {
		allowed[v.AsDuration()] = struct{}{}
	}
	for i := range len(in) + 1 {
		candidate := time.Duration(i) * time.Second
		if _, ok := allowed[candidate]; !ok {
			return candidate
		}
	}
	return 0
}

func bytesToStrings(in [][]byte) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = string(v)
	}
	return out
}
