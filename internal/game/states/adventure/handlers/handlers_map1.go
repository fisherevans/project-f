package handlers

import (
	"fisherevans.com/project/f/internal/game/events"
)

func init() {
	Register("map1", events.BasicHandlerBuilder[None]{
		OnInteract: func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
			if ctx.Id() != event.TargetId {
				return nil
			}
			return events.NewOutput().WithEffects(events.Effect{
				Dialogue: &events.EffectDialogue{
					Text: "You've interacted with me!",
				},
			})
		},
	}.CreateHandler)
}

func init() {
	type ExitChatterState struct {
		Ready bool
	}
	Register("exit_chatter", events.BasicHandlerBuilder[ExitChatterState]{
		Init: func(ctx events.EntityContext, world events.WorldStateReader, state ExitChatterState) *events.HandlerOutput {
			return events.NewOutput().WithState(ExitChatterState{
				Ready: true,
			})
		},
		TimerComplete: func(ctx events.EntityContext, world events.WorldStateReader, state ExitChatterState, event *events.EventTimerComplete) *events.HandlerOutput {
			if event.CreatedBy != ctx.Id() || event.TimerId != "reset" {
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
	}.CreateHandler)
}

func init() {
	Register("door_lever", events.BasicHandlerBuilder[None]{
		Init: func(ctx events.EntityContext, world events.WorldStateReader, state None) *events.HandlerOutput {
			return events.NewOutput().WithEffects(
				events.Effect{
					MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.Id()).
						WithMode("off").
						WithAnimations(map[string][]events.AnimationReference{
							"on": {
								{Tilesheet: "adventure/doors/lever_horizontal", Name: "on"},
							},
							"off": {
								{Tilesheet: "adventure/doors/lever_horizontal", Name: "off"},
							},
						}),
				},
				events.Effect{
					SetWorldState: events.NewSetWorldStateEffect("door_open", false),
				},
			)
		},
		OnInteract: func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
			if event.TargetId != ctx.Id() {
				return nil
			}

			if ctx.Mode() == "off" {
				return events.NewOutput().WithEffects(
					events.Effect{
						MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.Id()).WithMode("on"),
					},
					events.Effect{
						SetWorldState: events.NewSetWorldStateEffect("door_open", true),
					},
				)
			} else {
				return events.NewOutput().WithEffects(
					events.Effect{
						MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.Id()).WithMode("off"),
					},
					events.Effect{
						SetWorldState: events.NewSetWorldStateEffect("door_open", false),
					},
				)
			}
		},
	}.CreateHandler)
}

func init() {
	Register("door", events.BasicHandlerBuilder[None]{
		Init: func(ctx events.EntityContext, world events.WorldStateReader, state None) *events.HandlerOutput {
			return events.NewOutput().WithEffects(
				events.Effect{
					MutateBlockingPresence: events.NewMutateBlockingPresenceEffect(ctx.Id()).
						WithIsBlockingIngress(true),
					MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.Id()).
						WithMode("closed").
						WithAnimations(map[string][]events.AnimationReference{
							"closed": {
								{Tilesheet: "adventure/doors/shield_front_1", Name: "closed"},
								{Tilesheet: "adventure/doors/shield_front_1", Name: "waves"},
							},
							"open": {
								{Tilesheet: "adventure/doors/shield_front_1", Name: "open"},
							},
						}).
						WithLights(map[string][]events.LightConfig{
							"closed": {
								{Color: "#127fd7", Size: 1.5, Modifier: "pulse_slow"},
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
					MutateBlockingPresence: events.NewMutateBlockingPresenceEffect(ctx.Id()).
						WithIsBlockingIngress(false),
					MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.Id()).
						WithMode("open"),
				})
			} else {
				return events.NewOutput().WithEffects(events.Effect{
					MutateBlockingPresence: events.NewMutateBlockingPresenceEffect(ctx.Id()).
						WithIsBlockingIngress(true),
					MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.Id()).
						WithMode("closed"),
				})
			}
		},
	}.CreateHandler)
}
