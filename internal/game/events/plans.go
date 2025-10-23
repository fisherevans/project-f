package events

import (
	"fmt"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

var planIdCounter uint64
var effectIdCounter uint64

// PlanExecutor manages the execution of effect plans
type PlanExecutor struct {
	activePlans []*activePlan
}

type activePlan struct {
	source      EntityContext
	plan        *EffectPlan
	currentStep int
	waitingFor  map[string]bool // tracks completion IDs we're waiting for

	// For serial: index of next effect to dispatch (0-based)
	// For parallel: -1 if not started, 0 if all dispatched
	nextEffectIndex int
}

func NewPlanExecutor() *PlanExecutor {
	return &PlanExecutor{
		activePlans: make([]*activePlan, 0),
	}
}

// GenerateEffectId generates a unique sequential ID for effects
func (pe *PlanExecutor) GenerateEffectId(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, atomic.AddUint64(&effectIdCounter, 1))
}

// autoGenerateEffectIds auto-generates IDs for effects that need them
func (pe *PlanExecutor) autoGenerateEffectIds(effect *Effect) {
	if effect.Fade != nil && effect.Fade.FadeId == "" {
		effect.Fade.FadeId = pe.GenerateEffectId("fade")
	}
	if effect.Timer != nil && effect.Timer.TimerId == "" {
		effect.Timer.TimerId = pe.GenerateEffectId("timer")
	}
	if effect.Dialogue != nil && effect.Dialogue.DialogueId == "" {
		effect.Dialogue.DialogueId = pe.GenerateEffectId("dialogue")
	}
	if effect.Chatter != nil && effect.Chatter.ChatterId == "" {
		effect.Chatter.ChatterId = pe.GenerateEffectId("chatter")
	}
	if effect.TriggerCombat != nil && effect.TriggerCombat.CombatId == "" {
		effect.TriggerCombat.CombatId = pe.GenerateEffectId("combat")
	}
}

// StartPlan begins executing a new plan
func (pe *PlanExecutor) StartPlan(source EntityContext, plan *EffectPlan) {
	// Auto-generate plan ID if empty
	if plan.PlanId == "" {
		plan.PlanId = fmt.Sprintf("plan_%d", atomic.AddUint64(&planIdCounter, 1))
	}

	ap := &activePlan{
		source:          source,
		plan:            plan,
		currentStep:     0,
		waitingFor:      make(map[string]bool),
		nextEffectIndex: 0, // Start at beginning of first step
	}
	pe.activePlans = append(pe.activePlans, ap)
	log.Info().Str("planId", plan.PlanId).Msg("Plan started")
}

// GetNextEffects returns the effects to execute for all active plans
// Returns effects that should be processed and updates plan state
// Also cleans up completed plans
func (pe *PlanExecutor) GetNextEffects() []DispatchedEffect {
	defer pe.cleanupCompletedPlans()

	var effects []DispatchedEffect

	for _, ap := range pe.activePlans {
		// If we're waiting for effects to complete, don't dispatch more
		if len(ap.waitingFor) > 0 {
			continue
		}

		// Check if plan is complete
		if ap.currentStep >= len(ap.plan.Steps) {
			continue
		}

		step := ap.plan.Steps[ap.currentStep]

		// Determine if this step is serial or parallel
		var stepEffects []Effect
		isSerial := len(step.Serial) > 0
		if isSerial {
			stepEffects = step.Serial
		} else {
			stepEffects = step.Parallel
		}

		if len(stepEffects) == 0 {
			// Empty step, move to next
			ap.currentStep++
			ap.nextEffectIndex = 0
			continue
		}

		// Check if step is complete
		if ap.nextEffectIndex >= len(stepEffects) {
			// All effects in this step are done, move to next step
			ap.currentStep++
			ap.nextEffectIndex = 0
			continue
		}

		// Dispatch effects based on serial vs parallel
		if isSerial {
			// Serial: dispatch effects until we hit one with a completion ID
			for ap.nextEffectIndex < len(stepEffects) {
				effect := stepEffects[ap.nextEffectIndex]

				// Auto-generate IDs before checking for completion
				pe.autoGenerateEffectIds(&effect)

				completionIds := GetCompletionIds(effect)

				log.Info().
					Str("planId", ap.plan.PlanId).
					Int("effectIndex", ap.nextEffectIndex).
					Int("totalEffects", len(stepEffects)).
					Int("completionIds", len(completionIds)).
					Msg("Executing serial effect")

				for _, completionId := range completionIds {
					ap.waitingFor[completionId] = true
					log.Info().Str("planId", ap.plan.PlanId).Str("completionId", completionId).Msg("waiting for completion")
				}

				effects = append(effects, DispatchedEffect{
					Source: ap.source,
					Effect: effect,
				})

				ap.nextEffectIndex++

				// If this effect has completion IDs, stop and wait
				if len(completionIds) > 0 {
					log.Info().Str("planId", ap.plan.PlanId).Msg("Stopping serial execution - waiting for completion")
					break
				}
				// Otherwise, continue to next effect in same frame
			}
		} else {
			// Parallel: dispatch all effects at once
			for i := range stepEffects {
				effect := &stepEffects[i]

				// Auto-generate IDs before checking for completion
				pe.autoGenerateEffectIds(effect)

				completionIds := GetCompletionIds(*effect)
				for _, completionId := range completionIds {
					ap.waitingFor[completionId] = true
					log.Info().Str("planId", ap.plan.PlanId).Str("completionId", completionId).Msg("waiting for completion")
				}

				effects = append(effects, DispatchedEffect{
					Source: ap.source,
					Effect: *effect,
				})
			}
			// Mark all parallel effects as dispatched
			ap.nextEffectIndex = len(stepEffects)
		}
	}

	return effects
}

// MarkComplete marks an effect as complete (called when timers/overlays/etc finish)
func (pe *PlanExecutor) MarkComplete(completionId string) {
	log.Debug().Str("completionId", completionId).Msg("Effect marked complete")

	for _, ap := range pe.activePlans {
		if ap.waitingFor[completionId] {
			delete(ap.waitingFor, completionId)
			log.Info().
				Str("planId", ap.plan.PlanId).
				Str("completionId", completionId).
				Int("remaining", len(ap.waitingFor)).
				Msg("Effect completed in plan")
			// nextEffectIndex is already incremented, just need to clear waitingFor
			// Next call to GetNextEffects will dispatch the next effect
		}
	}
}

// cleanupCompletedPlans removes plans that have finished
func (pe *PlanExecutor) cleanupCompletedPlans() {
	var activePlans []*activePlan

	for _, ap := range pe.activePlans {
		if ap.currentStep < len(ap.plan.Steps) || len(ap.waitingFor) > 0 {
			activePlans = append(activePlans, ap)
		} else {
			log.Info().Str("planId", ap.plan.PlanId).Msg("Plan completed")
		}
	}

	pe.activePlans = activePlans
}

// GetCompletionIds extracts all completion IDs from an effect
// This is exported so external packages can compute the same IDs
func GetCompletionIds(effect Effect) []string {
	var ids []string

	if effect.Fade != nil && effect.Fade.FadeId != "" {
		ids = append(ids, MakeFadeCompletionId(effect.Fade.FadeId))
	}
	if effect.Timer != nil && effect.Timer.TimerId != "" {
		ids = append(ids, MakeTimerCompletionId(effect.Timer.TimerId))
	}
	if effect.Dialogue != nil && effect.Dialogue.DialogueId != "" {
		ids = append(ids, MakeDialogueCompletionId(effect.Dialogue.DialogueId))
	}
	if effect.Chatter != nil && effect.Chatter.ChatterId != "" {
		ids = append(ids, MakeChatterCompletionId(effect.Chatter.ChatterId))
	}
	if effect.TriggerCombat != nil && effect.TriggerCombat.CombatId != "" {
		ids = append(ids, MakeCombatCompletionId(effect.TriggerCombat.CombatId))
	}

	return ids
}

// Completion ID constructors - ensures consistency across packages
func MakeFadeCompletionId(fadeId string) string {
	return fmt.Sprintf("fade:%s", fadeId)
}

func MakeTimerCompletionId(timerId string) string {
	return fmt.Sprintf("timer:%s", timerId)
}

func MakeDialogueCompletionId(dialogueId string) string {
	return fmt.Sprintf("dialogue:%s", dialogueId)
}

func MakeChatterCompletionId(chatterId string) string {
	return fmt.Sprintf("chatter:%s", chatterId)
}

func MakeCombatCompletionId(combatId string) string {
	return fmt.Sprintf("combat:%s", combatId)
}

// MarkDialogueComplete marks a dialogue as complete
func (pe *PlanExecutor) MarkDialogueComplete(dialogueId string) {
	if dialogueId != "" {
		pe.MarkComplete(MakeDialogueCompletionId(dialogueId))
	}
}

// MarkChatterComplete marks a chatter as complete
func (pe *PlanExecutor) MarkChatterComplete(chatterId string) {
	if chatterId != "" {
		pe.MarkComplete(MakeChatterCompletionId(chatterId))
	}
}

// MarkFadeComplete marks a fade as complete
func (pe *PlanExecutor) MarkFadeComplete(fadeId string) {
	if fadeId != "" {
		pe.MarkComplete(MakeFadeCompletionId(fadeId))
		log.Debug().Str("fadeId", fadeId).Msg("Fade completed")
	}
}

// MarkTimerComplete marks a timer as complete
func (pe *PlanExecutor) MarkTimerComplete(timerId string) {
	if timerId != "" {
		pe.MarkComplete(MakeTimerCompletionId(timerId))
	}
}

func (pe *PlanExecutor) MarkCombatComplete(combatId string) {
	if combatId != "" {
		pe.MarkComplete(MakeCombatCompletionId(combatId))
	}

}
