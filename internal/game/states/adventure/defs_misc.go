package adventure

import (
	"fmt"
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
)

func init() {
	newRegistrarBuilder().byTile(tiles.RedCoin).
		registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {

			entity := system.RegisterEntity(params.EntityId, params.Location)
			entity.SetRenderer(NewBasicEntityRenderer(entity).
				WithAnimations(anim.RedCoin(atlas)).
				WithLights(NewLightWithModifier(colors.FromString("#f00"), 0.5, "pulse_slow")))
			AttachBlockIngressPresence(entity, false, NewImpassableImpedance())
			return entity, nil
		})
	newRegistrarBuilder().byTile(tiles.Rocket).
		registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {

			entity := system.RegisterEntity(params.EntityId, params.Location)
			entity.SetRenderer(NewBasicEntityRenderer(entity).
				WithAnimations(anim.NewStaticAnimation(tiles.Rocket.From(atlas))))
			AttachBlockIngressPresence(entity, true, NewImpassableImpedance())
			dest := "teleport:" + params.Properties.GetString("destination", "")
			requiredElythium := 2
			handler := NewBasicHandler(None{}).
				WithOnInteract(func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
					if event.TargetId != thisEntity.GetId() {
						return nil
					}
					if globals.Get(rpg.GlobalKeyElythium).AsInt(0) < requiredElythium {
						msg := fmt.Sprintf("You need %d Elythium to travel home!", requiredElythium)
						return NewOutput().WithEffects(NewDialogueEffect(msg))
					}
					return NewOutput().WithSerialPlan(
						NewDialogueEffect("You've managed to escape!"),
						NewMutateEntityBehaviorEffect(globals.Get(globalVariableNamePlayerId).AsString("unknown")).WithDisableBy("rocket"),
						NewYieldElythiumEffect(-requiredElythium),
						NewTeleportPlayerEffect().WithToReference(dest),
						NewMutateEntityBehaviorEffect(globals.Get(globalVariableNamePlayerId).AsString("unknown")).WithEnableBy("rocket"),
					)
				})
			return entity, handler.CreateHandler()
		})
	newRegistrarBuilder().byTile(tiles.DummyFightRobot).
		registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
			entity := system.RegisterEntity(params.EntityId, params.Location)
			entity.SetRenderer(NewBasicEntityRenderer(entity).
				WithAnimations(anim.NewStaticAnimation(atlas.GetSprite("primortals/dummy_entity"))))
			AttachBlockIngressPresence(entity, true, NewImpassableImpedance())
			var dummyQuips = []string{
				"Practice those steps - then try me.",
				"Come close - I don't bite... yet.",
				"Stretch first. I hate easy wins.",
				"Make a move, meatbag.",
				"You look fragile. Let's verify.",
				"Come fight me, big guy.",
			}
			quipTimerId := "dummy-quips-trigger"
			handler := NewBasicHandler(None{}).
				WithInit(func(thisEntity EntityReader, globals StateGlobalsReader, state None) *HandlerOutput {
					return NewOutput().WithEffects(NewTimerEffect(10 + rand.Float64()*10).WithTimerId(quipTimerId))
				}).
				WithTimerComplete(func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventTimerComplete) *HandlerOutput {
					if event.TimerId != quipTimerId {
						return nil
					}
					return NewOutput().WithEffects(
						NewChatterEffect(params.EntityId, 4, dummyQuips[rand.Intn(len(dummyQuips))]),
						NewTimerEffect(10+rand.Float64()*10).WithTimerId(quipTimerId))
				}).
				WithOnInteract(func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
					if thisEntity.GetId() != event.TargetId {
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
					return NewOutput().WithSerialPlan(
						NewDialogueEffect(message),
						NewMutateEntityBehaviorEffect(globals.Get(globalVariableNamePlayerId).AsString("unknown")).WithDisableBy("robot"),
						NewTriggerCombatEffect(rpg.CombatBGSpaceBase).
							WithOpponent(game.CombatOpponent{
								Type: rpg.Primortal_Dummy.Type,
							}),
						NewDialogueEffect("Well, butter my bolts... you actually did it."),
						NewMutateEntityBehaviorEffect(globals.Get(globalVariableNamePlayerId).AsString("unknown")).WithEnableBy("robot"),
					)
				})
			return entity, handler.CreateHandler()
		})
}
