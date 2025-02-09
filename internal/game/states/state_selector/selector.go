package state_selector

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"
)

type Destination struct {
	Name  string
	State func() game.State
}
type Selector struct {
	game.BaseState
	selected int
	states   []Destination
}

func New(destinations ...Destination) game.State {
	return &Selector{
		states: destinations,
	}
}

var textDrawer = text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))

func (s *Selector) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	switch ctx.Controls.DPad().JustPressedDirection(false) {
	case input.Up:
		s.selected--
		if s.selected < 0 {
			s.selected += len(s.states)
		}
	case input.Down:
		s.selected++
		if s.selected >= len(s.states) {
			s.selected -= len(s.states)
		}

	}

	if ctx.Controls.ButtonA().JustPressed() {
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
	textDrawer.Draw(target, pixel.IM.Moved(pixel.V(10, targetBounds.H()-10)))

	ctx.DebugBR("enter: select")
	ctx.DebugBR("w/s/up/down: change")
	ctx.DebugBR("esc: cancel")
}
