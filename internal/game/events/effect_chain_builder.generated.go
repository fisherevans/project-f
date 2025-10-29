// AUTO-GENERATED - DO NOT EDIT
// Generated at 2025-10-28T21:51:31-04:00 by go generate
// Source: internal/game/events/effects.go

package events

// NewEffect creates a new Effect that can be chained with With methods
func NewEffect() *Effect {
	return &Effect{}
}

// WithFunction sets the Function field and returns the Effect for chaining
func (e *Effect) WithFunction(v *EffectFunction) *Effect {
	e.Function = v
	return e
}

// WithDialogue sets the Dialogue field and returns the Effect for chaining
func (e *Effect) WithDialogue(v *EffectDialogue) *Effect {
	e.Dialogue = v
	return e
}

// WithChatter sets the Chatter field and returns the Effect for chaining
func (e *Effect) WithChatter(v *EffectChatter) *Effect {
	e.Chatter = v
	return e
}

// WithYieldElythium sets the YieldElythium field and returns the Effect for chaining
func (e *Effect) WithYieldElythium(v *EffectYieldElythium) *Effect {
	e.YieldElythium = v
	return e
}

// WithTimer sets the Timer field and returns the Effect for chaining
func (e *Effect) WithTimer(v *EffectTimer) *Effect {
	e.Timer = v
	return e
}

// WithMutateModeBasedEntity sets the MutateModeBasedEntity field and returns the Effect for chaining
func (e *Effect) WithMutateModeBasedEntity(v *EffectMutateModeBasedEntity) *Effect {
	e.MutateModeBasedEntity = v
	return e
}

// WithResetModeBasedEntityAnimation sets the ResetModeBasedEntityAnimation field and returns the Effect for chaining
func (e *Effect) WithResetModeBasedEntityAnimation(v *EffectResetModeBasedEntityAnimation) *Effect {
	e.ResetModeBasedEntityAnimation = v
	return e
}

// WithMutateBlockingPresence sets the MutateBlockingPresence field and returns the Effect for chaining
func (e *Effect) WithMutateBlockingPresence(v *EffectMutateBlockingPresence) *Effect {
	e.MutateBlockingPresence = v
	return e
}

// WithSetWorldState sets the SetWorldState field and returns the Effect for chaining
func (e *Effect) WithSetWorldState(v *EffectSetWorldState) *Effect {
	e.SetWorldState = v
	return e
}

// WithSetEntityLocation sets the SetEntityLocation field and returns the Effect for chaining
func (e *Effect) WithSetEntityLocation(v *EffectSetEntityLocation) *Effect {
	e.SetEntityLocation = v
	return e
}

// WithTeleportPlayer sets the TeleportPlayer field and returns the Effect for chaining
func (e *Effect) WithTeleportPlayer(v *EffectTeleportPlayer) *Effect {
	e.TeleportPlayer = v
	return e
}

// WithPlan sets the Plan field and returns the Effect for chaining
func (e *Effect) WithPlan(v *EffectPlan) *Effect {
	e.Plan = v
	return e
}

// WithMutateEntityBehavior sets the MutateEntityBehavior field and returns the Effect for chaining
func (e *Effect) WithMutateEntityBehavior(v *EffectMutateEntityBehavior) *Effect {
	e.MutateEntityBehavior = v
	return e
}

// WithFade sets the Fade field and returns the Effect for chaining
func (e *Effect) WithFade(v *EffectFade) *Effect {
	e.Fade = v
	return e
}

// WithDeactivateFade sets the DeactivateFade field and returns the Effect for chaining
func (e *Effect) WithDeactivateFade(v *EffectDeactivateFade) *Effect {
	e.DeactivateFade = v
	return e
}

// WithTriggerMovement sets the TriggerMovement field and returns the Effect for chaining
func (e *Effect) WithTriggerMovement(v *EffectTriggerMovement) *Effect {
	e.TriggerMovement = v
	return e
}

// WithResetMovement sets the ResetMovement field and returns the Effect for chaining
func (e *Effect) WithResetMovement(v *EffectResetMovement) *Effect {
	e.ResetMovement = v
	return e
}

// WithOverrideCamera sets the OverrideCamera field and returns the Effect for chaining
func (e *Effect) WithOverrideCamera(v *EffectOverrideCamera) *Effect {
	e.OverrideCamera = v
	return e
}

// WithPopCameraOverride sets the PopCameraOverride field and returns the Effect for chaining
func (e *Effect) WithPopCameraOverride(v *EffectPopCameraOverride) *Effect {
	e.PopCameraOverride = v
	return e
}

// WithMutateFollowCamera sets the MutateFollowCamera field and returns the Effect for chaining
func (e *Effect) WithMutateFollowCamera(v *EffectMutateFollowCamera) *Effect {
	e.MutateFollowCamera = v
	return e
}

// WithMutateNPC sets the MutateNPC field and returns the Effect for chaining
func (e *Effect) WithMutateNPC(v *EffectMutateNPC) *Effect {
	e.MutateNPC = v
	return e
}

// WithTriggerCombat sets the TriggerCombat field and returns the Effect for chaining
func (e *Effect) WithTriggerCombat(v *EffectTriggerCombat) *Effect {
	e.TriggerCombat = v
	return e
}

// WithEntityFaceDirection sets the EntityFaceDirection field and returns the Effect for chaining
func (e *Effect) WithEntityFaceDirection(v *EffectEntityFaceDirection) *Effect {
	e.EntityFaceDirection = v
	return e
}

// WithStartScriptedMotion sets the StartScriptedMotion field and returns the Effect for chaining
func (e *Effect) WithStartScriptedMotion(v *EffectStartScriptedMotion) *Effect {
	e.StartScriptedMotion = v
	return e
}

// WithOverrideEntityBehavior sets the OverrideEntityBehavior field and returns the Effect for chaining
func (e *Effect) WithOverrideEntityBehavior(v *EffectOverrideEntityBehavior) *Effect {
	e.OverrideEntityBehavior = v
	return e
}

// WithPopEntityBehaviorOverride sets the PopEntityBehaviorOverride field and returns the Effect for chaining
func (e *Effect) WithPopEntityBehaviorOverride(v *EffectPopEntityBehaviorOverride) *Effect {
	e.PopEntityBehaviorOverride = v
	return e
}

// With sets an effect field by inspecting the type and returns the Effect for chaining
func (e *Effect) With(v any) *Effect {
	switch val := v.(type) {
	case *EffectFunction:
		return e.WithFunction(val)
	case *EffectDialogue:
		return e.WithDialogue(val)
	case *EffectChatter:
		return e.WithChatter(val)
	case *EffectYieldElythium:
		return e.WithYieldElythium(val)
	case *EffectTimer:
		return e.WithTimer(val)
	case *EffectMutateModeBasedEntity:
		return e.WithMutateModeBasedEntity(val)
	case *EffectResetModeBasedEntityAnimation:
		return e.WithResetModeBasedEntityAnimation(val)
	case *EffectMutateBlockingPresence:
		return e.WithMutateBlockingPresence(val)
	case *EffectSetWorldState:
		return e.WithSetWorldState(val)
	case *EffectSetEntityLocation:
		return e.WithSetEntityLocation(val)
	case *EffectTeleportPlayer:
		return e.WithTeleportPlayer(val)
	case *EffectPlan:
		return e.WithPlan(val)
	case *EffectMutateEntityBehavior:
		return e.WithMutateEntityBehavior(val)
	case *EffectFade:
		return e.WithFade(val)
	case *EffectDeactivateFade:
		return e.WithDeactivateFade(val)
	case *EffectTriggerMovement:
		return e.WithTriggerMovement(val)
	case *EffectResetMovement:
		return e.WithResetMovement(val)
	case *EffectOverrideCamera:
		return e.WithOverrideCamera(val)
	case *EffectPopCameraOverride:
		return e.WithPopCameraOverride(val)
	case *EffectMutateFollowCamera:
		return e.WithMutateFollowCamera(val)
	case *EffectMutateNPC:
		return e.WithMutateNPC(val)
	case *EffectTriggerCombat:
		return e.WithTriggerCombat(val)
	case *EffectEntityFaceDirection:
		return e.WithEntityFaceDirection(val)
	case *EffectStartScriptedMotion:
		return e.WithStartScriptedMotion(val)
	case *EffectOverrideEntityBehavior:
		return e.WithOverrideEntityBehavior(val)
	case *EffectPopEntityBehaviorOverride:
		return e.WithPopEntityBehaviorOverride(val)
	default:
		// Unknown type, ignore
		return e
	}
}
