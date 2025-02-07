package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"math/rand"
)

type NPC struct {
	AnimatedMoveableEntity
	DoesMove  bool
	HorizOnly bool

	IdleChance      float64
	MaxIdleDuration float64
	idleDuration    float64
}

func (n *NPC) Update(ctx *game.Context, adv *State, timeDelta float64) {
	defer n.AnimatedMoveableEntity.Update(ctx, adv, timeDelta)
	if !n.DoesMove || n.Moving {
		return
	}
	if n.idleDuration > 0 {
		n.idleDuration -= timeDelta
		return
	}
	if rand.Float64() < n.IdleChance {
		n.idleDuration = rand.Float64() * n.MaxIdleDuration
		return
	}
	if n.TriggerMovement(adv, n.FacingDirection) {
		return
	}
	var dir input.Direction
	if n.HorizOnly {
		if n.FacingDirection == input.NotPressed {
			dir = input.Left
		} else {
			dir = n.FacingDirection.Opposite()
		}
	} else {
		dir = input.Directions[int(rand.Float64()*float64(len(input.Directions)))]
	}
	n.TriggerMovement(adv, dir)
}

func (n *NPC) Interact(ctx *game.Context, adv *State, source Entity) {
	//adv.actions.Add(NewChainedActions(
	//	NewChangeCameraAction(func(ctx *game.Context, s *State) Camera {
	//		return NewFollowCamera(n.EntityId, s.camera.CurrentLocation(), EntityCameraSpeedSlow)
	//	}),
	//	NewDelayAction(NewChangeCameraAction(func(ctx *game.Context, s *State) Camera {
	//		return NewFollowCamera(s.player.EntityId, s.camera.CurrentLocation(), EntityCameraSpeedSlow)
	//	}), 10)))
	adv.chatters.Add(newBasicEntityChatter(n.EntityId, 10, "Hi Mister?"))
}
