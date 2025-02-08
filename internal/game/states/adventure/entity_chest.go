package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
)

type EntityChest struct {
	InnateEntity
	hasItem bool
	item    string
}

func (e *EntityChest) Render(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityChest) Interact(ctx *game.Context, adv *State, source Entity) {
	triggerDialogue := func(msg string) {
		ctx.Notify("appending dialogue")
		adv.dialogues.Append(NewBasicDialogue(msg))
	}
	if e.hasItem {
		triggerDialogue(util.SingularItemFoundMessageFormats.Randomf(e.item))
		e.hasItem = false
	} else {
		triggerDialogue(util.EmptyChestMessages.Random())
	}
}
