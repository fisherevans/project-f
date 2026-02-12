package adventure

import (
	"fmt"
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

func init() {
	registerEventHandler("hq.chair", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				chairId := "hq.chair"
				chairFrontId := "hq.chair_front"
				pId := globals.Player().GetId()
				return NewOutput().WithSerialPlan(
					NewPushEntityBehaviorEffect(pId).WithScriptedMotion(EntityBehaviorScriptedMotion{}),
					NewStartScriptedMotionEffect(pId).WithToEntityId(chairFrontId),
					NewPopEntityBehaviorEffect(pId),
					NewEntityFaceDirectionEffect(event.SourceId).WithDirection(input.Down),
					NewChangePlayerRendererEffect("hidden"),
					NewMutateModeBasedEntityEffect(chairId).WithMode("enter"),
					NewWaitForAnimationComplete(chairId),
					NewTimerEffect(0.25),
					NewFunctionEffect(func(s *State) {
						game.SetActiveStateIntent(game.ComputerIntent{
							Background: s,
						})
					}),
					NewTimerEffect(1.0),
					NewMutateModeBasedEntityEffect(chairId).WithMode("close_eyes"),
					NewWaitForAnimationComplete(chairId),
					NewSelfDialogueEffect("now wake up.."),
					NewMutateModeBasedEntityEffect(chairId).WithMode("open_eyes"),
					NewWaitForAnimationComplete(chairId),
					NewMutateModeBasedEntityEffect(chairId).WithMode("exit"),
					NewWaitForAnimationComplete(chairId),
					NewPlaySoundEffect("adventure/beeps/success"),
					// get back out
					NewMutateModeBasedEntityEffect(chairId).WithMode("exit"),
					NewWaitForConditionEffect(func(s *State, td float64) bool {
						e, r, ok := GetModeBasedRenderer(s, chairId)
						if !ok {
							log.Fatal().Msgf("failed to get renderer for %s", chairId)
						}
						return r.getBasicEntityRenderer(ModeMetadataKey.Get(e)).AreAnimationsComplete()
					}),
					NewChangePlayerRendererEffect("human"),
					NewMutateModeBasedEntityEffect(chairId).WithMode(""),
				)
			},
		}.CreateHandler()
	})
	registerEventHandler("hq.animech", func(*util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				level := game.CurrentSave().Animech.Upgrades.GetLevel()
				msg := fmt.Sprintf("It's a shiny, level %d Animech.", level)
				return NewOutput().WithEffects(NewSelfDialogueEffect(msg))
			},
		}.CreateHandler()
	})
	registerEventHandler("hq.xenolog", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				return NewOutput().WithEffects(NewFunctionEffect(func(s *State) {
					s.openXenolog()
				}))
			},
		}.CreateHandler()
	})
	registerEventHandler("hq.bed", func(*util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				if err := game.CurrentSave().Save(); err != nil {
					return NewOutput().WithEffects(
						NewSelfDialogueEffect("oh no."),
						NewSelfDialogueEffect(err.Error()),
					)
				}
				return NewOutput().WithEffects(
					NewSelfDialogueEffect("SAVING... DON'T TURN OFF THE POWER."),
					NewSelfDialogueEffect("You saved the game."),
				)
			},
		}.CreateHandler()
	})
	registerEventHandler("hq.stars", func(*util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(thisEntity EntityReader, globals StateGlobalsReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(NewFunctionEffect(func(s *State) {
					log.Info().Msg("adding stars")
					topLeft, _ := s.entities.GetEntity("hq.stars.top_left")
					bottomLeft, _ := s.entities.GetEntity("hq.stars.bottom_left")
					topRight, _ := s.entities.GetEntity("hq.stars.top_right")
					startX := topLeft.GetPreciseLocation().X - 0.5
					endX := topRight.GetPreciseLocation().X + 0.5
					startY := bottomLeft.GetPreciseLocation().Y - 0.5
					endY := topLeft.GetPreciseLocation().Y + 0.5
					for i := 0; i < 100; i++ {
						x := rand.Float64()*(endX-startX) + startX
						y := rand.Float64()*(endY-startY) + startY
						s.addBackgroundFx(newStarFx(pixel.V(x, y), startX, endX))
						log.Info().Msgf("added star at %v", pixel.V(x, y))
					}
				}))
			},
		}.CreateHandler()
	})
}
