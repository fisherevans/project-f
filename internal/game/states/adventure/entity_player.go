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
		renderer := NewMovementBasedEntityRenderer(entityId, system)

		dashPlayerLight := NewDynamicLight(colors.HexString("#88f"), 1.5, "")
		normalPlayerLight := NewDynamicLight(colors.HexString("#888"), 1.5, "")
		for moveState, animations := range map[types.MoveState]map[input.Direction]*anim.AnimatedSprite{
			types.MoveStateIdle:    anim.AshaIdle(atlas),
			types.MoveStateWalking: anim.AshaWalk(atlas),
			types.MoveStateRunning: anim.AshaRun(atlas),
			types.MoveStateDashing: anim.Dash(atlas),
		} {
			for direction, animation := range animations {
				moveStateRenderer := NewBasicEntityRenderer().WithAnimations(animation).
					WithAnimationSpeedScaler(NewMoveAnimationSpeedScaler(system, entityId)).
					WithAnimationOriginOffset(pixel.V(0, 0.25))
				if moveState == types.MoveStateDashing {
					moveStateRenderer.WithLights(dashPlayerLight)
				} else {
					moveStateRenderer.WithLights(normalPlayerLight)
				}
				renderer.WithMovementStateRenderer(moveState, direction, moveStateRenderer)
			}
		}
		presence := newBlockIngressPresence(false)
		behavior := NewPlayerBehavior(entityId, system)
		position := system.RegisterEntity(entityId, location, presence, behavior, renderer, nil)
		position.MovementSpeeds = map[types.MoveState]float64{
			types.MoveStateWalking: characterSpeed,
			types.MoveStateRunning: characterSpeed * 1.75,
			types.MoveStateDashing: characterSpeed * 3,
		}
		// todo this seems gross
		system.state.player = entityId
		system.state.camera = NewFollowCamera(entityId, location.ToVec(), EntityCameraSpeedPlayerDefault)
		system.state.worldState.Set("player_id", entityId)
		return nil, nil
	})
}

type PlayerBehavior struct {
	*baseEntityBehavior
	intentDirection     input.Direction
	intentDuration      float64
	awaitingInteraction bool
}

func NewPlayerBehavior(id string, system *EntitySystem) *PlayerBehavior {
	return &PlayerBehavior{
		baseEntityBehavior: newBaseEntityBehavior(id, system),
	}
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
	dispatcher.ProcessEffects(events.Effect{
		TriggerMovement: events.NewTriggerMovementEffect(b.id).WithDirection(direction).WithMoveState(moveState),
	})
}

func (b *PlayerBehavior) triggerInteraction(p *EntityPosition, dispatcher Dispatcher) {
	b.awaitingInteraction = false
	interactLocation := p.Location.MovedDelta(p.FacingDirection.GetVector())
	for targetEntityId := range b.system.interactableEntityIds(interactLocation) {
		dispatcher.EmitEvents(events.EventOnInteract{
			SourceId:              b.id,
			SourceFacingDirection: p.FacingDirection,
			TargetId:              targetEntityId,
		})
	}
}

func (b *PlayerBehavior) Update(timeDelta float64, p *EntityPosition, dispatcher Dispatcher) {
	// trigger running or face new direction after movement
	if p.IsMoving() {
		// todo - make effects?
		if game.Controls[*State]().DPad().IsPressed() {
			b.intentDirection = game.Controls[*State]().DPad().GetDirection()
			b.intentDuration += timeDelta
		}
		if game.Controls[*State]().ButtonB().IsPressed() {
			if p.MovementState == types.MoveStateWalking {
				p.MovementState = types.MoveStateRunning
			}
		} else {
			if p.MovementState == types.MoveStateRunning {
				p.MovementState = types.MoveStateWalking
			}
		}
		if game.Controls[*State]().ButtonA().JustPressedOrRepeated() {
			b.awaitingInteraction = true
		}
		return
	}
	// face direction of intent after movement
	if b.intentDirection != input.NotPressed && b.intentDirection != p.FacingDirection {
		p.FacingDirection = b.intentDirection
	}
	// trigger movement if player is pressing a direction
	if game.Controls[*State]().DPad().IsPressed() {
		direction := game.Controls[*State]().DPad().GetDirection()
		p.FacingDirection = direction
		if b.intentDirection != direction {
			b.intentDirection = direction
			b.intentDuration = 0
		}
		b.intentDuration += timeDelta
	}
	if b.awaitingInteraction {
		b.triggerInteraction(p, dispatcher)
		return
	}
	// interact with item if player is pressing A
	if game.Controls[*State]().ButtonA().JustPressedOrRepeated() && p.FacingDirection != input.NotPressed {
		b.triggerInteraction(p, dispatcher)
	} else {
		b.triggerMovement(dispatcher)
	}

}
