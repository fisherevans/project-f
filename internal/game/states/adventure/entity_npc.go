package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"math/rand"
)

type NPC struct {
	MoveableEntity
	DoesMove      bool
	HorizOnly     bool
	LastDirection input.Direction
}

func (n *NPC) Update(ctx *game.Context, adv *State, timeDelta float64) {
	if !n.DoesMove || n.Moving {
		return
	}
	dx, dy := n.LastDirection.GetVector()
	if n.TriggerMovement(adv, dx, dy) {
		return
	}
	var dir input.Direction
	if n.HorizOnly {
		if n.LastDirection == input.NotPressed {
			dir = input.Left
		} else {
			dir = n.LastDirection.Opposite()
		}
	} else {
		dir = input.Directions[int(rand.Float64()*float64(len(input.Directions)))]
	}
	n.LastDirection = dir
	dx, dy = dir.GetVector()
	n.TriggerMovement(adv, dx, dy)
}

func (n *NPC) Sprite() *pixel.Sprite {
	if n.DoesMove {
		ref := resources.Sprites[resources.SpriteId{
			Tilesheet: "ui",
			Column:    3,
			Row:       2,
		}]
		return ref.Sprite
	}
	return resources.Sprites[npcRandomSpriteId].Sprite
}

func (n *NPC) Interact(ctx *game.Context, adv *State, source Entity) {
	adv.actions.Add(NewChainedActions(
		NewChangeCameraAction(func(ctx *game.Context, s *State) Camera {
			return NewFollowCamera(n.EntityId, s.camera.CurrentLocation(), EntityCameraSpeedSlow)
		}),
		NewDelayAction(NewChangeCameraAction(func(ctx *game.Context, s *State) Camera {
			return NewFollowCamera(s.player.EntityId, s.camera.CurrentLocation(), EntityCameraSpeedSlow)
		}), 10)))
	adv.chatters.Add(newBasicEntityChatter(n.EntityId, "sup, guy", 5))
}
