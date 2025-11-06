package audio

import (
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
)

const targetSR = 48000 // Target sample rate for all audio

// Bus represents an audio mixing bus with volume control
type Bus struct {
	mix *beep.Mixer     // add/remove sounds here
	vol *effects.Volume // runtime dB control
}

// System manages the audio graph with multiple buses
type System struct {
	mu sync.Mutex

	Master    *effects.Volume
	MasterTee *TeeStreamer // Tee for capturing audio samples
	Buses     struct {
		SFX, UI, Music, Amb *Bus
	}

	// Small SFX are kept decoded for low-latency polyphony.
	cache map[string]*beep.Buffer
}

// newSystem creates and initializes a newSystem audio system
func newSystem() (*System, error) {
	// Init speaker with larger buffer for stability
	sr := beep.SampleRate(targetSR)
	if err := speaker.Init(sr, sr.N(time.Second/10)); err != nil {
		return nil, err
	}

	mkBus := func() *Bus {
		m := &beep.Mixer{}
		m.Add(&silent{})                                      // keep-alive
		v := &effects.Volume{Streamer: m, Base: 2, Volume: 0} // 0 dB
		return &Bus{mix: m, vol: v}
	}

	sys := &System{cache: make(map[string]*beep.Buffer)}
	sys.Buses.SFX = mkBus()
	sys.Buses.UI = mkBus()
	sys.Buses.Music = mkBus()
	sys.Buses.Amb = mkBus()

	masterMix := &beep.Mixer{}
	masterMix.Add(sys.Buses.SFX.vol, sys.Buses.UI.vol, sys.Buses.Music.vol, sys.Buses.Amb.vol)

	sys.Master = &effects.Volume{Streamer: masterMix, Base: 2, Volume: 0} // master gain
	
	// Wrap master in a tee for audio capture
	sys.MasterTee = NewTeeStreamer(sys.Master)
	
	// Start the persistent graph once.
	speaker.Play(sys.MasterTee)
	return sys, nil
}

// SetBusGain sets the volume of a specific bus in dB
func (a *System) SetBusGain(bus *Bus, dB float64) {
	speaker.Lock()
	bus.vol.Volume = dB
	speaker.Unlock()
}

// SetMaster sets the master volume in dB
func (a *System) SetMaster(dB float64) {
	speaker.Lock()
	a.Master.Volume = dB
	speaker.Unlock()
}
