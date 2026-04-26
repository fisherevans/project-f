package adventure

import (
	"testing"
)

func newTestTC(globals StateGlobalsReader, handlerState map[string]any) *TemplateContext {
	if globals == nil {
		globals = newTestGlobals(nil)
	}
	if handlerState == nil {
		handlerState = make(map[string]any)
	}
	return &TemplateContext{
		SelfId:       "test_self",
		PlayerId:     "test_player",
		Globals:      globals,
		handlerState: handlerState,
	}
}

func TestConvertIfStep_TrueBranch(t *testing.T) {
	tc := newTestTC(nil, map[string]any{"ready": true})

	effects := convertIfStep(map[string]any{
		"when": "var.ready == true",
		"then": []any{
			map[string]any{"dialogue": "yes"},
		},
		"else": []any{
			map[string]any{"dialogue": "no"},
		},
	}, tc, nil)

	if len(effects) != 1 {
		t.Fatalf("got %d effects, want 1 (deferred batch)", len(effects))
	}

	deferred, ok := effects[0].(*EffectDeferredBatch)
	if !ok {
		t.Fatalf("effect = %T, want *EffectDeferredBatch", effects[0])
	}

	children := deferred.BuildEffects(nil, nil)
	if len(children) != 1 {
		t.Fatalf("got %d children, want 1", len(children))
	}
	d, ok := children[0].(*EffectDialogue)
	if !ok {
		t.Fatalf("child = %T, want *EffectDialogue", children[0])
	}
	if d.Text != "yes" {
		t.Errorf("text = %q, want %q", d.Text, "yes")
	}
}

func TestConvertIfStep_FalseBranch(t *testing.T) {
	tc := newTestTC(nil, map[string]any{"ready": false})

	effects := convertIfStep(map[string]any{
		"when": "var.ready == true",
		"then": []any{
			map[string]any{"dialogue": "yes"},
		},
		"else": []any{
			map[string]any{"dialogue": "no"},
		},
	}, tc, nil)

	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)
	if len(children) != 1 {
		t.Fatalf("got %d children, want 1", len(children))
	}
	d := children[0].(*EffectDialogue)
	if d.Text != "no" {
		t.Errorf("text = %q, want %q", d.Text, "no")
	}
}

func TestConvertIfStep_NoElse(t *testing.T) {
	tc := newTestTC(nil, map[string]any{"ready": false})

	effects := convertIfStep(map[string]any{
		"when": "var.ready == true",
		"then": []any{
			map[string]any{"dialogue": "yes"},
		},
	}, tc, nil)

	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)
	if len(children) != 0 {
		t.Errorf("got %d children, want 0 (no else branch)", len(children))
	}
}

func TestConvertIfStep_InvalidParams(t *testing.T) {
	tc := newTestTC(nil, nil)

	if effects := convertIfStep("not a map", tc, nil); effects != nil {
		t.Error("non-map params should return nil")
	}

	if effects := convertIfStep(map[string]any{}, tc, nil); effects != nil {
		t.Error("missing 'when' should return nil")
	}
}

func TestConvertSwitchStep_MatchingCase(t *testing.T) {
	tc := newTestTC(nil, map[string]any{"mode": "open"})

	effects := convertSwitchStep(map[string]any{
		"on": "str(var.mode)",
		"cases": []any{
			map[string]any{
				"value": "'closed'",
				"steps": []any{map[string]any{"dialogue": "it's closed"}},
			},
			map[string]any{
				"value": "'open'",
				"steps": []any{map[string]any{"dialogue": "it's open"}},
			},
		},
		"default": []any{
			map[string]any{"dialogue": "unknown state"},
		},
	}, tc, nil)

	if len(effects) != 1 {
		t.Fatalf("got %d effects, want 1", len(effects))
	}
	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)
	if len(children) != 1 {
		t.Fatalf("got %d children, want 1", len(children))
	}
	d := children[0].(*EffectDialogue)
	if d.Text != "it's open" {
		t.Errorf("text = %q, want %q", d.Text, "it's open")
	}
}

func TestConvertSwitchStep_DefaultCase(t *testing.T) {
	tc := newTestTC(nil, map[string]any{"mode": "unknown_value"})

	effects := convertSwitchStep(map[string]any{
		"on": "str(var.mode)",
		"cases": []any{
			map[string]any{
				"value": "'open'",
				"steps": []any{map[string]any{"dialogue": "open"}},
			},
		},
		"default": []any{
			map[string]any{"dialogue": "fallback"},
		},
	}, tc, nil)

	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)
	if len(children) != 1 {
		t.Fatalf("got %d children, want 1", len(children))
	}
	d := children[0].(*EffectDialogue)
	if d.Text != "fallback" {
		t.Errorf("text = %q, want %q", d.Text, "fallback")
	}
}

func TestConvertSwitchStep_NoMatch_NoDefault(t *testing.T) {
	tc := newTestTC(nil, map[string]any{"mode": "other"})

	effects := convertSwitchStep(map[string]any{
		"on": "str(var.mode)",
		"cases": []any{
			map[string]any{
				"value": "'open'",
				"steps": []any{map[string]any{"dialogue": "open"}},
			},
		},
	}, tc, nil)

	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)
	if len(children) != 0 {
		t.Errorf("got %d children, want 0 (no match, no default)", len(children))
	}
}

func TestConvertSwitchStep_InvalidParams(t *testing.T) {
	tc := newTestTC(nil, nil)

	if effects := convertSwitchStep("not a map", tc, nil); effects != nil {
		t.Error("non-map params should return nil")
	}
	if effects := convertSwitchStep(map[string]any{}, tc, nil); effects != nil {
		t.Error("missing 'on' should return nil")
	}
}

func TestConvertWhileStep_BasicLoop(t *testing.T) {
	state := map[string]any{"count": 0}
	tc := newTestTC(nil, state)

	effects := convertWhileStep(map[string]any{
		"when": "int(var.count) < 3",
		"steps": []any{
			map[string]any{"set_var": map[string]any{
				"key":   "count",
				"value": "int(var.count) + 1",
			}},
		},
	}, tc, nil)

	if len(effects) != 1 {
		t.Fatalf("got %d effects, want 1", len(effects))
	}
	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)

	// set_var is an EffectFunction, processed eagerly in while loop.
	// After 3 iterations, count should be 3.
	count, ok := state["count"]
	if !ok {
		t.Fatal("count not set")
	}
	if count != 3 {
		t.Errorf("count = %v, want 3", count)
	}
	// Each iteration produces one EffectFunction (set_var)
	if len(children) != 3 {
		t.Errorf("got %d children, want 3", len(children))
	}
}

func TestConvertWhileStep_MaxIterations(t *testing.T) {
	state := map[string]any{"count": 0}
	tc := newTestTC(nil, state)

	effects := convertWhileStep(map[string]any{
		"when": "true",
		"max":  5,
		"steps": []any{
			map[string]any{"set_var": map[string]any{
				"key":   "count",
				"value": "int(var.count) + 1",
			}},
		},
	}, tc, nil)

	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)

	count := state["count"]
	if count != 5 {
		t.Errorf("count = %v, want 5 (capped by max)", count)
	}
	if len(children) != 5 {
		t.Errorf("got %d children, want 5", len(children))
	}
}

func TestConvertWhileStep_ConditionFalseImmediately(t *testing.T) {
	tc := newTestTC(nil, map[string]any{"done": true})

	effects := convertWhileStep(map[string]any{
		"when": "var.done == false",
		"steps": []any{
			map[string]any{"dialogue": "should not appear"},
		},
	}, tc, nil)

	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)
	if len(children) != 0 {
		t.Errorf("got %d children, want 0 (condition false immediately)", len(children))
	}
}

func TestConvertWhileStep_InvalidParams(t *testing.T) {
	tc := newTestTC(nil, nil)

	if effects := convertWhileStep("not a map", tc, nil); effects != nil {
		t.Error("non-map params should return nil")
	}
	if effects := convertWhileStep(map[string]any{}, tc, nil); effects != nil {
		t.Error("missing 'when' should return nil")
	}
}

func TestConvertWhileStep_WithBlockingEffects(t *testing.T) {
	state := map[string]any{"count": 0}
	tc := newTestTC(nil, state)

	effects := convertWhileStep(map[string]any{
		"when": "int(var.count) < 2",
		"steps": []any{
			map[string]any{"dialogue": "iteration message"},
			map[string]any{"set_var": map[string]any{
				"key":   "count",
				"value": "int(var.count) + 1",
			}},
		},
	}, tc, nil)

	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)

	// 2 iterations, each with dialogue + set_var = 4 effects
	if len(children) != 4 {
		t.Errorf("got %d children, want 4", len(children))
	}

	// Verify types alternate: dialogue, set_var, dialogue, set_var
	for i, child := range children {
		if i%2 == 0 {
			if _, ok := child.(*EffectDialogue); !ok {
				t.Errorf("child[%d] = %T, want *EffectDialogue", i, child)
			}
		} else {
			if _, ok := child.(*EffectFunction); !ok {
				t.Errorf("child[%d] = %T, want *EffectFunction", i, child)
			}
		}
	}
}

func TestExtractSubSteps(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  int
	}{
		{"nil", nil, 0},
		{"not a list", "hello", 0},
		{"empty list", []any{}, 0},
		{"one step", []any{map[string]any{"dialogue": "hi"}}, 1},
		{"two steps", []any{
			map[string]any{"dialogue": "a"},
			map[string]any{"timer": 1.0},
		}, 2},
		{"non-map item skipped", []any{
			"not a map",
			map[string]any{"dialogue": "b"},
		}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := extractSubSteps(tt.input)
			if len(steps) != tt.want {
				t.Errorf("got %d steps, want %d", len(steps), tt.want)
			}
		})
	}
}
