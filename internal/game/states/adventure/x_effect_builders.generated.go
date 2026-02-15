// AUTO-GENERATED - DO NOT EDIT
// Generated at 2026-02-14T22:30:19-05:00 by go generate
// Source: internal/game/adventure/effect.go

package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/highlighter"
)

func NewFunctionEffect(fn RunnableFunction) *EffectFunction {
	return &EffectFunction{
		Fn: fn,
	}
}

func NewTimerEffect(durationSeconds float64) *EffectTimer {
	return &EffectTimer{
		DurationSeconds: durationSeconds,
	}
}

func (e *EffectTimer) WithTimerId(timerId string) *EffectTimer {
	e.TimerId = timerId
	return e
}

func NewSetWorldStateEffect(key string, value any) *EffectSetWorldState {
	return &EffectSetWorldState{
		Key: key,
		Value: value,
	}
}

func NewSetRunStateEffect(key string, value any) *EffectSetRunState {
	return &EffectSetRunState{
		Key: key,
		Value: value,
	}
}

func NewBatchEffect(effects []Effect) *EffectBatch {
	return &EffectBatch{
		Effects: effects,
	}
}

func (e *EffectBatch) WithBatchId(batchId string) *EffectBatch {
	e.BatchId = batchId
	return e
}

func (e *EffectBatch) WithExecuteInParallel(executeInParallel bool) *EffectBatch {
	e.ExecuteInParallel = &executeInParallel
	return e
}

func NewWaitForConditionEffect(check ConditionCheck) *EffectWaitForCondition {
	return &EffectWaitForCondition{
		Check: check,
	}
}

func (e *EffectWaitForCondition) WithConditionId(conditionId string) *EffectWaitForCondition {
	e.ConditionId = conditionId
	return e
}

func NewSendEventEffect(event any) *EffectSendEvent {
	return &EffectSendEvent{
		Event: event,
	}
}

func NewSendBroadcastEffect(broadcastId string, data any) *EffectSendBroadcast {
	return &EffectSendBroadcast{
		BroadcastId: broadcastId,
		Data: data,
	}
}

func NewPushTooltipEffect(message string) *EffectPushTooltip {
	return &EffectPushTooltip{
		Message: message,
	}
}

func NewSetHighlightSequenceEffect(targets []highlighter.Target) *EffectSetHighlightSequence {
	return &EffectSetHighlightSequence{
		Targets: targets,
	}
}

func NewPlaySoundEffect(sound string) *EffectPlaySound {
	return &EffectPlaySound{
		Sound: sound,
	}
}

func (e *EffectPlaySound) WithPlaybackId(playbackId string) *EffectPlaySound {
	e.PlaybackId = playbackId
	return e
}

func (e *EffectPlaySound) WithVolume(volume float64) *EffectPlaySound {
	e.Volume = &volume
	return e
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

func NewDeleteEntityEffect(entityId string) *EffectDeleteEntity {
	return &EffectDeleteEntity{
		EntityId: entityId,
	}
}

func NewRegisterEntityEffect() *EffectRegisterEntity {
	return &EffectRegisterEntity{
	}
}

func (e *EffectRegisterEntity) WithEntityId(entityId string) *EffectRegisterEntity {
	e.EntityId = entityId
	return e
}

func (e *EffectRegisterEntity) WithClass(class string) *EffectRegisterEntity {
	e.Class = &class
	return e
}

func (e *EffectRegisterEntity) WithSpriteId(spriteId resources.TilesheetSpriteId) *EffectRegisterEntity {
	e.SpriteId = &spriteId
	return e
}

func (e *EffectRegisterEntity) WithProperties(properties *util.Properties) *EffectRegisterEntity {
	e.Properties = &properties
	return e
}

func (e *EffectRegisterEntity) WithMapLocation(mapLocation MapLocation) *EffectRegisterEntity {
	e.MapLocation = &mapLocation
	return e
}

func (e *EffectRegisterEntity) WithEntityLocation(entityLocation string) *EffectRegisterEntity {
	e.EntityLocation = &entityLocation
	return e
}

func NewMutateModeBasedEntityEffect(entityId string) *EffectMutateModeBasedEntity {
	return &EffectMutateModeBasedEntity{
		EntityId: entityId,
	}
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

func NewSetEntityLocationEffect(entityId string) *EffectSetEntityLocation {
	return &EffectSetEntityLocation{
		EntityId: entityId,
	}
}

func (e *EffectSetEntityLocation) WithToReference(toReference string) *EffectSetEntityLocation {
	e.ToReference = &toReference
	return e
}

func (e *EffectSetEntityLocation) WithToLocation(toLocation MapLocation) *EffectSetEntityLocation {
	e.ToLocation = &toLocation
	return e
}

func (e *EffectSetEntityLocation) WithToEntityId(toEntityId string) *EffectSetEntityLocation {
	e.ToEntityId = &toEntityId
	return e
}

func NewChangePlayerRendererEffect(style string) *EffectChangePlayerRenderer {
	return &EffectChangePlayerRenderer{
		Style: style,
	}
}

func NewTeleportPlayerEffect() *EffectTeleportPlayer {
	return &EffectTeleportPlayer{
	}
}

func (e *EffectTeleportPlayer) WithToReference(toReference string) *EffectTeleportPlayer {
	e.ToReference = &toReference
	return e
}

func (e *EffectTeleportPlayer) WithToLocation(toLocation MapLocation) *EffectTeleportPlayer {
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

func (e *EffectTeleportPlayer) WithInterstitialEffects(interstitialEffects []Effect) *EffectTeleportPlayer {
	e.InterstitialEffects = &interstitialEffects
	return e
}

func NewAddSoundProviderEffect(entityId string) *EffectAddSoundProvider {
	return &EffectAddSoundProvider{
		EntityId: entityId,
	}
}

func (e *EffectAddSoundProvider) WithModeBase(modeBase ModeBaseSoundProviderConfig) *EffectAddSoundProvider {
	e.ModeBase = &modeBase
	return e
}

func NewEntityFaceDirectionEffect(entityId string) *EffectEntityFaceDirection {
	return &EffectEntityFaceDirection{
		EntityId: entityId,
	}
}

func (e *EffectEntityFaceDirection) WithDirection(direction input.Direction) *EffectEntityFaceDirection {
	e.Direction = &direction
	return e
}

func (e *EffectEntityFaceDirection) WithTargetEntity(targetEntity string) *EffectEntityFaceDirection {
	e.TargetEntity = &targetEntity
	return e
}

func NewStartScriptedMotionEffect(entityId string) *EffectStartScriptedMotion {
	return &EffectStartScriptedMotion{
		EntityId: entityId,
	}
}

func (e *EffectStartScriptedMotion) WithMotionId(motionId string) *EffectStartScriptedMotion {
	e.MotionId = motionId
	return e
}

func (e *EffectStartScriptedMotion) WithLocation(location MapLocation) *EffectStartScriptedMotion {
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

func NewPushEntityBehaviorEffect(entityId string) *EffectPushEntityBehavior {
	return &EffectPushEntityBehavior{
		EntityId: entityId,
	}
}

func (e *EffectPushEntityBehavior) WithScriptedMotion(scriptedMotion EntityBehaviorScriptedMotion) *EffectPushEntityBehavior {
	e.ScriptedMotion = &scriptedMotion
	return e
}

func (e *EffectPushEntityBehavior) WithFacingEntity(facingEntity EntityBehaviorFacingEntity) *EffectPushEntityBehavior {
	e.FacingEntity = &facingEntity
	return e
}

func NewPopEntityBehaviorEffect(entityId string) *EffectPopEntityBehavior {
	return &EffectPopEntityBehavior{
		EntityId: entityId,
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

func NewTriggerMovementEffect(entityId string) *EffectTriggerMovement {
	return &EffectTriggerMovement{
		EntityId: entityId,
	}
}

func (e *EffectTriggerMovement) WithDirection(direction input.Direction) *EffectTriggerMovement {
	e.Direction = &direction
	return e
}

func (e *EffectTriggerMovement) WithLocation(location MapLocation) *EffectTriggerMovement {
	e.Location = &location
	return e
}

func (e *EffectTriggerMovement) WithMoveState(moveState MoveState) *EffectTriggerMovement {
	e.MoveState = &moveState
	return e
}

func NewResetMovementEffect(entityId string) *EffectResetMovement {
	return &EffectResetMovement{
		EntityId: entityId,
	}
}

func NewYieldElythiumEffect(amount int) *EffectYieldElythium {
	return &EffectYieldElythium{
		Amount: amount,
	}
}

func NewTriggerCombatEffect(background string) *EffectTriggerCombat {
	return &EffectTriggerCombat{
		Background: background,
	}
}

func (e *EffectTriggerCombat) WithCombatId(combatId string) *EffectTriggerCombat {
	e.CombatId = combatId
	return e
}

func (e *EffectTriggerCombat) WithOpponent(opponent game.CombatOpponent) *EffectTriggerCombat {
	e.Opponent = &opponent
	return e
}

func (e *EffectTriggerCombat) WithPlayer(player game.CombatPlayer) *EffectTriggerCombat {
	e.Player = &player
	return e
}

func (e *EffectTriggerCombat) WithReward(reward game.CombatReward) *EffectTriggerCombat {
	e.Reward = &reward
	return e
}

func (e *EffectTriggerCombat) WithTrainingSequence(trainingSequence string) *EffectTriggerCombat {
	e.TrainingSequence = &trainingSequence
	return e
}

func NewLoadMapEffect(mapName string) *EffectLoadMap {
	return &EffectLoadMap{
		MapName: mapName,
	}
}

func (e *EffectLoadMap) WithWaypoint(waypoint string) *EffectLoadMap {
	e.Waypoint = &waypoint
	return e
}

func NewDialogueEffect(text string) *EffectDialogue {
	return &EffectDialogue{
		Text: text,
	}
}

func (e *EffectDialogue) WithDialogueId(dialogueId string) *EffectDialogue {
	e.DialogueId = dialogueId
	return e
}

func (e *EffectDialogue) WithStyle(style DialogueStyle) *EffectDialogue {
	e.Style = &style
	return e
}

func NewChatterEffect(entityId string, durationSeconds float64, message string) *EffectChatter {
	return &EffectChatter{
		EntityId: entityId,
		DurationSeconds: durationSeconds,
		Message: message,
	}
}

func (e *EffectChatter) WithChatterId(chatterId string) *EffectChatter {
	e.ChatterId = chatterId
	return e
}

func NewFadeEffect(durationSeconds float64, transitions int) *EffectFade {
	return &EffectFade{
		DurationSeconds: durationSeconds,
		Transitions: transitions,
	}
}

func (e *EffectFade) WithFadeId(fadeId string) *EffectFade {
	e.FadeId = fadeId
	return e
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

