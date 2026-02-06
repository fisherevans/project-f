package transition

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"github.com/gopxl/pixel/v2"
)

type flush struct {
	*baseState
	canvas      *shaders.Canvas
	elapsedTime float64
	duration    float64
}

func NewFlushTransition(intent game.TransitionFlushIntent) game.State {
	duration := intent.Duration
	if duration <= 0 {
		duration = 0.8 // Flush transitions should be relatively quick
	}

	canvas := shaders.NewCanvas(game.GameWidth, game.GameHeight)
	bs := newBaseState(intent.BaseTransitionIntent)
	s := &flush{
		baseState: bs,
		canvas:    canvas,
		duration:  duration,
	}
	s.canvas.SetupFlushTransition(s.fromSnapshot.Texture())
	return s
}

func (s *flush) OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64) {
	s.elapsedTime += timeDelta
	progress := min(s.elapsedTime/s.duration, 1.0)

	s.canvas.UpdateFlushTransition(progress)

	s.canvas.Clear(pixel.RGBA{A: 0})
	s.toSnapshot.Draw(s.canvas, pixel.IM.Moved(s.canvas.Bounds().Center()))

	if s.elapsedTime > s.duration {
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s.toState,
		})
	}
	s.canvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))
}
