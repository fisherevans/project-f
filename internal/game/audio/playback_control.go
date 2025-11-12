package audio

import (
	"math"
	"sync"
	"time"

	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/rs/zerolog/log"
)

// controlCommand represents the desired state for a sound
type controlCommand struct {
	// Fade adjustment control (for pause/resume/start/stop fades only)
	fadeFrom      float64
	fadeTo        float64
	fadeStartTime time.Time     // When fade started
	fadeDuration  time.Duration // How long to fade

	// Actions to perform after fade completes (or immediately if duration=0)
	thenStop  bool
	thenPause bool
}

// PlaybackControl allows controlling a playing sound (stopping, adjusting volume, etc.)
type PlaybackControl struct {
	system       *System
	volumeEffect *effects.Volume
	stoppable    *stoppableStreamer

	// Worker state
	mu            sync.Mutex
	workerCommand *controlCommand
	workerDone    chan struct{}
	workerStarted bool

	playbackVolume float64 // The target volume set by SetVolume
	fadeVolume     float64 // Temporary adjustment for fades (1 = no adjustment)
}

// startWorker starts the control worker goroutine if not already running
func (s *PlaybackControl) startWorker() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.workerStarted {
		return
	}
	s.workerStarted = true
	s.workerDone = make(chan struct{})
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
		defer ticker.Stop()
		defer close(s.workerDone)
		for {
			<-ticker.C
			s.workOnTick()
		}
	}()
}

// worker is the main control loop that processes commands
func (s *PlaybackControl) workOnTick() {
	// Check if sound finished naturally
	if !s.stoppable.IsPlaying() {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.workerCommand == nil {
		return
	}

	cmd := s.workerCommand
	now := time.Now()
	elapsed := now.Sub(cmd.fadeStartTime)

	// Check if fade is complete (or immediate if duration=0)
	fadeComplete := elapsed >= cmd.fadeDuration
	if fadeComplete {
		// Set final fade adjustment
		s.fadeVolume = cmd.fadeTo
		s.updateStreamerVolume()

		// Execute post-fade actions
		if cmd.thenPause && !s.stoppable.IsPaused() {
			s.stoppable.Pause()
		}
		if cmd.thenStop {
			s.stoppable.Stop()
			return
		}

		log.Info().Any("cmd", s.workerCommand).Msg("sound control work done")
		// Clear workerCommand
		s.workerCommand = nil
	} else {
		// Interpolate fade adjustment with exponential curve (sounds more natural for dB)
		progress := float64(elapsed) / float64(cmd.fadeDuration)
		// Use exponential interpolation for smoother perceived volume change
		s.fadeVolume = cmd.fadeFrom + (cmd.fadeTo-cmd.fadeFrom)*progress
		s.updateStreamerVolume()
	}
}

// StopImmediately stops the sound immediately without fading (may cause audio pops)
func (s *PlaybackControl) StopImmediately() {
	if s == nil || s.stoppable == nil {
		return
	}
	s.stoppable.Stop()
}

// Stop fades out the sound over 0.1s and then stops it
func (s *PlaybackControl) Stop() {
	s.FadeOutAndStop(defaultFadeDuration)
}

// FadeOutAndStop fades out using adjustment, then stops playback
func (s *PlaybackControl) FadeOutAndStop(duration time.Duration) {
	if s == nil || s.volumeEffect == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workerCommand = &controlCommand{
		fadeFrom:      s.fadeVolume,
		fadeTo:        0,
		fadeStartTime: time.Now(),
		fadeDuration:  duration,
		thenStop:      true,
	}
}

func (s *PlaybackControl) PauseImmediately() {
	if s == nil || s.stoppable == nil {
		return
	}
	s.stoppable.Pause()
}

func (s *PlaybackControl) Pause() {
	s.FadeOutAndPause(defaultFadeDuration)
}

func (s *PlaybackControl) FadeOutAndPause(duration time.Duration) {
	if s == nil || s.volumeEffect == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workerCommand = &controlCommand{
		fadeFrom:      s.fadeVolume,
		fadeTo:        0,
		fadeStartTime: time.Now(),
		fadeDuration:  duration,
		thenPause:     true,
	}
}

func (s *PlaybackControl) ResumeImmediately() {
	if s == nil || s.stoppable == nil {
		return
	}
	s.fadeVolume = 1
	s.updateStreamerVolume()
	s.stoppable.Resume()
}

func (s *PlaybackControl) Resume() {
	s.ResumeAndFadeIn(defaultFadeDuration)
}

func (s *PlaybackControl) ResumeAndFadeIn(duration time.Duration) {
	if s == nil || s.volumeEffect == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workerCommand = &controlCommand{
		fadeFrom:      0,
		fadeTo:        1,
		fadeStartTime: time.Now(),
		fadeDuration:  duration,
	}
	s.fadeVolume = 0
	s.updateStreamerVolume()
	s.stoppable.Resume()
}

// SetVolume immediately sets the base volume in dB (e.g., 0 = normal, -6 = half, +6 = double)
// This respects any ongoing fade adjustments (pause/resume fades)
func (s *PlaybackControl) SetVolume(volume float64) {
	if s == nil || s.volumeEffect == nil {
		return
	}
	s.mu.Lock()
	s.playbackVolume = volume
	s.mu.Unlock()
	s.updateStreamerVolume()
}

func (s *PlaybackControl) updateStreamerVolume() {
	speaker.Lock()
	finalVolume := s.playbackVolume * s.fadeVolume
	s.volumeEffect.Volume = math.Log(finalVolume) / math.Log(s.volumeEffect.Base)
	speaker.Unlock()
}

// IsPlaying returns true if the sound is still playing (not stopped and not finished)
func (s *PlaybackControl) IsPlaying() bool {
	if s == nil || s.stoppable == nil {
		return false
	}
	return s.stoppable.IsPlaying()
}

// Position returns the current playback position as a duration
func (s *PlaybackControl) Position() time.Duration {
	if s == nil || s.stoppable == nil {
		return 0
	}
	samples := s.stoppable.Position()
	return time.Duration(float64(samples) / float64(targetSR) * float64(time.Second))
}

// Duration returns the total duration of the sound (0 if unknown/infinite)
func (s *PlaybackControl) Duration() time.Duration {
	if s == nil || s.stoppable == nil {
		return 0
	}
	samples := s.stoppable.Length()
	if samples == 0 {
		return 0
	}
	return time.Duration(float64(samples) / float64(targetSR) * float64(time.Second))
}

// TimeRemaining returns how much time is left in the current iteration.
// For looping sounds, this only returns time left in the current loop iteration.
func (s *PlaybackControl) TimeRemaining() time.Duration {
	if s == nil || s.stoppable == nil {
		return 0
	}

	duration := s.Duration()
	if duration == 0 {
		return 0
	}

	// Time remaining in current iteration
	currentRemaining := duration - s.Position()
	if currentRemaining < 0 {
		currentRemaining = 0
	}

	return currentRemaining
}

// IsPaused returns true if the sound is currently paused
func (s *PlaybackControl) IsPaused() bool {
	if s == nil || s.stoppable == nil {
		return false
	}
	return s.stoppable.IsPaused()
}
