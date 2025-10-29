// AUTO-GENERATED - DO NOT EDIT
// Generated at 2025-10-28T21:51:31-04:00 by go generate
// Source: internal/game/events/effects.go

package events

// Effect represents a single effect that can be applied
// Only one effect type should be set per Effect instance
type Effect struct {
	Function *EffectFunction
	Dialogue *EffectDialogue
	Chatter *EffectChatter
	YieldElythium *EffectYieldElythium
	Timer *EffectTimer
	MutateModeBasedEntity *EffectMutateModeBasedEntity
	ResetModeBasedEntityAnimation *EffectResetModeBasedEntityAnimation
	MutateBlockingPresence *EffectMutateBlockingPresence
	SetWorldState *EffectSetWorldState
	SetEntityLocation *EffectSetEntityLocation
	TeleportPlayer *EffectTeleportPlayer
	Plan *EffectPlan
	MutateEntityBehavior *EffectMutateEntityBehavior
	Fade *EffectFade
	DeactivateFade *EffectDeactivateFade
	TriggerMovement *EffectTriggerMovement
	ResetMovement *EffectResetMovement
	OverrideCamera *EffectOverrideCamera
	PopCameraOverride *EffectPopCameraOverride
	MutateFollowCamera *EffectMutateFollowCamera
	MutateNPC *EffectMutateNPC
	TriggerCombat *EffectTriggerCombat
	EntityFaceDirection *EffectEntityFaceDirection
	StartScriptedMotion *EffectStartScriptedMotion
	OverrideEntityBehavior *EffectOverrideEntityBehavior
	PopEntityBehaviorOverride *EffectPopEntityBehaviorOverride
}
