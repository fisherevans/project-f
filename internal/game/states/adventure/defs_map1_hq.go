package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	registerEventHandler("captain_chair_front", func(props *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			EntityZoneActivity: func(ctx EntityContext, gameState GameState, state None, event *EventEntityZoneActivity) *HandlerOutput {
				pId := gameState.RunState().Get(runStateKeyPlayerId).AsString("unknown")
				if event.ZoneId != "captains_area" || pId != event.EntityId || !event.IsEntering {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewPushEntityBehaviorEffect(pId).WithScriptedMotion(EntityBehaviorScriptedMotion{}),
					NewStartScriptedMotionEffect(pId).WithToEntityId("captain_chair_front"),
					NewPopEntityBehaviorEffect(pId),
					NewEntityFaceDirectionEffect(pId).WithDirection(input.Right),
				)
			},
		}.CreateHandler()
	})
}
