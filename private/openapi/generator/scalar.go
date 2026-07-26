package generator

import (
	"fmt"
	"math/rand"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

type ScalarSynthesizer struct{}

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

	// 2. Default Value
	if schema.Default != nil {
		return schema.Default
	}

	// Determine type
	types := schema.Type.Slice()
	primaryType := ""
	if len(types) > 0 {
		primaryType = types[0]
	}

	// 3. Format-driven synthesis
	switch strings.ToLower(schema.Format) {
	case "uuid":
		return uuid.New().String()
	case "date-time":
		return time.Now().UTC().Format(time.RFC3339)
	case "date":
		return time.Now().UTC().Format("2006-01-02")
	case "email":
		return fmt.Sprintf("user-%d@fauxrpc.local", r.Intn(10000))
	case "hostname":
		return fmt.Sprintf("host-%d.fauxrpc.local", r.Intn(1000))
	case "ipv4":
		return fmt.Sprintf("192.168.%d.%d", r.Intn(255)+1, r.Intn(254)+1)
	case "ipv6":
		return "2001:db8::1"
	case "uri", "url":
		return fmt.Sprintf("https://fauxrpc.local/resource/%d", r.Intn(1000))
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
		val := gofakeit.Word()
		for len(val) < minLen {
			val += gofakeit.Word()
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

// Suppress unused import warning for mail if any
var _ = mail.ParseAddress
