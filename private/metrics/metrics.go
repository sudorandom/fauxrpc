package metrics

import (
	"sync"
	"time"
)

type Stats struct {
	TotalRequests     int64
	RequestsPerSecond int64
	Errors            int64
	ErrorRate         string
	UniqueServices    int
	UniqueMethods     int
	HTTPHost          string
	GoVersion         string
	FauxRpcVersion    string
	StartedAt         time.Time
	LastReset         time.Time
	RequestCounts     map[time.Time]int64
	mu                sync.Mutex
}

func (s *Stats) IncrementTotalRequests() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalRequests++
	now := time.Now().Truncate(time.Second)
	s.RequestCounts[now]++
}

func (s *Stats) IncrementErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Errors++
}

func (s *Stats) Copy() *Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := &Stats{
		TotalRequests:     s.TotalRequests,
		RequestsPerSecond: s.RequestsPerSecond,
		Errors:            s.Errors,
		ErrorRate:         s.ErrorRate,
		UniqueServices:    s.UniqueServices,
		UniqueMethods:     s.UniqueMethods,
		HTTPHost:          s.HTTPHost,
		GoVersion:         s.GoVersion,
		FauxRpcVersion:    s.FauxRpcVersion,
		StartedAt:         s.StartedAt,
		LastReset:         s.LastReset,
		RequestCounts:     make(map[time.Time]int64),
		mu:                sync.Mutex{},
	}
	for k, v := range s.RequestCounts {
		copy.RequestCounts[k] = v
	}
	return copy
}

func (s *Stats) Reset() {
	s.mu.Lock()
	s.TotalRequests = 0
	s.Errors = 0
	s.RequestCounts = make(map[time.Time]int64)
	s.LastReset = time.Now()
	s.mu.Unlock()
}

func (s *Stats) Uptime() time.Duration {
	return time.Since(s.StartedAt)
}
