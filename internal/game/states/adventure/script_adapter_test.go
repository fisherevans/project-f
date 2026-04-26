package adventure

import (
	"testing"
)

func TestScriptHandler_ProcessRules_FirstMatchWins(t *testing.T) {
	globals := newTestGlobals(map[string]any{
		"visit_count": 2,
	})

	def := &HandlerDef{
		OnInteractSelf: &HookDef{Rules: []*RuleDef{
			{
				When: &ConditionNode{
					Kind:   "expr",
					Params: "global.visit_count > 1",
				},
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "repeat visitor"},
				},
			},
			{
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "first visit"},
				},
			},
		}},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	output := handler.HandleEvent(
		&NullEntity{id: "npc"},
		globals,
		nil,
		&EventOnInteract{TargetId: "npc", SourceId: "player"},
	)

	if output == nil {
		t.Fatal("expected output, got nil")
	}
	if len(output.Effects) != 1 {
		t.Fatalf("got %d effects, want 1", len(output.Effects))
	}

	batch, ok := output.Effects[0].(*EffectBatch)
	if !ok {
		t.Fatalf("effect = %T, want *EffectBatch", output.Effects[0])
	}
	if len(batch.Effects) != 1 {
		t.Fatalf("batch has %d effects, want 1", len(batch.Effects))
	}
	d, ok := batch.Effects[0].(*EffectDialogue)
	if !ok {
		t.Fatalf("batch effect = %T, want *EffectDialogue", batch.Effects[0])
	}
	if d.Text != "repeat visitor" {
		t.Errorf("text = %q, want %q (first match should win)", d.Text, "repeat visitor")
	}
}

func TestScriptHandler_ProcessRules_NoMatch(t *testing.T) {
	globals := newTestGlobals(map[string]any{})

	def := &HandlerDef{
		OnInteractSelf: &HookDef{Rules: []*RuleDef{
			{
				When: &ConditionNode{
					Kind:   "expr",
					Params: "global.ready == true",
				},
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "ready"},
				},
			},
		}},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	output := handler.HandleEvent(
		&NullEntity{id: "npc"},
		globals,
		nil,
		&EventOnInteract{TargetId: "npc", SourceId: "player"},
	)

	if output != nil {
		t.Error("expected nil output when no rules match")
	}
}

func TestScriptHandler_ProcessRules_SetVar(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: &HookDef{Rules: []*RuleDef{
			{
				Steps: []*StepNode{
					{Kind: "set_var", Params: map[string]any{"key": "talked", "value": "true"}},
					{Kind: "dialogue", Params: "hello"},
				},
			},
		}},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	output := handler.HandleEvent(
		&NullEntity{id: "npc"},
		globals,
		nil,
		&EventOnInteract{TargetId: "npc", SourceId: "player"},
	)

	if output == nil {
		t.Fatal("expected output")
	}

	if len(output.Effects) == 0 {
		t.Fatal("expected effects")
	}
}

func TestScriptHandler_ProcessRules_SetVarUsedInCondition(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: &HookDef{Rules: []*RuleDef{
			{
				When: &ConditionNode{
					Kind:   "expr",
					Params: "var.talked == true",
				},
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "we already talked"},
				},
			},
			{
				Steps: []*StepNode{
					{Kind: "set_var", Params: map[string]any{"key": "talked", "value": "true"}},
					{Kind: "dialogue", Params: "first time"},
				},
			},
		}},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	entity := &NullEntity{id: "npc"}
	event := &EventOnInteract{TargetId: "npc", SourceId: "player"}

	output1 := handler.HandleEvent(entity, globals, nil, event)
	if output1 == nil {
		t.Fatal("first interaction: expected output")
	}
	batch1 := output1.Effects[0].(*EffectBatch)
	d1 := batch1.Effects[1].(*EffectDialogue)
	if d1.Text != "first time" {
		t.Errorf("first interaction: text = %q, want %q", d1.Text, "first time")
	}

	// set_var runs as an effect during plan execution, so output1.State
	// won't contain the mutation yet. Simulate post-execution state.
	output2 := handler.HandleEvent(entity, globals, map[string]any{"talked": true}, event)
	if output2 == nil {
		t.Fatal("second interaction: expected output")
	}
	batch2 := output2.Effects[0].(*EffectBatch)
	d2 := batch2.Effects[0].(*EffectDialogue)
	if d2.Text != "we already talked" {
		t.Errorf("second interaction: text = %q, want %q", d2.Text, "we already talked")
	}
}

func TestScriptHandler_VarInitialization(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		Var: map[string]any{
			"mood":  "neutral",
			"count": 0,
		},
		OnInteractSelf: &HookDef{Rules: []*RuleDef{
			{
				When: &ConditionNode{
					Kind:   "expr",
					Params: "var.mood == 'neutral'",
				},
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "I'm feeling neutral"},
				},
			},
		}},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	output := handler.HandleEvent(
		&NullEntity{id: "npc"},
		globals,
		nil,
		&EventOnInteract{TargetId: "npc", SourceId: "player"},
	)

	if output == nil {
		t.Fatal("expected output (var defaults should set mood=neutral)")
	}
	batch := output.Effects[0].(*EffectBatch)
	d := batch.Effects[0].(*EffectDialogue)
	if d.Text != "I'm feeling neutral" {
		t.Errorf("text = %q, want %q", d.Text, "I'm feeling neutral")
	}
}

func TestScriptHandler_Filter(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnBroadcast: &HookDef{Rules: []*RuleDef{
			{
				Filter: map[string]any{"id": "alert"},
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "alert received"},
				},
			},
		}},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	t.Run("matching broadcast", func(t *testing.T) {
		output := handler.HandleEvent(
			&NullEntity{id: "npc"}, globals, nil,
			&EventBroadcast{Id: "alert"},
		)
		if output == nil {
			t.Fatal("expected output for matching broadcast")
		}
	})

	t.Run("non-matching broadcast", func(t *testing.T) {
		output := handler.HandleEvent(
			&NullEntity{id: "npc"}, globals, nil,
			&EventBroadcast{Id: "other"},
		)
		if output != nil {
			t.Error("expected nil output for non-matching broadcast")
		}
	})
}

func TestScriptHandler_OnInit(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInit: &HookDef{Rules: []*RuleDef{
			{
				Steps: []*StepNode{
					{Kind: "set_world_state", Params: map[string]any{
						"key":   "initialized",
						"value": true,
					}},
				},
			},
		}},
	}

	handler := NewScriptHandler("npc", def, nil)
	output := handler.Init(&NullEntity{id: "npc"}, globals, nil)

	if output == nil {
		t.Fatal("expected output from on_init")
	}
	if len(output.Effects) != 1 {
		t.Fatalf("got %d effects, want 1", len(output.Effects))
	}
}

func TestScriptHandler_EmptyStepsSkipped(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: &HookDef{Rules: []*RuleDef{
			{
				Steps: []*StepNode{},
			},
			{
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "fallback"},
				},
			},
		}},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	output := handler.HandleEvent(
		&NullEntity{id: "npc"}, globals, nil,
		&EventOnInteract{TargetId: "npc", SourceId: "player"},
	)

	if output == nil {
		t.Fatal("expected output (empty steps rule should be skipped)")
	}
	batch := output.Effects[0].(*EffectBatch)
	d := batch.Effects[0].(*EffectDialogue)
	if d.Text != "fallback" {
		t.Errorf("text = %q, want %q (empty-steps rule should skip to next)", d.Text, "fallback")
	}
}

func TestScriptHandler_EmptyStepsRuleSkipped(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: &HookDef{Rules: []*RuleDef{
			{
				Steps: []*StepNode{},
			},
		}},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	output := handler.HandleEvent(
		&NullEntity{id: "npc"}, globals, nil,
		&EventOnInteract{TargetId: "npc", SourceId: "player"},
	)

	if output != nil {
		t.Error("empty-steps rule should produce no output")
	}
}

func TestScriptHandler_UnhandledEventType(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: &HookDef{Rules: []*RuleDef{
			{
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "hello"},
				},
			},
		}},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	output := handler.HandleEvent(
		&NullEntity{id: "npc"}, globals, nil,
		&EventBroadcast{Id: "something"},
	)

	if output != nil {
		t.Error("expected nil output for unhandled event type")
	}
}

func TestScriptHandler_AllMode(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnBroadcast: &HookDef{
			Mode: HookModeAll,
			Rules: []*RuleDef{
				{
					Filter: map[string]any{"id": "alert"},
					Steps: []*StepNode{
						{Kind: "dialogue", Params: "handler A"},
					},
				},
				{
					Filter: map[string]any{"id": "alert"},
					Steps: []*StepNode{
						{Kind: "dialogue", Params: "handler B"},
					},
				},
				{
					Filter: map[string]any{"id": "other"},
					Steps: []*StepNode{
						{Kind: "dialogue", Params: "should not match"},
					},
				},
			},
		},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	output := handler.HandleEvent(
		&NullEntity{id: "npc"}, globals, nil,
		&EventBroadcast{Id: "alert"},
	)

	if output == nil {
		t.Fatal("expected output")
	}
	if len(output.Effects) != 2 {
		t.Fatalf("got %d effects, want 2 (both matching rules)", len(output.Effects))
	}

	batchA := output.Effects[0].(*EffectBatch)
	dA := batchA.Effects[0].(*EffectDialogue)
	if dA.Text != "handler A" {
		t.Errorf("first effect: text = %q, want %q", dA.Text, "handler A")
	}

	batchB := output.Effects[1].(*EffectBatch)
	dB := batchB.Effects[0].(*EffectDialogue)
	if dB.Text != "handler B" {
		t.Errorf("second effect: text = %q, want %q", dB.Text, "handler B")
	}
}

func TestScriptHandler_AllMode_NoMatch(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnBroadcast: &HookDef{
			Mode: HookModeAll,
			Rules: []*RuleDef{
				{
					Filter: map[string]any{"id": "alert"},
					Steps: []*StepNode{
						{Kind: "dialogue", Params: "alert"},
					},
				},
			},
		},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	output := handler.HandleEvent(
		&NullEntity{id: "npc"}, globals, nil,
		&EventBroadcast{Id: "other"},
	)

	if output != nil {
		t.Error("expected nil output when no rules match in all mode")
	}
}
