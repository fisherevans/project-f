package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
)

func init() {
	handler := NewBasicHandler(None{})
	handler.WithOnInteract(func(ctx EntityContext, world WorldStateReader, state None, event *EventOnInteract) *HandlerOutput {
		if event.TargetId != ctx.EntityId() || ctx.GetMetadata(types.MetadataKeyMode) == "mined" {
			return nil
		}
		mode := "mined"
		return NewOutput().WithEffects(
			NewTimerEffect(3).WithTimerId("reset"),
			NewYieldElythiumEffect(3),
			NewMutateModeBasedEntityEffect(ctx.EntityId()).WithMode(mode),
		)
	})
	handler.WithTimerComplete(func(ctx EntityContext, world WorldStateReader, state None, event *EventTimerComplete) *HandlerOutput {
		if event.CreatedBy != ctx.EntityId() || event.TimerId != "reset" {
			return nil
		}
		return NewOutput().WithEffects(NewMutateModeBasedEntityEffect(ctx.EntityId()).WithMode("ready"))
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
