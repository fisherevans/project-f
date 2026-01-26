package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/gopxl/pixel/v2"
)

func init() {
	newRegistrarBuilder().byClass("NPC").byTile(tiles.NPC).registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
		entity := system.RegisterEntity(params.EntityId, params.Location)
		renderer := AttachMovementBasedEntityRenderer(entity)
		color := colors.HSLToRGBA(rand.Float64(), 1, 0.65)
		for moveState, animations := range map[MoveState]map[input.Direction]*anim.AnimatedSprite{
			MoveStateIdle:    anim.AshaIdle(atlas),
			MoveStateWalking: anim.AshaWalk(atlas),
			MoveStateRunning: anim.AshaRun(atlas),
		} {
			for direction, animation := range animations {
				cma := NewColorMaskAnimation(animation).WithColorMask(color)
				renderer.WithMovementStateRenderer(moveState, direction, NewBasicEntityRenderer(entity).
					WithAnimationSpeedScaler(NewMoveAnimationSpeedScaler(entity)).
					WithAnimationOriginOffset(pixel.V(0, 0.25)).
					WithColorMaskAnimations(cma))
			}
		}
		doesMove, horizOnly := true, false
		idleChance, maxIdle, speed := 0.05, 6.0, 2.0
		switch params.Properties.GetString("movement", "") {
		case "static":
			doesMove = false
		case "horiz":
			horizOnly = true
		}
		switch params.Properties.GetString("speed", "") {
		case "fast":
			speed = 4
		}
		idleFacingDirection := input.DirectionFromString(params.Properties.GetString("idle_facing_direction", ""))
		// to walk around and entity, it will take 2 extra moves - adding 1 extra here when walking keeps them walking behind each other
		AttachBlockIngressPresence(entity, true, NewMovementAwareImpedance(entity, ImpedanceHigh, ImpedanceBase*2))
		AttachNPCBehavior(entity, doesMove, horizOnly, idleChance, maxIdle, idleFacingDirection)
		entity.SetMovementSpeed(MoveStateWalking, speed)
		entity.AddSoundProvider(NewStepSoundProvider(entity, createStepSoundsSoft(), FootstepFalloff))
		if params.Properties.GetString("script_ref", "") != "" {
			return entity, nil
		}
		return entity, NewBasicHandler(None{}).
			WithOnInteract(func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if thisEntity.GetId() != event.TargetId {
					return nil
				}
				if !IsBehaviorType[*NPCBehavior](thisEntity) {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewPushEntityBehaviorEffect(thisEntity.GetId()).WithFacingEntityId(event.SourceId),
					NewChatterEffect(thisEntity.GetId(), 4, util.OneOffDialogues.Random()),
					NewPopEntityBehaviorEffect(thisEntity.GetId()),
				)
			}).
			CreateHandler()
	})
}

type NPCBehavior struct {
	entity Entity

	idleDuration float64

	DoesMove            bool
	HorizOnly           bool
	IdleChance          float64
	MaxIdleDuration     float64
	IdleFacingDirection input.Direction
}

func AttachNPCBehavior(entity Entity, horizOnly, doesMove bool, idleChance, maxIdleDuration float64, idleFacingDirection input.Direction) *NPCBehavior {
	b := &NPCBehavior{
		entity:              entity,
		DoesMove:            doesMove,
		HorizOnly:           horizOnly,
		IdleChance:          idleChance,
		MaxIdleDuration:     maxIdleDuration,
		IdleFacingDirection: idleFacingDirection,
	}
	entity.PushBehavior(b)
	return b
}

func (b *NPCBehavior) Reset() {
	b.idleDuration = 0
}

func (b *NPCBehavior) MovementComplete() {
	b.doMovement()
}

func (b *NPCBehavior) Update(timeDelta float64) {
	if b.idleDuration > 0 {
		b.idleDuration -= timeDelta
		return
	}
	b.doMovement()
}

func (b *NPCBehavior) doMovement() {
	if b.entity.IsMoving() {
		return
	}
	if b.IdleFacingDirection != input.NotPressed {
		b.entity.SetFacingDirection(b.IdleFacingDirection)
	}
	if !b.DoesMove {
		return
	}
	if rand.Float64() < b.IdleChance {
		b.idleDuration = rand.Float64() * b.MaxIdleDuration
		return
	}
	nextLocation := b.entity.GetLocation().Moved(b.entity.GetFacingDirection())
	isValid, _ := b.entity.GetSystem().isMovementValid(b.entity, nextLocation)
	if isValid {
		b.entity.GetSystem().state.ExecuteSystemEffects(NewTriggerMovementEffect(b.entity.GetId()).
			WithLocation(nextLocation))
		return
	}
	var dir input.Direction
	if b.HorizOnly {
		if b.entity.GetFacingDirection() == input.NotPressed {
			dir = input.Left
		} else {
			dir = b.entity.GetFacingDirection().Opposite()
		}
	} else {
		dir = input.Directions[int(rand.Float64()*float64(len(input.Directions)))]
	}
	b.entity.GetSystem().state.ExecuteSystemEffects(
		NewEntityFaceDirectionEffect(b.entity.GetId()).WithDirection(dir),
		NewTriggerMovementEffect(b.entity.GetId()).WithDirection(dir),
	)
}
