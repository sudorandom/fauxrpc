package generator

import (
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

type ScalarSynthesizer struct{}

var (
	syntheticTimeStart = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	syntheticTimeEnd   = time.Date(2050, time.January, 1, 0, 0, 0, 0, time.UTC)
)

func NewScalarSynthesizer() *ScalarSynthesizer {
	return &ScalarSynthesizer{}
}

func (s *ScalarSynthesizer) Synthesize(schema *openapi3.Schema, r *rand.Rand) any {
	if schema == nil {
		return nil
	}

	// 1. Explicit Example / Examples
	if schema.Example != nil {
		return schema.Example
	}
	if len(schema.Examples) > 0 {
		return schema.Examples[0]
	}

	// Determine type
	types := schema.Type.Slice()
	primaryType := ""
	if len(types) > 0 {
		primaryType = types[0]
	}

	// 3. Format-driven synthesis
	faker := gofakeit.New(uint64(r.Int63()) + 1)
	switch strings.ToLower(schema.Format) {
	case "uuid":
		return syntheticUUID(r)
	case "date-time":
		return syntheticTime(r).Format(time.RFC3339)
	case "date":
		return syntheticTime(r).Format("2006-01-02")
	case "email":
		return strings.ToLower(faker.Email())
	case "hostname":
		return strings.ToLower(faker.DomainName())
	case "ipv4":
		return faker.IPv4Address()
	case "ipv6":
		return faker.IPv6Address()
	case "uri", "url":
		return faker.URL()
	case "byte":
		return "Z3JQQyBhdXRvLWdlbmVyYXRlZCBieXRlcw=="
	case "int32", "int64":
		if primaryType == "" {
			primaryType = "integer"
		}
	case "float", "double":
		if primaryType == "" {
			primaryType = "number"
		}
	}

	// 4. Constraint-driven synthesis
	switch primaryType {
	case "string":
		if schema.Pattern != "" {
			if synth := s.synthesizePattern(schema.Pattern); synth != "" {
				return synth
			}
		}
		if len(schema.Enum) > 0 {
			idx := r.Intn(len(schema.Enum))
			return schema.Enum[idx]
		}
		minLen := 5
		if schema.MinLength > 0 {
			minLen = int(schema.MinLength)
		}
		maxLen := minLen + 10
		if schema.MaxLength != nil && *schema.MaxLength > 0 {
			maxLen = int(*schema.MaxLength)
			if maxLen < minLen {
				maxLen = minLen
			}
		}
		faker := gofakeit.New(uint64(r.Int63()) + 1)
		val := faker.Word()
		for len(val) < minLen {
			val += faker.Word()
		}
		if len(val) > maxLen {
			val = val[:maxLen]
		}
		return val

	case "integer":
		minVal := int64(1)
		if schema.Min != nil {
			minVal = int64(*schema.Min)
		}
		maxVal := minVal + 100
		if schema.Max != nil {
			maxVal = int64(*schema.Max)
		}
		if maxVal < minVal {
			maxVal = minVal
		}
		if maxVal == minVal {
			return minVal
		}
		return minVal + r.Int63n(maxVal-minVal+1)

	case "number":
		minVal := 1.0
		if schema.Min != nil {
			minVal = *schema.Min
		}
		maxVal := minVal + 100.0
		if schema.Max != nil {
			maxVal = *schema.Max
		}
		if maxVal < minVal {
			maxVal = minVal
		}
		if maxVal == minVal {
			return minVal
		}
		return minVal + r.Float64()*(maxVal-minVal)

	case "boolean":
		return r.Intn(2) == 1
	}

	return "synthetic_value"
}

func syntheticUUID(r *rand.Rand) string {
	raw := make([]byte, 16)
	_, _ = r.Read(raw)
	raw[6] = raw[6]&0x0f | 0x40 // UUID version 4
	raw[8] = raw[8]&0x3f | 0x80 // RFC 4122 variant
	value, _ := uuid.FromBytes(raw)
	return value.String()
}

func syntheticTime(r *rand.Rand) time.Time {
	span := syntheticTimeEnd.Sub(syntheticTimeStart)
	return syntheticTimeStart.Add(time.Duration(r.Int63n(int64(span))))
}

func (s *ScalarSynthesizer) synthesizePattern(pattern string) string {
	// Common regex pattern matching fallbacks
	if strings.Contains(pattern, "[0-9]") || strings.Contains(pattern, "\\d") {
		return "12345"
	}
	if strings.Contains(pattern, "[a-z]") || strings.Contains(pattern, "[A-Z]") {
		return "sample"
	}
	if _, err := regexp.Compile(pattern); err == nil {
		return "matched_pattern"
	}
	return "pattern_val"
}
