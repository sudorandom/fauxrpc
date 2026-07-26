package stub

import (
	"sort"
	"sync"

	"github.com/sudorandom/fauxrpc/private/engine"
)

type Registry interface {
	AddStub(rule StubRule)
	FindMatch(req *engine.NormalizedRequest) (*StubRule, bool)
	Clear()
	NumStubs() int
}

type inMemoryRegistry struct {
	mu        sync.RWMutex
	stubs     []StubRule
	evaluator *MatchEvaluator
}

func NewRegistry() Registry {
	return &inMemoryRegistry{
		stubs:     make([]StubRule, 0),
		evaluator: NewMatchEvaluator(),
	}
}

func (r *inMemoryRegistry) AddStub(rule StubRule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stubs = append(r.stubs, rule)
}

func (r *inMemoryRegistry) FindMatch(req *engine.NormalizedRequest) (*StubRule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	type matchedCandidate struct {
		rule        *StubRule
		specificity int
		priority    int
		index       int
	}

	var candidates []matchedCandidate

	for i := range r.stubs {
		rule := &r.stubs[i]
		matched, specificity := r.evaluator.Matches(rule, req)
		if matched {
			candidates = append(candidates, matchedCandidate{
				rule:        rule,
				specificity: specificity,
				priority:    rule.Priority,
				index:       i,
			})
		}
	}

	if len(candidates) == 0 {
		return nil, false
	}

	// Sort candidates: highest priority first, then highest specificity, then insertion order
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority > candidates[j].priority
		}
		if candidates[i].specificity != candidates[j].specificity {
			return candidates[i].specificity > candidates[j].specificity
		}
		return candidates[i].index < candidates[j].index
	})

	matched := cloneStubRule(*candidates[0].rule)
	return &matched, true
}

func cloneStubRule(rule StubRule) StubRule {
	cloned := rule
	cloned.Match.PathParams = cloneStringMap(rule.Match.PathParams)
	cloned.Match.QueryParams = cloneStringMap(rule.Match.QueryParams)
	cloned.Match.Headers = cloneStringMap(rule.Match.Headers)
	cloned.Response.Headers = cloneStringMap(rule.Response.Headers)
	cloned.Response.Body = cloneStubBody(rule.Response.Body)
	return cloned
}

func cloneStringMap(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}
	cloned := make(map[string]string, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func cloneStubBody(value any) any {
	switch value := value.(type) {
	case map[string]any:
		cloned := make(map[string]any, len(value))
		for key, item := range value {
			cloned[key] = cloneStubBody(item)
		}
		return cloned
	case []any:
		cloned := make([]any, len(value))
		for index, item := range value {
			cloned[index] = cloneStubBody(item)
		}
		return cloned
	case []byte:
		return append([]byte(nil), value...)
	default:
		return value
	}
}

func (r *inMemoryRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stubs = make([]StubRule, 0)
}

func (r *inMemoryRegistry) NumStubs() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.stubs)
}
