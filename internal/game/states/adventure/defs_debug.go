package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	registerDebugButton("debug_button.intro", "redo the intro tutorial", "#fff", func(globals StateGlobalsReader) []Effect {
		return []Effect{NewLoadMapEffect("intro").WithWaypoint("default")}
	})
	registerDebugButton("debug_button.map1", "enter the tech demo map", "#fff", func(globals StateGlobalsReader) []Effect {
		return []Effect{NewLoadMapEffect("map1").WithWaypoint("default")}
	})
	registerDebugButton("debug_button.hq", "travel to HQ", "#fff", func(globals StateGlobalsReader) []Effect {
		return []Effect{NewLoadMapEffect("hq").WithWaypoint("default")}
	})
}

func registerDebugButton(name string, pressAction string, colorMask string, onPress func(reader StateGlobalsReader) []Effect) {
	type debugButtonState struct {
		lastPress float64
	}
	registerEventHandler(name, func(props *util.Properties) EventHandler {
		return BasicHandlerBuilder[debugButtonState]{
			Init: func(thisEntity EntityReader, globals StateGlobalsReader, state debugButtonState) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithAnimations(map[string][]AnimationReference{
							"": {{
								Name:      "adventure/doors/button",
								ColorMask: util.Ptr(colorMask),
							}},
						}))
			},
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state debugButtonState, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				var effects []Effect
				if (game.TimeElapsed() - state.lastPress) > 0.5 {
					effects = []Effect{NewChatterEffect(thisEntity.GetId(), 3, "Double-press to "+pressAction)}
				} else {
					effects = onPress(globals)
				}
				return NewOutput().
					WithEffects(append([]Effect{
						NewResetModeBasedEntityAnimationEffect(thisEntity.GetId()),
					}, effects...)...).
					WithState(debugButtonState{
						lastPress: game.TimeElapsed(),
					})
			},
		}.CreateHandler()
	})

}
