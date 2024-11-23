package stubs

import (
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type StubEntry struct {
	Message protoreflect.ProtoMessage
	Error   *StatusError
	Rules   *Rules
}

var _ StubDatabase = (*stubDatabase)(nil)

type StubDatabase interface {
	GetStubs(protoreflect.FullName) []StubEntry
	ListStubs(protoreflect.FullName, string) map[protoreflect.FullName]map[string]StubEntry
	AddStub(protoreflect.FullName, string, StubEntry)
	RemoveStub(protoreflect.FullName, string)
	RemoveAllStubs()
}

type stubDatabase struct {
	stubs map[protoreflect.FullName]map[string]StubEntry

	mutex sync.RWMutex
}

func NewStubDatabase() *stubDatabase {
	return &stubDatabase{
		stubs: map[protoreflect.FullName]map[string]StubEntry{},
		mutex: sync.RWMutex{},
	}
}

func (db *stubDatabase) GetStubs(name protoreflect.FullName) []StubEntry {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	values, ok := db.stubs[name]
	if !ok {
		return nil
	}

	// Create a copy to avoid async access
	stubs := make([]StubEntry, 0, len(values))
	for _, stub := range values {
		stubs = append(stubs, stub)
	}

	return stubs
}

func (db *stubDatabase) ListStubs(filtername protoreflect.FullName, filterid string) map[protoreflect.FullName]map[string]StubEntry {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	result := make(map[protoreflect.FullName]map[string]StubEntry, len(db.stubs))
	for name, stubs := range db.stubs {
		if filtername != "" && name != filtername {
			continue
		}
		stubsResult := make(map[string]StubEntry, len(stubs))
		for id, stub := range stubs {
			if filterid != "" && id != filterid {
				continue
			}
			stubsResult[id] = stub
		}
		result[name] = stubsResult
	}
	return result
}

func (db *stubDatabase) AddStub(name protoreflect.FullName, id string, value StubEntry) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	if _, ok := db.stubs[name]; !ok {
		db.stubs[name] = map[string]StubEntry{}
	}
	db.stubs[name][id] = value
}

func (db *stubDatabase) RemoveStub(name protoreflect.FullName, id string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	stubs, ok := db.stubs[name]
	if !ok {
		return
	}
	delete(stubs, id)
}

func (db *stubDatabase) RemoveAllStubs() {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.stubs = map[protoreflect.FullName]map[string]StubEntry{}
}
