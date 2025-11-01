package adventure

import (
	"math/rand/v2"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	Register("control_camera", func(props *util.Properties) EventHandler {
		controlled := props.GetString("control_id", "")
		return BasicHandlerBuilder[None]{
			Init: func(ctx EntityContext, world WorldStateReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(ctx.EntityId()).WithIsBlockingIngress(false),
					NewPushEntityBehaviorEffect(controlled).
						WithScriptedMotion(EntityBehaviorScriptedMotion{}),
					NewSetWorldStateEffect("control_id", controlled),
					NewMutateBlockingPresenceEffect("control_reset").WithIsBlockingIngress(false),
				)
			},
			EntityZoneActivity: func(ctx EntityContext, world WorldStateReader, state None, event *EventEntityZoneActivity) *HandlerOutput {
				if event.ZoneId == "control" && world.GetAsString("player_id") == event.EntityId {
					if event.IsEntering {
						return NewOutput().WithEffects(
							NewOverrideCameraEffect().WithFollow(FollowCamera{
								EntityId: util.Ptr("control_camera"),
							}),
						)
					} else {
						return NewOutput().WithEffects(
							NewPopCameraOverrideEffect(true),
						)
					}
				}
				playerId := world.GetAsString("player_id")
				if ("reset_left" == event.ZoneId || "reset_right" == event.ZoneId) && event.IsEntering {
					quips := []string{
						"Why am I moving?!",
						"I didn't say 'go there'!",
						"This isn't me, I swear.",
						"Quit puppeteering me!",
						"My legs! They've gone rogue!",
						"What... what's happening?!?!",
						"This is deeply unsettling.",
						"Do you mind? I just want to go home!",
					}
					quip := quips[rand.IntN(len(quips))]
					return NewOutput().WithSerialPlan(
						NewMutateEntityBehaviorEffect(playerId).WithDisableBy("control"),
						NewOverrideCameraEffect().WithFollow(FollowCamera{EntityId: util.Ptr("controlled_npc")}),
						NewParallelPlan(
							NewSerialPlan(
								NewTimerEffect(1),
								NewStartScriptedMotionEffect(world.GetAsString("control_id")).WithToEntityId("control_reset"),
							),
							NewChatterEffect("controlled_npc", 4, quip),
						),
						NewMutateEntityBehaviorEffect(playerId).WithEnableBy("control"),
						NewPopCameraOverrideEffect(true),
						NewEntityFaceDirectionEffect("controlled_npc", input.Down))
				}
				return nil
			},
		}.CreateHandler()
	})
	Register("control_button", func(props *util.Properties) EventHandler {
		colorMask := props.GetString("color_mask", "#fff")
		action := props.GetString("action", "")
		return BasicHandlerBuilder[None]{
			Init: func(ctx EntityContext, world WorldStateReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithAnimations(map[string][]types.AnimationReference{
							"": {{
								Name:      "adventure/doors/button",
								ColorMask: util.Ptr(colorMask),
							}},
						}))
			},
			OnInteract: func(ctx EntityContext, world WorldStateReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != ctx.EntityId() {
					return nil
				}
				output := func(dir input.Direction) Effect {
					return NewStartScriptedMotionEffect(world.GetAsString("control_id")).
						WithRelative(RelativeLocation{
							Direction: dir,
							Steps:     3,
						})
				}
				effects := []Effect{
					NewResetModeBasedEntityAnimationEffect(ctx.EntityId()),
				}
				switch action {
				case "down":
					effects = append(effects, output(input.Down))
				case "right":
					effects = append(effects, output(input.Right))
				case "left":
					effects = append(effects, output(input.Left))
				}
				return NewOutput().WithEffects(effects...)
			},
		}.CreateHandler()
	})
}
