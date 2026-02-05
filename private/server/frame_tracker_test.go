package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestFrameTracker(t *testing.T) {
	ft := NewFrameTracker(2) // Limit 2, so capacity 4 (2*2)

	// Add 1
	ft.Add(wrapperspb.Int32(1))
	frames := ft.Frames()
	assert.Len(t, frames, 1)
	assertFrameValue(t, frames[0], 1)

	// Add 2
	ft.Add(wrapperspb.Int32(2))
	frames = ft.Frames()
	assert.Len(t, frames, 2)
	assertFrameValue(t, frames[0], 1)
	assertFrameValue(t, frames[1], 2)

	// Add 3
	ft.Add(wrapperspb.Int32(3))
	frames = ft.Frames()
	assert.Len(t, frames, 3)
	assertFrameValue(t, frames[0], 1)
	assertFrameValue(t, frames[1], 2)
	assertFrameValue(t, frames[2], 3)

	// Add 4
	ft.Add(wrapperspb.Int32(4))
	frames = ft.Frames()
	assert.Len(t, frames, 4)
	assertFrameValue(t, frames[0], 1)
	assertFrameValue(t, frames[1], 2)
	assertFrameValue(t, frames[2], 3)
	assertFrameValue(t, frames[3], 4)

	// Add 5. Should keep first 2 (1, 2) and last 2 (4, 5). 3 is dropped.
	ft.Add(wrapperspb.Int32(5))
	frames = ft.Frames()
	assert.Len(t, frames, 4)
	assertFrameValue(t, frames[0], 1)
	assertFrameValue(t, frames[1], 2)
	assertFrameValue(t, frames[2], 4)
	assertFrameValue(t, frames[3], 5)

	// Add 6. Should keep first 2 (1, 2) and last 2 (5, 6). 4 is dropped.
	ft.Add(wrapperspb.Int32(6))
	frames = ft.Frames()
	assert.Len(t, frames, 4)
	assertFrameValue(t, frames[0], 1)
	assertFrameValue(t, frames[1], 2)
	assertFrameValue(t, frames[2], 5)
	assertFrameValue(t, frames[3], 6)
}

func assertFrameValue(t *testing.T, frame json.RawMessage, expected int32) {
	var val int32
	err := json.Unmarshal(frame, &val)
	require.NoError(t, err)
	assert.Equal(t, expected, val)
}
