package server

import (
	"encoding/json"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type FrameTracker struct {
	mu     sync.Mutex
	frames []json.RawMessage
	limit  int
}

func NewFrameTracker(limit int) *FrameTracker {
	if limit <= 0 {
		limit = 10
	}
	return &FrameTracker{
		limit: limit,
	}
}

func (ft *FrameTracker) Add(msg proto.Message) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	b, err := protojson.Marshal(msg)
	if err != nil {
		return
	}

	if len(ft.frames) < ft.limit*2 {
		ft.frames = append(ft.frames, b)
	} else {
		// Remove the first element of the "last N" segment
		// ft.frames[:ft.limit] keeps the "first N"
		// ft.frames[ft.limit+1:] keeps the rest of "last N" except the one we are removing
		ft.frames = append(ft.frames[:ft.limit], ft.frames[ft.limit+1:]...)
		ft.frames = append(ft.frames, b)
	}
}

func (ft *FrameTracker) Frames() []json.RawMessage {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	result := make([]json.RawMessage, len(ft.frames))
	copy(result, ft.frames)
	return result
}
