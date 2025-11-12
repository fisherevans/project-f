package adventure

import (
	"fmt"
	"sync/atomic"
	"time"

	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	registerEventHandler("map1", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if thisEntity.GetId() != event.TargetId {
					return nil
				}
				return NewOutput().WithEffects(NewDialogueEffect("You've interacted with me!"))
			},
		}.CreateHandler()
	})
}

func init() {
	type ExitChatterState struct {
		Ready bool
	}
	registerEventHandler("exit_chatter", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[ExitChatterState]{
			Init: func(thisEntity EntityReader, globals rpg.GlobalsReader, state ExitChatterState) *HandlerOutput {
				return NewOutput().WithState(ExitChatterState{
					Ready: true,
				})
			},
			TimerComplete: func(thisEntity EntityReader, globals rpg.GlobalsReader, state ExitChatterState, event *EventTimerComplete) *HandlerOutput {
				if event.CreatedBy != thisEntity.GetId() || event.TimerId != "reset" {
					return nil
				}
				state.Ready = true
				return NewOutput().WithState(state)
			},
			EntityZoneActivity: func(thisEntity EntityReader, globals rpg.GlobalsReader, state ExitChatterState, event *EventEntityZoneActivity) *HandlerOutput {
				if event.ZoneId == "exit" &&
					event.IsEntering &&
					event.EntityId == globals.Get(globalVariableNamePlayerId).AsString("unknown") &&
					state.Ready {
					state.Ready = false
					return NewOutput().WithState(state).WithEffects(
						NewSerialPlan(
							NewWaitForConditionEffect(func(s *State, _ float64) bool {
								e, _ := s.entities.GetEntity(s.player)
								return !e.IsMoving()
							}),
							NewChatterEffect(globals.Get(globalVariableNamePlayerId).AsString("unknown"), 5, "I should turn around..."),
							NewTimerEffect(10).WithTimerId("reset"),
						),
					)
				}
				return nil
			},
		}.CreateHandler()
	})
}

const doorStateVariable = "door_state"
const doorOpen, doorClosed = "open", "closed"

func doorState(globals rpg.GlobalsReader) string {
	return globals.Get(doorStateVariable).AsString(doorClosed)
}

func init() {
	registerEventHandler("door_lever", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(thisEntity EntityReader, globals rpg.GlobalsReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithMode(doorState(globals)).
						WithAnimations(map[string][]AnimationReference{
							doorClosed: {
								{Name: "adventure/doors/lever_horizontal:on"},
							},
							doorOpen: {
								{Name: "adventure/doors/lever_horizontal:off"},
							},
						}),
				)
			},
			OnInteract: func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				newValue := doorOpen
				if doorState(globals) == doorOpen {
					newValue = doorClosed
				}
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).WithMode(newValue),
					NewSetWorldStateEffect(doorStateVariable, newValue),
				)
			},
		}.CreateHandler()
	})
}

func init() {
	registerEventHandler("door", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(thisEntity EntityReader, globals rpg.GlobalsReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(thisEntity.GetId()).
						WithIsBlockingIngress(doorState(globals) == doorClosed),
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithMode(doorState(globals)).
						WithAnimations(map[string][]AnimationReference{
							doorClosed: {
								{Name: "adventure/doors/shield_front_1:closed"},
								{Name: "adventure/doors/shield_front_1:waves"},
							},
							doorOpen: {
								{Name: "adventure/doors/shield_front_1:open"},
							},
						}).
						WithLights(map[string][]LightConfig{
							doorClosed: {
								{Color: "#127fd7", Size: 1.5, Modifier: util.Ptr("pulse_slow")},
							},
						}),
					NewAddSoundProviderEffect(thisEntity.GetId()).WithModeBase(NewModeBaseSoundProviderConfig().
						WithSoundOnEnter("closed", SoundEffect{
							Name:    "adventure/props/energy_door_hum",
							Loop:    true,
							Falloff: StandardFalloff,
							Volume:  0.9,
							FadeIn:  time.Millisecond * 1000,
						}).
						WithSoundOnEnter("closed", SoundEffect{
							Name:    "adventure/props/energy_door_on",
							Volume:  0.9,
							Falloff: StandardFalloff,
						}).
						WithSoundOnEnter("open", SoundEffect{
							Name:    "adventure/props/energy_door_off",
							Volume:  0.9,
							Falloff: StandardFalloff,
						})),
				)
			},
			GlobalVariableUpdated: func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventGlobalVariableUpdated) *HandlerOutput {
				if event.Key != doorStateVariable {
					return nil
				}
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(thisEntity.GetId()).
						WithIsBlockingIngress(doorState(globals) == doorClosed),
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithMode(doorState(globals)))
			},
		}.CreateHandler()
	})
}

var nextSpawnedEntityId = atomic.Int64{}

func init() {
	registerEventHandler("spawn_entity", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(thisEntity EntityReader, globals rpg.GlobalsReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithAnimations(map[string][]AnimationReference{
							"": {{
								Name: "adventure/doors/button",
							}},
						}))
			},
			OnInteract: func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if thisEntity.GetId() != event.TargetId {
					return nil
				}
				id := fmt.Sprintf("spawned-%d", nextSpawnedEntityId.Add(1))
				return NewOutput().WithEffects(
					NewResetModeBasedEntityAnimationEffect(thisEntity.GetId()),
					NewSerialPlan(
						NewRegisterEntityEffect().
							WithEntityId(id).
							WithClass("NPC").
							WithProperties(util.NewProps(map[string]any{
								"movement": "static",
							})).
							WithEntityLocation("spawned_entity_start"),
						NewPushEntityBehaviorEffect(id).WithScriptedMotion(EntityBehaviorScriptedMotion{}),
						NewStartScriptedMotionEffect(id).WithToEntityId("spawned_entity_end"),
						NewDeleteEntityEffect(id),
					),
				)
			},
		}.CreateHandler()
	})
}
