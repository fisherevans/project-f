package state_selector

import (
	"fmt"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/runtime/warming"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var atlas = resources.DefaultAtlas()

func init() {
	game.RegisterStateFactory(New)
}

type Selector struct {
	game.BaseState
	initialized bool
	selected    int
	states      []game.SelectIntentDestination
	vm          *resources.ScriptEnvironment
}

func New(intent game.SelectIntent) game.State {
	return &Selector{
		states: intent.Destinations,
		vm:     resources.NewScriptEnvironment(nil),
	}
}

var titleTextbox = textbox.NewInstance(
	atlas.GetFont(resources.FontNameAddStandard),
	tbcfg.NewConfig(0, 0, tbcfg.WithExpandMode(tbcfg.ExpandFit), tbcfg.Foreground(colors.White.RGBA)))

var optionTextbox = textbox.NewInstance(
	atlas.GetFont(resources.FontNameFF),
	tbcfg.NewConfig(0, 0, tbcfg.WithExpandMode(tbcfg.ExpandFit), tbcfg.Foreground(colors.White.RGBA)))

func (s *Selector) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	switch game.Controls[*Selector]().DPad().JustPressedDirection() {
	case input.Up:
		s.selected--
		if s.selected < 0 {
			s.selected = 0
		}
	case input.Down:
		s.selected++
		if s.selected > len(s.states) {
			s.selected = len(s.states)
		}

	}

	if game.Controls[*Selector]().ButtonA().JustPressed() {
		if s.selected == len(s.states) {
			s, err := s.vm.ScriptReference("test")
			if err != nil {
				panic(err)
			}
			err = s.OnUpdate(1, 1)
			if err != nil {
				panic(err)
			}
			return
		}
		game.SetActiveStateIntent(s.states[s.selected].Intent())
		return
	}

	if !s.initialized {
		s.initialized = true
		warming.Warmup(target)
		target.Clear(colors.Black.RGBA)
	}

	titleContent := titleTextbox.NewSimpleContent("Select a State:")
	titleTextbox.Render(target, pixel.IM.Moved(pixel.V(10, targetBounds.H()-15)), titleContent)

	for index, option := range s.states {
		str := fmt.Sprintf("%s", option.Name)
		if index == s.selected {
			str = "> " + str
		} else {
			str = "   " + str
		}
		optionContent := optionTextbox.NewSimpleContent(str)
		optionTextbox.Render(target, gfx.Moved(10, int(targetBounds.H())-35-15*index), optionContent)
	}

	thingText := "test thing"
	if s.selected == len(s.states) {
		thingText = "> " + thingText
	} else {
		thingText = "   " + thingText
	}
	thingContent := optionTextbox.NewSimpleContent(thingText)
	optionTextbox.Render(target, gfx.Moved(10, int(targetBounds.H())-35-15*len(s.states)), thingContent)

	game.DebugBR("enter: select")
	game.DebugBR("w/s/up/down: change")
	game.DebugBR("esc: cancel")
}
