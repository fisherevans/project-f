package transition

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
)

type fade struct {
	*baseState
	canvas      *shaders.Canvas
	elapsedTime float64
	duration    float64
}

func NewFadeTransition(intent game.TransitionFadeIntent) game.State {
	duration := intent.Duration
	if duration <= 0 {
		duration = 2
	}
	return &fade{
		baseState: newBaseState(intent.BaseTransitionIntent),
		canvas:    shaders.NewCanvas(game.GameWidth, game.GameHeight),
		duration:  duration,
	}
}

func (s *fade) OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64) {
	s.elapsedTime += timeDelta
	s.fromSnapshot.Draw(target, pixel.IM.Moved(targetBounds.Center()))
	s.toSnapshot.DrawColorMask(target, pixel.IM.Moved(targetBounds.Center()), colors.Alpha(min(s.elapsedTime/s.duration, 1.0)))
	if s.elapsedTime > s.duration {
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s.toState,
		})
	}
	s.canvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))
}
