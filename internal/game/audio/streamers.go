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
	s        beep.Streamer
	stopped  bool
	paused   bool
	finished bool
	totalLen int // total samples in the buffer (if known, 0 if unknown/infinite)
	position int // current sample position
	mu       sync.Mutex
}

func (ss *stoppableStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	ss.mu.Lock()
	stopped := ss.stopped
	paused := ss.paused
	ss.mu.Unlock()

	if stopped {
		ss.mu.Lock()
		ss.finished = true
		ss.mu.Unlock()
		return 0, false // Signal mixer to remove us
	}
	
	// When paused, output silence but keep streaming
	if paused {
		for i := range samples {
			samples[i][0] = 0
			samples[i][1] = 0
		}
		return len(samples), true
	}
	
	n, ok = ss.s.Stream(samples)
	
	ss.mu.Lock()
	if ok {
		ss.position += n
	} else {
		ss.finished = true
	}
	ss.mu.Unlock()
	
	return n, ok
}

func (ss *stoppableStreamer) Err() error {
	return ss.s.Err()
}

func (ss *stoppableStreamer) Stop() {
	ss.mu.Lock()
	ss.stopped = true
	ss.mu.Unlock()
}

// IsPlaying returns true if the sound is still playing (not stopped and not finished)
func (ss *stoppableStreamer) IsPlaying() bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return !ss.stopped && !ss.finished
}

// Position returns the current playback position in samples
func (ss *stoppableStreamer) Position() int {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.position
}

// Length returns the total length in samples (0 if unknown/infinite)
func (ss *stoppableStreamer) Length() int {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.totalLen
}

// Pause pauses playback (outputs silence but continues streaming)
func (ss *stoppableStreamer) Pause() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.paused = true
}

// Resume resumes playback after pause
func (ss *stoppableStreamer) Resume() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.paused = false
}

// IsPaused returns true if the sound is currently paused
func (ss *stoppableStreamer) IsPaused() bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.paused
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
