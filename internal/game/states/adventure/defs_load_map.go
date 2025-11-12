package adventure

import (
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/tiles"
)

func init() {
	newRegistrarBuilder().
		byTile(tiles.LoadMap).
		registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
			entity := system.RegisterEntity(params.EntityId, params.Location)
			AttachBlockIngressPresence(entity, true, NewImpassableImpedance())
			return entity, BasicHandlerBuilder[None]{
				OnInteract: func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
					if thisEntity.GetId() != event.TargetId {
						return nil
					}
					return NewOutput().WithEffects(
						NewLoadMapEffect(params.Properties.GetString("map_name", "")),
					)
				},
			}.CreateHandler()
		})
}
