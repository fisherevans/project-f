// AUTO-GENERATED - DO NOT EDIT
// Generated at 2025-10-30T16:28:58-04:00 by go generate
// Source: internal/game/events/effect.go

package events

import (
	"fmt"
	"sync/atomic"
)

var dialogueCounter atomic.Uint64
var chatterCounter atomic.Uint64
var timerCounter atomic.Uint64
var batchCounter atomic.Uint64
var fadeCounter atomic.Uint64
var triggerCombatCounter atomic.Uint64
var startScriptedMotionCounter atomic.Uint64
var registerEntityCounter atomic.Uint64

func (e *EffectFunction) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()


	return reporter.report()
}

func (e *EffectDialogue) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	if e.DialogueId == "" {
		id := dialogueCounter.Add(1)
		e.DialogueId = fmt.Sprintf("dialogue-%05d", id)
	}

	reporter.requireString("dialogueId", e.DialogueId)
	reporter.requireString("text", e.Text)

	return reporter.report()
}

func (e *EffectChatter) FillDefaultsAndValidate() error {
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

func (e *EffectYieldElythium) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()


	return reporter.report()
}

func (e *EffectTimer) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	if e.TimerId == "" {
		id := timerCounter.Add(1)
		e.TimerId = fmt.Sprintf("timer-%05d", id)
	}

	reporter.requireString("timerId", e.TimerId)

	return reporter.report()
}

func (e *EffectMutateModeBasedEntity) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectResetModeBasedEntityAnimation) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectMutateBlockingPresence) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectSetWorldState) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	reporter.requireString("key", e.Key)

	return reporter.report()
}

func (e *EffectSetEntityLocation) FillDefaultsAndValidate() error {
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

func (e *EffectTeleportPlayer) FillDefaultsAndValidate() error {
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

func (e *EffectBatch) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	if e.BatchId == "" {
		id := batchCounter.Add(1)
		e.BatchId = fmt.Sprintf("batch-%05d", id)
	}

	reporter.requireString("batchId", e.BatchId)
	for i, item := range e.Effects {
		if err := item.FillDefaultsAndValidate(); err != nil {
			reporter.sub("effects").sub(fmt.Sprintf("[%d]", i)).addf("", "%v", err)
		}
	}

	return reporter.report()
}

func (e *EffectMutateEntityBehavior) FillDefaultsAndValidate() error {
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

func (e *EffectFade) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	if e.FadeId == "" {
		id := fadeCounter.Add(1)
		e.FadeId = fmt.Sprintf("fade-%05d", id)
	}

	reporter.requireString("fadeId", e.FadeId)

	return reporter.report()
}

func (e *EffectDeactivateFade) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	reporter.requireString("fadeId", e.FadeId)

	return reporter.report()
}

func (e *EffectTriggerMovement) FillDefaultsAndValidate() error {
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

func (e *EffectResetMovement) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectOverrideCamera) FillDefaultsAndValidate() error {
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

func (e *EffectPopCameraOverride) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()


	return reporter.report()
}

func (e *EffectMutateFollowCamera) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()


	return reporter.report()
}

func (e *EffectMutateNPC) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectTriggerCombat) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	if e.CombatId == "" {
		id := triggerCombatCounter.Add(1)
		e.CombatId = fmt.Sprintf("trigger-combat-%05d", id)
	}

	reporter.requireString("combatId", e.CombatId)
	reporter.requireString("background", e.Background)

	return reporter.report()
}

func (e *EffectEntityFaceDirection) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectStartScriptedMotion) FillDefaultsAndValidate() error {
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

func (e *EffectPushEntityBehavior) FillDefaultsAndValidate() error {
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

func (e *EffectPopEntityBehavior) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectDeleteEntity) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectRegisterEntity) FillDefaultsAndValidate() error {
	reporter := newIssueReporter()

	if e.EntityId == "" {
		id := registerEntityCounter.Add(1)
		e.EntityId = fmt.Sprintf("register-entity-%05d", id)
	}

	// Validate one_of group: location
	locationCount := 0
	if e.MapLocation != nil {
		locationCount++
	}
	if e.EntityLocation != nil {
		locationCount++
	}
	if locationCount != 1 {
		reporter.addf("location", "exactly one of [mapLocation, entityLocation] must be set")
	}

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

