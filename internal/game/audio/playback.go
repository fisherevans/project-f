package audio

import (
	"path/filepath"
	"time"

	"fisherevans.com/project/f/assets"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/rs/zerolog/log"
)

const defaultFadeDuration = 250 * time.Millisecond

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

	OnComplete func()
}

// PlaySFX plays a cached SFX on the SFX bus (polyphonic)
// Returns a Stopper that can fade out and stop the sound early.
func (a *System) PlaySFX(name string, volume float64) *PlaybackControl {
	return a.PlaySoundOnBus(name, a.Buses.SFX, volume, nil)
}

// PlayUI plays a one-shot UI sound on UI bus. Returns a Stopper.
func (a *System) PlayUI(name string) *PlaybackControl {
	return a.PlaySoundOnBus(name, a.Buses.UI, 1, nil)
}

// PlaySoundOnBus plays a cached SFX on a specific bus with optional pitch/speed modulation.
// opts can be nil for default playback. Returns a Stopper.
func (a *System) PlaySoundOnBus(name string, bus *Bus, volume float64, opts *PlaybackOptions) *PlaybackControl {
	a.mu.Lock()
	buf := a.cache[name]
	a.mu.Unlock()
	if buf == nil {
		log.Warn().Str("name", name).Msgf("unable to find cached audio buffer")
		if opts.OnComplete != nil {
			opts.OnComplete()
		}
		return nil
	}
	return a.PlayOnBus(buf, bus, volume, opts)
}

// PlayMusic starts looping music from file; returns a stop func.
func (a *System) PlayMusic(assetPath string, volume float64, opts *PlaybackOptions) *PlaybackControl {
	f, err := assets.FS.Open(assetPath)
	defer f.Close()
	if err != nil {
		log.Error().Str("assetPath", assetPath).Err(err).Msg("failed to open music file")
		return nil
	}
	src, fmt, err := openDecode(f, filepath.Ext(assetPath))
	if err != nil {
		log.Error().Str("assetPath", assetPath).Err(err).Msg("failed to decode music file")
		return nil
	}
	defer src.Close()
	return a.PlayOnBus(newSourceBuffer(src, fmt), a.Buses.Music, volume, opts)
}

func (a *System) PlayOnBus(buf *beep.Buffer, bus *Bus, volume float64, opts *PlaybackOptions) *PlaybackControl {
	if volume < 0 {
		log.Warn().Float64("volume", volume).Any("opts", opts).Msg("volume is less than 0")
		volume = 0
	}
	if opts == nil {
		opts = &PlaybackOptions{}
	}

	// Get the buffer streamer (StreamSeeker for looping)
	bufStreamer := buf.Streamer(0, buf.Len())

	// Apply looping FIRST (requires StreamSeeker)
	var s beep.Streamer
	if opts.Loop {
		s = beep.Loop(-1, bufStreamer)
	} else {
		s = bufStreamer
	}

	// Apply pitch/speed modulation if requested (coupled via resampling)
	if opts.Speed > 0 && opts.Speed != 1.0 {
		originalRate := beep.SampleRate(targetSR)
		modulatedRate := beep.SampleRate(float64(targetSR) * opts.Speed)
		s = beep.Resample(4, originalRate, modulatedRate, s)
	}

	// Add completion callback if provided
	if opts.OnComplete != nil {
		s = beep.Seq(s, beep.Callback(opts.OnComplete))
	}

	// Wrap in stoppable so we can stop it mid-playback
	stoppable := &stoppableStreamer{
		s:        s,
		totalLen: buf.Len(),
	}

	// Wrap in volume control
	volumeEffect := &effects.Volume{
		Streamer: stoppable,
		Base:     2,
		Volume:   0, // Start at target volume
	}

	ctrl := &PlaybackControl{
		system:         a,
		volumeEffect:   volumeEffect,
		stoppable:      stoppable,
		fadeVolume:     1,
		playbackVolume: volume,
	}

	// Start fade-in if requested (using worker-based fade)
	if opts.FadeIn > 0 {
		ctrl.fadeVolume = 0
		// Fade adjustment from silence to 0 (= base volume)
		ctrl.workerCommand = &controlCommand{
			fadeFrom:      0,
			fadeTo:        1,
			fadeStartTime: time.Now(),
			fadeDuration:  opts.FadeIn,
		}
	}

	speaker.Lock()
	bus.mix.Add(volumeEffect)
	speaker.Unlock()

	ctrl.updateStreamerVolume()
	ctrl.startWorker()
	return ctrl
}
