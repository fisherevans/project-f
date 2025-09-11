package adventure

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
)

type EntityInterest struct {
	InnateEntity
	topic string
}

func (e *EntityInterest) RenderScene(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityInterest) Interact(ctx *game.Context, adv *State, source Entity) {
	var msg string
	switch e.topic {
	case "test-tube":
		msg = "The fluid inside the tube glows faintly, casting eerie shadows. Suspended within, the creature twitches, its form shifting between solid and vapor, as if undecided on existing. A single monitor flickers beside it, reading:\n{+u,+c:warm_5,+s}STABILITY{+w:20} {-w}FLUCTUATING{-*}\nWhatever this thing was, it's still {+r:0.05}trying to be{-r}."
	case "knight":
		msg = "The statue is worn, its {+o,+c:white}knight{-o,-c} frozen in time beneath alien dust. Strange {+o,+c:white}symbols {+s}flicker{-s} along{-o,-c} the armor, barely legible:\n{+c:grey_4,+u,+w:2}'To stand is to defy. To fall is to be forgotten.'{-*}\nA {+r}chill{-*} runs down your spine, as if the words were spoken aloud."
	}
	if msg == "" {
		return
	}
	adv.dialogues.Append(NewBasicDialogue(msg, nil))
}

type EntityAnimatedInterest struct {
	InnateEntity
	OnAnimation  *anim.AnimatedSprite
	OffAnimation *anim.AnimatedSprite
	OnMessage    string
	OffMessage   string

	toggled bool
}

func (e *EntityAnimatedInterest) Update(ctx *game.Context, adv *State, timeDelta float64) {
	if e.toggled && e.OnAnimation != nil {
		e.OnAnimation.Update(timeDelta)
	}
	if !e.toggled && e.OffAnimation != nil {
		e.OffAnimation.Update(timeDelta)
	}
}

func (e *EntityAnimatedInterest) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	if e.toggled && e.OnAnimation != nil {
		e.OnAnimation.Sprite().Draw(target, matrix)
	}
	if !e.toggled && e.OffAnimation != nil {
		e.OffAnimation.Sprite().Draw(target, matrix)
	}
}

func (e *EntityAnimatedInterest) Interact(ctx *game.Context, adv *State, source Entity) {
	msg := e.OffMessage
	if e.toggled {
		msg = e.OnMessage
	}
	if msg == "" {
		return
	}
	e.toggled = true
	adv.dialogues.Append(NewBasicDialogue(msg, func(ctx *game.Context, s *State) {
		e.toggled = false
	}))
}
