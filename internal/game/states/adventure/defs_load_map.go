package adventure

import (
	"fisherevans.com/project/f/internal/util/tiles"
)

func init() {
	newRegistrarBuilder().
		byTile(tiles.LoadMap).
		registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
			entity := system.RegisterEntity(params.EntityId, params.Location)
			system.SetDebugType(params.EntityId, "trigger")
			AttachBlockIngressPresence(entity, true, NewImpassableImpedance())
			return entity, BasicHandlerBuilder[None]{
				OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
					if thisEntity.GetId() != event.TargetId {
						return nil
					}
					mapName := params.Properties.GetString("map_name", "")
					waypoint := params.Properties.GetString("waypoint", "")
					return NewOutput().WithEffects(
						NewLoadMapEffect(mapName).WithWaypoint(waypoint),
					)
				},
			}.CreateHandler()
		})
}
