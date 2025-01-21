package multi_select

import (
	"fisherevans.com/project/f/game"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"
)

type Consumer[T any] func(T)

type MultiSelect[T any] struct {
	prompt      string
	selected    int
	options     []T
	parentState game.State
	consumer    Consumer[T]
}

func New[T any](prompt string, initialSelection int, options []T, parent game.State, onSelect Consumer[T]) game.State {
	return &MultiSelect[T]{
		prompt:      prompt,
		selected:    initialSelection,
		options:     options,
		parentState: parent,
		consumer:    onSelect,
	}
}

var textDrawer = text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))

func (m *MultiSelect[T]) OnTick(ctx *game.Context, win *opengl.Window, canvas *opengl.Canvas, timeDelta float64) {
	if win.JustPressed(pixel.KeyW) || win.JustPressed(pixel.KeyUp) {
		m.selected--
		if m.selected < 0 {
			m.selected += len(m.options)
		}
	}
	if win.JustPressed(pixel.KeyS) || win.JustPressed(pixel.KeyDown) {
		m.selected++
		if m.selected >= len(m.options) {
			m.selected -= len(m.options)
		}
	}

	if win.JustPressed(pixel.KeyEscape) {
		ctx.SwapActiveState(m.parentState)
	}

	if win.JustPressed(pixel.KeyEnter) {
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
	textDrawer.Draw(canvas, pixel.IM.Moved(pixel.V(10, canvas.Bounds().H()-10)))

	ctx.DebugBR("enter: select")
	ctx.DebugBR("w/s/up/down: change")
	ctx.DebugBR("esc: cancel")
}
