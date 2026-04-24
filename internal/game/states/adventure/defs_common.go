package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
)

func init() {
	newRegistrarBuilder().byClass("ModeBasedEntity").registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
		entity := system.RegisterEntity(params.EntityId, params.Location)
		AttachPresenceFromConfig(entity, params.Properties)
		renderer := AttachModeBasedEntityRenderer(entity)
		if params.SpriteId != nil {
			renderer.getBasicEntityRenderer("").WithAnimations(anim.NewStaticAnimation(atlas.GetTilesheetSpriteById(*params.SpriteId)))
		}
		renderer.WithPropConfigurations(params.Properties)
		return entity, nil
	})
}

// random_chatters, random_dialogues, door.run_state_based handlers migrated
// to assets/scripts/_shared/common.yaml

func init() {
	newRegistrarBuilder().byClass("DirectInteraction").registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
		entity := system.RegisterEntity(params.EntityId, params.Location)
		AttachBlockIngressPresence(entity, true, NewImpassableImpedance())
		target := params.Properties.GetString("target", "")
		return entity, NewBasicHandler(None{}).
			WithOnInteract(func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if thisEntity.GetId() != event.TargetId {
					return nil
				}
				return NewOutput().WithEffects(NewSendEventEffect(EventOnInteract{
					SourceId:              event.SourceId,
					TargetId:              target,
					SourceFacingDirection: event.SourceFacingDirection,
				}))
			}).CreateHandler()
	})
}
