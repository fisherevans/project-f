package events

import (
	"fmt"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

var planIdCounter uint64
var effectIdCounter uint64

// EffectDispatcher is a function that processes dispatched effects
type EffectDispatcher func(effects ...DispatchedEffect)

// PlanExecutor manages the execution of effect plans
type PlanExecutor struct {
	activePlans      []*activePlan
	effectDispatcher EffectDispatcher
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

func NewPlanExecutor(effectDispatcher EffectDispatcher) *PlanExecutor {
	return &PlanExecutor{
		activePlans:      make([]*activePlan, 0),
		effectDispatcher: effectDispatcher,
	}
}

// GenerateEffectId generates a unique sequential ID for effects
func (pe *PlanExecutor) GenerateEffectId(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, atomic.AddUint64(&effectIdCounter, 1))
}

// Note: autoGenerateEffectIds is no longer needed - each effect type now handles
// ID generation in its FillDefaultsAndValidate() method

// StartPlan begins executing a new plan, dispatching effects immediately via the effect dispatcher.
// This method executes all non-blocking effects synchronously and only queues the plan
// if it encounters blocking effects (those with completion IDs).
func (pe *PlanExecutor) StartPlan(source EntityContext, plan *EffectPlan) {
	effects := pe.ExecutePlanImmediately(source, plan)
	if len(effects) > 0 {
		pe.effectDispatcher(effects...)
	}
}

// ExecutePlanImmediately executes a plan synchronously, returning all effects that can be executed immediately.
// If the plan has blocking effects (those with completion IDs), it adds the plan to the executor queue
// and returns those effects as well. This allows system effects to execute non-blocking effects immediately
// while still properly queuing blocking effects.
func (pe *PlanExecutor) ExecutePlanImmediately(source EntityContext, plan *EffectPlan) []DispatchedEffect {
	// Auto-generate plan ID if empty
	if plan.PlanId == "" {
		plan.PlanId = fmt.Sprintf("plan_%d", atomic.AddUint64(&planIdCounter, 1))
	}

	var effects []DispatchedEffect

	ap := &activePlan{
		source:          source,
		plan:            plan,
		currentStep:     0,
		waitingFor:      make(map[string]bool),
		nextEffectIndex: 0,
	}

	log.Info().Str("planId", plan.PlanId).Msg("Executing plan immediately")

	// Process steps until we hit a blocking effect
	for ap.currentStep < len(ap.plan.Steps) {
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
			ap.currentStep++
			ap.nextEffectIndex = 0
			continue
		}

		// Process effects based on serial vs parallel
		if isSerial {
			// Serial: dispatch effects until we hit one with a completion ID
			for ap.nextEffectIndex < len(stepEffects) {
				effect := stepEffects[ap.nextEffectIndex]

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
					pe.activePlans = append(pe.activePlans, ap)
					return effects
				}
			}
			// All serial effects completed without blocking
			ap.currentStep++
			ap.nextEffectIndex = 0
		} else {
			// Parallel: dispatch all effects at once
			for i := range stepEffects {
				effect := &stepEffects[i]

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
			
			// If any parallel effects are blocking, queue the plan
			if len(ap.waitingFor) > 0 {
				pe.activePlans = append(pe.activePlans, ap)
				return effects
			}
			
			// All parallel effects completed without blocking
			ap.currentStep++
			ap.nextEffectIndex = 0
		}
	}
	
	// All steps completed
	return effects
}

// Update processes all active plans and dispatches their next effects.
// This is called by the main game loop each frame to continue executing queued plans.
func (pe *PlanExecutor) Update() {
	defer pe.cleanupCompletedPlans()

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

				pe.effectDispatcher(DispatchedEffect{
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
			// Parallel: dispatch all effects at once (only if not already dispatched)
			if ap.nextEffectIndex == 0 {
				for i := range stepEffects {
					effect := &stepEffects[i]

					completionIds := GetCompletionIds(*effect)
					for _, completionId := range completionIds {
						ap.waitingFor[completionId] = true
						log.Info().Str("planId", ap.plan.PlanId).Str("completionId", completionId).Msg("waiting for completion")
					}

					pe.effectDispatcher(DispatchedEffect{
						Source: ap.source,
						Effect: *effect,
					})
				}
				// Mark all parallel effects as dispatched
				ap.nextEffectIndex = len(stepEffects)
			}
		}
	}
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
			// Next call to Update will dispatch the next effect
		}
	}
}

// cleanupCompletedPlans removes plans that have finished and marks them as complete
func (pe *PlanExecutor) cleanupCompletedPlans() {
	var activePlans []*activePlan

	for _, ap := range pe.activePlans {
		if ap.currentStep < len(ap.plan.Steps) || len(ap.waitingFor) > 0 {
			activePlans = append(activePlans, ap)
		} else {
			log.Info().Str("planId", ap.plan.PlanId).Msg("Plan completed")
			// Mark the plan as complete so parent plans can continue
			if ap.plan.PlanId != "" {
				pe.MarkComplete(MakePlanCompletionId(ap.plan.PlanId))
			}
		}
	}

	pe.activePlans = activePlans
}

// GetCompletionIds extracts all completion IDs from an effect
// This is exported so external packages can compute the same IDs
func GetCompletionIds(effect Effect) []string {
	var ids []string

	switch e := effect.(type) {
	case *EffectPlan:
		// Nested plans should be waited for
		if e.PlanId != "" {
			ids = append(ids, MakePlanCompletionId(e.PlanId))
		}
	case *EffectFade:
		if e.FadeId != "" {
			ids = append(ids, MakeFadeCompletionId(e.FadeId))
		}
	case *EffectTimer:
		if e.TimerId != "" {
			ids = append(ids, MakeTimerCompletionId(e.TimerId))
		}
	case *EffectDialogue:
		if e.DialogueId != "" {
			ids = append(ids, MakeDialogueCompletionId(e.DialogueId))
		}
	case *EffectChatter:
		if e.ChatterId != "" {
			ids = append(ids, MakeChatterCompletionId(e.ChatterId))
		}
	case *EffectTriggerCombat:
		if e.CombatId != "" {
			ids = append(ids, MakeCombatCompletionId(e.CombatId))
		}
	case *EffectStartScriptedMotion:
		if e.MotionId != "" {
			ids = append(ids, MakeMotionCompletionId(e.MotionId))
		}
	}

	return ids
}

// Completion ID constructors - ensures consistency across packages
func MakePlanCompletionId(planId string) string {
	return fmt.Sprintf("plan:%s", planId)
}

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

func MakeMotionCompletionId(motionId string) string {
	return fmt.Sprintf("motion:%s", motionId)
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

func (pe *PlanExecutor) MarkMotionComplete(motionId string) {
	if motionId != "" {
		pe.MarkComplete(MakeMotionCompletionId(motionId))
	}
}
