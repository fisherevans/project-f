package runtime

import (
	"sync"
	"testing"
	"time"

	"fisherevans.com/project/f/internal/game/input"
)

func TestHarnessGameDelta(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(h *Harness)
		natural float64
		want    float64
	}{
		{"running passes natural through", func(h *Harness) {}, 0.016, 0.016},
		{"running applies speed", func(h *Harness) { h.SetTime(false, 2.0) }, 0.01, 0.02},
		{"paused with no budget yields zero", func(h *Harness) { h.SetTime(true, 0) }, 0.016, 0},
		{"paused with budget yields fixed dt", func(h *Harness) {
			h.SetTime(true, 0)
			h.mu.Lock()
			h.stepBudget = 1
			h.fixedDt = 0.05
			h.mu.Unlock()
		}, 0.016, 0.05},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHarness()
			tt.setup(h)
			got := h.gameDelta(tt.natural)
			if got != tt.want {
				t.Fatalf("gameDelta(%v) = %v, want %v", tt.natural, got, tt.want)
			}
		})
	}
}

func TestHarnessGameDeltaConsumesBudget(t *testing.T) {
	h := NewHarness()
	h.SetTime(true, 0)
	h.mu.Lock()
	h.stepBudget = 2
	h.fixedDt = 1.0 / 60.0
	h.mu.Unlock()

	if d := h.gameDelta(99); d != 1.0/60.0 {
		t.Fatalf("first step dt = %v, want %v", d, 1.0/60.0)
	}
	if d := h.gameDelta(99); d != 1.0/60.0 {
		t.Fatalf("second step dt = %v, want %v", d, 1.0/60.0)
	}
	if d := h.gameDelta(99); d != 0 {
		t.Fatalf("third call after budget drained = %v, want 0", d)
	}
}

func TestHarnessMergeInput(t *testing.T) {
	base := input.VirtualState{}
	tests := []struct {
		name   string
		inject func(h *Harness)
		in     input.VirtualState
		want   input.VirtualState
	}{
		{
			name:   "no injection passes through",
			inject: func(h *Harness) {},
			in:     input.VirtualState{A: true},
			want:   input.VirtualState{A: true},
		},
		{
			name:   "injects button",
			inject: func(h *Harness) { h.InjectInput(true, false, false, false, "", 1) },
			in:     base,
			want:   input.VirtualState{A: true},
		},
		{
			name:   "ORs with existing",
			inject: func(h *Harness) { h.InjectInput(false, true, false, false, "", 1) },
			in:     input.VirtualState{A: true},
			want:   input.VirtualState{A: true, B: true},
		},
		{
			name:   "injects direction",
			inject: func(h *Harness) { h.InjectInput(false, false, false, false, "Up", 1) },
			in:     base,
			want:   input.VirtualState{Dir: input.Up},
		},
		{
			name:   "empty direction does not override existing",
			inject: func(h *Harness) { h.InjectInput(true, false, false, false, "", 1) },
			in:     input.VirtualState{Dir: input.Left},
			want:   input.VirtualState{A: true, Dir: input.Left},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHarness()
			tt.inject(h)
			got := h.mergeInput(tt.in)
			if got != tt.want {
				t.Fatalf("mergeInput = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestHarnessInjectWindowDecrements(t *testing.T) {
	h := NewHarness()
	h.InjectInput(true, false, false, false, "", 2)
	if got := h.mergeInput(input.VirtualState{}); !got.A {
		t.Fatal("frame 1: expected A injected")
	}
	if got := h.mergeInput(input.VirtualState{}); !got.A {
		t.Fatal("frame 2: expected A injected")
	}
	if got := h.mergeInput(input.VirtualState{}); got.A {
		t.Fatal("frame 3: injection window should be exhausted")
	}
}

func TestHarnessClearInput(t *testing.T) {
	h := NewHarness()
	h.InjectInput(true, false, false, false, "Up", 10)
	h.ClearInput()
	if got := h.mergeInput(input.VirtualState{}); got.A || got.Dir != input.NotPressed {
		t.Fatalf("ClearInput should cancel injection, got %+v", got)
	}
}

// TestHarnessStepBlocksUntilDrained simulates the game loop draining the step
// budget on a separate goroutine and verifies Step returns only once all frames
// have been consumed.
func TestHarnessStepBlocksUntilDrained(t *testing.T) {
	h := NewHarness()
	h.SetTime(true, 0) // pause first so running-mode frames aren't counted

	var consumed int
	var mu sync.Mutex
	stop := make(chan struct{})
	// Simulated game loop: drain one step per "frame".
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			d := h.gameDelta(0.016)
			if d > 0 {
				mu.Lock()
				consumed++
				mu.Unlock()
			}
			time.Sleep(time.Millisecond)
		}
	}()

	done := make(chan struct{})
	go func() {
		h.Step(5, 1.0/60.0)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		close(stop)
		t.Fatal("Step did not return within 2s")
	}
	close(stop)

	mu.Lock()
	defer mu.Unlock()
	if consumed != 5 {
		t.Fatalf("consumed %d steps, want 5", consumed)
	}
}

// TestHarnessSetTimeReleasesWaitingStep verifies that resuming (SetTime run)
// unblocks a Step that is still waiting on its budget.
func TestHarnessSetTimeReleasesWaitingStep(t *testing.T) {
	h := NewHarness()
	done := make(chan struct{})
	go func() {
		h.Step(1000, 0) // large budget, never drained by a loop here
		close(done)
	}()

	// Give the Step goroutine a moment to start waiting, then resume.
	time.Sleep(20 * time.Millisecond)
	h.SetTime(false, 0)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("SetTime(run) did not release waiting Step")
	}
}

func TestHarnessHealth(t *testing.T) {
	h := NewHarness()
	ready, state, frame := h.Health()
	if ready || state != "" || frame != 0 {
		t.Fatalf("fresh harness Health = (%v,%q,%d), want (false,\"\",0)", ready, state, frame)
	}
	h.markReady()
	if ready, _, _ := h.Health(); !ready {
		t.Fatal("Health should report ready after markReady")
	}
}
