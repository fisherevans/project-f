package state_selector

import (
	"fisherevans.com/project/f/game"
	"fisherevans.com/project/f/game/states/adventure"
	"fisherevans.com/project/f/game/states/map_editor"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"
)

type destination struct {
	Name  string
	State func() game.State
}
type Selector struct {
	selected int
	states   []destination
}

func New() game.State {
	return &Selector{
		states: []destination{
			{
				Name: "Adventure",
				State: func() game.State {
					return adventure.New("dummy")
				},
			},
			{
				Name: "Map Editor",
				State: func() game.State {
					return map_editor.New()
				},
			},
		},
	}
}

var textDrawer = text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))

func (s *Selector) OnTick(ctx *game.Context, win *opengl.Window, canvas *opengl.Canvas, timeDelta float64) {
	if win.JustPressed(pixel.KeyW) || win.JustPressed(pixel.KeyUp) {
		s.selected--
		if s.selected < 0 {
			s.selected += len(s.states)
		}
	}
	if win.JustPressed(pixel.KeyS) || win.JustPressed(pixel.KeyDown) {
		s.selected++
		if s.selected >= len(s.states) {
			s.selected -= len(s.states)
		}
	}

	if win.JustPressed(pixel.KeyEnter) || win.JustPressed(pixel.KeySpace) {
		ctx.SwapActiveState(s.states[s.selected].State())
		return
	}

	textDrawer.Clear()
	textDrawer.WriteString("Select a State:\n")
	for index, option := range s.states {
		str := fmt.Sprintf("%s", option.Name)
		if index == s.selected {
			str = "> " + str
		}
		textDrawer.WriteString(fmt.Sprintf("%s\n", str))
	}
	textDrawer.Draw(canvas, pixel.IM.Moved(pixel.V(10, canvas.Bounds().H()-10)))

	ctx.DebugBR("enter: select")
	ctx.DebugBR("w/s/up/down: change")
	ctx.DebugBR("esc: cancel")
}
