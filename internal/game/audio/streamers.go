package audio

import (
	"sync"

	"github.com/gopxl/beep/v2"
)

// silent is an infinite silence streamer to keep mixers alive
type silent struct{}

func (s *silent) Stream(samples [][2]float64) (n int, ok bool) {
	return len(samples), true
}
func (s *silent) Err() error { return nil }

// stoppableStreamer wraps a streamer and allows it to be stopped,
// causing it to be automatically removed from the mixer
type stoppableStreamer struct {
	s       beep.Streamer
	stopped bool
	mu      sync.Mutex
}

func (ss *stoppableStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	ss.mu.Lock()
	stopped := ss.stopped
	ss.mu.Unlock()

	if stopped {
		return 0, false // Signal mixer to remove us
	}
	return ss.s.Stream(samples)
}

func (ss *stoppableStreamer) Err() error {
	return ss.s.Err()
}

func (ss *stoppableStreamer) Stop() {
	ss.mu.Lock()
	ss.stopped = true
	ss.mu.Unlock()
}

// TeeStreamer duplicates audio samples to a callback while passing them through
type TeeStreamer struct {
	s        beep.Streamer
	callback func([][2]float64)
	mu       sync.RWMutex
}

func NewTeeStreamer(s beep.Streamer) *TeeStreamer {
	return &TeeStreamer{s: s}
}

func (t *TeeStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = t.s.Stream(samples)
	
	if n > 0 {
		t.mu.RLock()
		cb := t.callback
		t.mu.RUnlock()
		
		if cb != nil {
			// Make a copy of the samples to avoid race conditions
			samplesCopy := make([][2]float64, n)
			copy(samplesCopy, samples[:n])
			cb(samplesCopy)
		}
	}
	
	return n, ok
}

func (t *TeeStreamer) Err() error {
	return t.s.Err()
}

// SetCallback sets the function to call with duplicated samples
func (t *TeeStreamer) SetCallback(cb func([][2]float64)) {
	t.mu.Lock()
	t.callback = cb
	t.mu.Unlock()
}
