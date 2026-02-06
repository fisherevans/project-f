package transition

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopxl/pixel/v2"
)

type swirl struct {
	*baseState
	canvas      *shaders.Canvas
	elapsedTime float64
	duration    float64
}

func NewSwirlTransition(intent game.TransitionSwirlIntent) game.State {
	duration := intent.Duration
	if duration <= 0 {
		duration = 3
	}

	canvas := shaders.NewCanvas(game.GameWidth, game.GameHeight)
	bs := newBaseState(intent.BaseTransitionIntent)
	s := &swirl{
		baseState: bs,
		canvas:    canvas,
		duration:  duration,
	}
	s.canvas.SetupSwirlTransition(
		s.fromSnapshot.Texture(),
		mgl32.Vec2{0.5, 0.5}, // center
		0.8,                  // radius (covering corners)
		20,                   // swirl radians
		0.6,                  // falloff
	)
	return s
}

func (s *swirl) OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64) {
	s.elapsedTime += timeDelta
	progress := min(s.elapsedTime/s.duration, 1.0)
	s.canvas.UpdateSwirlTransition(progress)

	s.canvas.Clear(pixel.RGBA{A: 0})
	s.toSnapshot.Draw(s.canvas, pixel.IM.Moved(s.canvas.Bounds().Center()))

	if s.elapsedTime > s.duration {
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s.toState,
		})
	}
	s.canvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))
}
