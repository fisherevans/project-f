package edit_obj

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/multi_select"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/text_entry"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/rs/zerolog/log"
	"reflect"
)

func NewEditEntity(win *opengl.Window, parent game.State, entityId string, entity *resources.Entity) game.State {
	var fields []Field
	fields = append(fields, NewStringField("Type", entity.Type, false, func(newValue string) {
		entity.Type = newValue
	}))
	metadataSetter := func(k string) func(string) {
		return func(v string) {
			if entity.Metadata == nil {
				entity.Metadata = make(map[string]interface{})
			}
			entity.Metadata[k] = v
		}
	}
	for k, v := range entity.Metadata {
		if reflect.TypeOf(v).Kind() == reflect.String {
			strValue := v.(string)
			fields = append(fields, NewStringField("  "+k, strValue, true, metadataSetter(k)))
		}
	}
	var actions []Action
	actions = append(actions, newBasicAction("Add Field", func(ctx *game.Context, editor *Editor) {
		typeEntryState := multi_select.New(win, "Field Type", 0, FieldTypes, editor, func(ctx *game.Context, typeIndex int) {
			switch FieldTypes[typeIndex] {
			case FieldTypeString:
				ctx.SwapActiveState(text_entry.New(win, "Field Name", "", editor, func(ctx *game.Context, name string) {
					if name == "" {
						log.Error().Msg("can't be empty")
						return
					}
					if entity.Metadata == nil {
						entity.Metadata = make(map[string]interface{})
					}
					if _, exists := entity.Metadata[name]; exists {
						log.Error().Msg("field " + name + " already exists")
						return
					}
					ctx.SwapActiveState(text_entry.New(win, "Field Value", "", editor, func(ctx *game.Context, value string) {
						if value == "" {
							log.Error().Msg("can't be empty")
							return
						}
						editor.fields = append(editor.fields, NewStringField("  "+name, value, false, metadataSetter(name)))
						ctx.SwapActiveState(editor)
					}))
				}))
			}
		})
		ctx.SwapActiveState(typeEntryState)
	}))
	return New(win, "Entity ID: "+entityId, fields, actions, parent, parent, nil)
}
