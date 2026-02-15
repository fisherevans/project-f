package xenolog

import (
	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

var (
	screenWidth  = 206
	screenHeight = 128

	atlas                  = resources.DefaultAtlas()
	deviceBackgroundSprite pixelutil.BoundedDrawable
	moveDuration           = 0.4

	screenColors = screen.Colors{
		Dark:      colors.XenoLogDark.RGBA,
		Clear:     colors.XenoLogClear.RGBA,
		Text:      colors.XenoLogText.RGBA,
		Highlight: colors.XenoLogHighlight.RGBA,
	}

	badgeButtonStyle = badges.ButtonColorStyle{
		Action:    screenColors.Text,
		Button:    screenColors.Dark,
		Highlight: screenColors.Highlight,
	}
)

func init() {
	resources.RunOnceInitialized(func() {
		deviceBackgroundSprite = atlas.GetSprite("xenolog/device")
		initializeXenologVariables()
		badgeAContinue = badges.Using(atlas).ButtonAction("A", "continue", badgeButtonStyle).Flipped()
	})
}

type State struct {
	game.BaseState
	background game.State
	transition *interp.TimedProgress
	exiting    bool
	exitData   any

	screen *screen.Instance[*State]
}

func New(i game.XenologIntent) game.State {
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
	s.screen = screen.NewScreen(s, renderContext, func(s *screen.Instance[*State]) screen.Menu {
		return newHomeMenu(s, i.PrimortalsEnabled)
	})
	s.screen.PushMenu(newHomeMenu(s.screen, i.PrimortalsEnabled), false)
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
		game.SetActiveStateIntentWithData(game.SwapStateIntent{
			State: s.background,
		}, s.exitData)
	}

	fullDeltaY := deviceBackgroundSprite.Bounds().H()
	dy := (1.0 - s.transition.Progress()) * fullDeltaY

	game.DebugBLf("transition: %f", s.transition.Progress())
	game.DebugBLf("dy: %f", dy)

	bgMatrix := pixel.IM.Moved(gfx.BottomLeft.Align(deviceBackgroundSprite)).
		Moved(pixel.V(0, -dy))
	deviceBackgroundSprite.Draw(target, bgMatrix)

	paddingX := float64(game.GameWidth-screenWidth) / 2.0
	paddingY := float64(game.GameHeight-screenHeight) / 2.0
	screenMatrix := pixel.IM.Moved(gfx.BottomLeft.Align(s.screen)).
		Moved(pixel.V(paddingX, paddingY)).
		Moved(pixel.V(0, -dy))
	s.screen.OnTick(screenMatrix, target, timeDelta)

	if game.Controls[*State]().ButtonStart().JustPressed() {
		s.ToggleOpenState()
	}
}

func (s *State) ToggleOpenState() {
	s.exiting = !s.exiting
	s.exitData = nil
	s.transition.Reverse()
}

func (s *State) CloseWithData(data any) {
	s.exitData = data
	if s.exiting {
		return
	}
	s.exiting = true
	s.transition.Reverse()
}
