package xenolog

import (
	"image/color"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
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
	deviceBackgroundSprite = atlas.GetSprite("xenolog/device")
)

type State struct {
	background game.State
	transition *interp.TimedProgress
	exiting    bool

	screen *Screen
}

func New(i game.XenologIntent) game.State {
	return &State{
		background: i.Background,
		transition: interp.NewTimedProgress(.4, interp.Smootherstep),
		screen:     NewScreen(),
	}
}

func (s *State) ClearColor() color.Color {
	return colors.Black.RGBA
}

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	s.transition.Update(timeDelta)

	if s.exiting || !s.transition.IsComplete() {
		s.background.OnTick(target, targetBounds, timeDelta) // render with 0 time delta
	}

	if s.exiting && s.transition.IsComplete() {
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s.background,
		})
	}

	fullDeltaY := deviceBackgroundSprite.Bounds().H()
	dy := (1.0 - s.transition.Progress()) * fullDeltaY

	game.DebugBL("transition: %f", s.transition.Progress())
	game.DebugBL("dy: %f", dy)

	bgMatrix := pixel.IM.Moved(gfx.BottomLeft.Align(deviceBackgroundSprite)).
		Moved(pixel.V(0, -dy))
	deviceBackgroundSprite.Draw(target, bgMatrix)

	paddingX := float64(game.GameWidth-screenWidth) / 2.0
	paddingY := float64(game.GameHeight-screenHeight) / 2.0
	screenMatrix := pixel.IM.Moved(gfx.BottomLeft.Align(s.screen)).
		Moved(pixel.V(paddingX, paddingY)).
		Moved(pixel.V(0, -dy))
	s.screen.OnTick(s, screenMatrix, target, timeDelta)

	if game.Controls[*State]().ButtonSelect().JustPressed() {
		s.exiting = !s.exiting
		s.transition.Reverse()
	}
}
