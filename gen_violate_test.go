package fauxrpc_test

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	testv1 "github.com/sudorandom/fauxrpc/private/gen/test/v1"
)

func violateTestMessage(t *testing.T) protoreflect.MessageDescriptor {
	t.Helper()
	md := testv1.File_test_v1_test_proto.Messages().ByName("ValidateRulesTest")
	require.NotNil(t, md)
	return md
}

// TestViolateRulesEveryField checks that each violable field fails validation on
// its own when ViolateRules is 1.0.
func TestViolateRulesEveryField(t *testing.T) {
	md := violateTestMessage(t)
	validator, err := protovalidate.New()
	require.NoError(t, err)

	fields := md.Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		t.Run(string(fd.Name()), func(t *testing.T) {
			// Repeat with different seeds: each run picks a different rule to
			// break, and every one of them must be a real violation.
			for seed := range uint64(25) {
				msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{
					MaxDepth:     5,
					Faker:        gofakeit.New(seed + 1),
					ViolateRules: 1,
				})
				require.NoError(t, err)

				err = validator.Validate(msg)
				require.Error(t, err, "seed %d produced a valid message", seed)

				var violations *protovalidate.ValidationError
				require.ErrorAs(t, err, &violations)
				found := false
				for _, v := range violations.Violations {
					elements := v.Proto.GetField().GetElements()
					if len(elements) > 0 && elements[0].GetFieldName() == string(fd.Name()) {
						found = true
						break
					}
				}
				assert.True(t, found, "seed %d: no violation reported for %s", seed, fd.Name())
			}
		})
	}
}

// preexistingGenGaps are rules the generator does not satisfy even with
// violations turned off. They are unrelated to ViolateRules and are listed here
// so TestViolateRulesDisabled can tell them apart from a leaking violation.
var preexistingGenGaps = map[string]bool{
	"sfixed64_lt":   true, // negative exclusive upper bounds
	"string_prefix": true, // string.prefix/suffix/contains are not honored
	"timestamp_lt":  true, // timestamp bounds without a lower bound
}

// TestViolateRulesDisabled makes sure that leaving the option at 0 keeps the
// generator satisfying every rule it knows how to satisfy.
func TestViolateRulesDisabled(t *testing.T) {
	md := violateTestMessage(t)
	validator, err := protovalidate.New()
	require.NoError(t, err)

	for seed := range uint64(50) {
		msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{
			MaxDepth:     5,
			Faker:        gofakeit.New(seed + 1),
			ViolateRules: 0,
		})
		require.NoError(t, err)

		err = validator.Validate(msg)
		if err == nil {
			continue
		}
		var violations *protovalidate.ValidationError
		require.ErrorAs(t, err, &violations)
		for _, v := range violations.Violations {
			elements := v.Proto.GetField().GetElements()
			require.NotEmpty(t, elements)
			name := elements[0].GetFieldName()
			assert.True(t, preexistingGenGaps[name], "seed %d: %s was violated with the option off", seed, name)
		}
	}
}

// violationRates generates n messages and returns how often each field ended up
// with a reported violation.
func violationRates(t *testing.T, md protoreflect.MessageDescriptor, probability float64, n uint64) map[string]float64 {
	t.Helper()
	validator, err := protovalidate.New()
	require.NoError(t, err)

	counts := map[string]int{}
	for seed := range n {
		msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{
			MaxDepth:     5,
			Faker:        gofakeit.New(seed + 1),
			ViolateRules: probability,
		})
		require.NoError(t, err)

		err = validator.Validate(msg)
		if err == nil {
			continue
		}
		var violations *protovalidate.ValidationError
		require.ErrorAs(t, err, &violations)
		// One bad value can trip several rules at once, so count each field at
		// most once per message.
		seen := map[string]bool{}
		for _, v := range violations.Violations {
			elements := v.Proto.GetField().GetElements()
			if len(elements) > 0 {
				seen[elements[0].GetFieldName()] = true
			}
		}
		for name := range seen {
			counts[name]++
		}
	}

	rates := make(map[string]float64, len(counts))
	for name, count := range counts {
		rates[name] = float64(count) / float64(n)
	}
	return rates
}

// TestViolateRulesProbabilityPerRule checks that the option is a per-rule
// probability: a field with one rule is violated at roughly that rate, and a
// field with three is violated at roughly 1-(1-p)^3.
func TestViolateRulesProbabilityPerRule(t *testing.T) {
	md := violateTestMessage(t)
	const probability = 0.5
	rates := violationRates(t, md, probability, 600)

	// Generous bounds: this is a sampled rate, not an exact count.
	for _, name := range []string{"string_uuid", "int32_gt", "string_required"} {
		assert.InDelta(t, probability, rates[name], 0.08, "%s carries one rule", name)
	}

	expectedMulti := 1 - (1-probability)*(1-probability)*(1-probability)
	assert.InDelta(t, expectedMulti, rates["string_multi_rule"], 0.08,
		"string_multi_rule carries three rules, so each one gets its own roll")
}

// TestViolateRulesProbabilityScales checks that the violation rate tracks the
// option across its range rather than being all-or-nothing.
func TestViolateRulesProbabilityScales(t *testing.T) {
	md := violateTestMessage(t)
	for _, probability := range []float64{0.1, 0.25, 0.5, 0.9} {
		rates := violationRates(t, md, probability, 400)
		assert.InDelta(t, probability, rates["string_uuid"], 0.08,
			"single-rule field at probability %.2f", probability)
	}
}

// TestViolateRulesUnvalidatedMessage checks that a message with no rules is
// still generated normally.
func TestViolateRulesUnvalidatedMessage(t *testing.T) {
	md := testv1.File_test_v1_test_proto.Messages().ByName("UnvalidatedTest")
	require.NotNil(t, md)

	msg, err := fauxrpc.NewMessage(md, fauxrpc.GenOptions{
		MaxDepth:     5,
		Faker:        gofakeit.New(1),
		ViolateRules: 1,
	})
	require.NoError(t, err)

	name := msg.ProtoReflect().Descriptor().Fields().ByName("name")
	assert.NotEmpty(t, msg.ProtoReflect().Get(name).String())
}
