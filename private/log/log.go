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
	Duration        int64           `json:"duration"`
	RequestHeaders  json.RawMessage `json:"requestHeaders"`
	ResponseHeaders json.RawMessage `json:"responseHeaders"`
	RequestBody     json.RawMessage `json:"requestBody"`
	ResponseBody    json.RawMessage `json:"responseBody"`
}

// RingBuffer is a simple thread-safe ring buffer for LogEntry objects.
type RingBuffer struct {
	mu       sync.RWMutex
	entries  []*LogEntry
	size     int
	position int
	isFull   bool
}

// NewRingBuffer creates a new RingBuffer with the given size.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		entries: make([]*LogEntry, size),
		size:    size,
	}
}

// Add adds a new LogEntry to the buffer.
func (rb *RingBuffer) Add(entry *LogEntry) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.entries[rb.position] = entry
	rb.position++
	if !rb.isFull && rb.position == rb.size {
		rb.isFull = true
	}
	if rb.position == rb.size {
		rb.position = 0
	}
}

// GetAll returns all the entries in the buffer, in order from newest to oldest.
func (rb *RingBuffer) GetAll() []*LogEntry {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	var result []*LogEntry
	if !rb.isFull {
		for i := rb.position - 1; i >= 0; i-- {
			result = append(result, rb.entries[i])
		}
		return result
	}

	// When buffer is full, position is the oldest entry.
	// Start from one before position and go backwards.
	for i := 0; i < rb.size; i++ {
		idx := (rb.position - 1 - i + rb.size) % rb.size
		result = append(result, rb.entries[idx])
	}
	return result
}

// Logger manages the log entries and subscriptions.
type Logger struct {
	buffer      *RingBuffer
	mu          sync.RWMutex
	subscribers map[chan *LogEntry]struct{}
}

// NewLogger creates a new Logger.
func NewLogger(bufferSize int) *Logger {
	return &Logger{
		buffer:      NewRingBuffer(bufferSize),
		subscribers: make(map[chan *LogEntry]struct{}),
	}
}

// Log adds a new entry and notifies subscribers.
func (l *Logger) Log(entry *LogEntry) {
	l.buffer.Add(entry)
	l.mu.RLock()
	defer l.mu.RUnlock()
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

// GetHistory returns the historical log entries from the ring buffer.
func (l *Logger) GetHistory() []*LogEntry {
	return l.buffer.GetAll()
}
