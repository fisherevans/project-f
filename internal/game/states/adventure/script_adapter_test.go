package adventure

import (
	"testing"
)

func TestScriptHandler_ProcessRules_FirstMatchWins(t *testing.T) {
	globals := newTestGlobals(map[string]any{
		"visit_count": 2,
	})

	def := &HandlerDef{
		OnInteractSelf: []*RuleDef{
			{
				When: &ConditionNode{
					Kind:   "global_gt",
					Params: map[string]any{"visit_count": 1},
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
		},
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

	// The effect should be a serial plan containing the first matching rule's dialogue
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
		OnInteractSelf: []*RuleDef{
			{
				When: &ConditionNode{
					Kind:   "global_eq",
					Params: map[string]any{"ready": true},
				},
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "ready"},
				},
			},
		},
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

func TestScriptHandler_ProcessRules_SetState(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: []*RuleDef{
			{
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "hello"},
				},
				SetState: map[string]any{
					"talked": true,
				},
			},
		},
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

	// Verify handler state was mutated
	state, ok := output.State.(map[string]any)
	if !ok {
		t.Fatalf("state = %T, want map[string]any", output.State)
	}
	if state["talked"] != true {
		t.Errorf("talked = %v, want true", state["talked"])
	}
}

func TestScriptHandler_ProcessRules_SetStateUsedInCondition(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: []*RuleDef{
			{
				When: &ConditionNode{
					Kind:   "handler_state_eq",
					Params: map[string]any{"talked": true},
				},
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "we already talked"},
				},
			},
			{
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "first time"},
				},
				SetState: map[string]any{"talked": true},
			},
		},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	entity := &NullEntity{id: "npc"}
	event := &EventOnInteract{TargetId: "npc", SourceId: "player"}

	// First interaction
	output1 := handler.HandleEvent(entity, globals, nil, event)
	if output1 == nil {
		t.Fatal("first interaction: expected output")
	}
	batch1 := output1.Effects[0].(*EffectBatch)
	d1 := batch1.Effects[0].(*EffectDialogue)
	if d1.Text != "first time" {
		t.Errorf("first interaction: text = %q, want %q", d1.Text, "first time")
	}

	// Second interaction (pass state back)
	output2 := handler.HandleEvent(entity, globals, output1.State, event)
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
		OnInteractSelf: []*RuleDef{
			{
				When: &ConditionNode{
					Kind:   "handler_state_eq",
					Params: map[string]any{"mood": "neutral"},
				},
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "I'm feeling neutral"},
				},
			},
		},
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
		OnBroadcast: []*RuleDef{
			{
				Filter: map[string]any{"id": "alert"},
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "alert received"},
				},
			},
		},
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
		OnInit: []*RuleDef{
			{
				Steps: []*StepNode{
					{Kind: "set_world_state", Params: map[string]any{
						"key":   "initialized",
						"value": true,
					}},
				},
			},
		},
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
		OnInteractSelf: []*RuleDef{
			{
				Steps: []*StepNode{},
			},
			{
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "fallback"},
				},
			},
		},
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

func TestScriptHandler_SetStateOnlyRule(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: []*RuleDef{
			{
				Steps:    []*StepNode{},
				SetState: map[string]any{"seen": true},
			},
		},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	output := handler.HandleEvent(
		&NullEntity{id: "npc"}, globals, nil,
		&EventOnInteract{TargetId: "npc", SourceId: "player"},
	)

	if output == nil {
		t.Fatal("expected output (set_state only rule should still produce output)")
	}
	state := output.State.(map[string]any)
	if state["seen"] != true {
		t.Error("seen should be true")
	}
}

func TestScriptHandler_UnhandledEventType(t *testing.T) {
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: []*RuleDef{
			{
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "hello"},
				},
			},
		},
	}

	handler := NewScriptHandler("npc", def, nil)
	handler.Init(&NullEntity{id: "npc"}, globals, nil)

	// Send an event type that this handler doesn't handle
	output := handler.HandleEvent(
		&NullEntity{id: "npc"}, globals, nil,
		&EventBroadcast{Id: "something"},
	)

	if output != nil {
		t.Error("expected nil output for unhandled event type")
	}
}
