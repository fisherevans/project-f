package runtime

import "fisherevans.com/project/f/internal/game"

// These methods make *Instance satisfy debugapi.RuntimeControl. The loop-coupled
// calls delegate to the harness; Reset publishes the configured reset intent and
// must be invoked on the game thread (the debug API wraps it in the command
// queue).

func (i *Instance) Health() (bool, string, int64) {
	if i.harness == nil {
		return false, "", 0
	}
	return i.harness.Health()
}

func (i *Instance) CapturePNG(layer string) ([]byte, error) {
	return i.harness.CapturePNG(layer)
}

func (i *Instance) SetTime(paused bool, speed float64) {
	i.harness.SetTime(paused, speed)
}

func (i *Instance) Step(frames int, dt float64) {
	i.harness.Step(frames, dt)
}

func (i *Instance) InjectInput(a, b, start, sel bool, dir string, frames int) {
	i.harness.InjectInput(a, b, start, sel, dir, frames)
}

func (i *Instance) ClearInput() {
	i.harness.ClearInput()
}

func (i *Instance) Reset() {
	if i.harness != nil {
		i.harness.reseedIfDeterministic()
	}
	game.SetActiveStateIntent(i.resetIntentFactory())
}

func (i *Instance) SetDeterminism(on bool, seed uint64) {
	i.harness.SetDeterminism(on, seed)
}
