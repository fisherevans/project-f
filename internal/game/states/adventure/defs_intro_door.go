package adventure

import (
	"fisherevans.com/project/f/internal/util"
)

func init() {
	registerEventHandler("intro.door.run_state_based", func(props *util.Properties) EventHandler {
		variable := props.GetString("run_state_key", "")
		stateValue := func(gs StateGlobalsReader) string {
			v := gs.Get(variable)
			if !v.Exists() {
				return doorClosed
			}
			if b, ok := v.Value().(bool); ok {
				if b {
					return doorOpen
				} else {
					return doorClosed
				}
			}
			return gs.Get(variable).AsString(doorClosed)
		}
		var anims map[string][]AnimationReference
		var lights = map[string][]LightConfig{}
		switch props.GetString("style", "shield_door") {
		case "pillar":
			anims = map[string][]AnimationReference{
				doorClosed: {
					{Name: "space_base_pillar:down"},
				},
				doorOpen: {
					{Name: "space_base_pillar:up"},
				},
			}
		case "shield_door":
			fallthrough
		default:
			anims = map[string][]AnimationReference{
				doorClosed: {
					{Name: "adventure/doors/shield_front_1:closed"},
					{Name: "adventure/doors/shield_front_1:waves"},
				},
				doorOpen: {
					{Name: "adventure/doors/shield_front_1:open"},
				},
			}
			lights = map[string][]LightConfig{
				doorClosed: {
					{Color: "#127fd7", Size: 1.5, Modifier: util.Ptr("pulse_slow")},
				},
			}
		}
		return BasicHandlerBuilder[None]{
			Init: func(thisEntity EntityReader, globals StateGlobalsReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(thisEntity.GetId()).
						WithIsBlockingIngress(stateValue(globals) == doorClosed),
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithMode(stateValue(globals)).
						WithAnimations(anims).
						WithLights(lights),
				)
			},
			GlobalVariableUpdated: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventGlobalVariableUpdated) *HandlerOutput {
				if event.Key != variable {
					return nil
				}
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(thisEntity.GetId()).
						WithIsBlockingIngress(stateValue(globals) == doorClosed),
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithMode(stateValue(globals)))
			},
		}.CreateHandler()
	})
}
