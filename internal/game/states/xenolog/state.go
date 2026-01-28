package xenolog

import (
	"image/color"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

var atlas *resources.Atlas
var deviceBackgroundSprite pixelutil.BoundedDrawable
var moveDuration = 0.4

func init() {
	resources.RunOnceInitialized(func() {
		atlas = resources.DefaultAtlas()
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

	screen *Screen
}

func New(i game.XenologIntent) game.State {
	s := &State{
		background: i.Background,
		transition: interp.NewTimedProgress(moveDuration, interp.Smootherstep),
	}
	s.screen = NewScreen(s, i.PrimortalsEnabled)
	return s
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
	s.screen.OnTick(s, screenMatrix, target, timeDelta)

	if game.Controls[*State]().ButtonStart().JustPressed() {
		s.Close()
	}
}

func (s *State) Close() {
	s.exiting = !s.exiting
	s.transition.Reverse()
}
