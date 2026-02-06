package startup

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

var pressFor = 1.4

type ControlsState struct {
	game.BaseState
	exitStateIntent any
	buffer          *shaders.Canvas
	pressedDuration float64
}

func NewControls(i game.StartupControlsIntent) game.State {
	return &ControlsState{
		exitStateIntent: i.ExitStateIntent,
		buffer:          shaders.NewCanvas(game.GameWidth, game.GameHeight),
	}
}

func (s *ControlsState) ClearColor() pixel.RGBA {
	return pixel.RGB(0.1, 0.1, 0.1)
}

type area struct {
	x, y, width, height int
}

var (
	areaDpad   = area{x: 12, y: 62, width: 65, height: 65}
	areaSelect = area{x: 67, y: 22, width: 35, height: 23}
	areaStart  = area{x: 109, y: 22, width: 35, height: 23}
	areaB      = area{x: 152, y: 71, width: 34, height: 34}
	areaA      = area{x: 194, y: 91, width: 34, height: 34}

	progressFrameWidth  = 86
	progressFrameHeight = 14
	progressFrame       *frames.Instance

	smallText *textbox.Instance
	titleText *textbox.Instance
)

func (s *ControlsState) OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64) {
	if game.Controls[*ControlsState]().ButtonA().IsPressed() {
		s.pressedDuration += timeDelta
	} else {
		s.pressedDuration -= timeDelta
	}
	s.pressedDuration = max(0.0, min(pressFor, s.pressedDuration))
	if s.pressedDuration >= pressFor {
		game.SetActiveStateIntent(s.exitStateIntent)
	}

	s.buffer.Clear(colors.Alpha(0))
	s.buffer.SetComposeMethod(pixel.ComposeOver)
	atlas.GetSprite("startup/controls").Draw(s.buffer, pixel.IM.Moved(targetBounds.Center()))

	hightlight := colors.FromString("#e69f1a")
	drawArea := func(a area) {
		s.buffer.SetComposeMethod(pixel.ComposeMultiply)
		gfx.DrawRect(atlas, s.buffer, gfx.Moved(a.x, a.y), gfx.BottomLeft, a.width, a.height, hightlight)
	}

	if game.Controls[*ControlsState]().ButtonA().IsPressed() {
		drawArea(areaA)
	}

	if game.Controls[*ControlsState]().ButtonB().IsPressed() {
		drawArea(areaB)
	}

	if game.Controls[*ControlsState]().ButtonStart().IsPressed() {
		drawArea(areaStart)
	}

	if game.Controls[*ControlsState]().ButtonSelect().IsPressed() {
		drawArea(areaSelect)
	}

	if game.Controls[*ControlsState]().DPad().IsPressed() {
		drawArea(areaDpad)
	}

	s.buffer.Draw(target, pixel.IM.Moved(targetBounds.Center()))

	progressCenter := gfx.Moved(194, 18)
	lerp := interp.Smoothstep(s.pressedDuration / pressFor)
	bg := pixel.RGB(0.5, 0.5, 0.5)
	mg := colors.Lerp(bg, hightlight, lerp)
	fg := pixel.RGB(1, 1, 1)
	progressFrame.Draw(target, pixel.R(0, 0, float64(progressFrameWidth), float64(progressFrameHeight)), progressCenter, frames.WithColor(bg))
	pressedWidth := int((lerp) * float64(progressFrameWidth))
	progressFrame.Draw(target, pixel.R(0, 0, float64(pressedWidth), float64(progressFrameHeight)), progressCenter, frames.WithColor(mg))
	smallText.NewSimpleContent("Hold 'A' to continue").Render(target, progressCenter, tbcfg.Foreground(fg))

	subtitleColor := pixel.RGB(0.8, 0.8, 0.8)
	titleText.NewSimpleContent("CONTROL BINDINGS").Render(target, gfx.Moved(game.GameWidth/2, game.GameHeight-16), tbcfg.Foreground(colors.White.RGBA))
	smallText.NewSimpleContent("This game was design to be played").Render(target, gfx.Moved(game.GameWidth/2, game.GameHeight-27), tbcfg.Foreground(subtitleColor))
	smallText.NewSimpleContent("as if you were playing the GameBoy").Render(target, gfx.Moved(game.GameWidth/2, game.GameHeight-35), tbcfg.Foreground(subtitleColor))
}
