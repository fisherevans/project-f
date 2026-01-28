package state_selector

import (
	"fmt"

	"fisherevans.com/project/f/internal/game/audio"
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

var atlas *resources.Atlas

var titleTextbox *textbox.Instance
var optionTextbox *textbox.Instance

func init() {
	resources.RunOnceInitialized(func() {
		atlas = resources.DefaultAtlas()
		titleTextbox = textbox.NewInstance(
			atlas.GetFont(resources.FontNameAddStandard),
			tbcfg.NewConfig(0, 0, tbcfg.WithExpandMode(tbcfg.ExpandFit), tbcfg.Foreground(colors.White.RGBA)))
		optionTextbox = textbox.NewInstance(
			atlas.GetFont(resources.FontNameFF),
			tbcfg.NewConfig(0, 0, tbcfg.WithExpandMode(tbcfg.ExpandFit), tbcfg.Foreground(colors.White.RGBA)))
	})
}

type Selector struct {
	game.BaseState
	initialized bool
	selected    int
	states      []game.SelectIntentDestination
}

func New(intent game.SelectIntent) game.State {
	return &Selector{
		states: intent.Destinations,
	}
}

func (s *Selector) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	switch game.Controls[*Selector]().DPad().JustPressedDirection() {
	case input.Up:
		s.selected--
		if s.selected < 0 {
			s.selected = 0
		}
		audio.GetSystem().PlayUI("ui/click")
	case input.Down:
		s.selected++
		if s.selected >= len(s.states) {
			s.selected = len(s.states) - 1
		}
		audio.GetSystem().PlayUI("ui/click")

	}

	if game.Controls[*Selector]().ButtonA().JustPressed() {
		game.SetActiveStateIntent(s.states[s.selected].Intent())
		return
	}

	if !s.initialized {
		s.initialized = true
		warming.Warmup(target)
		target.Clear(colors.Black.RGBA)
	}

	titleContent := titleTextbox.NewSimpleContent("Select a state:")
	titleContent.Render(target, pixel.IM.Moved(pixel.V(10, targetBounds.H()-15)))

	for index, option := range s.states {
		str := fmt.Sprintf("%s", option.Name)
		if index == s.selected {
			str = "> " + str
		} else {
			str = "   " + str
		}
		optionContent := optionTextbox.NewSimpleContent(str)
		optionContent.Render(target, gfx.Moved(10, int(targetBounds.H())-35-15*index))
	}

	game.DebugBRf("enter: select")
	game.DebugBRf("w/s/up/down: change")
	game.DebugBRf("esc: cancel")
}
