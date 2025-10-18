package menu

import (
	"image/color"
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
)

var atlas = resources.DefaultAtlas()

var (
	margin         = 6.0
	padding        = 5.0
	transitionTime = 0.25
)

type State struct {
	background game.State
	transition float64

	main  Main
	batch *pixel.Batch
}

func New(i game.MenuIntent) game.State {
	return &State{
		background: i.Background,
		transition: 0,
		batch:      atlas.NewBatch(),

		main: createMain(),
	}
}

func (s *State) ClearColor() color.Color {
	return colors.Black.RGBA
}

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	if s.background != nil {
		s.background.OnTick(target, targetBounds, 0)
	}

	s.batch.Clear()

	s.transition = math.Min(1, s.transition+timeDelta*(1.0/transitionTime))

	fadeColor := colors.WithAlphaTodoFix(colors.Black.RGBA, 0.7*s.transition)
	gfx.DrawRect(atlas, s.batch, pixel.IM, gfx.BottomLeft, game.GameWidth, game.GameHeight, fadeColor)

	matrix := pixel.IM.Moved(pixel.V(-100.0*(1.0-s.transition), 0))
	s.main.Render(s, matrix, s.batch, targetBounds)

	s.batch.Draw(target)
}
