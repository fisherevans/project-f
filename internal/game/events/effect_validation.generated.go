// AUTO-GENERATED - DO NOT EDIT
// Generated at 2025-10-23T15:01:45-04:00 by go generate
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

func (e *EffectMutateEntity) Validate() error {
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
		if err := e.ToLocation.Validate(); err != nil {
			reporter.sub("toLocation").addf("", "%v", err)
		}
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
		if err := e.ToLocation.Validate(); err != nil {
			reporter.sub("toLocation").addf("", "%v", err)
		}
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

	return reporter.report()
}

func (e *EffectBlockInput) Validate() error {
	reporter := newIssueReporter()


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

	reporter.requireString("entityId", e.EntityId)

	return reporter.report()
}

func (e *EffectSetFollowCamera) Validate() error {
	reporter := newIssueReporter()

	// Validate one_of group: target
	targetCount := 0
	if e.EntityId != nil {
		targetCount++
	}
	if targetCount != 1 {
		reporter.addf("target", "exactly one of [entityId] must be set")
	}


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
	if e.MutateEntity != nil {
		effectCount++
		if err := e.MutateEntity.Validate(); err != nil {
			reporter.sub("mutateEntity").addf("", "%v", err)
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
	if e.BlockInput != nil {
		effectCount++
		if err := e.BlockInput.Validate(); err != nil {
			reporter.sub("blockInput").addf("", "%v", err)
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
	if e.SetFollowCamera != nil {
		effectCount++
		if err := e.SetFollowCamera.Validate(); err != nil {
			reporter.sub("setFollowCamera").addf("", "%v", err)
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

	reporter.requirePositive("effectCount", float64(effectCount))
	return reporter.report()
}
