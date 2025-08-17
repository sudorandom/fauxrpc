package log

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	l := NewLogger(10)

	_, ch, unsubscribe := l.Subscribe()
	defer unsubscribe()

	entry := &LogEntry{
		ID:        "test-id",
		Timestamp: time.Now(),
	}

	l.Log(entry)

	select {
	case received := <-ch:
		if received.ID != entry.ID {
			t.Errorf("Expected log entry with id %s, but got %s", entry.ID, received.ID)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timed out waiting for log entry")
	}
}
