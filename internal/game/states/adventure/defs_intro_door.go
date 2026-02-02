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
		var sounds = map[string][]SoundEffect{}
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
		case "sliding_metal":
			anims = map[string][]AnimationReference{
				doorClosed: {
					{Name: "adventure/doors/sliding_steel:closed"},
				},
				doorOpen: {
					{Name: "adventure/doors/sliding_steel:open"},
				},
			}
			sounds = map[string][]SoundEffect{
				doorOpen: {
					{
						Name:    "adventure/sealed_door_opens",
						Falloff: StandardFalloff,
						Volume:  1,
					},
				},
				doorClosed: {
					{
						Name:    "adventure/sealed_door_opens",
						Falloff: StandardFalloff,
						Volume:  1,
					},
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
			sounds = map[string][]SoundEffect{
				"closed": {
					{
						Name:    "adventure/props/energy_door_on",
						Volume:  1.0,
						Falloff: StandardFalloff,
					},
					{
						Name:          "adventure/props/energy_door_hum",
						Loop:          true,
						Falloff:       AmbientFalloff,
						Volume:        0.5,
						FadeInSeconds: 1,
					},
				},
				"open": {
					{
						Name:    "adventure/props/energy_door_off",
						Volume:  1.0,
						Falloff: StandardFalloff,
					},
				},
			}
		}
		return BasicHandlerBuilder[None]{
			Init: func(thisEntity EntityReader, globals StateGlobalsReader, state None) *HandlerOutput {
				soundProvider := NewModeBaseSoundProviderConfig()
				for onMode, effects := range sounds {
					for _, effect := range effects {
						soundProvider = soundProvider.WithSoundOnEnter(onMode, effect)
					}
				}
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(thisEntity.GetId()).
						WithIsBlockingIngress(stateValue(globals) == doorClosed),
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithMode(stateValue(globals)).
						WithAnimations(anims).
						WithLights(lights),
					NewAddSoundProviderEffect(thisEntity.GetId()).WithModeBase(soundProvider),
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
