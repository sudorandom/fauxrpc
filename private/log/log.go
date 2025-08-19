package log

import (
	"encoding/json"
	"sync"
	"time"
)

// LogEntry represents a single log entry for a request.
type LogEntry struct {
	ID              string          `json:"id"`
	Timestamp       time.Time       `json:"timestamp"`
	Service         string          `json:"service"`
	Method          string          `json:"method"`
	Status          int             `json:"status"`
	Duration        time.Duration   `json:"duration"`
	RequestHeaders  json.RawMessage `json:"requestHeaders"`
	ResponseHeaders json.RawMessage `json:"responseHeaders"`
	RequestBody     json.RawMessage `json:"requestBody"`
	ResponseBody    json.RawMessage `json:"responseBody"`
}

// Logger manages the log entries and subscriptions.
type Logger struct {
	mu          sync.RWMutex
	subscribers map[chan *LogEntry]struct{}
}

// NewLogger creates a new Logger.
func NewLogger() *Logger {
	return &Logger{
		subscribers: make(map[chan *LogEntry]struct{}),
	}
}

// Log adds a new entry and notifies subscribers.
func (l *Logger) Log(entry *LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for ch := range l.subscribers {
		// Use a non-blocking send to avoid blocking the logger
		// if a subscriber is slow.
		select {
		case ch <- entry:
		default:
		}
	}
}

// Subscribe adds a new subscriber channel.
func (l *Logger) Subscribe() (chan *LogEntry, func()) {
	l.mu.Lock()
	defer l.mu.Unlock()

	ch := make(chan *LogEntry, 100) // Buffered channel
	l.subscribers[ch] = struct{}{}

	unsubscribe := func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		delete(l.subscribers, ch)
		close(ch)
	}
	return ch, unsubscribe
}
