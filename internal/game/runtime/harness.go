package runtime

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util/rng"
)

// Harness is the runtime control surface for the dev debug API. It lets an
// external driver pause and single-step game logic, inject synthetic controller
// input, capture rendered frames, and read coarse runtime status.
//
// All mutators are safe to call from the HTTP server goroutine. The game loop
// reads harness state once per frame on the game thread; frame capture is
// serviced after compositing so a captured PNG reflects the frame that was just
// rendered.
type Harness struct {
	mu   sync.Mutex
	cond *sync.Cond // signaled when stepBudget drains to zero

	// time control
	paused     bool
	speed      float64
	stepBudget int     // remaining fixed-dt updates to run while paused
	fixedDt    float64 // dt applied per stepped frame

	// input injection
	injected     input.VirtualState
	injectFrames int // remaining frames to re-apply injected state

	// determinism
	deterministic bool
	seed          uint64

	// capture targets, set once the loop has built its canvases
	sceneCanvas  *shaders.Canvas
	scaledCanvas *shaders.Canvas
	shotReqs     chan *shotRequest

	// status
	ready     atomic.Bool
	frame     atomic.Int64
	stateName atomic.Value // string
}

type shotRequest struct {
	layer string // "scene" | "scaled"
	done  chan shotResult
}

type shotResult struct {
	png []byte
	err error
}

const defaultStepDt = 1.0 / 60.0

func NewHarness() *Harness {
	h := &Harness{
		speed:    1.0,
		fixedDt:  defaultStepDt,
		shotReqs: make(chan *shotRequest, 8),
	}
	h.cond = sync.NewCond(&h.mu)
	h.stateName.Store("")
	return h
}

// --- loop-side hooks (game thread) ---

func (h *Harness) setCanvases(scene, scaled *shaders.Canvas) {
	h.mu.Lock()
	h.sceneCanvas = scene
	h.scaledCanvas = scaled
	h.mu.Unlock()
}

func (h *Harness) setScaledCanvas(scaled *shaders.Canvas) {
	h.mu.Lock()
	h.scaledCanvas = scaled
	h.mu.Unlock()
}

func (h *Harness) markReady() { h.ready.Store(true) }

// gameDelta maps the natural (wallclock) game delta to the effective delta for
// this frame, honoring pause, single-step, and the harness speed multiplier.
func (h *Harness) gameDelta(natural float64) float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.paused {
		return natural * h.speed
	}
	if h.stepBudget > 0 {
		h.stepBudget--
		if h.stepBudget == 0 {
			h.cond.Broadcast()
		}
		return h.fixedDt
	}
	return 0
}

// mergeInput ORs any active injected input into the per-frame virtual control
// state and decrements the injection window.
func (h *Harness) mergeInput(vc input.VirtualState) input.VirtualState {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.injectFrames <= 0 {
		return vc
	}
	h.injectFrames--
	in := h.injected
	vc.A = vc.A || in.A
	vc.B = vc.B || in.B
	vc.Start = vc.Start || in.Start
	vc.Select = vc.Select || in.Select
	if in.Dir != input.NotPressed {
		vc.Dir = in.Dir
	}
	return vc
}

// endFrame is called after compositing. It services any pending capture
// requests, advances the frame counter, and records the active state name.
func (h *Harness) endFrame() {
	h.frame.Add(1)
	if s := game.GetActiveState(); s != nil {
		h.stateName.Store(reflect.TypeOf(s).String())
	}
	for {
		select {
		case req := <-h.shotReqs:
			req.done <- h.capture(req.layer)
		default:
			return
		}
	}
}

func (h *Harness) capture(layer string) shotResult {
	h.mu.Lock()
	scene, scaled := h.sceneCanvas, h.scaledCanvas
	h.mu.Unlock()
	var c *shaders.Canvas
	switch layer {
	case "", "scene":
		c = scene
	case "scaled":
		c = scaled
	default:
		return shotResult{err: fmt.Errorf("unknown layer %q (want scene or scaled)", layer)}
	}
	if c == nil {
		return shotResult{err: fmt.Errorf("canvas not ready")}
	}
	png, err := CanvasToPNG(c.Canvas)
	return shotResult{png: png, err: err}
}

// --- API-side controls (server goroutine), satisfy debugapi.RuntimeControl ---

func (h *Harness) Health() (bool, string, int64) {
	name, _ := h.stateName.Load().(string)
	return h.ready.Load(), name, h.frame.Load()
}

func (h *Harness) SetTime(paused bool, speed float64) {
	h.mu.Lock()
	h.paused = paused
	if speed > 0 {
		h.speed = speed
	}
	if !paused {
		h.stepBudget = 0
		h.cond.Broadcast() // release any waiting Step
	}
	h.mu.Unlock()
}

// Step advances exactly `frames` fixed-dt game updates while paused, and blocks
// until they have all been applied (one per rendered frame). This makes the call
// synchronous from the driver's perspective: when it returns, the game state and
// the rendered canvas reflect the stepped frames.
func (h *Harness) Step(frames int, dt float64) {
	if frames <= 0 {
		frames = 1
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.paused = true
	if dt > 0 {
		h.fixedDt = dt
	}
	h.stepBudget += frames
	for h.stepBudget > 0 {
		h.cond.Wait()
	}
}

func (h *Harness) InjectInput(a, b, start, sel bool, dir string, frames int) {
	if frames <= 0 {
		frames = 1
	}
	h.mu.Lock()
	h.injected = input.VirtualState{
		A:      a,
		B:      b,
		Start:  start,
		Select: sel,
		Dir:    input.DirectionFromString(dir),
	}
	h.injectFrames = frames
	h.mu.Unlock()
}

func (h *Harness) ClearInput() {
	h.mu.Lock()
	h.injected = input.VirtualState{}
	h.injectFrames = 0
	h.mu.Unlock()
}

// SetDeterminism enables or disables deterministic mode. When enabled, the
// shared RNG (internal/util/rng) is seeded to `seed` immediately and re-seeded
// on every Reset, so a fixed sequence of operations (reset, pause, step N,
// screenshot) produces byte-identical frames across runs.
func (h *Harness) SetDeterminism(on bool, seed uint64) {
	h.mu.Lock()
	h.deterministic = on
	h.seed = seed
	h.mu.Unlock()
	if on {
		rng.SetSeed(seed)
	}
}

// reseedIfDeterministic restarts the RNG stream from the configured seed when
// deterministic mode is on. Called on Reset so each reset is reproducible.
func (h *Harness) reseedIfDeterministic() {
	h.mu.Lock()
	on, seed := h.deterministic, h.seed
	h.mu.Unlock()
	if on {
		rng.SetSeed(seed)
	}
}

// Deterministic reports whether deterministic mode is on and the active seed.
func (h *Harness) Deterministic() (bool, uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.deterministic, h.seed
}

// CapturePNG requests a frame capture serviced by the game loop and blocks
// until the PNG is ready.
func (h *Harness) CapturePNG(layer string) ([]byte, error) {
	req := &shotRequest{layer: layer, done: make(chan shotResult, 1)}
	select {
	case h.shotReqs <- req:
	default:
		return nil, fmt.Errorf("capture queue full")
	}
	res := <-req.done
	return res.png, res.err
}
