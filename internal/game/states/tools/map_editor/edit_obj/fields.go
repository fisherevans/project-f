package edit_obj

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/multi_select"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/text_entry"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/rs/zerolog/log"
)

type FieldType string

const (
	FieldTypeString FieldType = "string"
)

var FieldTypes = []FieldType{
	FieldTypeString,
}

type Field interface {
	Name() string
	Value() any
	Edit(win *opengl.Window, parent game.State) game.State
	Apply()
}

type baseStringField struct {
	name, value string
	apply       func(string)
}

func newBaseStringField(name, value string, apply func(string)) baseStringField {
	return baseStringField{
		name:  name,
		value: value,
		apply: apply,
	}
}

func (s *baseStringField) Name() string {
	return s.name
}

func (s *baseStringField) Value() any {
	return s.value
}

func (s *baseStringField) Apply() {
	s.apply(s.value)
}

type StringField struct {
	baseStringField
	allowEmpty bool
}

func NewStringField(name, value string, allowEmpty bool, apply func(string)) Field {
	return &StringField{
		baseStringField: newBaseStringField(name, value, apply),
		allowEmpty:      allowEmpty,
	}
}

func (s *StringField) Edit(win *opengl.Window, parent game.State) game.State {
	return text_entry.New(win, s.name, s.value, parent, func(ctx *game.Context, newValue string) {
		if newValue == "" && !s.allowEmpty {
			log.Error().Msg("can't be empty")
			return
		}
		s.value = newValue
		ctx.SwapActiveState(parent)
	})
}

type StringEnumField struct {
	baseStringField
	options []string
}

func NewStringEnumField(name, value string, options []string, apply func(string)) *StringEnumField {
	return &StringEnumField{
		baseStringField: newBaseStringField(name, value, apply),
		options:         options,
	}
}

func (s *StringEnumField) Edit(win *opengl.Window, parent game.State) game.State {
	var selected int
	for index, option := range s.options {
		if s.value == option {
			selected = index
			break
		}
	}
	return multi_select.New(win, s.name, selected, s.options, parent, func(ctx *game.Context, newIndex int) {
		s.value = s.options[selected]
		ctx.SwapActiveState(parent)
	})
}
