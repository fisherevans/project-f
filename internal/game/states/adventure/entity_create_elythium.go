package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
)

func init() {
	handler := events.NewBasicHandler(None{})
	handler.WithOnInteract(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
		if event.TargetId != ctx.EntityId() || ctx.GetMetadata(types.MetadataKeyMode) == "mined" {
			return nil
		}
		mode := "mined"
		return events.NewOutput().WithEffects(
			events.NewTimerEffect(3).WithTimerId("reset"),
			events.NewYieldElythiumEffect(3),
			events.NewMutateModeBasedEntityEffect(ctx.EntityId()).WithMode(mode),
		)
	})
	handler.WithTimerComplete(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventTimerComplete) *events.HandlerOutput {
		if event.CreatedBy != ctx.EntityId() || event.TimerId != "reset" {
			return nil
		}
		return events.NewOutput().WithEffects(events.NewMutateModeBasedEntityEffect(ctx.EntityId()).WithMode("ready"))
	})
	targetRegistration().
		byTile(tiles.Elythium).
		registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) (events.EntityContext, events.EventHandler) {
			entity := system.RegisterEntity(entityId, location)

			AttachModeBasedEntityRenderer(entity).WithMode("ready").
				WithModeRenderer("ready", NewBasicEntityRenderer().WithAnimations(
					anim.LoadTilesheetAnimation(atlas, "adventure/entities/elythium/crystals", "default"),
					anim.LoadTilesheetAnimation(atlas, "adventure/entities/elythium/crystals_sparkle", "default")).
					WithLights(NewLight(colors.HexString("#f06"), 1.5))).
				WithModeRenderer("mined", NewBasicEntityRenderer().WithAnimations(
					anim.LoadTilesheetAnimation(atlas, "adventure/entities/elythium/crystals_rock", "default")))

			AttachBlockIngressPresence(entity, true)

			return entity.GetEntityContext(), handler.CreateHandler()
		})
}
