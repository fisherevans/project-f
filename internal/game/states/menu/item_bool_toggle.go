package menu

import (
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

type itemBoolToggle struct {
	state                      *State
	label, trueText, falseText string
	field                      *bool
	onChangeListener           func(bool)
}

func newItemBoolToggle(state *State, label, trueText, falseText string, field *bool) *itemBoolToggle {
	return &itemBoolToggle{
		state:     state,
		label:     label,
		trueText:  trueText,
		falseText: falseText,
		field:     field,
	}
}

func (i *itemBoolToggle) withListener(l func(bool)) *itemBoolToggle {
	i.onChangeListener = l
	return i
}

func (i *itemBoolToggle) Height() int {
	return 12
}

func (i *itemBoolToggle) Render(leftCenter pixel.Matrix, target pixel.Target, isHighlighted bool) {
	mask := colors.FromString("#555")
	text := i.label + "{-*}: "
	if *i.field {
		text += i.trueText
	} else {
		text += i.falseText
	}
	prefix := ""
	if isHighlighted {
		mask = colors.FromString("#111")
		prefix = "{+u}"
	}
	content := mainText.NewComplexContent(prefix + text)
	content.Render(target, leftCenter, tbcfg.Foreground(mask), tbcfg.RenderFrom(gfx.LeftCenter))
}

func (i *itemBoolToggle) set(b bool) {
	if *i.field == b {
		return
	}
	*i.field = b
	if i.onChangeListener != nil {
		i.onChangeListener(b)
	}
}

func (i *itemBoolToggle) OnSelect() {
	i.set(!*i.field)
}

func (i *itemBoolToggle) OnDPadLeft() {
	i.set(false)
}

func (i *itemBoolToggle) OnDPadRight() {
	i.set(true)
}
