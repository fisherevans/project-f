package multi_select

import (
	"fisherevans.com/project/f/internal/game"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"
)

type Consumer[T any] func(T)

type MultiSelect[T any] struct {
	win         *opengl.Window
	prompt      string
	selected    int
	options     []T
	parentState game.State
	consumer    Consumer[T]
}

func New[T any](win *opengl.Window, prompt string, initialSelection int, options []T, parent game.State, onSelect Consumer[T]) game.State {
	return &MultiSelect[T]{
		win:         win,
		prompt:      prompt,
		selected:    initialSelection,
		options:     options,
		parentState: parent,
		consumer:    onSelect,
	}
}

var textDrawer = text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))

func (m *MultiSelect[T]) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	if m.win.JustPressed(pixel.KeyW) || m.win.JustPressed(pixel.KeyUp) {
		m.selected--
		if m.selected < 0 {
			m.selected += len(m.options)
		}
	}
	if m.win.JustPressed(pixel.KeyS) || m.win.JustPressed(pixel.KeyDown) {
		m.selected++
		if m.selected >= len(m.options) {
			m.selected -= len(m.options)
		}
	}

	if m.win.JustPressed(pixel.KeyEscape) {
		ctx.SwapActiveState(m.parentState)
	}

	if m.win.JustPressed(pixel.KeyEnter) {
		m.consumer(m.options[m.selected])
		ctx.SwapActiveState(m.parentState)
	}

	textDrawer.Clear()
	textDrawer.WriteString(fmt.Sprintf("%s:\n", m.prompt))
	for index, option := range m.options {
		str := fmt.Sprintf("%s", option)
		if index == m.selected {
			str = "> " + str
		}
		textDrawer.WriteString(fmt.Sprintf("%s\n", str))
	}
	textDrawer.Draw(target, pixel.IM.Moved(pixel.V(10, targetBounds.H()-10)))

	ctx.DebugBR("enter: select")
	ctx.DebugBR("w/s/up/down: change")
	ctx.DebugBR("esc: cancel")
}
