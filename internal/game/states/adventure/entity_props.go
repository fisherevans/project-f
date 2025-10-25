package adventure

import (
	"fmt"
	"math/rand"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
)

func init() {
	targetRegistration().byTile(tiles.RedCoin).
		registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) events.EventHandler {
			renderer := NewBasicEntityRenderer().
				WithAnimations(anim.RedCoin(atlas)).
				WithLights(NewDynamicLight(colors.FromString("#f00"), 0.5, "pulse_slow"))
			presence := newBlockIngressPresence(false)
			system.RegisterEntity(entityId, location, presence, nil, renderer, nil)
			return nil
		})
	targetRegistration().byTile(tiles.Rocket).
		registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) events.EventHandler {
			renderer := NewBasicEntityRenderer().
				WithAnimations(anim.NewStaticAnimation(tiles.Rocket.From(atlas)))
			presence := newBlockIngressPresence(true)
			system.RegisterEntity(entityId, location, presence, nil, renderer, nil)
			dest := "teleport:" + mapEntity.GetStringMetadata("destination", "")
			requiredElythium := 2
			handler := events.NewBasicHandler(None{}).
				WithOnInteract(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
					if event.TargetId != ctx.Id() {
						return nil
					}
					if world.GetRun().Elythium < requiredElythium {
						msg := fmt.Sprintf("You need %d Elythium to travel home!", requiredElythium)
						return events.NewOutput().WithEffects(events.Effect{
							Dialogue: &events.EffectDialogue{
								Text: msg,
							},
						})
					}
					return events.NewOutput().WithSerialPlan(
						events.Effect{
							Dialogue: &events.EffectDialogue{
								Text: "You've managed to escape!",
							},
						},
						events.Effect{
							YieldElythium: &events.EffectYieldElythium{
								Amount: -requiredElythium,
							},
							TeleportPlayer: &events.EffectTeleportPlayer{
								ToReference: &dest,
							},
						},
					)
				})
			return handler.CreateHandler()
		})
	targetRegistration().byTile(tiles.DummyFightRobot).
		registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) events.EventHandler {
			renderer := NewBasicEntityRenderer().
				WithAnimations(anim.NewStaticAnimation(atlas.GetSprite("primortals/dummy_entity")))
			presence := newBlockIngressPresence(true)
			system.RegisterEntity(entityId, location, presence, nil, renderer, nil)
			var dummyQuips = []string{
				"Practice those steps - then try me.",
				"Come close - I don't bite... yet.",
				"Stretch first. I hate easy wins.",
				"Make a move, meatbag.",
				"You look fragile. Let's verify.",
				"Come fight me, big guy.",
			}
			quipTimerId := "dummy-quips-trigger"
			handler := events.NewBasicHandler(None{}).
				WithInit(func(ctx events.EntityContext, world events.WorldStateReader, state None) *events.HandlerOutput {
					return events.NewOutput().WithEffects(events.Effect{
						Timer: &events.EffectTimer{
							TimerId:         quipTimerId,
							DurationSeconds: 10 + rand.Float64()*10,
						},
					})
				}).
				WithTimerComplete(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventTimerComplete) *events.HandlerOutput {
					if event.TimerId != quipTimerId {
						return nil
					}
					return events.NewOutput().WithEffects(events.Effect{
						Chatter: &events.EffectChatter{
							EntityId:        string(entityId),
							DurationSeconds: 4,
							Message:         dummyQuips[rand.Intn(len(dummyQuips))],
						},
						Timer: &events.EffectTimer{
							TimerId:         quipTimerId,
							DurationSeconds: 10 + rand.Float64()*10,
						},
					})
				}).
				WithOnInteract(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
					if ctx.Id() != event.TargetId {
						return nil
					}
					stutters := []string{"Beep.", "Boop.", "Die.", "Die!", "{+u}DIE!!!{-u"}
					var message string
					for id, s := range stutters {
						if id > 0 {
							message += fmt.Sprintf("{+w:%d} {-w}", id*3)
						}
						message += s
					}
					return events.NewOutput().WithSerialPlan(
						events.Effect{
							Dialogue: &events.EffectDialogue{
								Text: message,
							},
						},
						events.Effect{
							TriggerCombat: &events.EffectTriggerCombat{
								Opponent:   &rpg.Primortal_Dummy.Type,
								Background: "combat/background_space_base",
							},
						},
						events.Effect{
							Dialogue: &events.EffectDialogue{
								Text: "Well, butter my bolts... you actually did it.",
							},
						},
					)
				})
			return handler.CreateHandler()
		})
}
