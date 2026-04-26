package adventure

import (
	"fmt"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

var deferredBatchCounter atomic.Uint64

// EffectDeferredBatch defers child effect construction to Process time, but
// still blocks the parent plan via its CompletionID. This is used by control
// flow (if/switch) and custom actions where the child effects depend on
// runtime state (expression evaluation, parameter resolution) that isn't
// available at script load time.
type EffectDeferredBatch struct {
	BatchId      string
	ScopeId      string
	BuildEffects func(source EntityReader, s *State) []Effect
}

func (e *EffectDeferredBatch) CompletionID() string {
	if e.BatchId == "" {
		return ""
	}
	return "batch:" + e.BatchId
}

func (e *EffectDeferredBatch) FillDefaultsAndValidate() error {
	if e.BatchId == "" {
		id := deferredBatchCounter.Add(1)
		e.BatchId = fmt.Sprintf("deferred-%05d", id)
	}
	return nil
}

func (e *EffectDeferredBatch) Process(source EntityReader, s *State) bool {
	effects := e.BuildEffects(source, s)
	if len(effects) == 0 {
		s.planExecutor.MarkComplete(e.CompletionID())
		return true
	}
	batch := NewSerialPlan(effects...)
	batch.BatchId = e.BatchId
	batch.ScopeId = e.ScopeId
	if err := batch.FillDefaultsAndValidate(); err != nil {
		log.Error().Err(err).Str("batchId", e.BatchId).Msg("deferred batch: invalid child effects")
		s.planExecutor.MarkComplete(e.CompletionID())
		return true
	}
	s.planExecutor.StartPlan(source, batch)
	return true
}
