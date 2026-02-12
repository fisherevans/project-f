package computer

import (
	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

var (
	atlas                  = resources.DefaultAtlas()
	deviceBackgroundSprite pixelutil.BoundedDrawable
	moveDuration           = 0.4

	screenWidth  = 220
	screenHeight = 140

	screenColors = screen.Colors{
		Dark:      colors.MapComputerDark.RGBA,
		Clear:     colors.MapComputerClear.RGBA,
		Text:      colors.MapComputerText.RGBA,
		Highlight: colors.MapComputerHighlight.RGBA,
	}
)

func init() {
	resources.RunOnceInitialized(func() {
		deviceBackgroundSprite = atlas.GetSprite("computer/monitor")
	})
}

type State struct {
	game.BaseState
	background game.State
	transition *interp.TimedProgress
	exiting    bool

	screen *screen.Instance
}

func New(i game.ComputerIntent) game.State {
	s := &State{
		background: i.Background,
		transition: interp.NewTimedProgress(moveDuration, interp.Smootherstep),
	}
	renderContext := screen.RenderContext{
		Atlas:  atlas,
		Width:  screenWidth,
		Height: screenHeight,
		Colors: screenColors,
	}
	s.screen = screen.NewScreen(s, renderContext, newMapMenu)
	return s
}

func (s *State) ClearColor() pixel.RGBA {
	return colors.Black.RGBA
}

func (s *State) OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64) {
	s.transition.Update(timeDelta)

	if s.exiting || !s.transition.IsComplete() {
		s.background.OnTick(target, targetBounds, timeDelta) // render with 0 time delta
	}

	if s.exiting && s.transition.IsComplete() {
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s.background,
		})
	}

	fullDeltaX := deviceBackgroundSprite.Bounds().W()
	dx := (1.0 - s.transition.Progress()) * fullDeltaX

	game.DebugBLf("transition: %f", s.transition.Progress())
	game.DebugBLf("dx: %f", dx)

	bgMatrix := pixel.IM.Moved(gfx.BottomLeft.Align(deviceBackgroundSprite)).
		Moved(pixel.V(-dx, 0))
	deviceBackgroundSprite.Draw(target, bgMatrix)

	paddingX := float64(game.GameWidth-screenWidth) / 2.0
	paddingY := float64(game.GameHeight-screenHeight) / 2.0
	screenMatrix := pixel.IM.Moved(gfx.BottomLeft.Align(s.screen)).
		Moved(pixel.V(paddingX, paddingY)).
		Moved(pixel.V(-dx, 0))
	s.screen.OnTick(screenMatrix, target, timeDelta)

	if game.Controls[*State]().ButtonStart().JustPressed() {
		s.Close()
	}
}

func (s *State) Close() {
	s.exiting = !s.exiting
	s.transition.Reverse()
}
