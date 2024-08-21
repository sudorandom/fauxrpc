package stubs

import (
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type stubDB map[protoreflect.FullName]map[string]protoreflect.ProtoMessage

type StubDatabase interface {
	GetStubs(protoreflect.FullName) []protoreflect.ProtoMessage
	ListStubs(protoreflect.FullName, string) map[protoreflect.FullName]map[string]protoreflect.ProtoMessage
	AddStub(protoreflect.FullName, string, protoreflect.ProtoMessage)
	RemoveStub(protoreflect.FullName, string)
	RemoveAllStubs()
}

type stubDatabase struct {
	stubs stubDB

	mutex sync.RWMutex
}

func NewStubDatabase() *stubDatabase {
	return &stubDatabase{
		stubs: map[protoreflect.FullName]map[string]protoreflect.ProtoMessage{},
		mutex: sync.RWMutex{},
	}
}

func (db *stubDatabase) GetStubs(name protoreflect.FullName) []protoreflect.ProtoMessage {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	values, ok := db.stubs[name]
	if !ok {
		return nil
	}

	// Create a copy to avoid async access
	stubs := make([]protoreflect.ProtoMessage, 0, len(values))
	for _, stub := range values {
		stubs = append(stubs, stub)
	}

	return stubs
}

func (db *stubDatabase) ListStubs(filtername protoreflect.FullName, filterid string) map[protoreflect.FullName]map[string]protoreflect.ProtoMessage {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	result := make(map[protoreflect.FullName]map[string]protoreflect.ProtoMessage, len(db.stubs))
	for name, stubs := range db.stubs {
		if filtername != "" && name != filtername {
			continue
		}
		stubsResult := make(map[string]protoreflect.ProtoMessage, len(stubs))
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

func (db *stubDatabase) AddStub(name protoreflect.FullName, id string, value protoreflect.ProtoMessage) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	if _, ok := db.stubs[name]; !ok {
		db.stubs[name] = map[string]protoreflect.ProtoMessage{}
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
	db.stubs = map[protoreflect.FullName]map[string]protoreflect.ProtoMessage{}
}
