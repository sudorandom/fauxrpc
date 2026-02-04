package stubs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestRemoveStubLeak(t *testing.T) {
	db := NewStubDatabase()

	key := StubKey{
		Name: "test.Message",
		ID:   "1",
	}
	entry := StubEntry{
		Key:      key,
		Priority: 10,
	}

	db.AddStub(entry)

	// Verify it was added
	assert.Equal(t, 1, db.NumStubs())
	assert.Len(t, db.stubIndex, 1)
	assert.Len(t, db.stubIndex[key.Name], 1)
	assert.Len(t, db.stubIndex[key.Name][0].Entries, 1)

	// Remove the stub
	db.RemoveStub(key)

	// Verify it was removed from stubsByKey
	assert.Equal(t, 0, db.NumStubs())

	// Verify NO LEAK: check if it is gone from stubIndex
	priorityEntries, ok := db.stubIndex[key.Name]
	assert.False(t, ok, "stubIndex should not contain the key name")
	assert.Empty(t, priorityEntries, "priorityEntries should be empty")
}

func BenchmarkGetStubsPrioritizedWithChurn(b *testing.B) {
	db := NewStubDatabase()
	// Add stubs
	totalStubs := 10000
	removeStubs := 5000

	// Use a single message name to force them into the same list in stubIndex
	msgName := protoreflect.FullName("test.Message")

	for i := 0; i < totalStubs; i++ {
		db.AddStub(StubEntry{
			Key: StubKey{
				Name: msgName,
				ID:   fmt.Sprintf("%d", i),
			},
			Priority: 1,
		})
	}

	// Remove half of them
	for i := 0; i < removeStubs; i++ {
		db.RemoveStub(StubKey{
			Name: msgName,
			ID:   fmt.Sprintf("%d", i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.GetStubsPrioritized(msgName)
	}
}
