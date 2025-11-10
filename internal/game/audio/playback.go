package audio

import (
	"os"
	"path/filepath"
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
}

// Stopper allows stopping a sound with a fade-out to avoid pops
type Stopper struct {
	sys       *System
	vol       *effects.Volume
	stoppable *stoppableStreamer
}

// Stop fades out the sound over 0.1s and then stops it
func (s *Stopper) Stop() {
	if s == nil || s.stoppable == nil {
		return
	}
	go func() {
		speaker.Lock()
		currentVol := s.vol.Volume
		speaker.Unlock()
		s.sys.fade(s.vol, currentVol, -60, fadeOutDuration)
		s.stoppable.Stop()
	}()
}

// PlaySFX plays a cached SFX on the SFX bus (polyphonic). `gainDB` adjusts per call, e.g., -6 for softer.
// Returns a Stopper that can fade out and stop the sound early.
func (a *System) PlaySFX(name string, gainDB float64) *Stopper {
	return a.PlaySoundOnBus(name, a.Buses.SFX, gainDB, nil)
}

// PlayUI plays a one-shot UI sound on UI bus. Returns a Stopper.
func (a *System) PlayUI(name string) *Stopper { return a.PlaySoundOnBus(name, a.Buses.UI, 0, nil) }

// PlaySilenceOnBus plays silence for a duration on a bus, then calls onComplete.
func (a *System) PlaySilenceOnBus(bus *Bus, duration time.Duration, onComplete func()) *Stopper {
	sr := beep.SampleRate(targetSR)
	silence := beep.Silence(sr.N(duration))

	// Add completion callback
	s := beep.Seq(silence, beep.Callback(onComplete))

	// Wrap in stoppable
	stoppable := &stoppableStreamer{s: s}

	speaker.Lock()
	bus.mix.Add(stoppable)
	speaker.Unlock()

	return &Stopper{
		sys:       a,
		vol:       nil, // no volume control for silence
		stoppable: stoppable,
	}
}

// PlaySoundOnBus plays a cached SFX on a specific bus with optional pitch/speed modulation.
// opts can be nil for default playback. Returns a Stopper.
func (a *System) PlaySoundOnBus(name string, bus *Bus, gainDB float64, opts *PlaybackOptions) *Stopper {
	return a.PlaySoundOnBusWithCallback(name, bus, gainDB, opts, nil)
}

// PlaySoundOnBusWithCallback plays a sound and calls onComplete when it finishes naturally (not when stopped).
func (a *System) PlaySoundOnBusWithCallback(name string, bus *Bus, gainDB float64, opts *PlaybackOptions, onComplete func()) *Stopper {
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
	var s beep.Streamer = buf.Streamer(0, buf.Len())

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
	stoppable := &stoppableStreamer{s: s}

	// Wrap in volume control for both initial gain and fade-out
	vol := &effects.Volume{
		Streamer: stoppable,
		Base:     2,
		Volume:   gainDB,
	}

	speaker.Lock()
	bus.mix.Add(vol)
	speaker.Unlock()

	return &Stopper{
		sys:       a,
		vol:       vol,
		stoppable: stoppable,
	}
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
