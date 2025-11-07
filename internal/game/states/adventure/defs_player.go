package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/gopxl/pixel/v2"
)

func init() {
	newRegistrarBuilder().byTile(tiles.Player).registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
		entity := system.RegisterEntity(params.EntityId, params.Location)
		renderer := AttachMovementBasedEntityRenderer(entity)

		dashPlayerLight := NewLight(colors.HexString("#88f"), 1.5)
		normalPlayerLight := NewLight(colors.HexString("#888"), 1.5)
		for moveState, animations := range map[MoveState]map[input.Direction]*anim.AnimatedSprite{
			MoveStateIdle:    anim.AshaIdle(atlas),
			MoveStateWalking: anim.AshaWalk(atlas),
			MoveStateRunning: anim.AshaRun(atlas),
			MoveStateDashing: anim.Dash(atlas),
		} {
			for direction, animation := range animations {
				moveStateRenderer := NewBasicEntityRenderer(entity).WithAnimations(animation).
					WithAnimationSpeedScaler(NewMoveAnimationSpeedScaler(entity)).
					WithAnimationOriginOffset(pixel.V(0, 0.25))
				if moveState == MoveStateDashing {
					moveStateRenderer.WithLights(dashPlayerLight)
				} else {
					moveStateRenderer.WithLights(normalPlayerLight)
				}
				renderer.WithMovementStateRenderer(moveState, direction, moveStateRenderer)
			}
		}
		AttachBlockIngressPresence(entity, false, NewStaticImpedance(ImpedanceHigh))
		AttachPlayerBehavior(entity)
		entity.SetMovementSpeed(MoveStateWalking, characterSpeed)
		entity.SetMovementSpeed(MoveStateRunning, characterSpeed*1.75)
		entity.SetMovementSpeed(MoveStateDashing, characterSpeed*3)
		// todo this seems gross
		system.state.player = params.EntityId
		system.state.camera = NewFollowCamera(params.EntityId, params.Location.ToVec(), EntityCameraSpeedMedium)
		system.state.runState.Set(runStateKeyPlayerId, system.state.player)
		return nil, nil
	})
}

type PlayerBehavior struct {
	entity Entity

	intentDirection     input.Direction
	intentDuration      float64
	awaitingInteraction bool
}

func AttachPlayerBehavior(entity Entity) *PlayerBehavior {
	b := &PlayerBehavior{
		entity: entity,
	}
	entity.PushBehavior(b)
	return b
}

func (b *PlayerBehavior) Reset() {
	b.intentDirection = input.NotPressed
	b.intentDuration = 0
	b.awaitingInteraction = false
}

func (b *PlayerBehavior) MovementComplete() {
	b.triggerMovement()
}

func (b *PlayerBehavior) triggerMovement() {
	doTrigger := game.Controls[*State]().DPad().IsPressed() && b.intentDuration > 0.075
	if !doTrigger {
		return
	}
	direction := game.Controls[*State]().DPad().GetDirection()
	moveState := MoveStateWalking
	if game.Controls[*State]().ButtonB().IsPressed() {
		moveState = MoveStateRunning
	}
	b.entity.GetSystem().state.ExecuteSystemEffects(
		NewTriggerMovementEffect(b.entity.GetId()).
			WithDirection(direction).
			WithMoveState(moveState))
}

func (b *PlayerBehavior) triggerInteraction() {
	b.awaitingInteraction = false
	interactLocation := b.entity.GetLocation().MovedDelta(b.entity.GetFacingDirection().GetVector())
	b.entity.InteractsWith(interactLocation)
}

func (b *PlayerBehavior) Update(timeDelta float64) {
	// trigger running or face new direction after movement
	if b.entity.IsMoving() {
		if game.Controls[*State]().DPad().IsPressed() {
			b.intentDirection = game.Controls[*State]().DPad().GetDirection()
			b.intentDuration += timeDelta
		}
		if game.Controls[*State]().ButtonB().IsPressed() {
			if b.entity.GetMovementState() == MoveStateWalking {
				b.entity.AlterMovementState(MoveStateRunning)
			}
		} else {
			if b.entity.GetMovementState() == MoveStateRunning {
				b.entity.AlterMovementState(MoveStateWalking)
			}
		}
		if game.Controls[*State]().ButtonA().JustPressedOrRepeated() {
			b.awaitingInteraction = true
		}
		return
	}
	// face direction of intent after movement
	if b.intentDirection != input.NotPressed && b.intentDirection != b.entity.GetFacingDirection() {
		b.entity.SetFacingDirection(b.intentDirection)
	}
	// trigger movement if player is pressing a direction
	if game.Controls[*State]().DPad().IsPressed() {
		direction := game.Controls[*State]().DPad().GetDirection()
		b.entity.SetFacingDirection(b.intentDirection)
		if b.intentDirection != direction {
			b.intentDirection = direction
			b.intentDuration = 0
		}
		b.intentDuration += timeDelta
	}
	if b.awaitingInteraction {
		b.triggerInteraction()
		return
	}
	// interact with item if player is pressing A
	if game.Controls[*State]().ButtonA().JustPressedOrRepeated() && b.entity.GetFacingDirection() != input.NotPressed {
		b.triggerInteraction()
	} else {
		b.triggerMovement()
	}

}
