package engine

import (
	"sync"
)

type mockStubRegistry struct {
	mu    sync.RWMutex
	stubs []StubRule
}

func newMockStubRegistry() *mockStubRegistry {
	return &mockStubRegistry{
		stubs: make([]StubRule, 0),
	}
}

func (r *mockStubRegistry) AddStub(rule StubRule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stubs = append(r.stubs, rule)
}

func (r *mockStubRegistry) FindMatch(req *NormalizedRequest) (*StubRule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i := range r.stubs {
		rule := &r.stubs[i]
		if rule.Target.OperationID != "" && rule.Target.OperationID == req.OperationID {
			if id, ok := req.PathParams["id"]; ok {
				if rule.Match.PathParams["id"] == id {
					return rule, true
				}
			}
		}
	}
	return nil, false
}

func (r *mockStubRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stubs = nil
}

func (r *mockStubRegistry) NumStubs() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.stubs)
}
