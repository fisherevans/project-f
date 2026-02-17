package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

func loadPlayerEntity(params NewEntityParams, system *EntitySystem, defaultPlayerMode string) (Entity, EventHandler) {
	entity := system.RegisterEntity(params.EntityId, params.Location)
	switch defaultPlayerMode {
	case "human":
		playerHumanRenderer(entity)
	case "animech":
		playerAnimechRenderer(entity)
	default:
		log.Fatal().Msgf("invalid player mode: %s", defaultPlayerMode)
	}
	AttachBlockIngressPresence(entity, false, NewStaticImpedance(ImpedanceHigh))
	AttachPlayerBehavior(entity, system.state.Controls)
	entity.SetMovementSpeed(MoveStateWalking, characterSpeed)
	entity.SetMovementSpeed(MoveStateRunning, characterSpeed*1.75)
	entity.SetMovementSpeed(MoveStateDashing, characterSpeed*3)
	entity.AddSoundProvider(NewStepSoundProvider(entity, createStepSoundsHard(), FootstepFalloff))
	// todo this seems gross
	system.state.player = params.EntityId
	system.state.camera = NewSimpleEntityCamera(params.EntityId, params.Location.ToVec(), EntityCameraSpeedMedium, true)
	system.state.globals.Set(globalVariableNamePlayerId, system.state.player)
	return nil, nil
}

func playerHumanRenderer(entity Entity) *MovementBasedEntityRenderer {
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
	return renderer
}

func playerAnimechRenderer(entity Entity) *MovementBasedEntityRenderer {
	renderer := AttachMovementBasedEntityRenderer(entity)
	dashPlayerLight := NewLight(colors.HexString("#88f"), 1.5)
	normalPlayerLight := NewLight(colors.HexString("#888"), 1.5)
	for moveState, animations := range map[MoveState]map[input.Direction]*anim.AnimatedSprite{
		MoveStateIdle:    anim.AnimechIdle(atlas),
		MoveStateWalking: anim.AnimechWalk(atlas),
		MoveStateRunning: anim.AnimechRun(atlas),
		MoveStateDashing: anim.Dash(atlas),
	} {
		for direction, animation := range animations {
			moveStateRenderer := NewBasicEntityRenderer(entity).WithAnimations(animation).
				WithAnimationSpeedScaler(NewMoveAnimationSpeedScaler(entity)).
				WithAnimationOriginOffset(pixel.V(0, 0.375))
			if moveState == MoveStateDashing {
				moveStateRenderer.WithLights(dashPlayerLight)
			} else {
				moveStateRenderer.WithLights(normalPlayerLight)
			}
			renderer.WithMovementStateRenderer(moveState, direction, moveStateRenderer)
		}
	}
	return renderer
}

func playerHiddenRenderer(entity Entity) *MovementBasedEntityRenderer {
	renderer := AttachMovementBasedEntityRenderer(entity)
	return renderer
}

type PlayerBehavior struct {
	entity Entity

	intentDirection     input.Direction
	intentDuration      float64
	awaitingInteraction bool

	controls func() *input.Controls
}

func AttachPlayerBehavior(entity Entity, controls func() *input.Controls) *PlayerBehavior {
	b := &PlayerBehavior{
		entity:   entity,
		controls: controls,
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
	doTrigger := b.controls().DPad().IsPressed() && b.intentDuration > 0.075
	if !doTrigger {
		return
	}
	direction := b.controls().DPad().GetDirection()
	moveState := MoveStateWalking
	if b.controls().ButtonB().IsPressed() {
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

func (b *PlayerBehavior) Update(timeDelta float64, globals StateGlobalsReader) {
	// trigger running or face new direction after movement
	if b.entity.IsMoving() {
		if b.controls().DPad().IsPressed() {
			b.intentDirection = b.controls().DPad().GetDirection()
			b.intentDuration += timeDelta
		}
		if b.controls().ButtonB().IsPressed() {
			if b.entity.GetMovementState() == MoveStateWalking {
				b.entity.AlterMovementState(MoveStateRunning)
			}
		} else {
			if b.entity.GetMovementState() == MoveStateRunning {
				b.entity.AlterMovementState(MoveStateWalking)
			}
		}
		if b.controls().ButtonA().JustPressedOrRepeated() {
			b.awaitingInteraction = true
		}
		return
	}
	// face direction of intent after movement
	if b.intentDirection != input.NotPressed && b.intentDirection != b.entity.GetFacingDirection() {
		b.entity.SetFacingDirection(b.intentDirection)
	}
	// trigger movement if player is pressing a direction
	if b.controls().DPad().IsPressed() {
		direction := b.controls().DPad().GetDirection()
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
	if b.controls().ButtonA().JustPressedOrRepeated() && b.entity.GetFacingDirection() != input.NotPressed {
		b.triggerInteraction()
	} else {
		b.triggerMovement()
	}
}
