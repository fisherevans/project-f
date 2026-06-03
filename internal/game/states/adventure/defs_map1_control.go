package adventure

import (
	"fisherevans.com/project/f/internal/util/rng"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	registerEventHandler("control_camera", func(props *util.Properties) EventHandler {
		controlled := props.GetString("control_id", "")
		return BasicHandlerBuilder[None]{
			Init: func(thisEntity EntityReader, globals StateGlobalsReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(thisEntity.GetId()).WithIsBlockingIngress(false),
					NewPushEntityBehaviorEffect(controlled).
						WithScriptedMotion(EntityBehaviorScriptedMotion{}),
					NewSetGlobalEffect("control_id", controlled),
					NewMutateBlockingPresenceEffect("control_reset").WithIsBlockingIngress(false),
				)
			},
			EntityZoneActivity: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventEntityZoneActivity) *HandlerOutput {
				if event.ZoneId == "control" && globals.Get(globalVariableNamePlayerId).AsString("unknown") == event.EntityId {
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
				playerId := globals.Get(globalVariableNamePlayerId).AsString("unknown")
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
					quip := quips[rng.IntN(len(quips))]
					return NewOutput().WithSerialPlan(
						NewMutateEntityBehaviorEffect(playerId).WithDisableBy("control"),
						NewOverrideCameraEffect().WithFollow(FollowCamera{EntityId: util.Ptr("controlled_npc")}),
						NewParallelPlan(
							NewSerialPlan(
								NewTimerEffect(1),
								NewStartScriptedMotionEffect(globals.Get("control_id").AsString("")).WithToEntityId("control_reset"),
							),
							NewChatterEffect("controlled_npc", 4, quip),
						),
						NewMutateEntityBehaviorEffect(playerId).WithEnableBy("control"),
						NewPopCameraOverrideEffect(true),
						NewEntityFaceDirectionEffect("controlled_npc").WithDirection(input.Down))
				}
				return nil
			},
		}.CreateHandler()
	})
	registerEventHandler("control_button", func(props *util.Properties) EventHandler {
		colorMask := props.GetString("color_mask", "#fff")
		action := props.GetString("action", "")
		return BasicHandlerBuilder[None]{
			Init: func(thisEntity EntityReader, globals StateGlobalsReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithAnimations(map[string][]AnimationReference{
							"": {{
								Name:      "adventure/doors/button",
								ColorMask: util.Ptr(colorMask),
							}},
						}))
			},
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				output := func(dir input.Direction) Effect {
					return NewStartScriptedMotionEffect(globals.Get("control_id").AsString("")).
						WithRelative(RelativeLocation{
							Direction: dir,
							Steps:     3,
						})
				}
				effects := []Effect{
					NewResetModeBasedEntityAnimationEffect(thisEntity.GetId()),
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
