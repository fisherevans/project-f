package handlers

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	Register("captain_chair_front", func(props *util.Properties) events.EventHandler {
		return events.BasicHandlerBuilder[None]{
			EntityZoneActivity: func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventEntityZoneActivity) *events.HandlerOutput {
				pId := world.GetAsString("player_id")
				if event.ZoneId != "captains_area" || pId != event.EntityId || !event.IsEntering {
					return nil
				}
				return events.NewOutput().WithSerialPlan(
					events.NewPushEntityBehaviorEffect(pId).WithScriptedMotion(events.EntityBehaviorScriptedMotion{}),
					events.NewStartScriptedMotionEffect(pId).WithToEntityId("captain_chair_front"),
					events.NewPopEntityBehaviorEffect(pId),
					events.NewEntityFaceDirectionEffect(pId, input.Right),
				)
			},
		}.CreateHandler()
	})
}
