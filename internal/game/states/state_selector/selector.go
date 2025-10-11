package state_selector

import (
	"fmt"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
)

func init() {
	game.RegisterStateFactory(New)
}

type Selector struct {
	game.BaseState
	selected int
	states   []game.SelectIntentDestination
}

func New(intent game.SelectIntent) game.State {
	return &Selector{
		states: intent.Destinations,
	}
}

var titleDrawer = text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))
var optionDrawer = text.New(pixel.ZV, resources.CreateFont(resources.FontNameM5x7).Atlas)

func (s *Selector) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	switch game.Controls[*Selector]().DPad().JustPressedDirection() {
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

	if game.Controls[*Selector]().ButtonA().JustPressed() {
		game.SetActiveStateIntent(s.states[s.selected].Intent())
		return
	}

	titleDrawer.Clear()
	titleDrawer.WriteString("Select a State:")
	titleDrawer.Draw(target, pixel.IM.Moved(pixel.V(10, targetBounds.H()-15)))

	optionDrawer.Clear()
	for index, option := range s.states {
		str := fmt.Sprintf("%s", option.Name)
		if index == s.selected {
			str = "> " + str
		} else {
			str = "   " + str
		}
		optionDrawer.WriteString(fmt.Sprintf("%s\n", str))
	}
	optionDrawer.Draw(target, pixel.IM.Moved(pixel.V(10, targetBounds.H()-35)))

	game.DebugBR("enter: select")
	game.DebugBR("w/s/up/down: change")
	game.DebugBR("esc: cancel")
}
