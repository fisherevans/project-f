//go:build js && wasm

package audio

import (
	"io"
	"math"
	"path/filepath"
	"sync"
	"syscall/js"
	"time"

	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/assets"
)

// ---- System + Bus ---------------------------------------------------------

type Bus struct {
	gain js.Value // GainNode routed into master
}

type System struct {
	mu sync.Mutex

	ctx         js.Value // AudioContext (suspended until first gesture)
	master      js.Value // master GainNode -> destination
	cache       map[string]js.Value
	pendingLoad sync.WaitGroup

	Buses struct {
		SFX, UI, Music, Amb *Bus
	}
}

var system *System

func GetSystem() *System { return system }

// WaitUntilReady blocks until all in-flight Preloads have completed decoding.
func WaitUntilReady() {
	if system == nil {
		return
	}
	system.pendingLoad.Wait()
}

func init() {
	audioCtxCtor := js.Global().Get("AudioContext")
	if !audioCtxCtor.Truthy() {
		audioCtxCtor = js.Global().Get("webkitAudioContext")
	}
	if !audioCtxCtor.Truthy() {
		log.Warn().Msg("WebAudio unavailable; audio disabled")
		system = &System{cache: map[string]js.Value{}}
		system.Buses.SFX = &Bus{}
		system.Buses.UI = &Bus{}
		system.Buses.Music = &Bus{}
		system.Buses.Amb = &Bus{}
		return
	}

	ctx := audioCtxCtor.New()
	master := ctx.Call("createGain")
	master.Get("gain").Set("value", 1.0)
	master.Call("connect", ctx.Get("destination"))

	system = &System{
		ctx:    ctx,
		master: master,
		cache:  map[string]js.Value{},
	}
	system.Buses.SFX = system.newBus()
	system.Buses.UI = system.newBus()
	system.Buses.Music = system.newBus()
	system.Buses.Amb = system.newBus()

	installGestureUnlock(ctx)
	installVisibilityHandler(ctx)
	installWebVolumeBridge(system)
	installResumeAudioBridge(ctx)
}

// installWebVolumeBridge exposes the web page's volume slider as a global
// function that drives the same master gain the game itself writes to via
// SetMaster(). Linear in [0, 1] — 1.0 = 0 dB, 0 = silenced. Any in-game UI
// that reads or writes the master value shares this source of truth.
func installWebVolumeBridge(sys *System) {
	fn := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		v := args[0].Float()
		if v < 0 {
			v = 0
		} else if v > 1 {
			v = 1
		}
		var dB float64
		if v <= 0 {
			dB = -60
		} else {
			dB = 20 * math.Log10(v)
		}
		sys.SetMaster(dB)
		return nil
	})
	js.Global().Set("primortalSetMasterVolume", fn)
}

// installVisibilityHandler suspends the AudioContext when the tab is hidden
// and resumes it when shown again — prevents music/ambient drifting while
// backgrounded and avoids audio glitches from the blocked rAF loop.
func installVisibilityHandler(ctx js.Value) {
	doc := js.Global().Get("document")
	cb := js.FuncOf(func(this js.Value, args []js.Value) any {
		hidden := doc.Get("hidden").Bool()
		state := ctx.Get("state").String()
		if hidden && state == "running" {
			ctx.Call("suspend")
		} else if !hidden && state == "suspended" {
			ctx.Call("resume")
		}
		return nil
	})
	doc.Call("addEventListener", "visibilitychange", cb)
}

func (a *System) newBus() *Bus {
	g := a.ctx.Call("createGain")
	g.Get("gain").Set("value", 1.0)
	g.Call("connect", a.master)
	return &Bus{gain: g}
}

// installResumeAudioBridge exposes primortalResumeAudio() so the JS "press to
// play" handler can explicitly resume the AudioContext as part of the same
// user gesture, rather than relying on event bubbling reaching document.
func installResumeAudioBridge(ctx js.Value) {
	fn := js.FuncOf(func(this js.Value, args []js.Value) any {
		if ctx.Get("state").String() == "suspended" {
			ctx.Call("resume")
		}
		return nil
	})
	js.Global().Set("primortalResumeAudio", fn)
}

// installGestureUnlock resumes the AudioContext on the first user interaction.
// Browsers block audio output until a gesture — clicking, a keypress, or touch.
func installGestureUnlock(ctx js.Value) {
	doc := js.Global().Get("document")
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, args []js.Value) any {
		if ctx.Get("state").String() == "suspended" {
			ctx.Call("resume")
		}
		doc.Call("removeEventListener", "keydown", handler)
		doc.Call("removeEventListener", "click", handler)
		doc.Call("removeEventListener", "touchstart", handler)
		handler.Release()
		return nil
	})
	doc.Call("addEventListener", "keydown", handler)
	doc.Call("addEventListener", "click", handler)
	doc.Call("addEventListener", "touchstart", handler)
}

// SetBusGain sets the gain on a bus in decibels.
func (a *System) SetBusGain(bus *Bus, dB float64) {
	if bus == nil || !bus.gain.Truthy() {
		return
	}
	bus.gain.Get("gain").Set("value", dbToLinear(dB))
}

// SetMaster sets the master gain in decibels.
func (a *System) SetMaster(dB float64) {
	if !a.master.Truthy() {
		return
	}
	a.master.Get("gain").Set("value", dbToLinear(dB))
}

func dbToLinear(dB float64) float64 {
	if dB <= -60 {
		return 0
	}
	return math.Pow(10, dB/20)
}

// ---- Preload --------------------------------------------------------------

// Preload reads bytes from reader, copies them into a JS ArrayBuffer, and
// asynchronously decodes into an AudioBuffer. The cache entry is populated on
// decode success.
func (a *System) Preload(name, extension string, reader io.ReadCloser) error {
	defer reader.Close()
	if !a.ctx.Truthy() {
		return nil
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	u8 := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(u8, data)
	buf := u8.Get("buffer")

	a.pendingLoad.Add(1)

	var onOk, onErr js.Func
	onOk = js.FuncOf(func(this js.Value, args []js.Value) any {
		defer a.pendingLoad.Done()
		defer onOk.Release()
		defer onErr.Release()
		if len(args) > 0 {
			a.mu.Lock()
			a.cache[name] = args[0]
			a.mu.Unlock()
		}
		return nil
	})
	onErr = js.FuncOf(func(this js.Value, args []js.Value) any {
		defer a.pendingLoad.Done()
		defer onOk.Release()
		defer onErr.Release()
		errMsg := ""
		if len(args) > 0 {
			errMsg = args[0].String()
		}
		log.Warn().Str("name", name).Str("err", errMsg).Msg("decodeAudioData failed")
		return nil
	})
	a.ctx.Call("decodeAudioData", buf).Call("then", onOk, onErr)
	return nil
}

// ---- Playback -------------------------------------------------------------

type PlaybackOptions struct {
	Speed         float64
	Loop          bool
	FadeInSeconds float64
	OnComplete    func()
}

type PlaybackControl struct {
	mu sync.Mutex

	system  *System
	source  js.Value // AudioBufferSourceNode
	gain    js.Value // per-playback GainNode
	buffer  js.Value // AudioBuffer (for duration)
	startAt float64  // ctx.currentTime when started
	loop    bool
	stopped bool

	targetVolume float64 // in dB
	onEndedCb    js.Func
	onComplete   func()
}

func (a *System) PlaySFX(name string, volume float64) *PlaybackControl {
	return a.PlaySoundOnBus(name, a.Buses.SFX, volume, nil)
}

func (a *System) PlayUI(name string) *PlaybackControl {
	return a.PlaySoundOnBus(name, a.Buses.UI, 1, nil)
}

func (a *System) PlaySoundOnBus(name string, bus *Bus, volume float64, opts *PlaybackOptions) *PlaybackControl {
	a.mu.Lock()
	audioBuf, ok := a.cache[name]
	a.mu.Unlock()
	if !ok || !audioBuf.Truthy() {
		if opts != nil && opts.OnComplete != nil {
			opts.OnComplete()
		}
		return nil
	}
	return a.playBuffer(audioBuf, bus, volume, opts)
}

// PlayMusic reads the asset path from the embedded FS, decodes it, and plays.
// Loading is async; the returned control is a shell that becomes live once
// decode finishes.
func (a *System) PlayMusic(assetPath string, opts *PlaybackOptions) *PlaybackControl {
	if opts == nil {
		opts = &PlaybackOptions{}
	}
	if !a.ctx.Truthy() {
		if opts.OnComplete != nil {
			opts.OnComplete()
		}
		return &PlaybackControl{system: a}
	}

	data, err := assets.FS.ReadFile(assetPath)
	if err != nil {
		log.Error().Str("assetPath", assetPath).Err(err).Msg("failed to read music asset")
		if opts.OnComplete != nil {
			opts.OnComplete()
		}
		return &PlaybackControl{system: a}
	}
	_ = filepath.Ext(assetPath) // decodeAudioData sniffs format itself

	ctrl := &PlaybackControl{system: a, targetVolume: 0}
	u8 := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(u8, data)
	buf := u8.Get("buffer")

	var onOk, onErr js.Func
	onOk = js.FuncOf(func(this js.Value, args []js.Value) any {
		defer onOk.Release()
		defer onErr.Release()
		if len(args) == 0 {
			return nil
		}
		a.startOnControl(ctrl, args[0], a.Buses.Music, 1.0, opts)
		return nil
	})
	onErr = js.FuncOf(func(this js.Value, args []js.Value) any {
		defer onOk.Release()
		defer onErr.Release()
		log.Warn().Str("assetPath", assetPath).Msg("music decodeAudioData failed")
		if opts.OnComplete != nil {
			opts.OnComplete()
		}
		return nil
	})
	a.ctx.Call("decodeAudioData", buf).Call("then", onOk, onErr)
	return ctrl
}

func (a *System) playBuffer(audioBuf js.Value, bus *Bus, volume float64, opts *PlaybackOptions) *PlaybackControl {
	ctrl := &PlaybackControl{system: a}
	a.startOnControl(ctrl, audioBuf, bus, volume, opts)
	return ctrl
}

// startOnControl wires an AudioBufferSourceNode + GainNode and starts playback
// on the given control. Safe to call after the control has been handed to the
// caller (for async music loads).
func (a *System) startOnControl(ctrl *PlaybackControl, audioBuf js.Value, bus *Bus, volume float64, opts *PlaybackOptions) {
	if opts == nil {
		opts = &PlaybackOptions{}
	}
	if bus == nil {
		bus = a.Buses.SFX
	}

	ctrl.mu.Lock()
	if ctrl.stopped {
		ctrl.mu.Unlock()
		return
	}
	ctrl.buffer = audioBuf

	source := a.ctx.Call("createBufferSource")
	source.Set("buffer", audioBuf)
	if opts.Loop {
		source.Set("loop", true)
		ctrl.loop = true
	}
	if opts.Speed > 0 && opts.Speed != 1.0 {
		source.Get("playbackRate").Set("value", opts.Speed)
	}

	gain := a.ctx.Call("createGain")
	targetLinear := volume
	if volume < 0 {
		targetLinear = 0
	}
	if opts.FadeInSeconds > 0 {
		now := a.ctx.Get("currentTime").Float()
		gain.Get("gain").Set("value", 0)
		gain.Get("gain").Call("linearRampToValueAtTime", targetLinear, now+opts.FadeInSeconds)
	} else {
		gain.Get("gain").Set("value", targetLinear)
	}

	source.Call("connect", gain)
	gain.Call("connect", bus.gain)

	ctrl.source = source
	ctrl.gain = gain
	ctrl.targetVolume = targetLinear
	ctrl.startAt = a.ctx.Get("currentTime").Float()

	if opts.OnComplete != nil {
		ctrl.onComplete = opts.OnComplete
		ctrl.onEndedCb = js.FuncOf(func(this js.Value, args []js.Value) any {
			if ctrl.onComplete != nil {
				ctrl.onComplete()
			}
			ctrl.onEndedCb.Release()
			return nil
		})
		source.Set("onended", ctrl.onEndedCb)
	}

	source.Call("start", 0)
	ctrl.mu.Unlock()
}

// ---- PlaybackControl methods ---------------------------------------------

func (p *PlaybackControl) Stop()                                  { p.stopIn(defaultFadeSeconds) }
func (p *PlaybackControl) StopImmediately()                       { p.stopIn(0) }
func (p *PlaybackControl) FadeOutAndStop(duration time.Duration)  { p.stopIn(duration.Seconds()) }
func (p *PlaybackControl) Pause()                                 { p.stopIn(defaultFadeSeconds) }
func (p *PlaybackControl) PauseImmediately()                      { p.stopIn(0) }
func (p *PlaybackControl) FadeOutAndPause(duration time.Duration) { p.stopIn(duration.Seconds()) }
func (p *PlaybackControl) Resume()                                {}
func (p *PlaybackControl) ResumeImmediately()                     {}
func (p *PlaybackControl) ResumeAndFadeIn(duration time.Duration) {}

func (p *PlaybackControl) SetVolume(volume float64) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.gain.Truthy() {
		return
	}
	if volume < 0 {
		volume = 0
	}
	p.gain.Get("gain").Set("value", volume)
	p.targetVolume = volume
}

func (p *PlaybackControl) IsPlaying() bool {
	if p == nil {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return !p.stopped && p.source.Truthy()
}

func (p *PlaybackControl) IsPaused() bool { return false }

func (p *PlaybackControl) Position() time.Duration {
	if p == nil || !p.system.ctx.Truthy() {
		return 0
	}
	now := p.system.ctx.Get("currentTime").Float()
	return time.Duration((now - p.startAt) * float64(time.Second))
}

func (p *PlaybackControl) Duration() time.Duration {
	if p == nil || !p.buffer.Truthy() {
		return 0
	}
	sec := p.buffer.Get("duration").Float()
	return time.Duration(sec * float64(time.Second))
}

func (p *PlaybackControl) TimeRemaining() time.Duration {
	d := p.Duration()
	pos := p.Position()
	if pos >= d {
		return 0
	}
	return d - pos
}

const defaultFadeSeconds = 0.25

func (p *PlaybackControl) stopIn(fadeSeconds float64) {
	if p == nil {
		return
	}
	p.mu.Lock()
	if p.stopped || !p.source.Truthy() {
		p.stopped = true
		p.mu.Unlock()
		return
	}
	p.stopped = true
	source := p.source
	gain := p.gain
	ctx := p.system.ctx
	p.mu.Unlock()

	if fadeSeconds > 0 && gain.Truthy() && ctx.Truthy() {
		now := ctx.Get("currentTime").Float()
		gain.Get("gain").Call("cancelScheduledValues", now)
		gain.Get("gain").Call("setValueAtTime", gain.Get("gain").Get("value"), now)
		gain.Get("gain").Call("linearRampToValueAtTime", 0, now+fadeSeconds)
		source.Call("stop", now+fadeSeconds)
	} else {
		source.Call("stop", 0)
	}
}

// ---- Talker (stubbed; speech ticks play via PlaySFX paths elsewhere) -----

type Talker struct{}

func (a *System) NewTalker(speed, volume, throttle float64) *Talker { return &Talker{} }
func (t *Talker) Speak(text string, rate float64)                   {}

type SpeechGenerator struct{}

func (a *System) CreateSpeechGenerator() *SpeechGenerator { return &SpeechGenerator{} }
func (sg *SpeechGenerator) Play(text string)              {}
