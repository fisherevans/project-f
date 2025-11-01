package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	Register("captain_chair_front", func(props *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			EntityZoneActivity: func(ctx EntityContext, world WorldStateReader, state None, event *EventEntityZoneActivity) *HandlerOutput {
				pId := world.GetAsString("player_id")
				if event.ZoneId != "captains_area" || pId != event.EntityId || !event.IsEntering {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewPushEntityBehaviorEffect(pId).WithScriptedMotion(EntityBehaviorScriptedMotion{}),
					NewStartScriptedMotionEffect(pId).WithToEntityId("captain_chair_front"),
					NewPopEntityBehaviorEffect(pId),
					NewEntityFaceDirectionEffect(pId, input.Right),
				)
			},
		}.CreateHandler()
	})
}
