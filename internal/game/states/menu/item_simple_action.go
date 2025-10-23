package menu

import (
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

type simpleSelectOption struct {
	state  *State
	label  string
	action func(s *State)
}

func newSimpleSelectionOption(state *State, label string, action func(s *State)) *simpleSelectOption {
	return &simpleSelectOption{
		state:  state,
		label:  label,
		action: action,
	}
}

func (s *simpleSelectOption) Height() int {
	return 12
}

func (s *simpleSelectOption) Render(leftCenter pixel.Matrix, target pixel.Target, isHighlighted bool) {
	mask := colors.FromString("#555")
	prefix := ""
	if isHighlighted {
		mask = colors.FromString("#111")
		prefix = "{+u}"
	}
	content := mainText.NewComplexContent(prefix + s.label)
	mainText.Render(target, leftCenter, content, tbcfg.Foreground(mask), tbcfg.RenderFrom(gfx.LeftCenter))
}

func (s *simpleSelectOption) OnSelect() {
	if s.action != nil {
		s.action(s.state)
	}
}

func (s *simpleSelectOption) OnDPadLeft() {}

func (s *simpleSelectOption) OnDPadRight() {}
