package handlers

import (
	"math/rand/v2"

	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	Register("control_camera", func(props *util.Properties) events.EventHandler {
		controlled := props.GetString("control_id", "")
		return events.BasicHandlerBuilder[None]{
			Init: func(ctx events.EntityContext, world events.WorldStateReader, state None) *events.HandlerOutput {
				return events.NewOutput().WithEffects(events.Effect{
					MutateBlockingPresence: events.NewMutateBlockingPresenceEffect(ctx.EntityId()).WithIsBlockingIngress(false),
					OverrideEntityBehavior: events.NewOverrideEntityBehaviorEffect(controlled).
						WithScriptedMotion(events.EntityBehaviorScriptedMotion{}),
					SetWorldState: events.NewSetWorldStateEffect("control_id", controlled),
				}, events.Effect{
					MutateBlockingPresence: events.NewMutateBlockingPresenceEffect("control_reset").WithIsBlockingIngress(false),
				})
			},
			EntityZoneActivity: func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventEntityZoneActivity) *events.HandlerOutput {
				if event.ZoneId == "control" && world.GetAsString("player_id") == event.EntityId {
					if event.IsEntering {
						return events.NewOutput().WithEffects(events.Effect{
							OverrideCamera: events.NewOverrideCameraEffect().WithFollow(events.FollowCamera{
								EntityId: util.Ptr("control_camera"),
							}),
						})
					} else {
						return events.NewOutput().WithEffects(events.Effect{
							PopCameraOverride: events.NewPopCameraOverrideEffect(true),
						})
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
					return events.NewOutput().WithEffects(events.Effect{
						Plan: events.NewPlanEffect("", []events.PlanStep{{Serial: []events.Effect{
							{
								MutateEntityBehavior: events.NewMutateEntityBehaviorEffect(playerId).WithDisableBy("control"),
								StartScriptedMotion: events.NewStartScriptedMotionEffect("", world.GetAsString("control_id")).
									WithToEntityId("control_reset"),
								OverrideCamera: events.NewOverrideCameraEffect().WithFollow(events.FollowCamera{
									EntityId: util.Ptr("controlled_npc"),
								}),
								Chatter: events.NewChatterEffect("", "controlled_npc", 4, quip),
							},
							{
								MutateEntityBehavior: events.NewMutateEntityBehaviorEffect(playerId).WithEnableBy("control"),
								PopCameraOverride:    events.NewPopCameraOverrideEffect(true),
								EntityFaceDirection:  events.NewEntityFaceDirectionEffect("controlled_npc", input.Down),
							},
						}}}),
					})
				}
				return nil
			},
		}.CreateHandler()
	})
	Register("control_button", func(props *util.Properties) events.EventHandler {
		colorMask := props.GetString("color_mask", "#fff")
		action := props.GetString("action", "")
		return events.BasicHandlerBuilder[None]{
			Init: func(ctx events.EntityContext, world events.WorldStateReader, state None) *events.HandlerOutput {
				return events.NewOutput().WithEffects(events.Effect{
					MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithAnimations(map[string][]types.AnimationReference{
							"": {{
								Name:      "adventure/doors/button",
								ColorMask: util.Ptr(colorMask),
							}},
						}),
				})
			},
			OnInteract: func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
				if event.TargetId != ctx.EntityId() {
					return nil
				}
				output := func(dir input.Direction) events.Effect {
					return events.Effect{
						StartScriptedMotion: events.NewStartScriptedMotionEffect("", world.GetAsString("control_id")).
							WithRelative(events.RelativeLocation{
								Direction: dir,
								Steps:     3,
							}),
					}
				}
				effects := []events.Effect{{
					ResetModeBasedEntityAnimation: events.NewResetModeBasedEntityAnimationEffect(ctx.EntityId()),
				}}
				switch action {
				case "down":
					effects = append(effects, output(input.Down))
				case "right":
					effects = append(effects, output(input.Right))
				case "left":
					effects = append(effects, output(input.Left))
				}
				return events.NewOutput().WithEffects(effects...)
			},
		}.CreateHandler()
	})
}
