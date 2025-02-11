package text_entry

import (
	"fisherevans.com/project/f/internal/game"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"
	"regexp"
)

type TextConsumer func(*game.Context, string)

type TextEntry struct {
	game.BaseState
	win           *opengl.Window
	prompt, input string
	backState     game.State
	consumer      TextConsumer
}

func New(win *opengl.Window, prompt, initialText string, backState game.State, onSelect TextConsumer) game.State {
	return &TextEntry{
		win:       win,
		prompt:    prompt,
		input:     initialText,
		backState: backState,
		consumer:  onSelect,
	}
}

var textDrawer = text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))

var invalidChars = regexp.MustCompile("[^a-zA-Z0-9_\\-]")

func (s *TextEntry) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	s.input += string(invalidChars.ReplaceAll([]byte(s.win.Typed()), []byte("")))

	if s.win.JustPressed(pixel.KeyBackspace) || s.win.Repeated(pixel.KeyBackspace) && len(s.input) > 0 {
		s.input = s.input[:len(s.input)-1]
	}

	if s.win.JustPressed(pixel.KeyEscape) {
		ctx.SwapActiveState(s.backState)
	}
	if s.win.JustPressed(pixel.KeyEnter) {
		s.consumer(ctx, s.input)
	}

	textDrawer.Clear()
	textDrawer.WriteString(fmt.Sprintf("%s:\n%s", s.prompt, s.input))
	textDrawer.Draw(target, pixel.IM.Moved(pixel.V(10, textDrawer.LineHeight+10)))

	ctx.DebugBR("enter: select text")
	ctx.DebugBR("esc: cancel")
}
