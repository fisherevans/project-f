package transition

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"github.com/gopxl/pixel/v2"
)

type glitch struct {
	*baseState
	canvas      *shaders.Canvas
	elapsedTime float64
	duration    float64
}

func NewGlitchTransition(intent game.TransitionGlitchIntent) game.State {
	duration := intent.Duration
	if duration <= 0 {
		duration = 1.0 // Glitch transitions are usually fast
	}

	canvas := shaders.NewCanvas(game.GameWidth, game.GameHeight)
	bs := newBaseState(intent.BaseTransitionIntent)
	s := &glitch{
		baseState: bs,
		canvas:    canvas,
		duration:  duration,
	}
	s.canvas.SetupGlitchTransition(s.fromSnapshot.Texture())
	return s
}

func (s *glitch) OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64) {
	s.elapsedTime += timeDelta
	progress := min(s.elapsedTime/s.duration, 1.0)

	// We use the absolute time for the shader's randomness, not just the transition time
	s.canvas.UpdateGlitchTransition(progress, game.TimeElapsed())

	s.canvas.Clear(pixel.RGBA{A: 0})
	s.toSnapshot.Draw(s.canvas, pixel.IM.Moved(s.canvas.Bounds().Center()))

	if s.elapsedTime > s.duration {
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s.toState,
		})
	}
	s.canvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))
}
