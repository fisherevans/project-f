package adventure

import (
	"fmt"
	"sync/atomic"

	"fisherevans.com/project/f/internal/util"
)

func init() {
	registerEventHandler("map1", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(ctx EntityContext, gameState GameState, state None, event *EventOnInteract) *HandlerOutput {
				if ctx.EntityId() != event.TargetId {
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
			Init: func(ctx EntityContext, gameState GameState, state ExitChatterState) *HandlerOutput {
				return NewOutput().WithState(ExitChatterState{
					Ready: true,
				})
			},
			TimerComplete: func(ctx EntityContext, gameState GameState, state ExitChatterState, event *EventTimerComplete) *HandlerOutput {
				if event.CreatedBy != ctx.EntityId() || event.TimerId != "reset" {
					return nil
				}
				state.Ready = true
				return NewOutput().WithState(state)
			},
			EntityZoneActivity: func(ctx EntityContext, gameState GameState, state ExitChatterState, event *EventEntityZoneActivity) *HandlerOutput {
				if event.ZoneId == "exit" &&
					event.IsEntering &&
					event.EntityId == gameState.RunState().Get(runStateKeyPlayerId).AsString("unknown") &&
					state.Ready {
					state.Ready = false
					return NewOutput().WithState(state).WithEffects(
						NewSerialPlan(
							NewWaitForConditionEffect(func(s *State, _ float64) bool {
								e, _ := s.entities.GetEntity(s.player)
								return !e.IsMoving()
							}),
							NewChatterEffect(gameState.RunState().Get(runStateKeyPlayerId).AsString("unknown"), 5, "I should turn around..."),
							NewTimerEffect(10).WithTimerId("reset"),
						),
					)
				}
				return nil
			},
		}.CreateHandler()
	})
}

const doorStateKey = "door_state"
const doorOpen, doorClosed = "open", "closed"

func doorState(gameState GameState) string {
	return gameState.WorldState().Get(doorStateKey).AsString(doorClosed)
}

func init() {
	registerEventHandler("door_lever", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(ctx EntityContext, gameState GameState, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithMode(doorState(gameState)).
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
			OnInteract: func(ctx EntityContext, gameState GameState, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != ctx.EntityId() {
					return nil
				}
				newValue := doorOpen
				if doorState(gameState) == doorOpen {
					newValue = doorClosed
				}
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(ctx.EntityId()).WithMode(newValue),
					NewSetWorldStateEffect(doorStateKey, newValue),
				)
			},
		}.CreateHandler()
	})
}

func init() {
	registerEventHandler("door", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(ctx EntityContext, gameState GameState, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(ctx.EntityId()).
						WithIsBlockingIngress(doorState(gameState) == doorClosed),
					NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithMode(doorState(gameState)).
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
				)
			},
			WorldStateUpdated: func(ctx EntityContext, gameState GameState, state None, event *EventWorldStateUpdated) *HandlerOutput {
				if event.Key != doorStateKey {
					return nil
				}
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(ctx.EntityId()).
						WithIsBlockingIngress(doorState(gameState) == doorClosed),
					NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithMode(doorState(gameState)))
			},
		}.CreateHandler()
	})
}

var nextSpawnedEntityId = atomic.Int64{}

func init() {
	registerEventHandler("spawn_entity", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(ctx EntityContext, gameState GameState, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithAnimations(map[string][]AnimationReference{
							"": {{
								Name: "adventure/doors/button",
							}},
						}))
			},
			OnInteract: func(ctx EntityContext, gameState GameState, state None, event *EventOnInteract) *HandlerOutput {
				if ctx.EntityId() != event.TargetId {
					return nil
				}
				id := fmt.Sprintf("spawned-%d", nextSpawnedEntityId.Add(1))
				return NewOutput().WithEffects(
					NewResetModeBasedEntityAnimationEffect(ctx.EntityId()),
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
