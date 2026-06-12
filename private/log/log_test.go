package log

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	l := NewLogger()

	ch, unsubscribe := l.Subscribe()
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

func TestLogger_History(t *testing.T) {
	l := NewLogger()

	// Log 12 entries
	for i := 1; i <= 12; i++ {
		l.Log(&LogEntry{
			ID:        string(rune('A' + i)),
			Timestamp: time.Now(),
		})
	}

	history, ch, unsubscribe := l.SubscribeWithHistory()
	defer unsubscribe()

	// History should have exactly 10 entries (items 3 to 12)
	if len(history) != 10 {
		t.Fatalf("Expected history length to be 10, got %d", len(history))
	}

	expectedFirstID := string(rune('A' + 3))
	if history[0].ID != expectedFirstID {
		t.Errorf("Expected first history item ID to be %s, got %s", expectedFirstID, history[0].ID)
	}

	// Log a new entry and check if it is sent to channel
	newEntry := &LogEntry{
		ID:        "new-id",
		Timestamp: time.Now(),
	}
	l.Log(newEntry)

	select {
	case received := <-ch:
		if received.ID != newEntry.ID {
			t.Errorf("Expected log entry with id %s, but got %s", newEntry.ID, received.ID)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timed out waiting for new log entry")
	}
}
