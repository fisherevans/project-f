package adventure

import "fisherevans.com/project/f/internal/util"

type FocusedSequenceBuilder struct {
	focusedEntity                          string
	playerId                               string
	moveCamera                             bool
	facePlayer                             bool
	preEffects, middleEffects, postEffects []Effect
}

func NewFocusedSequenceBuilder(focusedEntity, playerId string) *FocusedSequenceBuilder {
	return &FocusedSequenceBuilder{
		focusedEntity: focusedEntity,
		playerId:      playerId,
		postEffects:   []Effect{},
	}
}

func (b *FocusedSequenceBuilder) WithFacePlayer(facePlayer bool) *FocusedSequenceBuilder {
	b.facePlayer = facePlayer
	return b
}

func (b *FocusedSequenceBuilder) WithMoveCamera(moveCamera bool) *FocusedSequenceBuilder {
	b.moveCamera = moveCamera
	return b
}

func (b *FocusedSequenceBuilder) WithPreEffects(effects ...Effect) *FocusedSequenceBuilder {
	b.preEffects = append(b.preEffects, effects...)
	return b
}

func (b *FocusedSequenceBuilder) WithMiddleEffects(effects ...Effect) *FocusedSequenceBuilder {
	b.middleEffects = append(b.middleEffects, effects...)
	return b
}

func (b *FocusedSequenceBuilder) WithPostEffects(effects ...Effect) *FocusedSequenceBuilder {
	b.postEffects = append(b.postEffects, effects...)
	return b
}

func (b *FocusedSequenceBuilder) Build() *HandlerOutput {
	var effects []Effect
	effects = append(effects, b.preEffects...)
	effects = append(effects,
		NewMutateEntityBehaviorEffect(b.playerId).WithDisableBy(b.focusedEntity),
		NewPushEntityBehaviorEffect(b.focusedEntity).WithFacingEntityId(b.playerId),
	)
	if b.facePlayer {
		effects = append(effects, NewEntityFaceDirectionEffect(b.playerId).WithTargetEntity(b.focusedEntity))
	}
	if b.moveCamera {
		effects = append(effects,
			NewOverrideCameraEffect().WithFollow(FollowCamera{EntityId: util.Ptr(b.focusedEntity)}),
			NewWaitForConditionEffect(CameraWithinDistanceToTarget(1)))
	}
	effects = append(effects, b.middleEffects...)
	if b.moveCamera {
		effects = append(effects, NewPopCameraOverrideEffect(true))
	}
	if b.facePlayer {
		effects = append(effects, NewResetMovementEffect(b.playerId))
	}
	effects = append(effects,
		NewPopEntityBehaviorEffect(b.focusedEntity),
		NewMutateEntityBehaviorEffect(b.playerId).WithEnableBy(b.focusedEntity))
	effects = append(effects, b.postEffects...)
	return NewOutput().WithSerialPlan(effects...)
}
