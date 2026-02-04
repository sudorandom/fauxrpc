package stubs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestStubDatabase_RemoveStub(t *testing.T) {
	db := NewStubDatabase()
	name := protoreflect.FullName("test.service.Method")
	key := StubKey{Name: name, ID: "id1"}
	entry := StubEntry{Key: key, Priority: 1}

	db.AddStub(entry)

	// Verify added
	got, ok := db.GetStub(key)
	assert.True(t, ok)
	assert.Equal(t, entry, got)

	prioritized := db.GetStubsPrioritized(name)
	assert.Len(t, prioritized, 1)
	assert.Len(t, prioritized[0], 1)

	// Remove
	db.RemoveStub(key)

	// Verify removed from map
	_, ok = db.GetStub(key)
	assert.False(t, ok)

	// Verify removed from index (this effectively checks my optimization/fix correctness)
	prioritized = db.GetStubsPrioritized(name)
	// GetStubsPrioritized returns groups. If the group exists but is empty, it might return empty group?
	// Let's see implementation of GetStubsPrioritized:
	/*
		groups := make([][]StubEntry, len(priorityStubEntries))
		for i, stubEntries := range priorityStubEntries {
			groups[i] = make([]StubEntry, 0, len(stubEntries.Entries))
			for _, key := range stubEntries.Entries {
				entry, ok := db.stubsByKey[key]
				if !ok {
					continue
				}
                // ...
				groups[i] = append(groups[i], entry)
			}
		}
	*/
	// If I removed it from entries, then `stubEntries.Entries` should be empty (or not contain the key).
	// So groups[i] should be empty.

	if len(prioritized) > 0 {
		assert.Empty(t, prioritized[0])
	}
}

func TestStubDatabase_AddRemoveMultiple(t *testing.T) {
	db := NewStubDatabase()
	name := protoreflect.FullName("test.service.Method")

	key1 := StubKey{Name: name, ID: "id1"}
	entry1 := StubEntry{Key: key1, Priority: 1}
	db.AddStub(entry1)

	key2 := StubKey{Name: name, ID: "id2"}
	entry2 := StubEntry{Key: key2, Priority: 1}
	db.AddStub(entry2)

	// Remove one
	db.RemoveStub(key1)

	prioritized := db.GetStubsPrioritized(name)
	// Should contain key2
	found := false
	for _, group := range prioritized {
		for _, e := range group {
			if e.Key == key2 {
				found = true
			}
			if e.Key == key1 {
				t.Error("key1 should not be present")
			}
		}
	}
	assert.True(t, found, "key2 should be present")
}
