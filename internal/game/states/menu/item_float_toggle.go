package menu

import (
	"fmt"

	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

type itemFloatSlider struct {
	state          *State
	label          string
	format         string
	min, max, step float64
	field          *float64

	onChangeListener func(float64)
}

func newItemFloatSlider(state *State, label, format string, min, max, step float64, field *float64) *itemFloatSlider {
	return &itemFloatSlider{
		state:  state,
		label:  label,
		format: format,
		min:    min,
		max:    max,
		step:   step,
		field:  field,
	}
}

func (i *itemFloatSlider) withListener(l func(float65 float64)) *itemFloatSlider {
	i.onChangeListener = l
	return i
}

func (i *itemFloatSlider) Height() int {
	return 12
}

func (i *itemFloatSlider) Render(leftCenter pixel.Matrix, target pixel.Target, isHighlighted bool) {
	mask := colors.FromString("#555")
	text := i.label + "{-*}: "
	text += fmt.Sprintf(i.format, *i.field)
	prefix := ""
	if isHighlighted {
		mask = colors.FromString("#111")
		prefix = "{+u}"
	}
	content := mainText.NewComplexContent(prefix + text)
	content.Render(target, leftCenter, tbcfg.Foreground(mask), tbcfg.RenderFrom(gfx.LeftCenter))
}

func (i *itemFloatSlider) shift(delta float64) {
	newVal := *i.field + delta
	if newVal < i.min {
		newVal = i.min
	} else if newVal > i.max {
		newVal = i.max
	}
	*i.field = newVal
	if i.onChangeListener != nil {
		i.onChangeListener(newVal)
	}
}

func (i *itemFloatSlider) OnSelect() {
}

func (i *itemFloatSlider) OnDPadLeft() {
	i.shift(-i.step)
}

func (i *itemFloatSlider) OnDPadRight() {
	i.shift(i.step)
}
