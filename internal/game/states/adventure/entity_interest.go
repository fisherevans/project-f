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
		msg = "The fluid inside the tube glows faintly, casting eerie shadows. Suspended within, the creature twitches, its form shifting between solid and vapor, as if undecided on existing. A single monitor flickers beside it, reading:\n{+u,+c:warm_5,+s}STABILITY{+w:20} {-w}FLUCTUATING{-*}\nWhatever this thing was, it's still {+r:0.05}trying to be{-r}."
	case "knight":
		msg = "The statue is worn, its knight frozen in time beneath alien dust. Strange symbols flicker along the armor, barely legible:\n{+c:grey_4,+u,+w:2}'To stand is to defy. To fall is to be forgotten.'{-*}\nA {+r}chill{-*} runs down your spine, as if the words were spoken aloud."
	}
	if msg == "" {
		return
	}
	adv.dialogues.Append(NewBasicDialogue(msg))
}
