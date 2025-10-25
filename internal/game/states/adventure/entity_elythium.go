package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/tiles"
)

func init() {
	handler := events.NewBasicHandler(None{})
	handler.WithOnInteract(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
		if event.TargetId != ctx.Id() || ctx.Mode() == "mined" {
			return nil
		}
		mode := "mined"
		return events.NewOutput().WithEffects(
			events.Effect{
				Timer: &events.EffectTimer{
					TimerId:         "reset",
					DurationSeconds: 3,
				},
				YieldElythium: &events.EffectYieldElythium{
					Amount: 3,
				},
				MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.Id()).WithMode(mode),
			},
		)
	})
	handler.WithTimerComplete(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventTimerComplete) *events.HandlerOutput {
		if event.CreatedBy != ctx.Id() || event.TimerId != "reset" {
			return nil
		}
		return events.NewOutput().WithEffects(events.Effect{
			MutateModeBasedEntity: events.NewMutateModeBasedEntityEffect(ctx.Id()).WithMode("ready"),
		})
	})
	targetRegistration().
		byTile(tiles.Elythium).
		registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) events.EventHandler {
			renderer := NewModeBasedEntityRenderer(entityId, system)
			renderer.WithModeRenderer("ready", NewBasicEntityRenderer().WithAnimations(
				anim.Load(atlas, "adventure/entities/elythium/crystals", "default"),
				anim.Load(atlas, "adventure/entities/elythium/crystals_sparkle", "default")))
			renderer.WithModeRenderer("mind", NewBasicEntityRenderer().WithAnimations(
				anim.Load(atlas, "adventure/entities/elythium/crystals_rock", "default")))
			system.RegisterEntity(entityId, location, nil, nil, renderer, nil)
			return handler.CreateHandler()
		})
}
