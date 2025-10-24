package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/gopxl/pixel/v2"
)

func init() {
	registerDynamicEntity().byClass("NPC").byTile(tiles.NPC).register(func(entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler) {
		e := &NPC{
			AnimatedMoveableEntity: AnimatedMoveableEntity{
				MoveableEntity: MoveableEntity{
					BaseEntity: BaseEntity{
						id:             entityId,
						isInteractable: true,
					},
					CurrentLocation: location,
					MoveSpeeds: map[MoveState]float64{
						MoveStateWalking: 2,
					},
					Passable: newPassablePreventIngress(true),
				},
				Animations: map[MoveState]map[input.Direction]*anim.AnimatedSprite{
					MoveStateIdle:    anim.AshaIdle(atlas),
					MoveStateWalking: anim.AshaWalk(atlas),
					MoveStateRunning: anim.AshaRun(atlas),
				},
				ColorMask: pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64()),
			},
			DoesMove:        true,
			IdleChance:      0.05,
			MaxIdleDuration: 6,
		}
		switch mapEntity.GetStringMetadata("movement", "") {
		case "static":
			e.DoesMove = false
		case "horiz":
			e.HorizOnly = true
		}
		switch mapEntity.GetStringMetadata("speed", "") {
		case "fast":
			e.MoveableEntity.MoveSpeeds[MoveStateWalking] = 4
		}
		return e, events.NewBasicHandler(None{}).
			WithOnInteract(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
				if ctx.Id() != event.TargetId {
					return nil
				}
				npcCtx, ok := ctx.(NPCEntityContext)
				if !ok {
					return nil
				}
				if npcCtx.IsTalking() {
					return nil
				}
				return events.NewOutput().WithSerialPlan(
					events.Effect{
						MutateNPC: &events.EffectMutateNPC{
							EntityId:          ctx.Id(),
							TalkingAtEntityId: util.Ptr(world.GetAsString("player_id")),
						},
						Chatter: &events.EffectChatter{
							EntityId:        ctx.Id(),
							DurationSeconds: 5,
							Message:         util.OneOffDialogues.Random(),
						},
					},
					events.Effect{
						MutateNPC: &events.EffectMutateNPC{
							EntityId:  ctx.Id(),
							IsTalking: util.Ptr(false),
						},
					},
				)
			}).
			CreateHandler()
	})
}

type NPCEntityContext interface {
	events.EntityContext
	IsTalking() bool
}

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

func (n *NPC) IsTalking() bool {
	return n.Talking
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
