package startup

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/game/shaders/bloom"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"
)

type CopyrightState struct {
	game.BaseState
	elapsed float64

	canvas *shaders.Canvas
	bloom  *bloom.Helper
}

func NewCopyright(_ game.StartupCopyrightsIntent) game.State {
	s := &CopyrightState{
		canvas: shaders.NewCanvas(game.GameWidth, game.GameHeight),
		bloom: bloom.NewHelper(
			game.GameWidth, game.GameWidth,
			bloom.DefaultBrightnessConfig(),
			bloom.DefaultBlurConfig(),
			bloom.DefaultBlendConfig().WithIntensity(0.5)),
	}
	return s
}

const copyrightDisplayTime = 4

var copyrightFadeKeys = interp.NewKeys().
	WithDefaultFunction(interp.Smootherstep).
	WithKey(0, 0).
	WithKey(0.1, 1).
	WithKey(0.9, 1).
	WithKey(1, 0)

var copyrightMoveKeys = interp.NewKeys().
	WithDefaultFunction(interp.Smootherstep).
	WithKey(0, 1).
	WithKey(0.2, 0).
	WithKey(1, 0)

func (s *CopyrightState) OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64) {
	s.canvas.Clear(colors.Black.RGBA)

	if game.Controls[*CopyrightState]().ButtonB().JustPressed() {
		game.SetActiveStateIntent(game.StartupDeviceIntent{})
	}

	s.elapsed += timeDelta
	if s.elapsed > copyrightDisplayTime || game.Controls[*CopyrightState]().ButtonA().JustPressed() {
		game.SetActiveStateIntent(game.StartupDeveloperIntent{})
	}

	progress := interp.Smootherstep(min(s.elapsed/copyrightDisplayTime, 1.0))

	mask := colors.WithAlpha(colors.White.RGBA, copyrightFadeKeys.Interpolate(progress))
	dy := -10 * copyrightMoveKeys.Interpolate(progress)

	m := centerMatrix.Moved(pixel.V(0, dy))
	atlas.GetSprite("startup/copyright").DrawColorMask(s.canvas, m, mask)
	generatedBloom := s.bloom.ApplyBloom(s.canvas)
	generatedBloom.Draw(s.canvas, centerMatrix)

	s.canvas.Draw(target, centerMatrix)
}
