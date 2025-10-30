package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/gopxl/pixel/v2"
)

func init() {
	targetRegistration().byTile(tiles.Player).registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) (events.EntityContext, events.EventHandler) {
		entity := system.RegisterEntity(entityId, location)
		renderer := AttachMovementBasedEntityRenderer(entity)

		dashPlayerLight := NewLight(colors.HexString("#88f"), 1.5)
		normalPlayerLight := NewLight(colors.HexString("#888"), 1.5)
		for moveState, animations := range map[types.MoveState]map[input.Direction]*anim.AnimatedSprite{
			types.MoveStateIdle:    anim.AshaIdle(atlas),
			types.MoveStateWalking: anim.AshaWalk(atlas),
			types.MoveStateRunning: anim.AshaRun(atlas),
			types.MoveStateDashing: anim.Dash(atlas),
		} {
			for direction, animation := range animations {
				moveStateRenderer := NewBasicEntityRenderer().WithAnimations(animation).
					WithAnimationSpeedScaler(NewMoveAnimationSpeedScaler(entity)).
					WithAnimationOriginOffset(pixel.V(0, 0.25))
				if moveState == types.MoveStateDashing {
					moveStateRenderer.WithLights(dashPlayerLight)
				} else {
					moveStateRenderer.WithLights(normalPlayerLight)
				}
				renderer.WithMovementStateRenderer(moveState, direction, moveStateRenderer)
			}
		}
		AttachBlockIngressPresence(entity, false)
		AttachPlayerBehavior(entity)
		entity.SetMovementSpeed(types.MoveStateWalking, characterSpeed)
		entity.SetMovementSpeed(types.MoveStateRunning, characterSpeed*1.75)
		entity.SetMovementSpeed(types.MoveStateDashing, characterSpeed*3)
		// todo this seems gross
		system.state.player = entityId
		system.state.camera = NewFollowCamera(entityId, location.ToVec(), EntityCameraSpeedPlayerDefault)
		system.state.worldState.Set("player_id", entityId)
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

func (b *PlayerBehavior) MovementComplete(dispatcher Dispatcher) {
	b.triggerMovement(dispatcher)
}

func (b *PlayerBehavior) triggerMovement(dispatcher Dispatcher) {
	doTrigger := game.Controls[*State]().DPad().IsPressed() && b.intentDuration > 0.075
	if !doTrigger {
		return
	}
	direction := game.Controls[*State]().DPad().GetDirection()
	moveState := types.MoveStateWalking
	if game.Controls[*State]().ButtonB().IsPressed() {
		moveState = types.MoveStateRunning
	}
	dispatcher.ProcessEffects(
		events.NewTriggerMovementEffect(b.entity.GetId()).
			WithDirection(direction).
			WithMoveState(moveState))
}

func (b *PlayerBehavior) triggerInteraction(dispatcher Dispatcher) {
	b.awaitingInteraction = false
	interactLocation := b.entity.GetLocation().MovedDelta(b.entity.GetFacingDirection().GetVector())
	b.entity.InteractsWith(interactLocation)
}

func (b *PlayerBehavior) Update(timeDelta float64, dispatcher Dispatcher) {
	// trigger running or face new direction after movement
	if b.entity.IsMoving() {
		// todo - make effects?
		if game.Controls[*State]().DPad().IsPressed() {
			b.intentDirection = game.Controls[*State]().DPad().GetDirection()
			b.intentDuration += timeDelta
		}
		if game.Controls[*State]().ButtonB().IsPressed() {
			if b.entity.GetMovementState() == types.MoveStateWalking {
				b.entity.AlterMovementState(types.MoveStateRunning)
			}
		} else {
			if b.entity.GetMovementState() == types.MoveStateRunning {
				b.entity.AlterMovementState(types.MoveStateWalking)
			}
		}
		if game.Controls[*State]().ButtonA().JustPressedOrRepeated() {
			b.awaitingInteraction = true
		}
		return
	}
	// face direction of intent after movement
	if b.intentDirection != input.NotPressed && b.intentDirection != b.entity.GetFacingDirection() {
		dispatcher.ProcessEffects(events.NewEntityFaceDirectionEffect(b.entity.GetId(), b.intentDirection))
	}
	// trigger movement if player is pressing a direction
	if game.Controls[*State]().DPad().IsPressed() {
		direction := game.Controls[*State]().DPad().GetDirection()
		dispatcher.ProcessEffects(events.NewEntityFaceDirectionEffect(b.entity.GetId(), b.intentDirection))
		if b.intentDirection != direction {
			b.intentDirection = direction
			b.intentDuration = 0
		}
		b.intentDuration += timeDelta
	}
	if b.awaitingInteraction {
		b.triggerInteraction(dispatcher)
		return
	}
	// interact with item if player is pressing A
	if game.Controls[*State]().ButtonA().JustPressedOrRepeated() && b.entity.GetFacingDirection() != input.NotPressed {
		b.triggerInteraction(dispatcher)
	} else {
		b.triggerMovement(dispatcher)
	}

}
