// AUTO-GENERATED - DO NOT EDIT
// Generated at 2025-10-26T09:38:00-04:00 by go generate
// Source: internal/game/events/effects.go

package events

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"github.com/gopxl/pixel/v2"
)

func NewFunctionEffect(fn RunnableFunction) *EffectFunction {
	return &EffectFunction{
		Fn: fn,
	}
}

func NewDialogueEffect(dialogueId string, text string) *EffectDialogue {
	return &EffectDialogue{
		DialogueId: dialogueId,
		Text: text,
	}
}

func NewChatterEffect(chatterId string, entityId string, durationSeconds float64, message string) *EffectChatter {
	return &EffectChatter{
		ChatterId: chatterId,
		EntityId: entityId,
		DurationSeconds: durationSeconds,
		Message: message,
	}
}

func NewYieldElythiumEffect(amount int) *EffectYieldElythium {
	return &EffectYieldElythium{
		Amount: amount,
	}
}

func NewTimerEffect(timerId string, durationSeconds float64) *EffectTimer {
	return &EffectTimer{
		TimerId: timerId,
		DurationSeconds: durationSeconds,
	}
}

func NewMutateModeBasedEntityEffect(entityId string) *EffectMutateModeBasedEntity {
	return &EffectMutateModeBasedEntity{
		EntityId: entityId,
	}
}

func (e *EffectMutateModeBasedEntity) WithMode(mode string) *EffectMutateModeBasedEntity {
	e.Mode = &mode
	return e
}

func (e *EffectMutateModeBasedEntity) WithAnimations(animations map[string][]AnimationReference) *EffectMutateModeBasedEntity {
	e.Animations = &animations
	return e
}

func (e *EffectMutateModeBasedEntity) WithLights(lights map[string][]LightConfig) *EffectMutateModeBasedEntity {
	e.Lights = &lights
	return e
}

func (e *EffectMutateModeBasedEntity) WithAnimationColorMasks(animationColorMasks map[string]pixel.RGBA) *EffectMutateModeBasedEntity {
	e.AnimationColorMasks = &animationColorMasks
	return e
}

func NewResetModeBasedEntityAnimationEffect(entityId string) *EffectResetModeBasedEntityAnimation {
	return &EffectResetModeBasedEntityAnimation{
		EntityId: entityId,
	}
}

func NewMutateBlockingPresenceEffect(entityId string) *EffectMutateBlockingPresence {
	return &EffectMutateBlockingPresence{
		EntityId: entityId,
	}
}

func (e *EffectMutateBlockingPresence) WithIsBlockingIngress(isBlockingIngress bool) *EffectMutateBlockingPresence {
	e.IsBlockingIngress = &isBlockingIngress
	return e
}

func NewSetWorldStateEffect(key string, value any) *EffectSetWorldState {
	return &EffectSetWorldState{
		Key: key,
		Value: value,
	}
}

func NewSetEntityLocationEffect(entityId string) *EffectSetEntityLocation {
	return &EffectSetEntityLocation{
		EntityId: entityId,
	}
}

func (e *EffectSetEntityLocation) WithToReference(toReference string) *EffectSetEntityLocation {
	e.ToReference = &toReference
	return e
}

func (e *EffectSetEntityLocation) WithToLocation(toLocation Location) *EffectSetEntityLocation {
	e.ToLocation = &toLocation
	return e
}

func (e *EffectSetEntityLocation) WithToEntityId(toEntityId string) *EffectSetEntityLocation {
	e.ToEntityId = &toEntityId
	return e
}

func NewTeleportPlayerEffect() *EffectTeleportPlayer {
	return &EffectTeleportPlayer{
	}
}

func (e *EffectTeleportPlayer) WithToReference(toReference string) *EffectTeleportPlayer {
	e.ToReference = &toReference
	return e
}

func (e *EffectTeleportPlayer) WithToLocation(toLocation Location) *EffectTeleportPlayer {
	e.ToLocation = &toLocation
	return e
}

func (e *EffectTeleportPlayer) WithToEntityId(toEntityId string) *EffectTeleportPlayer {
	e.ToEntityId = &toEntityId
	return e
}

func (e *EffectTeleportPlayer) WithExitDirection(exitDirection input.Direction) *EffectTeleportPlayer {
	e.ExitDirection = &exitDirection
	return e
}

func (e *EffectTeleportPlayer) WithTransitionStyle(transitionStyle string) *EffectTeleportPlayer {
	e.TransitionStyle = &transitionStyle
	return e
}

func NewPlanEffect(planId string, steps []PlanStep) *EffectPlan {
	return &EffectPlan{
		PlanId: planId,
		Steps: steps,
	}
}

func NewMutateEntityBehaviorEffect(entityId string) *EffectMutateEntityBehavior {
	return &EffectMutateEntityBehavior{
		EntityId: entityId,
	}
}

func (e *EffectMutateEntityBehavior) WithDisableBy(disableBy string) *EffectMutateEntityBehavior {
	e.DisableBy = &disableBy
	return e
}

func (e *EffectMutateEntityBehavior) WithEnableBy(enableBy string) *EffectMutateEntityBehavior {
	e.EnableBy = &enableBy
	return e
}

func (e *EffectMutateEntityBehavior) WithReset(reset bool) *EffectMutateEntityBehavior {
	e.Reset = &reset
	return e
}

func NewFadeEffect(fadeId string, durationSeconds float64, transitions int) *EffectFade {
	return &EffectFade{
		FadeId: fadeId,
		DurationSeconds: durationSeconds,
		Transitions: transitions,
	}
}

func (e *EffectFade) WithAutoDeactivate(autoDeactivate bool) *EffectFade {
	e.AutoDeactivate = &autoDeactivate
	return e
}

func (e *EffectFade) WithFromColor(fromColor string) *EffectFade {
	e.FromColor = &fromColor
	return e
}

func (e *EffectFade) WithToColor(toColor string) *EffectFade {
	e.ToColor = &toColor
	return e
}

func NewDeactivateFadeEffect(fadeId string) *EffectDeactivateFade {
	return &EffectDeactivateFade{
		FadeId: fadeId,
	}
}

func NewTriggerMovementEffect(entityId string) *EffectTriggerMovement {
	return &EffectTriggerMovement{
		EntityId: entityId,
	}
}

func (e *EffectTriggerMovement) WithDirection(direction input.Direction) *EffectTriggerMovement {
	e.Direction = &direction
	return e
}

func (e *EffectTriggerMovement) WithLocation(location Location) *EffectTriggerMovement {
	e.Location = &location
	return e
}

func (e *EffectTriggerMovement) WithMoveState(moveState types.MoveState) *EffectTriggerMovement {
	e.MoveState = &moveState
	return e
}

func NewResetMovementEffect(entityId string) *EffectResetMovement {
	return &EffectResetMovement{
		EntityId: entityId,
	}
}

func NewOverrideCameraEffect() *EffectOverrideCamera {
	return &EffectOverrideCamera{
	}
}

func (e *EffectOverrideCamera) WithFollow(follow FollowCamera) *EffectOverrideCamera {
	e.Follow = &follow
	return e
}

func NewPopCameraOverrideEffect(maintainCurrentLocation bool) *EffectPopCameraOverride {
	return &EffectPopCameraOverride{
		MaintainCurrentLocation: maintainCurrentLocation,
	}
}

func NewMutateFollowCameraEffect() *EffectMutateFollowCamera {
	return &EffectMutateFollowCamera{
	}
}

func (e *EffectMutateFollowCamera) WithFollowEntityId(followEntityId string) *EffectMutateFollowCamera {
	e.FollowEntityId = &followEntityId
	return e
}

func (e *EffectMutateFollowCamera) WithResetPosition(resetPosition bool) *EffectMutateFollowCamera {
	e.ResetPosition = &resetPosition
	return e
}

func NewMutateNPCEffect(entityId string) *EffectMutateNPC {
	return &EffectMutateNPC{
		EntityId: entityId,
	}
}

func (e *EffectMutateNPC) WithTalkingAtEntityId(talkingAtEntityId string) *EffectMutateNPC {
	e.TalkingAtEntityId = &talkingAtEntityId
	return e
}

func NewTriggerCombatEffect(combatId string, background string) *EffectTriggerCombat {
	return &EffectTriggerCombat{
		CombatId: combatId,
		Background: background,
	}
}

func (e *EffectTriggerCombat) WithOpponent(opponent rpg.PrimortalType) *EffectTriggerCombat {
	e.Opponent = &opponent
	return e
}

func NewEntityFaceDirectionEffect(entityId string, direction input.Direction) *EffectEntityFaceDirection {
	return &EffectEntityFaceDirection{
		EntityId: entityId,
		Direction: direction,
	}
}

func NewStartScriptedMotionEffect(motionId string, entityId string) *EffectStartScriptedMotion {
	return &EffectStartScriptedMotion{
		MotionId: motionId,
		EntityId: entityId,
	}
}

func (e *EffectStartScriptedMotion) WithLocation(location Location) *EffectStartScriptedMotion {
	e.Location = &location
	return e
}

func (e *EffectStartScriptedMotion) WithToEntityId(toEntityId string) *EffectStartScriptedMotion {
	e.ToEntityId = &toEntityId
	return e
}

func (e *EffectStartScriptedMotion) WithRelative(relative RelativeLocation) *EffectStartScriptedMotion {
	e.Relative = &relative
	return e
}

func NewOverrideEntityBehaviorEffect(entityId string) *EffectOverrideEntityBehavior {
	return &EffectOverrideEntityBehavior{
		EntityId: entityId,
	}
}

func (e *EffectOverrideEntityBehavior) WithScriptedMotion(scriptedMotion EntityBehaviorScriptedMotion) *EffectOverrideEntityBehavior {
	e.ScriptedMotion = &scriptedMotion
	return e
}

func NewPopEntityBehaviorOverrideEffect(entityId string) *EffectPopEntityBehaviorOverride {
	return &EffectPopEntityBehaviorOverride{
		EntityId: entityId,
	}
}

