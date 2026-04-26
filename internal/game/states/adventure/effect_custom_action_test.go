package adventure

import (
	"testing"
)

func TestConvertCustomActionStep_BasicCall(t *testing.T) {
	oldActions := scriptCustomActions
	scriptCustomActions = map[string]*CustomActionDef{
		"greet": {
			Description: "Say hello",
			Steps: []*StepNode{
				{Kind: "dialogue", Params: "Hello!"},
			},
		},
	}
	defer func() { scriptCustomActions = oldActions }()

	tc := newTestTC(nil, nil)

	effects := convertCustomActionStep(map[string]any{
		"name": "greet",
	}, tc, nil)

	if len(effects) != 1 {
		t.Fatalf("got %d effects, want 1", len(effects))
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
	if d.Text != "Hello!" {
		t.Errorf("text = %q, want %q", d.Text, "Hello!")
	}
}

func TestConvertCustomActionStep_StringShorthand(t *testing.T) {
	oldActions := scriptCustomActions
	scriptCustomActions = map[string]*CustomActionDef{
		"greet": {
			Steps: []*StepNode{
				{Kind: "dialogue", Params: "Hello!"},
			},
		},
	}
	defer func() { scriptCustomActions = oldActions }()

	tc := newTestTC(nil, nil)

	effects := convertCustomActionStep("greet", tc, nil)
	if len(effects) != 1 {
		t.Fatalf("got %d effects, want 1", len(effects))
	}
	_, ok := effects[0].(*EffectDeferredBatch)
	if !ok {
		t.Fatalf("effect = %T, want *EffectDeferredBatch", effects[0])
	}
}

func TestConvertCustomActionStep_WithParams(t *testing.T) {
	oldActions := scriptCustomActions
	scriptCustomActions = map[string]*CustomActionDef{
		"greet_named": {
			Params: []*CustomActionParam{
				{Name: "name", Description: "who to greet", Default: "stranger"},
			},
			Steps: []*StepNode{
				{Kind: "dialogue", Params: "Hello, {{param.name}}!"},
			},
		},
	}
	defer func() { scriptCustomActions = oldActions }()

	tc := newTestTC(nil, nil)

	t.Run("explicit param", func(t *testing.T) {
		effects := convertCustomActionStep(map[string]any{
			"name": "greet_named",
			"with": map[string]any{
				"name": "'World'",
			},
		}, tc, nil)

		deferred := effects[0].(*EffectDeferredBatch)
		children := deferred.BuildEffects(nil, nil)
		if len(children) != 1 {
			t.Fatalf("got %d children, want 1", len(children))
		}
	})

	t.Run("default param", func(t *testing.T) {
		effects := convertCustomActionStep(map[string]any{
			"name": "greet_named",
		}, tc, nil)

		deferred := effects[0].(*EffectDeferredBatch)
		children := deferred.BuildEffects(nil, nil)
		if len(children) != 1 {
			t.Fatalf("got %d children, want 1", len(children))
		}
	})
}

func TestConvertCustomActionStep_DepthLimit(t *testing.T) {
	oldActions := scriptCustomActions
	scriptCustomActions = map[string]*CustomActionDef{
		"recurse": {
			Steps: []*StepNode{
				{Kind: "custom_action", Params: map[string]any{"name": "recurse"}},
			},
		},
	}
	defer func() { scriptCustomActions = oldActions }()

	tc := newTestTC(nil, nil)

	effects := convertCustomActionStep(map[string]any{
		"name": "recurse",
	}, tc, nil)

	// First level should produce a deferred batch
	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)
	if len(children) != 1 {
		t.Fatalf("got %d children at level 1, want 1", len(children))
	}

	// Recurse through levels until depth limit
	current := children
	depth := 1
	for depth < maxCustomActionDepth+5 {
		if len(current) == 0 {
			break
		}
		inner, ok := current[0].(*EffectDeferredBatch)
		if !ok {
			break
		}
		current = inner.BuildEffects(nil, nil)
		depth++
	}

	if depth > maxCustomActionDepth+1 {
		t.Errorf("recursion went to depth %d, should have stopped at %d", depth, maxCustomActionDepth)
	}
}

func TestConvertCustomActionStep_UnknownAction(t *testing.T) {
	oldActions := scriptCustomActions
	scriptCustomActions = map[string]*CustomActionDef{}
	defer func() { scriptCustomActions = oldActions }()

	tc := newTestTC(nil, nil)
	effects := convertCustomActionStep(map[string]any{
		"name": "nonexistent",
	}, tc, nil)

	if effects != nil {
		t.Errorf("unknown action should return nil, got %d effects", len(effects))
	}
}

func TestConvertCustomActionStep_MissingName(t *testing.T) {
	tc := newTestTC(nil, nil)

	if effects := convertCustomActionStep(map[string]any{}, tc, nil); effects != nil {
		t.Error("missing name should return nil")
	}
	if effects := convertCustomActionStep(42, tc, nil); effects != nil {
		t.Error("non-string/map params should return nil")
	}
}

func TestConvertCustomActionStep_SharedHandlerState(t *testing.T) {
	state := map[string]any{"count": 0}
	oldActions := scriptCustomActions
	scriptCustomActions = map[string]*CustomActionDef{
		"increment": {
			Steps: []*StepNode{
				{Kind: "set_var", Params: map[string]any{
					"key":   "count",
					"value": "int(var.count) + 1",
				}},
			},
		},
	}
	defer func() { scriptCustomActions = oldActions }()

	tc := newTestTC(nil, state)

	effects := convertCustomActionStep(map[string]any{
		"name": "increment",
	}, tc, nil)

	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)

	if len(children) != 1 {
		t.Fatalf("got %d children, want 1", len(children))
	}

	fn, ok := children[0].(*EffectFunction)
	if !ok {
		t.Fatalf("child = %T, want *EffectFunction", children[0])
	}

	fn.Process(nil, nil)

	if state["count"] != 1 {
		t.Errorf("count = %v, want 1 (custom action should share handler state)", state["count"])
	}
}

func TestConvertCustomActionStep_NestedCustomAction(t *testing.T) {
	oldActions := scriptCustomActions
	scriptCustomActions = map[string]*CustomActionDef{
		"inner": {
			Steps: []*StepNode{
				{Kind: "dialogue", Params: "from inner"},
			},
		},
		"outer": {
			Steps: []*StepNode{
				{Kind: "dialogue", Params: "before inner"},
				{Kind: "custom_action", Params: map[string]any{"name": "inner"}},
				{Kind: "dialogue", Params: "after inner"},
			},
		},
	}
	defer func() { scriptCustomActions = oldActions }()

	tc := newTestTC(nil, nil)

	effects := convertCustomActionStep(map[string]any{
		"name": "outer",
	}, tc, nil)

	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)

	if len(children) != 3 {
		t.Fatalf("got %d children, want 3 (dialogue, custom_action deferred, dialogue)", len(children))
	}

	// First child: dialogue "before inner"
	d1, ok := children[0].(*EffectDialogue)
	if !ok {
		t.Fatalf("child[0] = %T, want *EffectDialogue", children[0])
	}
	if d1.Text != "before inner" {
		t.Errorf("child[0] text = %q", d1.Text)
	}

	// Second child: deferred batch for inner custom action
	innerDeferred, ok := children[1].(*EffectDeferredBatch)
	if !ok {
		t.Fatalf("child[1] = %T, want *EffectDeferredBatch", children[1])
	}
	innerChildren := innerDeferred.BuildEffects(nil, nil)
	if len(innerChildren) != 1 {
		t.Fatalf("inner got %d children, want 1", len(innerChildren))
	}
	innerD, ok := innerChildren[0].(*EffectDialogue)
	if !ok {
		t.Fatalf("inner child = %T, want *EffectDialogue", innerChildren[0])
	}
	if innerD.Text != "from inner" {
		t.Errorf("inner text = %q", innerD.Text)
	}

	// Third child: dialogue "after inner"
	d3, ok := children[2].(*EffectDialogue)
	if !ok {
		t.Fatalf("child[2] = %T, want *EffectDialogue", children[2])
	}
	if d3.Text != "after inner" {
		t.Errorf("child[2] text = %q", d3.Text)
	}
}
