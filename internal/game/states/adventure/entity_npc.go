package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
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

func (n *NPC) Update(adv *State, timeDelta float64) {
	defer n.AnimatedMoveableEntity.Update(adv, timeDelta)
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

func (n *NPC) Interact(adv *State, source Entity) {
	if n.Talking {
		return
	}
	playerId := string(adv.player.Id)
	bFalse := false
	adv.AddSystemEffect(events.Effect{
		Plan: &events.EffectPlan{
			Steps: []events.PlanStep{
				{
					Serial: []events.Effect{
						{
							MutateNPC: &events.EffectMutateNPC{
								EntityId:          string(n.Id),
								TalkingAtEntityId: &playerId,
							},
							Chatter: &events.EffectChatter{
								EntityId:        string(n.Id),
								DurationSeconds: 5,
								Message:         util.OneOffDialogues.Random(),
							},
						},
						{
							MutateNPC: &events.EffectMutateNPC{
								EntityId:  string(n.Id),
								IsTalking: &bFalse,
							},
						},
					},
				},
			},
		},
	})
}
