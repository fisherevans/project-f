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

// PlaySFX plays a cached SFX on the SFX bus (polyphonic). `gainDB` adjusts per call, e.g., -6 for softer.
func (a *System) PlaySFX(name string, gainDB float64) {
	a.PlaySoundOnBus(name, a.Buses.SFX, gainDB)
}

// PlayUI plays a one-shot UI sound on UI bus
func (a *System) PlayUI(name string) { a.PlaySoundOnBus(name, a.Buses.UI, 0) }

// PlaySoundOnBus plays a cached SFX on a specific bus
func (a *System) PlaySoundOnBus(name string, bus *Bus, gainDB float64) {
	a.mu.Lock()
	buf := a.cache[name]
	a.mu.Unlock()
	if buf == nil {
		log.Warn().Str("name", name).Msgf("unable to find cached audio buffer")
		return
	}
	var s beep.Streamer = buf.Streamer(0, buf.Len())
	// Optional per-instance volume (fast)
	if gainDB != 0 {
		s = &effects.Volume{Streamer: s, Base: 2, Volume: gainDB}
	}
	speaker.Lock()
	bus.mix.Add(s)
	speaker.Unlock()
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
