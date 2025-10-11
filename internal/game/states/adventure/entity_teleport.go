package adventure

import (
	"fmt"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/util/tiles"

	"fisherevans.com/project/f/internal/game/anim"
)

type EntityTeleport struct {
	InnateEntity
	RequiredElythium int
	Destination      TeleportReference
}

func (e *EntityTeleport) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	tiles.Rocket.From(atlas).Draw(target, matrix)
}

func (e *EntityTeleport) Interact(adv *State, source Entity) {
	player, IsPlayer := source.(*Player)
	if !IsPlayer {
		log.Warn().Msg("EntityTeleport interact: source is not a Player")
		return
	}

	if adv.hud.ElythiumCount < e.RequiredElythium {
		adv.dialogues.Append(NewBasicDialogue(fmt.Sprintf("You need %d Elythium to travel home!", e.RequiredElythium), nil))
		return
	}

	destination, exists := adv.teleports[e.Destination]
	if !exists {
		log.Warn().Msgf("EntityTeleport interact: destination '%s' not found", e.Destination)
		return
	}

	adv.dialogues.Append(NewBasicDialogue("You've managed to escape!", func(s *State) {
		adv.hud.ElythiumCount -= e.RequiredElythium
		adv.teleport(player, destination)
	}))
}

type EntityAnimatedTeleport struct {
	InnateEntity
	OnAnimation  *anim.AnimatedSprite
	OffAnimation *anim.AnimatedSprite
	OnMessage    string
	OffMessage   string

	toggled bool
}

func (e *EntityAnimatedTeleport) Update(adv *State, timeDelta float64) {
	if e.toggled && e.OnAnimation != nil {
		e.OnAnimation.Update(timeDelta)
	}
	if !e.toggled && e.OffAnimation != nil {
		e.OffAnimation.Update(timeDelta)
	}
}

func (e *EntityAnimatedTeleport) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	if e.toggled && e.OnAnimation != nil {
		e.OnAnimation.Sprite().Draw(target, matrix)
	}
	if !e.toggled && e.OffAnimation != nil {
		e.OffAnimation.Sprite().Draw(target, matrix)
	}
}

func (e *EntityAnimatedTeleport) Interact(adv *State, source Entity) {
	msg := e.OffMessage
	if e.toggled {
		msg = e.OnMessage
	}
	if msg == "" {
		return
	}
	e.toggled = true
	adv.dialogues.Append(NewBasicDialogue(msg, func(s *State) {
		e.toggled = false
	}))
}
