package confirm

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"
	"image/color"
)

type OnConfirmed func(ctx *game.Context)

type Confirmation struct {
	game.BaseState
	progress    float64
	win         *opengl.Window
	prompt      []string
	backState   game.State
	onConfirmed OnConfirmed
}

func New(win *opengl.Window, prompt []string, backState game.State, onConfirmed OnConfirmed) game.State {
	return &Confirmation{
		win:         win,
		prompt:      prompt,
		backState:   backState,
		onConfirmed: onConfirmed,
	}
}

var textDrawer = text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))

func (s *Confirmation) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	if s.win.JustPressed(pixel.KeyEscape) {
		ctx.SwapActiveState(s.backState)
	}

	if s.win.Pressed(pixel.KeyEnter) {
		if s.win.Pressed(pixel.KeyLeftShift) {
			s.progress += 1
		}
		s.progress += timeDelta
	} else {
		s.progress = 0
	}

	if s.progress > 1 {
		s.onConfirmed(ctx)
	}

	textDrawer.Clear()
	for _, line := range s.prompt {
		textDrawer.WriteString(fmt.Sprintf("%s\n", line))
	}
	gb := uint8(util.Clamp(0, int((1.0-s.progress)*255), 255))
	mask := color.RGBA{255, gb, gb, 255}

	renderHeight := textDrawer.LineHeight*float64(len(s.prompt)) + 10

	if renderHeight > targetBounds.H() {
		target = s.win
	}
	textDrawer.DrawColorMask(target, pixel.IM.Moved(pixel.V(10, textDrawer.LineHeight*float64(len(s.prompt))+10)), mask)

	ctx.DebugBR("enter: confirm")
	ctx.DebugBR("esc: cancel")
}
