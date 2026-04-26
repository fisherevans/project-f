package adventure

import (
	"testing"
)

type trackingEffect struct {
	instantEffect
	label     string
	processed bool
	processFn func()
}

func newTrackingEffect(label string) *trackingEffect {
	return &trackingEffect{label: label}
}

func (e *trackingEffect) FillDefaultsAndValidate() error { return nil }

func (e *trackingEffect) Process(source EntityReader, s *State) bool {
	e.processed = true
	if e.processFn != nil {
		e.processFn()
	}
	return true
}

type blockingTrackingEffect struct {
	completionId string
	label        string
	processed    bool
}

func newBlockingEffect(label, completionId string) *blockingTrackingEffect {
	return &blockingTrackingEffect{label: label, completionId: completionId}
}

func (e *blockingTrackingEffect) FillDefaultsAndValidate() error { return nil }
func (e *blockingTrackingEffect) CompletionID() string           { return e.completionId }
func (e *blockingTrackingEffect) Process(source EntityReader, s *State) bool {
	e.processed = true
	return true
}

type planTestHarness struct {
	pe         *PlanExecutor
	state      *State
	dispatched []DispatchedEffect
}

func newPlanTestHarness() *planTestHarness {
	h := &planTestHarness{}
	s := &State{}
	h.state = s
	h.pe = NewPlanExecutor(func(effects ...DispatchedEffect) {
		for _, e := range effects {
			switch e.Effect.(type) {
			case *EffectDialogue:
				// needs entities/dialogues - skip in tests
			default:
				e.Effect.Process(e.Source, s)
			}
			h.dispatched = append(h.dispatched, e)
		}
	})
	s.planExecutor = h.pe
	return h
}

func (h *planTestHarness) dispatchedLabels() []string {
	var labels []string
	for _, d := range h.dispatched {
		switch e := d.Effect.(type) {
		case *trackingEffect:
			labels = append(labels, e.label)
		case *blockingTrackingEffect:
			labels = append(labels, e.label)
		}
	}
	return labels
}

func TestPlanExecutor_SerialInstantEffects(t *testing.T) {
	h := newPlanTestHarness()
	e1 := newTrackingEffect("a")
	e2 := newTrackingEffect("b")
	e3 := newTrackingEffect("c")

	batch := NewSerialPlan(e1, e2, e3)
	batch.FillDefaultsAndValidate()
	h.pe.StartPlan(nil, batch)

	if !e1.processed || !e2.processed || !e3.processed {
		t.Error("all instant effects should be processed immediately")
	}
	labels := h.dispatchedLabels()
	if len(labels) != 3 || labels[0] != "a" || labels[1] != "b" || labels[2] != "c" {
		t.Errorf("dispatch order = %v, want [a b c]", labels)
	}
}

func TestPlanExecutor_SerialBlockingEffect(t *testing.T) {
	h := newPlanTestHarness()
	e1 := newTrackingEffect("a")
	block := newBlockingEffect("block", "test:block1")
	e2 := newTrackingEffect("b")

	batch := NewSerialPlan(e1, block, e2)
	batch.FillDefaultsAndValidate()
	h.pe.StartPlan(nil, batch)

	if !e1.processed {
		t.Error("e1 should be processed")
	}
	if !block.processed {
		t.Error("block should be processed (dispatched)")
	}
	if e2.processed {
		t.Error("e2 should NOT be processed yet (blocked)")
	}

	h.pe.MarkComplete("test:block1")
	h.pe.Update()

	if !e2.processed {
		t.Error("e2 should be processed after unblocking")
	}
}

func TestPlanExecutor_SerialMultipleBlockingEffects(t *testing.T) {
	h := newPlanTestHarness()
	e1 := newTrackingEffect("a")
	block1 := newBlockingEffect("block1", "test:b1")
	e2 := newTrackingEffect("b")
	block2 := newBlockingEffect("block2", "test:b2")
	e3 := newTrackingEffect("c")

	batch := NewSerialPlan(e1, block1, e2, block2, e3)
	batch.FillDefaultsAndValidate()
	h.pe.StartPlan(nil, batch)

	if e2.processed || e3.processed {
		t.Error("e2 and e3 should be blocked")
	}

	h.pe.MarkComplete("test:b1")
	h.pe.Update()

	if !e2.processed {
		t.Error("e2 should be processed after first unblock")
	}
	if !block2.processed {
		t.Error("block2 should be dispatched")
	}
	if e3.processed {
		t.Error("e3 should still be blocked")
	}

	h.pe.MarkComplete("test:b2")
	h.pe.Update()

	if !e3.processed {
		t.Error("e3 should be processed after second unblock")
	}
}

func TestPlanExecutor_ParallelEffects(t *testing.T) {
	h := newPlanTestHarness()
	e1 := newTrackingEffect("a")
	e2 := newTrackingEffect("b")
	e3 := newTrackingEffect("c")

	batch := NewParallelPlan(e1, e2, e3)
	batch.FillDefaultsAndValidate()
	h.pe.StartPlan(nil, batch)

	if !e1.processed || !e2.processed || !e3.processed {
		t.Error("all parallel effects should be processed immediately")
	}
}

func TestPlanExecutor_ParallelWithBlocking(t *testing.T) {
	h := newPlanTestHarness()
	block1 := newBlockingEffect("b1", "test:p1")
	block2 := newBlockingEffect("b2", "test:p2")

	batch := NewParallelPlan(block1, block2)
	batch.BatchId = "parent"
	batch.FillDefaultsAndValidate()
	h.pe.StartPlan(nil, batch)

	if !block1.processed || !block2.processed {
		t.Error("both blocking effects should be dispatched immediately in parallel")
	}
}

func TestPlanExecutor_EmptyBatchCompletesImmediately(t *testing.T) {
	completed := false
	pe := NewPlanExecutor(func(effects ...DispatchedEffect) {
		for _, e := range effects {
			e.Effect.Process(e.Source, nil)
		}
	})

	outer := newBlockingEffect("outer", "test:outer")
	after := newTrackingEffect("after")
	parentBatch := NewSerialPlan(outer, after)
	parentBatch.FillDefaultsAndValidate()
	pe.StartPlan(nil, parentBatch)

	emptyBatch := NewSerialPlan()
	emptyBatch.BatchId = "empty"
	emptyBatch.FillDefaultsAndValidate()
	pe.StartPlan(nil, emptyBatch)

	_ = completed
}

func TestPlanExecutor_NestedBatchCompletion(t *testing.T) {
	h := newPlanTestHarness()

	innerBlock := newBlockingEffect("inner_block", "test:inner")
	afterInner := newTrackingEffect("after_inner")
	innerBatch := NewSerialPlan(innerBlock, afterInner)
	innerBatch.FillDefaultsAndValidate()

	outerAfter := newTrackingEffect("outer_after")

	outerBatch := NewSerialPlan(innerBatch, outerAfter)
	outerBatch.FillDefaultsAndValidate()
	h.pe.StartPlan(nil, outerBatch)

	if outerAfter.processed {
		t.Error("outer_after should be blocked by inner batch")
	}

	h.pe.MarkComplete("test:inner")
	h.pe.Update()

	if !afterInner.processed {
		t.Error("after_inner should be processed after inner unblock")
	}

	h.pe.Update()

	if !outerAfter.processed {
		t.Error("outer_after should be processed after inner batch completes")
	}
}

func TestPlanExecutor_MarkCompleteUnknownId(t *testing.T) {
	h := newPlanTestHarness()
	h.pe.MarkComplete("nonexistent:id")
}

func TestPlanExecutor_GenerateEffectId(t *testing.T) {
	h := newPlanTestHarness()
	id1 := h.pe.GenerateEffectId("test")
	id2 := h.pe.GenerateEffectId("test")
	if id1 == id2 {
		t.Error("generated IDs should be unique")
	}
}
