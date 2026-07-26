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

	return candidates[0].rule, true
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
