package handlers

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	Register("map1", func(_ *util.Properties) events.EventHandler {
		return events.BasicHandlerBuilder[None]{
			OnInteract: func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
				if ctx.EntityId() != event.TargetId {
					return nil
				}
				return events.NewOutput().WithEffects(events.Effect{
					Dialogue: &events.EffectDialogue{
						Text: "You've interacted with me!",
					},
				})
			},
		}.CreateHandler()
	})
}

func init() {
	type ExitChatterState struct {
		Ready bool
	}
	Register("exit_chatter", func(_ *util.Properties) events.EventHandler {
		return events.BasicHandlerBuilder[ExitChatterState]{
			Init: func(ctx events.EntityContext, world events.WorldStateReader, state ExitChatterState) *events.HandlerOutput {
				return events.NewOutput().WithState(ExitChatterState{
					Ready: true,
				})
			},
			TimerComplete: func(ctx events.EntityContext, world events.WorldStateReader, state ExitChatterState, event *events.EventTimerComplete) *events.HandlerOutput {
				if event.CreatedBy != ctx.EntityId() || event.TimerId != "reset" {
					return nil
				}
				state.Ready = true
				return events.NewOutput().WithState(state)
			},
			EntityZoneActivity: func(ctx events.EntityContext, world events.WorldStateReader, state ExitChatterState, event *events.EventEntityZoneActivity) *events.HandlerOutput {
				if event.ZoneId == "exit" &&
					event.IsEntering &&
					event.EntityId == world.GetAsString("player_id") &&
					state.Ready {
					state.Ready = false
					return events.NewOutput().WithState(state).WithEffects(
						events.Effect{
							Timer: events.NewTimerEffect("reset", 10),
						},
						events.Effect{
							Chatter: events.NewChatterEffect("", world.GetAsString("player_id"), 5, "I should turn around..."),
						},
					)
				}
				return nil
			},
		}.CreateHandler()
	})
}

func init() {
	Register("door_lever", func(_ *util.Properties) events.EventHandler {
		return events.BasicHandlerBuilder[None]{
			Init: func(ctx events.EntityContext, world events.WorldStateReader, state None) *events.HandlerOutput {
				return events.NewOutput().WithEffects(
					events.Effect{
						MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.EntityId()).
							WithMode("off").
							WithAnimations(map[string][]types.AnimationReference{
								"on": {
									{Name: "adventure/doors/lever_horizontal:on"},
								},
								"off": {
									{Name: "adventure/doors/lever_horizontal:off"},
								},
							}),
					},
					events.Effect{
						SetWorldState: events.NewSetWorldStateEffect("door_open", false),
					},
				)
			},
			OnInteract: func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
				if event.TargetId != ctx.EntityId() {
					return nil
				}

				if ctx.GetStringMetadata(types.MetadataKeyMode) == "off" {
					return events.NewOutput().WithEffects(
						events.Effect{
							MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.EntityId()).WithMode("on"),
						},
						events.Effect{
							SetWorldState: events.NewSetWorldStateEffect("door_open", true),
						},
					)
				} else {
					return events.NewOutput().WithEffects(
						events.Effect{
							MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.EntityId()).WithMode("off"),
						},
						events.Effect{
							SetWorldState: events.NewSetWorldStateEffect("door_open", false),
						},
					)
				}
			},
		}.CreateHandler()
	})
}

func init() {
	Register("door", func(_ *util.Properties) events.EventHandler {
		return events.BasicHandlerBuilder[None]{
			Init: func(ctx events.EntityContext, world events.WorldStateReader, state None) *events.HandlerOutput {
				return events.NewOutput().WithEffects(
					events.Effect{
						MutateBlockingPresence: events.NewMutateBlockingPresenceEffect(ctx.EntityId()).
							WithIsBlockingIngress(true),
						MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.EntityId()).
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
					},
				)
			},
			WorldStateUpdated: func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventWorldStateUpdated) *events.HandlerOutput {
				if event.Key != "door_open" {
					return nil
				}
				if doorOpen, ok := event.NewValue.(bool); ok && doorOpen {
					return events.NewOutput().WithEffects(events.Effect{
						MutateBlockingPresence: events.NewMutateBlockingPresenceEffect(ctx.EntityId()).
							WithIsBlockingIngress(false),
						MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.EntityId()).
							WithMode("open"),
					})
				} else {
					return events.NewOutput().WithEffects(events.Effect{
						MutateBlockingPresence: events.NewMutateBlockingPresenceEffect(ctx.EntityId()).
							WithIsBlockingIngress(true),
						MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.EntityId()).
							WithMode("closed"),
					})
				}
			},
		}.CreateHandler()
	})
}
