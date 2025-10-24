package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
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
				MutateEntity: &events.EffectMutateEntity{
					EntityId: ctx.Id(),
					Mode:     &mode,
				},
				YieldElythium: &events.EffectYieldElythium{
					Amount: 3,
				},
			},
		)
	})
	handler.WithTimerComplete(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventTimerComplete) *events.HandlerOutput {
		if event.CreatedBy != ctx.Id() || event.TimerId != "reset" {
			return nil
		}
		return events.NewOutput().WithEffects(events.Effect{
			MutateEntity: &events.EffectMutateEntity{
				EntityId: ctx.Id(),
				Mode:     util.Ptr("ready"),
			},
		})
	})
	registerDynamicEntity().
		byTile(tiles.Elythium).
		register(func(s *State, entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler) {
			e := NewDynamicEntity(entityId, location).
				WithMode("ready").
				WithLights("ready", &Light{
					RenderDetails: LightRenderDetails{
						SizeScale: 1.5,
						ColorMask: colors.HexString("#f06"),
					},
					Modifiers: []LightModifier{
						&LightModifierPulse{
							PeriodSeconds:       4,
							SizeIntensity:       0.1,
							BrightnessIntensity: 0.4,
						},
					},
				}).
				WithAnimations("ready",
					anim.Load(atlas, "adventure/entities/elythium/crystals", "default"),
					anim.Load(atlas, "adventure/entities/elythium/crystals_sparkle", "default")).
				WithAnimations("mined",
					anim.Load(atlas, "adventure/entities/elythium/crystals_rock", "default"))
			return e, handler.CreateHandler()
		})
}
