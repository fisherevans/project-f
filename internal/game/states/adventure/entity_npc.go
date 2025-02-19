package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
	"math/rand"
)

type NPC struct {
	AnimatedMoveableEntity
	DoesMove  bool
	HorizOnly bool

	Talking        bool
	TalkingTowards EntityId

	IdleChance      float64
	MaxIdleDuration float64
	idleDuration    float64
}

func (n *NPC) Update(ctx *game.Context, adv *State, timeDelta float64) {
	defer n.AnimatedMoveableEntity.Update(ctx, adv, timeDelta)
	if n.IsMoving() {
		return
	}
	if n.Talking {
		ent, exists := adv.entities[n.TalkingTowards]
		if exists {
			n.FacingDirection = DirectionTowards(n.RenderMapLocation(), ent.RenderMapLocation())
		}
		return
	}
	if !n.DoesMove {
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
	if n.TriggerMovement(adv, n.GetFacingLocation(), MoveStateWalking) {
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
	n.FacingDirection = dir
	n.TriggerMovement(adv, n.GetLocationInDirection(dir), MoveStateWalking)
}

func (n *NPC) Interact(ctx *game.Context, adv *State, source Entity) {
	if n.Talking {
		return
	}
	n.Talking = true
	n.TalkingTowards = source.GetEntityId()
	duration := 5.
	adv.chatters.Add(newBasicEntityChatter(n.EntityId, duration, util.OneOffDialogues.Random()))
	adv.actions.Add(NewDelayAction(NewSimpleAction(func(ctx *game.Context, _ *State) {
		ctx.Notify("npc %s is no longer talking", n.EntityId)
		n.Talking = false
	}), duration))
}
