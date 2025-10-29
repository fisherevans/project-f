// AUTO-GENERATED - DO NOT EDIT
// Generated at 2025-10-28T21:51:31-04:00 by go generate
// Source: internal/game/events/effect.go

package events

import (
	"fmt"
	"sync/atomic"
)

var dialogueCounter atomic.Uint64
var chatterCounter atomic.Uint64
var timerCounter atomic.Uint64
var planCounter atomic.Uint64
var fadeCounter atomic.Uint64
var triggerCombatCounter atomic.Uint64
var startScriptedMotionCounter atomic.Uint64

func (p *PlanStep) Validate() error {
	reporter := newIssueReporter()

	// Validate one_of group: type
	typeCount := 0
	if len(p.Serial) > 0 {
		typeCount++
		for i, item := range p.Serial {
			if err := item.Validate(); err != nil {
				reporter.sub("serial").sub(fmt.Sprintf("[%d]", i)).addf("", "%v", err)
			}
		}
	}
	if len(p.Parallel) > 0 {
		typeCount++
		for i, item := range p.Parallel {
			if err := item.Validate(); err != nil {
				reporter.sub("parallel").sub(fmt.Sprintf("[%d]", i)).addf("", "%v", err)
			}
		}
	}
	if typeCount != 1 {
		reporter.addf("type", "exactly one of [serial, parallel] must be set")
	}


	return reporter.report()
}

func (e *EffectFunction) Validate() error {
	reporter := newIssueReporter()


	return reporter.report()
}

func (e *EffectDialogue) Validate() error {
	reporter := newIssueReporter()

	if e.DialogueId == "" {
		id := dialogueCounter.Add(1)
		e.DialogueId = fmt.Sprintf("dialogue-%05d", id)
	}

	reporter.requireString("dialogueId", e.DialogueId)
	reporter.requireString("text", e.Text)

	return reporter.report()
}

func (e *EffectChatter) Validate() error {
	reporter := newIssueReporter()

	if e.ChatterId == "" {
		id := chatterCounter.Add(1)
		e.ChatterId = fmt.Sprintf("chatter-%05d", id)
	}

	reporter.requireString("chatterId", e.ChatterId)
	reporter.requireString("entityId", e.EntityId)
	reporter.requireString("message", e.Message)

	return reporter.report()
}

func (e *EffectYieldElythium) Validate() error {
	reporter := newIssueReporter()


	return reporter.report()
}

func (e *EffectTimer) Validate() error {
	reporter := newIssueReporter()

	if e.TimerId == "" {
		id := timerCounter.Add(1)
		e.TimerId = fmt.Sprintf("timer-%05d", id)
	}

	reporter.requireString("timerId", e.TimerId)

	return reporter.report()
}

func (e *EffectMutateModeBasedEntity) Validate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectResetModeBasedEntityAnimation) Validate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectMutateBlockingPresence) Validate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectSetWorldState) Validate() error {
	reporter := newIssueReporter()

	reporter.requireString("key", e.Key)

	return reporter.report()
}

func (e *EffectSetEntityLocation) Validate() error {
	reporter := newIssueReporter()

	// Validate one_of group: destination
	destinationCount := 0
	if e.ToReference != nil {
		destinationCount++
	}
	if e.ToLocation != nil {
		destinationCount++
	}
	if e.ToEntityId != nil {
		destinationCount++
	}
	if destinationCount != 1 {
		reporter.addf("destination", "exactly one of [toReference, toLocation, toEntityId] must be set")
	}

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectTeleportPlayer) Validate() error {
	reporter := newIssueReporter()

	// Validate one_of group: destination
	destinationCount := 0
	if e.ToReference != nil {
		destinationCount++
	}
	if e.ToLocation != nil {
		destinationCount++
	}
	if e.ToEntityId != nil {
		destinationCount++
	}
	if destinationCount != 1 {
		reporter.addf("destination", "exactly one of [toReference, toLocation, toEntityId] must be set")
	}


	return reporter.report()
}

func (e *EffectPlan) Validate() error {
	reporter := newIssueReporter()

	if e.PlanId == "" {
		id := planCounter.Add(1)
		e.PlanId = fmt.Sprintf("plan-%05d", id)
	}

	reporter.requireString("planId", e.PlanId)
	for i, item := range e.Steps {
		if err := item.Validate(); err != nil {
			reporter.sub("steps").sub(fmt.Sprintf("[%d]", i)).addf("", "%v", err)
		}
	}

	return reporter.report()
}

func (e *EffectMutateEntityBehavior) Validate() error {
	reporter := newIssueReporter()

	// Validate one_of group: enablement
	enablementCount := 0
	if e.DisableBy != nil {
		enablementCount++
	}
	if e.EnableBy != nil {
		enablementCount++
	}
	if enablementCount != 1 {
		reporter.addf("enablement", "exactly one of [disableBy, enableBy] must be set")
	}

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectFade) Validate() error {
	reporter := newIssueReporter()

	if e.FadeId == "" {
		id := fadeCounter.Add(1)
		e.FadeId = fmt.Sprintf("fade-%05d", id)
	}

	reporter.requireString("fadeId", e.FadeId)

	return reporter.report()
}

func (e *EffectDeactivateFade) Validate() error {
	reporter := newIssueReporter()

	reporter.requireString("fadeId", e.FadeId)

	return reporter.report()
}

func (e *EffectTriggerMovement) Validate() error {
	reporter := newIssueReporter()

	// Validate one_of group: to
	toCount := 0
	if e.Direction != nil {
		toCount++
	}
	if e.Location != nil {
		toCount++
	}
	if toCount != 1 {
		reporter.addf("to", "exactly one of [direction, location] must be set")
	}

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectResetMovement) Validate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectOverrideCamera) Validate() error {
	reporter := newIssueReporter()

	// Validate one_of group: type
	typeCount := 0
	if e.Follow != nil {
		typeCount++
	}
	if typeCount != 1 {
		reporter.addf("type", "exactly one of [follow] must be set")
	}


	return reporter.report()
}

func (e *EffectPopCameraOverride) Validate() error {
	reporter := newIssueReporter()


	return reporter.report()
}

func (e *EffectMutateFollowCamera) Validate() error {
	reporter := newIssueReporter()


	return reporter.report()
}

func (e *EffectMutateNPC) Validate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectTriggerCombat) Validate() error {
	reporter := newIssueReporter()

	if e.CombatId == "" {
		id := triggerCombatCounter.Add(1)
		e.CombatId = fmt.Sprintf("trigger-combat-%05d", id)
	}

	reporter.requireString("combatId", e.CombatId)
	reporter.requireString("background", e.Background)

	return reporter.report()
}

func (e *EffectEntityFaceDirection) Validate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectStartScriptedMotion) Validate() error {
	reporter := newIssueReporter()

	if e.MotionId == "" {
		id := startScriptedMotionCounter.Add(1)
		e.MotionId = fmt.Sprintf("start-scripted-motion-%05d", id)
	}

	// Validate one_of group: target
	targetCount := 0
	if e.Location != nil {
		targetCount++
	}
	if e.ToEntityId != nil {
		targetCount++
	}
	if e.Relative != nil {
		targetCount++
	}
	if targetCount != 1 {
		reporter.addf("target", "exactly one of [location, toEntityId, relative] must be set")
	}

	reporter.requireString("motionId", e.MotionId)
	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectOverrideEntityBehavior) Validate() error {
	reporter := newIssueReporter()

	// Validate one_of group: type
	typeCount := 0
	if e.ScriptedMotion != nil {
		typeCount++
	}
	if typeCount != 1 {
		reporter.addf("type", "exactly one of [scriptedMotion] must be set")
	}

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectPopEntityBehaviorOverride) Validate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

// Validate checks that the effect has exactly one effect type set and calls its validator
func (e Effect) Validate() error {
	reporter := newIssueReporter()
	effectCount := 0

	if e.Function != nil {
		effectCount++
		if err := e.Function.Validate(); err != nil {
			reporter.sub("function").addf("", "%v", err)
		}
	}
	if e.Dialogue != nil {
		effectCount++
		if err := e.Dialogue.Validate(); err != nil {
			reporter.sub("dialogue").addf("", "%v", err)
		}
	}
	if e.Chatter != nil {
		effectCount++
		if err := e.Chatter.Validate(); err != nil {
			reporter.sub("chatter").addf("", "%v", err)
		}
	}
	if e.YieldElythium != nil {
		effectCount++
		if err := e.YieldElythium.Validate(); err != nil {
			reporter.sub("yieldElythium").addf("", "%v", err)
		}
	}
	if e.Timer != nil {
		effectCount++
		if err := e.Timer.Validate(); err != nil {
			reporter.sub("timer").addf("", "%v", err)
		}
	}
	if e.MutateModeBasedEntity != nil {
		effectCount++
		if err := e.MutateModeBasedEntity.Validate(); err != nil {
			reporter.sub("mutateModeBasedEntity").addf("", "%v", err)
		}
	}
	if e.ResetModeBasedEntityAnimation != nil {
		effectCount++
		if err := e.ResetModeBasedEntityAnimation.Validate(); err != nil {
			reporter.sub("resetModeBasedEntityAnimation").addf("", "%v", err)
		}
	}
	if e.MutateBlockingPresence != nil {
		effectCount++
		if err := e.MutateBlockingPresence.Validate(); err != nil {
			reporter.sub("mutateBlockingPresence").addf("", "%v", err)
		}
	}
	if e.SetWorldState != nil {
		effectCount++
		if err := e.SetWorldState.Validate(); err != nil {
			reporter.sub("setWorldState").addf("", "%v", err)
		}
	}
	if e.SetEntityLocation != nil {
		effectCount++
		if err := e.SetEntityLocation.Validate(); err != nil {
			reporter.sub("setEntityLocation").addf("", "%v", err)
		}
	}
	if e.TeleportPlayer != nil {
		effectCount++
		if err := e.TeleportPlayer.Validate(); err != nil {
			reporter.sub("teleportPlayer").addf("", "%v", err)
		}
	}
	if e.Plan != nil {
		effectCount++
		if err := e.Plan.Validate(); err != nil {
			reporter.sub("plan").addf("", "%v", err)
		}
	}
	if e.MutateEntityBehavior != nil {
		effectCount++
		if err := e.MutateEntityBehavior.Validate(); err != nil {
			reporter.sub("mutateEntityBehavior").addf("", "%v", err)
		}
	}
	if e.Fade != nil {
		effectCount++
		if err := e.Fade.Validate(); err != nil {
			reporter.sub("fade").addf("", "%v", err)
		}
	}
	if e.DeactivateFade != nil {
		effectCount++
		if err := e.DeactivateFade.Validate(); err != nil {
			reporter.sub("deactivateFade").addf("", "%v", err)
		}
	}
	if e.TriggerMovement != nil {
		effectCount++
		if err := e.TriggerMovement.Validate(); err != nil {
			reporter.sub("triggerMovement").addf("", "%v", err)
		}
	}
	if e.ResetMovement != nil {
		effectCount++
		if err := e.ResetMovement.Validate(); err != nil {
			reporter.sub("resetMovement").addf("", "%v", err)
		}
	}
	if e.OverrideCamera != nil {
		effectCount++
		if err := e.OverrideCamera.Validate(); err != nil {
			reporter.sub("overrideCamera").addf("", "%v", err)
		}
	}
	if e.PopCameraOverride != nil {
		effectCount++
		if err := e.PopCameraOverride.Validate(); err != nil {
			reporter.sub("popCameraOverride").addf("", "%v", err)
		}
	}
	if e.MutateFollowCamera != nil {
		effectCount++
		if err := e.MutateFollowCamera.Validate(); err != nil {
			reporter.sub("mutateFollowCamera").addf("", "%v", err)
		}
	}
	if e.MutateNPC != nil {
		effectCount++
		if err := e.MutateNPC.Validate(); err != nil {
			reporter.sub("mutateNPC").addf("", "%v", err)
		}
	}
	if e.TriggerCombat != nil {
		effectCount++
		if err := e.TriggerCombat.Validate(); err != nil {
			reporter.sub("triggerCombat").addf("", "%v", err)
		}
	}
	if e.EntityFaceDirection != nil {
		effectCount++
		if err := e.EntityFaceDirection.Validate(); err != nil {
			reporter.sub("entityFaceDirection").addf("", "%v", err)
		}
	}
	if e.StartScriptedMotion != nil {
		effectCount++
		if err := e.StartScriptedMotion.Validate(); err != nil {
			reporter.sub("startScriptedMotion").addf("", "%v", err)
		}
	}
	if e.OverrideEntityBehavior != nil {
		effectCount++
		if err := e.OverrideEntityBehavior.Validate(); err != nil {
			reporter.sub("overrideEntityBehavior").addf("", "%v", err)
		}
	}
	if e.PopEntityBehaviorOverride != nil {
		effectCount++
		if err := e.PopEntityBehaviorOverride.Validate(); err != nil {
			reporter.sub("popEntityBehaviorOverride").addf("", "%v", err)
		}
	}

	reporter.requirePositive("effectCount", float64(effectCount))
	return reporter.report()
}
