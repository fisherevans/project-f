package adventure

import (
	"fmt"
	"sync/atomic"

	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	Register("map1", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(ctx EntityContext, world WorldStateReader, state None, event *EventOnInteract) *HandlerOutput {
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
	Register("exit_chatter", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[ExitChatterState]{
			Init: func(ctx EntityContext, world WorldStateReader, state ExitChatterState) *HandlerOutput {
				return NewOutput().WithState(ExitChatterState{
					Ready: true,
				})
			},
			TimerComplete: func(ctx EntityContext, world WorldStateReader, state ExitChatterState, event *EventTimerComplete) *HandlerOutput {
				if event.CreatedBy != ctx.EntityId() || event.TimerId != "reset" {
					return nil
				}
				state.Ready = true
				return NewOutput().WithState(state)
			},
			EntityZoneActivity: func(ctx EntityContext, world WorldStateReader, state ExitChatterState, event *EventEntityZoneActivity) *HandlerOutput {
				if event.ZoneId == "exit" &&
					event.IsEntering &&
					event.EntityId == world.GetAsString("player_id") &&
					state.Ready {
					state.Ready = false
					return NewOutput().WithState(state).WithEffects(
						NewSerialPlan(
							NewWaitForConditionEffect(func(s *State, _ float64) bool {
								e, _ := s.entities.GetEntity(s.player)
								return !e.IsMoving()
							}),
							NewChatterEffect(world.GetAsString("player_id"), 5, "I should turn around..."),
							NewTimerEffect(10).WithTimerId("reset"),
						),
					)
				}
				return nil
			},
		}.CreateHandler()
	})
}

func init() {
	Register("door_lever", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(ctx EntityContext, world WorldStateReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithMode("off").
						WithAnimations(map[string][]types.AnimationReference{
							"on": {
								{Name: "adventure/doors/lever_horizontal:on"},
							},
							"off": {
								{Name: "adventure/doors/lever_horizontal:off"},
							},
						}),
					NewSetWorldStateEffect("door_open", false),
				)
			},
			OnInteract: func(ctx EntityContext, world WorldStateReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != ctx.EntityId() {
					return nil
				}

				if ctx.GetStringMetadata(types.MetadataKeyMode) == "off" {
					return NewOutput().WithEffects(
						NewMutateModeBasedEntityEffect(ctx.EntityId()).WithMode("on"),
						NewSetWorldStateEffect("door_open", true),
					)
				} else {
					return NewOutput().WithEffects(
						NewMutateModeBasedEntityEffect(ctx.EntityId()).WithMode("off"),
						NewSetWorldStateEffect("door_open", false),
					)
				}
			},
		}.CreateHandler()
	})
}

func init() {
	Register("door", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(ctx EntityContext, world WorldStateReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(ctx.EntityId()).
						WithIsBlockingIngress(true),
					NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithMode("closed").
						WithAnimations(map[string][]types.AnimationReference{
							"closed": {
								{Name: "adventure/doors/shield_front_1:closed"},
								{Name: "adventure/doors/shield_front_1:waves"},
							},
							"open": {
								{Name: "adventure/doors/shield_front_1:open"},
							},
						}).
						WithLights(map[string][]types.LightConfig{
							"closed": {
								{Color: "#127fd7", Size: 1.5, Modifier: util.Ptr("pulse_slow")},
							},
						}),
				)
			},
			WorldStateUpdated: func(ctx EntityContext, world WorldStateReader, state None, event *EventWorldStateUpdated) *HandlerOutput {
				if event.Key != "door_open" {
					return nil
				}
				if doorOpen, ok := event.NewValue.(bool); ok && doorOpen {
					return NewOutput().WithEffects(
						NewMutateBlockingPresenceEffect(ctx.EntityId()).
							WithIsBlockingIngress(false),
						NewMutateModeBasedEntityEffect(ctx.EntityId()).
							WithMode("open"))
				} else {
					return NewOutput().WithEffects(
						NewMutateBlockingPresenceEffect(ctx.EntityId()).
							WithIsBlockingIngress(true),
						NewMutateModeBasedEntityEffect(ctx.EntityId()).
							WithMode("closed"))
				}
			},
		}.CreateHandler()
	})
}

var nextSpawnedEntityId = atomic.Int64{}

func init() {
	Register("spawn_entity", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(ctx EntityContext, world WorldStateReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithAnimations(map[string][]types.AnimationReference{
							"": {{
								Name: "adventure/doors/button",
							}},
						}))
			},
			OnInteract: func(ctx EntityContext, world WorldStateReader, state None, event *EventOnInteract) *HandlerOutput {
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
