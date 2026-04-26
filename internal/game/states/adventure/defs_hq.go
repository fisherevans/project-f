package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

func init() {
	registerEventHandler("hq.chair", func(_ *util.Properties) EventHandler {
		chairId := "hq.chair"
		chairFrontId := "hq.chair_front"
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				pId := globals.Player().GetId()
				return NewOutput().WithSerialPlan(
					NewPushEntityBehaviorEffect(pId).WithScriptedMotion(EntityBehaviorScriptedMotion{}),
					NewStartScriptedMotionEffect(pId).WithToEntityId(chairFrontId),
					NewPopEntityBehaviorEffect(pId),
					NewEntityFaceDirectionEffect(event.SourceId).WithDirection(input.Down),
					NewMutateEntityBehaviorEffect(pId).WithDisableBy("computer"),
					NewChangePlayerRendererEffect("hidden"),
					NewMutateModeBasedEntityEffect(chairId).WithMode("enter"),
					NewWaitForAnimationComplete(chairId),
					NewTimerEffect(0.25),
					NewFunctionEffect(func(_ EntityReader, s *State) {
						game.SetActiveStateIntent(game.ComputerIntent{
							Background: s,
						})
					}),
				)
			},
			OnStateEnter: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnStateEnter) *HandlerOutput {
				data, ok := event.Data.(game.ComputerReturnData)
				if !ok {
					return nil
				}
				pId := globals.Player().GetId()
				// just exiting, not lanching
				if data.PlanetName == "" || data.Waypoint == "" {
					return NewOutput().WithSerialPlan(
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
						NewMutateEntityBehaviorEffect(pId).WithEnableBy("computer"),
					)
				}
				fadeId := "hq.chair_fade"
				fadeDuration := 0.75
				animechId := "hq.animech"
				// planet chosen, time to launch!
				return NewOutput().WithSerialPlan(
					NewMutateModeBasedEntityEffect(chairId).WithMode("close_eyes"),
					NewWaitForAnimationComplete(chairId),
					NewFadeEffect(fadeDuration, 1).
						WithFadeId(fadeId).
						WithAutoDeactivate(false).
						WithFromColor("#00000000").
						WithToColor("#000000FF"),
					NewMutateFollowCameraEffect().
						WithFollowEntityId(animechId).
						WithResetPosition(true),
					NewParallelPlan(
						NewFadeEffect(fadeDuration, 1).
							WithAutoDeactivate(true).
							WithFromColor("#000000FF").
							WithToColor("#00000000"),
						NewDeactivateFadeEffect(fadeId),
					),
					NewMutateModeBasedEntityEffect(animechId).WithMode("descend"),
					NewWaitForAnimationComplete(animechId),
					NewLoadMapEffect(data.PlanetName).WithWaypoint(data.Waypoint).WithTravelPlanetName(data.PlanetSpriteName),
				)
			},
		}.CreateHandler()
	})
	// hq.animech, hq.xenolog, hq.bed handlers migrated to assets/scripts/hq/main.yaml
	registerEventHandler("hq.stars", func(*util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			Init: func(thisEntity EntityReader, globals StateGlobalsReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(NewFunctionEffect(func(_ EntityReader, s *State) {
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
					}
				}))
			},
		}.CreateHandler()
	})
}
