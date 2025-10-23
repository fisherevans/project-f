package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util"
)

type EntityChest struct {
	InnateEntity
	hasItem bool
	item    string
	Passable
}

func (e *EntityChest) RenderScene(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityChest) Interact(adv *State, source Entity) {
	triggerDialogue := func(msg string) {
		game.DebugNotification("appending dialogue")
		adv.AddSystemEffect(events.Effect{
			Dialogue: &events.EffectDialogue{
				Text: msg,
			},
		})
	}
	if e.hasItem {
		triggerDialogue(util.SingularItemFoundMessageFormats.Randomf(e.item))
		e.hasItem = false
	} else {
		triggerDialogue(util.EmptyChestMessages.Random())
	}
}
