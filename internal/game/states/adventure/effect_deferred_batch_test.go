package adventure

import (
	"testing"
)

func TestEffectDeferredBatch_FillDefaults(t *testing.T) {
	e := &EffectDeferredBatch{
		BuildEffects: func(source EntityReader, s *State) []Effect {
			return nil
		},
	}

	if err := e.FillDefaultsAndValidate(); err != nil {
		t.Fatalf("FillDefaultsAndValidate error: %v", err)
	}

	if e.BatchId == "" {
		t.Error("BatchId should be auto-generated")
	}
	if e.CompletionID() == "" {
		t.Error("CompletionID should be non-empty after fill")
	}
}

func TestEffectDeferredBatch_UniqueIds(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		e := &EffectDeferredBatch{
			BuildEffects: func(source EntityReader, s *State) []Effect { return nil },
		}
		e.FillDefaultsAndValidate()
		if ids[e.BatchId] {
			t.Fatalf("duplicate BatchId: %s", e.BatchId)
		}
		ids[e.BatchId] = true
	}
}

func TestEffectDeferredBatch_EmptyBuildEffects(t *testing.T) {
	h := newPlanTestHarness()

	deferred := &EffectDeferredBatch{
		BuildEffects: func(source EntityReader, s *State) []Effect {
			return nil
		},
	}
	deferred.FillDefaultsAndValidate()

	// When BuildEffects returns nil, the deferred batch should mark itself complete
	// The Process method calls MarkComplete internally
	after := newTrackingEffect("after")
	batch := NewSerialPlan(deferred, after)
	batch.FillDefaultsAndValidate()
	h.pe.StartPlan(nil, batch)

	// The deferred batch's Process should mark it complete immediately
	// since BuildEffects returned nil, allowing "after" to proceed
	h.pe.Update()

	if !after.processed {
		t.Error("after should be processed (empty deferred batch should complete immediately)")
	}
}

func TestEffectDeferredBatch_WithChildEffects(t *testing.T) {
	h := newPlanTestHarness()

	innerTracked := newTrackingEffect("inner")
	deferred := &EffectDeferredBatch{
		BuildEffects: func(source EntityReader, s *State) []Effect {
			return []Effect{innerTracked}
		},
	}
	deferred.FillDefaultsAndValidate()

	after := newTrackingEffect("after")
	batch := NewSerialPlan(deferred, after)
	batch.FillDefaultsAndValidate()
	h.pe.StartPlan(nil, batch)

	if !innerTracked.processed {
		t.Error("inner effect should be processed")
	}

	// After the inner plan completes, the deferred batch should complete,
	// allowing "after" to proceed
	h.pe.Update()

	if !after.processed {
		t.Error("after should be processed once deferred batch's children complete")
	}
}

func TestEffectDeferredBatch_WithBlockingChild(t *testing.T) {
	h := newPlanTestHarness()

	block := newBlockingEffect("child_block", "test:child_block")
	afterBlock := newTrackingEffect("after_block")
	deferred := &EffectDeferredBatch{
		BuildEffects: func(source EntityReader, s *State) []Effect {
			return []Effect{block, afterBlock}
		},
	}
	deferred.FillDefaultsAndValidate()

	outerAfter := newTrackingEffect("outer_after")
	batch := NewSerialPlan(deferred, outerAfter)
	batch.FillDefaultsAndValidate()
	h.pe.StartPlan(nil, batch)

	if !block.processed {
		t.Error("blocking child should be dispatched")
	}
	if afterBlock.processed {
		t.Error("after_block should be blocked")
	}
	if outerAfter.processed {
		t.Error("outer_after should be blocked by deferred batch")
	}

	h.pe.MarkComplete("test:child_block")
	h.pe.Update()

	if !afterBlock.processed {
		t.Error("after_block should be processed after unblock")
	}

	h.pe.Update()

	if !outerAfter.processed {
		t.Error("outer_after should be processed after deferred batch completes")
	}
}
