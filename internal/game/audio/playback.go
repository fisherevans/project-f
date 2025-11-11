package audio

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/rs/zerolog/log"
)

const SilenceGainDbThreshold = -40.0
const fadeOutDuration = 100 * time.Millisecond

// PlaybackOptions contains optional modulation parameters for sound playback.
type PlaybackOptions struct {
	// Speed multiplies the playback rate (affects both pitch and speed together).
	// 1.0 = normal, 2.0 = double speed/pitch, 0.5 = half speed/pitch.
	// Values must be > 0. Typical range: 0.5 to 2.0
	Speed float64

	// Loop specifies whether to loop the sound infinitely
	Loop bool

	// FadeIn specifies the duration to fade in from silence.
	// 0 = no fade (default), starts at target volume immediately
	// > 0 = fade in from -60dB to target volume over this duration
	FadeIn time.Duration
}

// controlCommand represents the desired state for a sound
type controlCommand struct {
	// Fade adjustment control (for pause/resume/start/stop fades only)
	fadeAdjTarget   float64       // Target fade adjustment
	fadeAdjStart    float64       // Starting fade adjustment
	fadeStartTime   time.Time     // When fade started
	fadeDuration    time.Duration // How long to fade

	// Actions to perform after fade completes (or immediately if duration=0)
	thenStop  bool
	thenPause bool
}

// SoundControl allows controlling a playing sound (stopping, adjusting volume, etc.)
type SoundControl struct {
	sys       *System
	vol       *effects.Volume
	stoppable *stoppableStreamer

	// Worker state
	mu            sync.Mutex
	cmd           *controlCommand
	workerDone    chan struct{}
	workerStarted bool
	
	// Volume management: baseVolume + fadeAdjustment = actual volume
	baseVolume     float64 // The target volume set by SetVolume
	fadeAdjustment float64 // Temporary adjustment for fades (0 = no adjustment)
}

// startWorker starts the control worker goroutine if not already running
func (s *SoundControl) startWorker() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.workerStarted {
		return
	}
	s.workerStarted = true
	s.workerDone = make(chan struct{})
	go s.worker()
}

// worker is the main control loop that processes commands
func (s *SoundControl) worker() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()
	defer close(s.workerDone)

	for {
		<-ticker.C

		// Check if sound finished naturally
		if !s.stoppable.IsPlaying() {
			return
		}

		s.mu.Lock()

		if s.cmd == nil {
			s.mu.Unlock()
			continue
		}

		cmd := s.cmd
		now := time.Now()
		elapsed := now.Sub(cmd.fadeStartTime)

		// Check if fade is complete (or immediate if duration=0)
		fadeComplete := elapsed >= cmd.fadeDuration

		if fadeComplete {
			// Set final fade adjustment
			s.fadeAdjustment = cmd.fadeAdjTarget
			actualVol := s.baseVolume + s.fadeAdjustment
			
			speaker.Lock()
			s.vol.Volume = actualVol
			speaker.Unlock()

			log.Debug().
				Float64("fadeAdj", s.fadeAdjustment).
				Float64("baseVol", s.baseVolume).
				Float64("actualVol", actualVol).
				Bool("thenPause", cmd.thenPause).
				Bool("thenStop", cmd.thenStop).
				Msg("Fade complete")

			// Execute post-fade actions
			if cmd.thenStop {
				s.mu.Unlock()
				s.stoppable.Stop()
				return
			}
			if cmd.thenPause && !s.stoppable.IsPaused() {
				log.Debug().Msg("Executing pause after fade")
				s.stoppable.Pause()
			}

			// Clear command
			s.cmd = nil
		} else {
			// Interpolate fade adjustment with exponential curve (sounds more natural for dB)
			progress := float64(elapsed) / float64(cmd.fadeDuration)
			// Use exponential interpolation for smoother perceived volume change
			s.fadeAdjustment = cmd.fadeAdjStart + (cmd.fadeAdjTarget-cmd.fadeAdjStart)*progress
			actualVol := s.baseVolume + s.fadeAdjustment
			
			speaker.Lock()
			s.vol.Volume = actualVol
			speaker.Unlock()
		}

		s.mu.Unlock()
	}
}

// Stop fades out the sound over 0.1s and then stops it
func (s *SoundControl) Stop() {
	s.FadeOutAndStop(fadeOutDuration)
}

// FadeOutAndStop fades out using adjustment, then stops playback
func (s *SoundControl) FadeOutAndStop(duration time.Duration) {
	if s == nil || s.vol == nil {
		return
	}
	s.startWorker()

	s.mu.Lock()
	currentFadeAdj := s.fadeAdjustment
	// Fade adjustment to silence (relative to base volume)
	targetFadeAdj := -60 - s.baseVolume
	
	s.cmd = &controlCommand{
		fadeAdjTarget: targetFadeAdj,
		fadeAdjStart:  currentFadeAdj,
		fadeStartTime: time.Now(),
		fadeDuration:  duration,
		thenStop:      true,
	}
	s.mu.Unlock()
}

// StopImmediately stops the sound immediately without fading (may cause audio pops)
func (s *SoundControl) StopImmediately() {
	if s == nil || s.stoppable == nil {
		return
	}
	s.stoppable.Stop()
}

// SetVolume immediately sets the base volume in dB (e.g., 0 = normal, -6 = half, +6 = double)
// This respects any ongoing fade adjustments (pause/resume fades)
func (s *SoundControl) SetVolume(gainDB float64) {
	if s == nil || s.vol == nil {
		return
	}
	
	s.mu.Lock()
	s.baseVolume = gainDB
	actualVolume := gainDB + s.fadeAdjustment
	s.mu.Unlock()
	
	speaker.Lock()
	s.vol.Volume = actualVolume
	speaker.Unlock()
}


// GetVolume returns the current volume in dB
func (s *SoundControl) GetVolume() float64 {
	if s == nil || s.vol == nil {
		return 0
	}
	speaker.Lock()
	defer speaker.Unlock()
	return s.vol.Volume
}

// IsPlaying returns true if the sound is still playing (not stopped and not finished)
func (s *SoundControl) IsPlaying() bool {
	if s == nil || s.stoppable == nil {
		return false
	}
	return s.stoppable.IsPlaying()
}

// Position returns the current playback position as a duration
func (s *SoundControl) Position() time.Duration {
	if s == nil || s.stoppable == nil {
		return 0
	}
	samples := s.stoppable.Position()
	return time.Duration(float64(samples) / float64(targetSR) * float64(time.Second))
}

// Duration returns the total duration of the sound (0 if unknown/infinite)
func (s *SoundControl) Duration() time.Duration {
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
func (s *SoundControl) TimeRemaining() time.Duration {
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

// Pause pauses playback immediately (outputs silence)
func (s *SoundControl) Pause() {
	if s == nil || s.stoppable == nil {
		return
	}
	s.stoppable.Pause()
}

// Resume resumes playback immediately after pause
func (s *SoundControl) Resume() {
	if s == nil || s.stoppable == nil {
		return
	}
	s.stoppable.Resume()
}

// PauseWithFade fades out over the given duration, then pauses
// Uses fade adjustment so SetVolume calls don't interfere
func (s *SoundControl) PauseWithFade(duration time.Duration) {
	if s == nil || s.vol == nil {
		return
	}
	s.startWorker()

	s.mu.Lock()
	currentFadeAdj := s.fadeAdjustment
	// Fade adjustment from current to -60 (relative to base volume)
	targetFadeAdj := -60 - s.baseVolume
	
	log.Debug().
		Float64("baseVol", s.baseVolume).
		Float64("currentFadeAdj", currentFadeAdj).
		Float64("targetFadeAdj", targetFadeAdj).
		Float64("duration_ms", float64(duration.Milliseconds())).
		Msg("PauseWithFade called")
	
	s.cmd = &controlCommand{
		fadeAdjTarget: targetFadeAdj,
		fadeAdjStart:  currentFadeAdj,
		fadeStartTime: time.Now(),
		fadeDuration:  duration,
		thenPause:     true,
	}
	s.mu.Unlock()
}

// ResumeWithFade resumes playback immediately and fades in to the base volume
// Ignores the targetDB parameter and uses the current base volume instead
func (s *SoundControl) ResumeWithFade(targetDB float64, duration time.Duration) {
	if s == nil || s.vol == nil {
		return
	}
	
	// Resume immediately
	s.stoppable.Resume()
	
	s.mu.Lock()
	// Start with max fade adjustment (silence), fade to 0 (no adjustment = base volume)
	s.fadeAdjustment = -60 - s.baseVolume
	currentFadeAdj := s.fadeAdjustment
	
	log.Debug().
		Float64("baseVol", s.baseVolume).
		Float64("startFadeAdj", currentFadeAdj).
		Float64("duration_ms", float64(duration.Milliseconds())).
		Msg("ResumeWithFade called")
	
	s.cmd = &controlCommand{
		fadeAdjTarget: 0, // Fade adjustment back to 0 (= base volume)
		fadeAdjStart:  currentFadeAdj,
		fadeStartTime: time.Now(),
		fadeDuration:  duration,
	}
	s.mu.Unlock()
	
	// Set initial volume (base + fade adjustment = silence)
	speaker.Lock()
	s.vol.Volume = s.baseVolume + currentFadeAdj
	speaker.Unlock()
}

// IsPaused returns true if the sound is currently paused
func (s *SoundControl) IsPaused() bool {
	if s == nil || s.stoppable == nil {
		return false
	}
	return s.stoppable.IsPaused()
}

// CancelFade cancels any fade currently in progress, leaving volume at current value
func (s *SoundControl) CancelFade() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.cmd = nil
	s.mu.Unlock()
}

// PlaySFX plays a cached SFX on the SFX bus (polyphonic). `gainDB` adjusts per call, e.g., -6 for softer.
// Returns a Stopper that can fade out and stop the sound early.
func (a *System) PlaySFX(name string, gainDB float64) *SoundControl {
	return a.PlaySoundOnBus(name, a.Buses.SFX, gainDB, nil)
}

// PlayUI plays a one-shot UI sound on UI bus. Returns a Stopper.
func (a *System) PlayUI(name string) *SoundControl { return a.PlaySoundOnBus(name, a.Buses.UI, 0, nil) }

// PlaySilenceOnBus plays silence for a duration on a bus, then calls onComplete.
func (a *System) PlaySilenceOnBus(bus *Bus, duration time.Duration, onComplete func()) *SoundControl {
	sr := beep.SampleRate(targetSR)
	silence := beep.Silence(sr.N(duration))

	// Add completion callback
	s := beep.Seq(silence, beep.Callback(onComplete))

	// Wrap in stoppable
	stoppable := &stoppableStreamer{s: s}

	speaker.Lock()
	bus.mix.Add(stoppable)
	speaker.Unlock()

	return &SoundControl{
		sys:       a,
		vol:       nil, // no volume control for silence
		stoppable: stoppable,
	}
}

// PlaySoundOnBus plays a cached SFX on a specific bus with optional pitch/speed modulation.
// opts can be nil for default playback. Returns a Stopper.
func (a *System) PlaySoundOnBus(name string, bus *Bus, gainDB float64, opts *PlaybackOptions) *SoundControl {
	return a.PlaySoundOnBusWithCallback(name, bus, gainDB, opts, nil)
}

// PlaySoundOnBusWithCallback plays a sound and calls onComplete when it finishes naturally (not when stopped).
func (a *System) PlaySoundOnBusWithCallback(name string, bus *Bus, gainDB float64, opts *PlaybackOptions, onComplete func()) *SoundControl {
	if gainDB < SilenceGainDbThreshold {
		return nil
	}
	a.mu.Lock()
	buf := a.cache[name]
	a.mu.Unlock()
	if buf == nil {
		log.Warn().Str("name", name).Msgf("unable to find cached audio buffer")
		if onComplete != nil {
			onComplete()
		}
		return nil
	}

	// Get the buffer streamer (StreamSeeker for looping)
	bufStreamer := buf.Streamer(0, buf.Len())
	
	// Apply looping FIRST (requires StreamSeeker)
	var s beep.Streamer
	if opts != nil && opts.Loop {
		s = beep.Loop(-1, bufStreamer)
	} else {
		s = bufStreamer
	}

	// Apply pitch/speed modulation if requested (coupled via resampling)
	if opts != nil && opts.Speed > 0 && opts.Speed != 1.0 {
		originalRate := beep.SampleRate(targetSR)
		modulatedRate := beep.SampleRate(float64(targetSR) * opts.Speed)
		s = beep.Resample(4, originalRate, modulatedRate, s)
	}

	// Add completion callback if provided
	if onComplete != nil {
		s = beep.Seq(s, beep.Callback(onComplete))
	}

	// Wrap in stoppable so we can stop it mid-playback
	stoppable := &stoppableStreamer{
		s:        s,
		totalLen: buf.Len(),
	}

	// Wrap in volume control
	vol := &effects.Volume{
		Streamer: stoppable,
		Base:     2,
		Volume:   gainDB, // Start at target volume
	}

	speaker.Lock()
	bus.mix.Add(vol)
	speaker.Unlock()

	ctrl := &SoundControl{
		sys:        a,
		vol:        vol,
		stoppable:  stoppable,
		baseVolume: gainDB, // Initialize with the starting gain
	}
	
	// Start fade-in if requested (using worker-based fade)
	if opts != nil && opts.FadeIn > 0 {
		// Set initial fade adjustment to silence
		fadeAdj := -60 - gainDB
		
		ctrl.mu.Lock()
		ctrl.fadeAdjustment = fadeAdj
		ctrl.mu.Unlock()
		
		speaker.Lock()
		vol.Volume = gainDB + fadeAdj // Start silent
		speaker.Unlock()
		
		log.Debug().
			Float64("gainDB", gainDB).
			Float64("fadeAdj", fadeAdj).
			Float64("duration_ms", float64(opts.FadeIn.Milliseconds())).
			Msg("Starting fade-in")
		
		// Fade adjustment from silence to 0 (= base volume)
		ctrl.startWorker()
		ctrl.mu.Lock()
		ctrl.cmd = &controlCommand{
			fadeAdjTarget: 0, // Fade to no adjustment
			fadeAdjStart:  fadeAdj,
			fadeStartTime: time.Now(),
			fadeDuration:  opts.FadeIn,
		}
		ctrl.mu.Unlock()
	}
	
	return ctrl
}

// PlayMusic starts looping music from file; returns a stop func.
func (a *System) PlayMusic(path string, fadeIn time.Duration) (stop func()) {
	// TODO: Update to use embedded resources
	f, err := os.Open(path)
	if err != nil {
		log.Error().Str("path", path).Err(err).Msg("failed to open music file")
		return func() {}
	}

	src, fmt, err := openDecode(f, filepath.Ext(path))
	if err != nil {
		f.Close()
		log.Error().Str("path", path).Err(err).Msg("failed to decode music file")
		return func() {}
	}

	loop, err := beep.Loop2(src) // infinite
	if err != nil {
		log.Error().Str("path", path).Err(err).Msg("failed to loop music")
		src.Close()
		return func() {}
	}

	// Resample if needed
	str := beep.Streamer(loop)
	if int(fmt.SampleRate) != targetSR {
		str = beep.Resample(5, fmt.SampleRate, beep.SampleRate(targetSR), str)
	}

	// Wrap in stoppable so we can remove it from mixer
	stoppable := &stoppableStreamer{s: str}

	// start low, fade in
	vol := &effects.Volume{
		Streamer: stoppable,
		Base:     2,
		Volume:   -12,
	}

	// Add to music bus
	speaker.Lock()
	a.Buses.Music.mix.Add(vol)
	speaker.Unlock()

	if fadeIn > 0 {
		go a.fade(vol, -12, 0, fadeIn)
	}

	// Provide a stop with fade-out and resource close.
	return func() {
		go func() {
			a.fade(vol, vol.Volume, -24, 300*time.Millisecond)
			stoppable.Stop()                  // Causes Stream() to return (0, false), triggering auto-removal
			time.Sleep(50 * time.Millisecond) // Give mixer time to remove it
			src.Close()
		}()
	}
}

// fade performs a linear volume fade in dB
func (a *System) fade(v *effects.Volume, fromDB, toDB float64, d time.Duration) {
	steps := 30
	delay := d / time.Duration(steps)
	for i := 0; i <= steps; i++ {
		x := fromDB + (toDB-fromDB)*float64(i)/float64(steps)
		speaker.Lock()
		v.Volume = x
		speaker.Unlock()
		time.Sleep(delay)
	}
}

// DuckMusic temporarily lowers music volume while SFX plays
func (a *System) DuckMusic(amountDB float64, forDur time.Duration) {
	go func() {
		speaker.Lock()
		old := a.Buses.Music.vol.Volume
		a.Buses.Music.vol.Volume = old - amountDB
		speaker.Unlock()
		time.Sleep(forDur)
		speaker.Lock()
		a.Buses.Music.vol.Volume = old
		speaker.Unlock()
	}()
}
