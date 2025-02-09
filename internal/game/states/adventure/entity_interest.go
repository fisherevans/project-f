package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"github.com/gopxl/pixel/v2"
)

type EntityInterest struct {
	InnateEntity
	topic string
}

func (e *EntityInterest) Render(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityInterest) Interact(ctx *game.Context, adv *State, source Entity) {
	var msg string
	switch e.topic {
	case "test-tube":
		msg = "The fluid inside the tube glows faintly, casting eerie shadows. Suspended within, the creature twitches, its form shifting between solid and vapor, as if undecided on existing. A single monitor flickers beside it, reading: 'STABILITY: FLUCTUATING' in ominous red text. Whatever this thing was, it's still trying to be."
	}
	if msg == "" {
		return
	}
	adv.dialogues.Append(NewBasicDialogue(msg))
}
