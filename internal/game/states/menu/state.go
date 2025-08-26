package menu

import (
	"image/color"
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
)

var atlas = resources.CreateAtlas(resources.AtlasFilter{
	FontNames: []string{
		resources.FontNameM5x7,
		resources.FontNameM3x6,
		resources.FontNameAddStandard,
		resources.FontNameFF,
		resources.FontName3x5,
	},
})

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

func New(background game.State) *State {
	return &State{
		background: background,
		transition: 0,
		batch:      atlas.NewBatch(),

		main: NewMain(),
	}
}

func (s *State) ClearColor() color.Color {
	return colors.Black.RGBA
}

func (s *State) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	if s.background != nil {
		s.background.OnTick(ctx.WithNoControls(), target, targetBounds, 0)
	}

	s.batch.Clear()

	s.transition = math.Min(1, s.transition+timeDelta*(1.0/transitionTime))

	fadeColor := colors.WithAlpha(colors.Black.RGBA, 0.7*s.transition)
	gfx.DrawRect(atlas, s.batch, pixel.IM, gfx.BottomLeft, game.GameWidth, game.GameHeight, fadeColor)

	matrix := pixel.IM.Moved(pixel.V(-100.0*(1.0-s.transition), 0))
	s.main.Render(s, ctx, matrix, s.batch, targetBounds)

	s.batch.Draw(target)
}
