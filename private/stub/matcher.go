package stub

import (
	"strings"

	"github.com/sudorandom/fauxrpc/private/engine"
	"github.com/tidwall/gjson"
)

// Type aliases from pkg/engine
type MatchTarget = engine.TargetRef
type StubRule = engine.StubRule
type StubMatch = engine.StubMatch
type StubResponse = engine.StubResponse

// MatchEvaluator checks if a StubRule matches a NormalizedRequest and calculates specificity.
type MatchEvaluator struct{}

func NewMatchEvaluator() *MatchEvaluator {
	return &MatchEvaluator{}
}

// Matches checks if the stub targets the request and satisfies all predicates.
func (e *MatchEvaluator) Matches(rule *StubRule, req *engine.NormalizedRequest) (bool, int) {
	if !e.targetMatches(rule.Target, req) {
		return false, 0
	}

	specificity := 0

	// 1. Path parameters matching
	for k, expectedVal := range rule.Match.PathParams {
		actualVal, exists := req.PathParams[k]
		if !exists || actualVal != expectedVal {
			return false, 0
		}
		specificity++
	}

	// 2. Query parameters matching
	for k, expectedVal := range rule.Match.QueryParams {
		actualVal, exists := req.QueryParams[k]
		if !exists || actualVal != expectedVal {
			return false, 0
		}
		specificity++
	}

	// 3. Header names are case-insensitive; values must match exactly.
	for k, expectedVal := range rule.Match.Headers {
		matched := false
		for _, actualVal := range req.Headers.Values(k) {
			if actualVal == expectedVal {
				matched = true
				break
			}
		}
		if !matched {
			return false, 0
		}
		specificity++
	}

	// 4. Body expression evaluation via gjson
	if rule.Match.BodyExpression != "" {
		if len(req.Body) == 0 {
			return false, 0
		}
		res := gjson.GetBytes(req.Body, rule.Match.BodyExpression)
		if !res.Exists() || !e.isTruthful(res) {
			return false, 0
		}
		specificity++
	}

	return true, specificity
}

func (e *MatchEvaluator) targetMatches(target MatchTarget, req *engine.NormalizedRequest) bool {
	if target.OperationID != "" && req.OperationID != "" {
		if target.OperationID == req.OperationID {
			return true
		}
	}
	if target.Service != "" && target.Method != "" {
		if target.Service == req.Service && target.Method == req.Method {
			return true
		}
	}
	if target.Path != "" {
		if target.Path == req.Path {
			if target.HTTPMethod == "" || strings.EqualFold(target.HTTPMethod, req.HTTPMethod) {
				return true
			}
		}
	}
	return false
}

func (e *MatchEvaluator) isTruthful(res gjson.Result) bool {
	switch res.Type {
	case gjson.True:
		return true
	case gjson.False:
		return false
	case gjson.Number:
		return res.Num != 0
	case gjson.String:
		return res.Str != "" && res.Str != "false"
	case gjson.JSON:
		return true
	case gjson.Null:
		return false
	default:
		return res.Bool()
	}
}
