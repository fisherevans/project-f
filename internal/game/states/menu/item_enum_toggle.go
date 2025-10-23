package menu

import (
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

type itemEnumToggleOption[T any] struct {
	enumVal T
	label   string
}

func newEnumToggleOption[T any](label string, val T) itemEnumToggleOption[T] {
	return itemEnumToggleOption[T]{
		enumVal: val,
		label:   label,
	}
}

type itemEnumToggle[T comparable] struct {
	state   *State
	label   string
	options []itemEnumToggleOption[T]
	field   *T

	onChangeListener func(T)
}

func newItemEnumToggle[T comparable](state *State, label string, options []itemEnumToggleOption[T], field *T) *itemEnumToggle[T] {
	return &itemEnumToggle[T]{
		state:   state,
		label:   label,
		options: options,
		field:   field,
	}
}

func (i *itemEnumToggle[T]) withListener(l func(T)) *itemEnumToggle[T] {
	i.onChangeListener = l
	return i
}

func (i *itemEnumToggle[T]) Height() int {
	return 12
}

func (i *itemEnumToggle[T]) Render(leftCenter pixel.Matrix, target pixel.Target, isHighlighted bool) {
	mask := colors.FromString("#555")
	text := i.label + "{-*}: "
	current := *i.field
	for _, opt := range i.options {
		if opt.enumVal == current {
			text += opt.label
			break
		}
	}
	prefix := ""
	if isHighlighted {
		mask = colors.FromString("#111")
		prefix = "{+u}"
	}
	content := mainText.NewComplexContent(prefix + text)
	mainText.Render(target, leftCenter, content, tbcfg.Foreground(mask), tbcfg.RenderFrom(gfx.LeftCenter))
}

func (i *itemEnumToggle[T]) selectDelta(delta int) {
	if delta == 0 {
		return
	}
	current := *i.field
	index := 0
	for optIndex, opt := range i.options {
		if opt.enumVal == current {
			index = optIndex
			break
		}
	}
	index += delta
	if index < 0 {
		index = len(i.options) - 1
	} else if index >= len(i.options) {
		index = 0
	}
	*i.field = i.options[index].enumVal
	if i.onChangeListener != nil {
		i.onChangeListener(i.options[index].enumVal)
	}
}

func (i *itemEnumToggle[T]) OnSelect() {
	i.selectDelta(1)
}

func (i *itemEnumToggle[T]) OnDPadLeft() {
	i.selectDelta(-1)
}

func (i *itemEnumToggle[T]) OnDPadRight() {
	i.selectDelta(1)
}
