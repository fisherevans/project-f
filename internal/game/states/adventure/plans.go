package adventure

import (
	"fmt"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

var planIdCounter uint64
var effectIdCounter uint64

// EffectDispatcher is a function that processes dispatched effects
type EffectDispatcher func(effects ...DispatchedEffect)

// PlanExecutor manages the execution of effect batches
type PlanExecutor struct {
	activeBatches    []*activeBatch
	effectDispatcher EffectDispatcher
}

type activeBatch struct {
	source          EntityContext
	batch           *EffectBatch
	waitingFor      map[string]bool // tracks completion IDs we're waiting for
	nextEffectIndex int             // index of next effect to dispatch
}

func NewPlanExecutor(effectDispatcher EffectDispatcher) *PlanExecutor {
	return &PlanExecutor{
		activeBatches:    make([]*activeBatch, 0),
		effectDispatcher: effectDispatcher,
	}
}

// GenerateEffectId generates a unique sequential ID for effects
func (pe *PlanExecutor) GenerateEffectId(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, atomic.AddUint64(&effectIdCounter, 1))
}

// Note: autoGenerateEffectIds is no longer needed - each effect type now handles
// ID generation in its FillDefaultsAndValidate() method

// StartPlan begins executing a new batch, dispatching effects immediately via the effect dispatcher.
// This method executes all non-blocking effects synchronously and only queues the batch
// if it encounters blocking effects (those with completion IDs).
func (pe *PlanExecutor) StartPlan(source EntityContext, batch *EffectBatch) {
	effects := pe.ExecuteBatchImmediately(source, batch)
	if len(effects) > 0 {
		pe.effectDispatcher(effects...)
	}
}

// ExecuteBatchImmediately executes a batch synchronously, returning all effects that can be executed immediately.
// If the batch has blocking effects (those with completion IDs), it adds the batch to the executor queue
// and returns those effects as well. This allows system effects to execute non-blocking effects immediately
// while still properly queuing blocking effects.
func (pe *PlanExecutor) ExecuteBatchImmediately(source EntityContext, batch *EffectBatch) []DispatchedEffect {
	// Auto-generate batch ID if empty
	if batch.BatchId == "" {
		batch.BatchId = fmt.Sprintf("plan_%d", atomic.AddUint64(&planIdCounter, 1))
	}

	var effects []DispatchedEffect

	ab := &activeBatch{
		source:          source,
		batch:           batch,
		waitingFor:      make(map[string]bool),
		nextEffectIndex: 0,
	}

	log.Info().Str("batchId", batch.BatchId).Msg("Executing batch immediately")

	if len(batch.Effects) == 0 {
		return effects
	}

	// Process effects based on serial vs parallel
	if !batch.IsParallel() {
		// Serial: dispatch effects until we hit one with a completion ID
		for ab.nextEffectIndex < len(batch.Effects) {
			effect := batch.Effects[ab.nextEffectIndex]

			completionIds := GetCompletionIds(effect)

			log.Info().
				Str("batchId", batch.BatchId).
				Int("effectIndex", ab.nextEffectIndex).
				Int("totalEffects", len(batch.Effects)).
				Int("completionIds", len(completionIds)).
				Msg("Executing serial effect")

			for _, completionId := range completionIds {
				ab.waitingFor[completionId] = true
				log.Info().Str("batchId", batch.BatchId).Str("completionId", completionId).Msg("waiting for completion")
			}

			effects = append(effects, DispatchedEffect{
				Source: ab.source,
				Effect: effect,
			})

			ab.nextEffectIndex++

			// If this effect has completion IDs, stop and wait
			if len(completionIds) > 0 {
				log.Info().Str("batchId", batch.BatchId).Msg("Stopping serial execution - waiting for completion")
				pe.activeBatches = append(pe.activeBatches, ab)
				return effects
			}
		}
		// All serial effects completed without blocking
	} else {
		// Parallel: dispatch all effects at once
		for i := range batch.Effects {
			effect := &batch.Effects[i]

			completionIds := GetCompletionIds(*effect)
			for _, completionId := range completionIds {
				ab.waitingFor[completionId] = true
				log.Info().Str("batchId", batch.BatchId).Str("completionId", completionId).Msg("waiting for completion")
			}

			effects = append(effects, DispatchedEffect{
				Source: ab.source,
				Effect: *effect,
			})
		}

		// Mark all parallel effects as dispatched
		ab.nextEffectIndex = len(batch.Effects)

		// If any parallel effects are blocking, queue the batch
		if len(ab.waitingFor) > 0 {
			pe.activeBatches = append(pe.activeBatches, ab)
			return effects
		}
	}

	// All effects completed without blocking
	return effects
}

// Update processes all active batches and dispatches their next effects.
// This is called by the main game loop each frame to continue executing queued batches.
func (pe *PlanExecutor) Update() {
	defer pe.cleanupCompletedBatches()

	for _, ab := range pe.activeBatches {
		// If we're waiting for effects to complete, don't dispatch more
		if len(ab.waitingFor) > 0 {
			continue
		}

		// Check if batch is complete
		if ab.nextEffectIndex >= len(ab.batch.Effects) {
			continue
		}

		// Dispatch effects based on serial vs parallel
		if !ab.batch.IsParallel() {
			// Serial: dispatch effects until we hit one with a completion ID
			for ab.nextEffectIndex < len(ab.batch.Effects) {
				effect := ab.batch.Effects[ab.nextEffectIndex]

				completionIds := GetCompletionIds(effect)

				log.Info().
					Str("batchId", ab.batch.BatchId).
					Int("effectIndex", ab.nextEffectIndex).
					Int("totalEffects", len(ab.batch.Effects)).
					Int("completionIds", len(completionIds)).
					Msg("Executing serial effect")

				for _, completionId := range completionIds {
					ab.waitingFor[completionId] = true
					log.Info().Str("batchId", ab.batch.BatchId).Str("completionId", completionId).Msg("waiting for completion")
				}

				pe.effectDispatcher(DispatchedEffect{
					Source: ab.source,
					Effect: effect,
				})

				ab.nextEffectIndex++

				// If this effect has completion IDs, stop and wait
				if len(completionIds) > 0 {
					log.Info().Str("batchId", ab.batch.BatchId).Msg("Stopping serial execution - waiting for completion")
					break
				}
				// Otherwise, continue to next effect in same frame
			}
		} else {
			// Parallel: dispatch all effects at once (only if not already dispatched)
			if ab.nextEffectIndex == 0 {
				for i := range ab.batch.Effects {
					effect := &ab.batch.Effects[i]

					completionIds := GetCompletionIds(*effect)
					for _, completionId := range completionIds {
						ab.waitingFor[completionId] = true
						log.Info().Str("batchId", ab.batch.BatchId).Str("completionId", completionId).Msg("waiting for completion")
					}

					pe.effectDispatcher(DispatchedEffect{
						Source: ab.source,
						Effect: *effect,
					})
				}
				// Mark all parallel effects as dispatched
				ab.nextEffectIndex = len(ab.batch.Effects)
			}
		}
	}
}

// MarkComplete marks an effect as complete (called when timers/overlays/etc finish)
func (pe *PlanExecutor) MarkComplete(completionId string) {
	log.Info().Str("completionId", completionId).Msg("Effect marked complete")

	for _, ab := range pe.activeBatches {
		if ab.waitingFor[completionId] {
			delete(ab.waitingFor, completionId)
			log.Info().
				Str("batchId", ab.batch.BatchId).
				Str("completionId", completionId).
				Int("remaining", len(ab.waitingFor)).
				Msg("Effect completed in batch")
			// nextEffectIndex is already incremented, just need to clear waitingFor
			// Next call to Update will dispatch the next effect
		}
	}
}

// cleanupCompletedBatches removes batches that have finished and marks them as complete
func (pe *PlanExecutor) cleanupCompletedBatches() {
	var activeBatches []*activeBatch

	for _, ab := range pe.activeBatches {
		if ab.nextEffectIndex < len(ab.batch.Effects) || len(ab.waitingFor) > 0 {
			activeBatches = append(activeBatches, ab)
		} else {
			log.Info().Str("batchId", ab.batch.BatchId).Msg("Batch completed")
			// Mark the batch as complete so parent batches can continue
			if ab.batch.BatchId != "" {
				pe.MarkComplete(MakePlanCompletionId(ab.batch.BatchId))
			}
		}
	}

	pe.activeBatches = activeBatches
}

// GetCompletionIds extracts all completion IDs from an effect
// This is exported so external packages can compute the same IDs
func GetCompletionIds(effect Effect) []string {
	var ids []string

	switch e := effect.(type) {
	case *EffectBatch:
		// Nested batches should be waited for
		if e.BatchId != "" {
			ids = append(ids, MakePlanCompletionId(e.BatchId))
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
	case *EffectWaitForCondition:
		if e.ConditionId != "" {
			ids = append(ids, MakeConditionCompleteId(e.ConditionId))
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

func MakeConditionCompleteId(motionId string) string {
	return fmt.Sprintf("condition:%s", motionId)
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

func (pe *PlanExecutor) MarkConditionComplete(conditionId string) {
	if conditionId != "" {
		pe.MarkComplete(MakeConditionCompleteId(conditionId))
	}
}
