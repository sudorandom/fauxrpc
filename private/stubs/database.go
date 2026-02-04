package stubs

import (
	"sort"
	"sync"

	"github.com/sudorandom/fauxrpc/protocel"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type StubKey struct {
	Name protoreflect.FullName
	ID   string
}

func (e StubKey) GetName() protoreflect.FullName {
	return e.Name
}

func (e StubKey) GetID() string {
	return e.ID
}

type StubEntry struct {
	Key              StubKey
	Message          protoreflect.ProtoMessage
	CELMessage       protocel.CELMessage
	CELContentString string
	Error            *StatusError
	ActiveIf         *ActiveIf
	Priority         int
}

type PriorityStubEntries struct {
	Priority int
	Entries  []StubKey
}

var _ StubDatabase = (*stubDatabase)(nil)

type StubDatabase interface {
	GetStub(StubKey) (StubEntry, bool)
	GetStubs() []StubEntry
	GetStubsPrioritized(protoreflect.FullName) [][]StubEntry
	AddStub(StubEntry)
	RemoveStub(StubKey)
	RemoveAllStubs()
	NumStubs() int
}

type stubDatabase struct {
	stubsByKey map[StubKey]StubEntry
	stubIndex  map[protoreflect.FullName][]PriorityStubEntries
	mutex      sync.RWMutex
}

func NewStubDatabase() *stubDatabase {
	return &stubDatabase{
		stubsByKey: map[StubKey]StubEntry{},
		stubIndex:  map[protoreflect.FullName][]PriorityStubEntries{},
		mutex:      sync.RWMutex{},
	}
}

func (db *stubDatabase) GetStub(key StubKey) (StubEntry, bool) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	entry, ok := db.stubsByKey[key]
	if !ok {
		return entry, false
	}

	return entry, true
}

func (db *stubDatabase) GetStubs() []StubEntry {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	// Create a copy to avoid async access
	entries := make([]StubEntry, 0, len(db.stubsByKey))
	for _, stub := range db.stubsByKey {
		entries = append(entries, stub)
	}

	return entries
}

func (db *stubDatabase) GetStubsPrioritized(name protoreflect.FullName) [][]StubEntry {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	priorityStubEntries, ok := db.stubIndex[name]
	if !ok {
		return nil
	}

	groups := make([][]StubEntry, len(priorityStubEntries))
	for i, stubEntries := range priorityStubEntries {
		groups[i] = make([]StubEntry, 0, len(stubEntries.Entries))
		for _, key := range stubEntries.Entries {
			entry, ok := db.stubsByKey[key]
			if !ok {
				continue
			}
			if entry.Priority != stubEntries.Priority {
				continue
			}
			groups[i] = append(groups[i], entry)
		}
	}

	return groups
}

func (db *stubDatabase) AddStub(entry StubEntry) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.stubsByKey[entry.Key] = entry
	groups := db.stubIndex[entry.Key.Name]
	var foundGroup bool
	for i, group := range groups {
		if entry.Priority == group.Priority {
			foundGroup = true
			groups[i].Entries = append(groups[i].Entries, entry.Key)
			break
		}
	}
	if !foundGroup {
		groups = append(groups, PriorityStubEntries{
			Priority: entry.Priority,
			Entries:  []StubKey{entry.Key},
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Priority > groups[j].Priority
	})
	db.stubIndex[entry.Key.Name] = groups
}

func (db *stubDatabase) RemoveStub(key StubKey) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	entry, ok := db.stubsByKey[key]
	if !ok {
		return
	}
	delete(db.stubsByKey, key)

	groups := db.stubIndex[key.Name]
	for i, group := range groups {
		if group.Priority == entry.Priority {
			for j, k := range group.Entries {
				if k == key {
					groups[i].Entries = append(group.Entries[:j], group.Entries[j+1:]...)
					break
				}
			}
			break
		}
	}
}

func (db *stubDatabase) RemoveAllStubs() {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.stubsByKey = map[StubKey]StubEntry{}
	db.stubIndex = map[protoreflect.FullName][]PriorityStubEntries{}
}

func (db *stubDatabase) NumStubs() int {
	return len(db.stubsByKey)
}
