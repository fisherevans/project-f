//go:build !js

package audio

import (
	"math"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
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

	// Async loading state (for music that loads in background)
	loading       bool
	loadingFailed bool
	pendingStop   bool // Stop requested during loading
	pendingPause  bool // Pause requested during loading
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
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.loading {
		s.pendingStop = true
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	if s.stoppable == nil {
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
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loading {
		s.pendingStop = true
		return
	}
	if s.volumeEffect == nil {
		return
	}
	s.workerCommand = &controlCommand{
		fadeFrom:      s.fadeVolume,
		fadeTo:        0,
		fadeStartTime: time.Now(),
		fadeDuration:  duration,
		thenStop:      true,
	}
}

func (s *PlaybackControl) PauseImmediately() {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.loading {
		s.pendingPause = true
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	if s.stoppable == nil {
		return
	}
	s.stoppable.Pause()
}

func (s *PlaybackControl) Pause() {
	s.FadeOutAndPause(defaultFadeDuration)
}

func (s *PlaybackControl) FadeOutAndPause(duration time.Duration) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loading {
		s.pendingPause = true
		return
	}
	if s.volumeEffect == nil {
		return
	}
	s.workerCommand = &controlCommand{
		fadeFrom:      s.fadeVolume,
		fadeTo:        0,
		fadeStartTime: time.Now(),
		fadeDuration:  duration,
		thenPause:     true,
	}
}

func (s *PlaybackControl) ResumeImmediately() {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.loading {
		s.pendingPause = false
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	if s.stoppable == nil {
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
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loading {
		s.pendingPause = false
		return
	}
	if s.volumeEffect == nil {
		return
	}
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
	if s == nil {
		return
	}
	s.mu.Lock()
	s.playbackVolume = volume
	if s.loading {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	if s.volumeEffect == nil {
		return
	}
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
	if s == nil {
		return false
	}
	s.mu.Lock()
	if s.loading {
		s.mu.Unlock()
		return true // Still loading, consider it "playing"
	}
	s.mu.Unlock()
	if s.stoppable == nil {
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
	if s == nil {
		return false
	}
	s.mu.Lock()
	if s.loading {
		pending := s.pendingPause
		s.mu.Unlock()
		return pending
	}
	s.mu.Unlock()
	if s.stoppable == nil {
		return false
	}
	return s.stoppable.IsPaused()
}

// finishAsyncLoad completes async loading by initializing the streamer and applying pending operations
func (s *PlaybackControl) finishAsyncLoad(buf *beep.Buffer, bus *Bus, opts *PlaybackOptions) {
	s.mu.Lock()

	// Check if stop was requested during loading
	if s.pendingStop {
		s.loading = false
		s.mu.Unlock()
		if opts.OnComplete != nil {
			opts.OnComplete()
		}
		return
	}

	// Capture pending state before releasing lock
	pendingPause := s.pendingPause
	s.mu.Unlock()

	// Get the buffer streamer (StreamSeeker for looping)
	bufStreamer := buf.Streamer(0, buf.Len())

	// Apply looping FIRST (requires StreamSeeker)
	var str beep.Streamer
	if opts.Loop {
		str = beep.Loop(-1, bufStreamer)
	} else {
		str = bufStreamer
	}

	// Apply pitch/speed modulation if requested (coupled via resampling)
	if opts.Speed > 0 && opts.Speed != 1.0 {
		originalRate := beep.SampleRate(targetSR)
		modulatedRate := beep.SampleRate(float64(targetSR) * opts.Speed)
		str = beep.Resample(4, originalRate, modulatedRate, str)
	}

	// Add completion callback if provided
	if opts.OnComplete != nil {
		str = beep.Seq(str, beep.Callback(opts.OnComplete))
	}

	// Wrap in stoppable so we can stop it mid-playback
	stoppable := &stoppableStreamer{
		s:        str,
		totalLen: buf.Len(),
	}

	// Calculate initial volume based on fade state
	s.mu.Lock()
	initialVolume := s.playbackVolume * s.fadeVolume
	initialVolumeDB := math.Log(initialVolume) / math.Log(2.0)
	s.mu.Unlock()

	// Wrap in volume control with correct initial volume
	volumeEffect := &effects.Volume{
		Streamer: stoppable,
		Base:     2,
		Volume:   initialVolumeDB,
	}

	// Add to bus mixer (no locks held) - streaming starts NOW
	speaker.Lock()
	bus.mix.Add(volumeEffect)
	speaker.Unlock()

	// Now acquire lock to update state
	s.mu.Lock()
	s.stoppable = stoppable
	s.volumeEffect = volumeEffect

	// Apply pending pause if requested during loading
	if pendingPause {
		s.stoppable.Pause()
	}

	// Set up fade-in if fadeVolume is 0 (indicating fade-in was requested)
	// Start time is NOW since we just added the streamer to the mixer
	if s.fadeVolume == 0 && opts.FadeInSeconds > 0 {
		s.workerCommand = &controlCommand{
			fadeFrom:      0,
			fadeTo:        1,
			fadeStartTime: time.Now(),
			fadeDuration:  time.Duration(opts.FadeInSeconds * float64(time.Second)),
		}
	}

	// Mark loading complete
	s.loading = false

	// Start worker if not already started
	if !s.workerStarted {
		s.workerStarted = true
		s.workerDone = make(chan struct{})
		s.mu.Unlock()
		go func() {
			ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
			defer ticker.Stop()
			defer close(s.workerDone)
			for {
				<-ticker.C
				s.workOnTick()
			}
		}()
	} else {
		s.mu.Unlock()
	}
}
