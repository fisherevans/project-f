package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
)

func init() {
	handler := NewBasicHandler(None{})
	handler.WithOnInteract(func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
		if event.TargetId != thisEntity.GetId() || ModeMetadataKey.Get(thisEntity) == "mined" {
			return nil
		}
		newMode := "mined"
		return NewOutput().WithEffects(
			NewPlaySoundEffect("adventure/elythium_breaking"),
			NewTimerEffect(3).WithTimerId("reset"),
			NewYieldElythiumEffect(3),
			NewMutateModeBasedEntityEffect(thisEntity.GetId()).WithMode(newMode),
		)
	})
	handler.WithTimerComplete(func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventTimerComplete) *HandlerOutput {
		if event.CreatedBy != thisEntity.GetId() || event.TimerId != "reset" {
			return nil
		}
		return NewOutput().WithEffects(NewMutateModeBasedEntityEffect(thisEntity.GetId()).WithMode("ready"))
	})
	newRegistrarBuilder().
		byTile(tiles.Elythium).
		registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
			entity := system.RegisterEntity(params.EntityId, params.Location)

			AttachModeBasedEntityRenderer(entity).WithMode("ready").
				WithModeRenderer("ready", NewBasicEntityRenderer(entity).WithAnimations(
					anim.LoadTilesheetAnimation(atlas, "adventure/entities/elythium/crystals", "default"),
					anim.LoadTilesheetAnimation(atlas, "adventure/entities/elythium/crystals_sparkle", "default")).
					WithLights(NewLight(colors.HexString("#f06"), 1.5))).
				WithModeRenderer("mined", NewBasicEntityRenderer(entity).WithAnimations(
					anim.LoadTilesheetAnimation(atlas, "adventure/entities/elythium/crystals_rock", "default")))

			AttachBlockIngressPresence(entity, true, NewImpassableImpedance())

			return entity, handler.CreateHandler()
		})
}
