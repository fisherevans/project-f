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
