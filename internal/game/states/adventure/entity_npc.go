package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/gopxl/pixel/v2"
)

func init() {
	targetRegistration().byClass("NPC").byTile(tiles.NPC).registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) (events.EntityContext, events.EventHandler) {
		renderer := NewMovementBasedEntityRenderer(entityId, system)
		color := colors.HSLToRGBA(rand.Float64(), 1, 1)
		for moveState, animations := range map[types.MoveState]map[input.Direction]*anim.AnimatedSprite{
			types.MoveStateIdle:    anim.AshaIdle(atlas),
			types.MoveStateWalking: anim.AshaWalk(atlas),
			types.MoveStateRunning: anim.AshaRun(atlas),
		} {
			for direction, animation := range animations {
				renderer.WithMovementStateRenderer(moveState, direction, NewBasicEntityRenderer().
					WithAnimationSpeedScaler(NewMoveAnimationSpeedScaler(system, entityId)).
					WithAnimationOriginOffset(pixel.V(0, 0.25)).
					WithAnimations(animation).
					WithAnimationColorMask(color))
			}
		}
		doesMove, horizOnly := true, false
		idleChance, maxIdle, speed := 0.05, 6.0, 2.0
		switch mapEntity.GetStringMetadata("movement", "") {
		case "static":
			doesMove = false
		case "horiz":
			horizOnly = true
		}
		switch mapEntity.GetStringMetadata("speed", "") {
		case "fast":
			speed = 4
		}
		presence := newBlockIngressPresence(true)
		behavior := NewNPCBehavior(entityId, system, doesMove, horizOnly, idleChance, maxIdle)
		position := system.RegisterEntity(entityId, location, presence, behavior, renderer, nil)
		position.MovementSpeeds = map[types.MoveState]float64{
			types.MoveStateWalking: speed,
		}
		return behavior.GenerateEntityContext(), events.NewBasicHandler(None{}).
			WithOnInteract(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
				if ctx.EntityId() != event.TargetId {
					return nil
				}
				if ctx.GetBoolMetadata(types.MetadataKeyIsTalking) {
					return nil
				}
				return events.NewOutput().WithSerialPlan(
					events.Effect{
						MutateNPC: &events.EffectMutateNPC{
							EntityId:          ctx.EntityId(),
							TalkingAtEntityId: util.Ptr(world.GetAsString("player_id")),
						},
						Chatter: &events.EffectChatter{
							EntityId:        ctx.EntityId(),
							DurationSeconds: 5,
							Message:         util.OneOffDialogues.Random(),
						},
					},
					events.Effect{
						MutateNPC: &events.EffectMutateNPC{
							EntityId:          ctx.EntityId(),
							TalkingAtEntityId: util.Ptr(""),
						},
					},
				)
			}).
			CreateHandler()
	})
}

type NPCBehavior struct {
	id        string
	system    *EntitySystem
	isEnabled func() bool

	talkingTowards string
	idleDuration   float64

	DoesMove        bool
	HorizOnly       bool
	IdleChance      float64
	MaxIdleDuration float64
}

func NewNPCBehavior(id string, system *EntitySystem, doesMove, horizOnly bool, idleChance, maxIdleDuration float64) *NPCBehavior {
	return &NPCBehavior{
		id:              id,
		system:          system,
		isEnabled:       func() bool { return true },
		DoesMove:        doesMove,
		HorizOnly:       horizOnly,
		IdleChance:      idleChance,
		MaxIdleDuration: maxIdleDuration,
	}
}

func (b *NPCBehavior) GenerateEntityContext() *events.BasicEntityContext {
	return events.NewBasicEntityContext(b.id).WithMetadata(types.MetadataKeyIsTalking, func() any {
		return b.talkingTowards != ""
	})
}

func (b *NPCBehavior) MovementComplete(dispatcher Dispatcher) {
}

func (b *NPCBehavior) Update(timeDelta float64, p *EntityPosition, dispatcher Dispatcher) {
	if p.IsMoving() {
		return
	}
	if b.talkingTowards != "" {
		talkingTowardsEnt, exists := b.system.positions[b.talkingTowards]
		if exists {
			// todo effect?
			p.FacingDirection = DirectionTowards(p.PreciseLocation(), talkingTowardsEnt.PreciseLocation())
		}
		return
	}
	if !b.DoesMove {
		return
	}
	if b.idleDuration > 0 {
		b.idleDuration -= timeDelta
		return
	}
	if rand.Float64() < b.IdleChance {
		b.idleDuration = rand.Float64() * b.MaxIdleDuration
		return
	}
	dispatcher.ProcessEffects(events.Effect{
		TriggerMovement: events.NewTriggerMovementEffect(b.id).WithDirection(p.FacingDirection),
	})
	// todo event when trigger effect fails maybe?
	// used to return here if trigger worked
	if true {
		return
	}
	var dir input.Direction
	if b.HorizOnly {
		if p.FacingDirection == input.NotPressed {
			dir = input.Left
		} else {
			dir = p.FacingDirection.Opposite()
		}
	} else {
		dir = input.Directions[int(rand.Float64()*float64(len(input.Directions)))]
	}
	p.FacingDirection = dir
	dispatcher.ProcessEffects(events.Effect{
		TriggerMovement: events.NewTriggerMovementEffect(b.id).WithDirection(p.FacingDirection),
	})
}
