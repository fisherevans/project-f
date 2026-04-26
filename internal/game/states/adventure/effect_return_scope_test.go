package adventure

import (
	"testing"
)

func TestReturnScope_IsReturned(t *testing.T) {
	t.Run("nil scope", func(t *testing.T) {
		var s *ReturnScope
		if s.isReturned() {
			t.Error("nil scope should not be returned")
		}
	})

	t.Run("fresh scope", func(t *testing.T) {
		s := &ReturnScope{Id: "test"}
		if s.isReturned() {
			t.Error("fresh scope should not be returned")
		}
	})

	t.Run("returned scope", func(t *testing.T) {
		s := &ReturnScope{Id: "test", returned: true}
		if !s.isReturned() {
			t.Error("returned scope should be returned")
		}
	})
}

func TestEffectReturn_CompletionID(t *testing.T) {
	e := newReturnEffect(nil)
	if e.CompletionID() == "" {
		t.Error("return effect should have a completionID")
	}
}

func TestReturn_AtHandlerLevel(t *testing.T) {
	h := newPlanTestHarness()
	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: []*RuleDef{
			{
				Steps: []*StepNode{
					{Kind: "dialogue", Params: "before return"},
					{Kind: "return", Params: true},
					{Kind: "dialogue", Params: "after return"},
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
		t.Fatal("expected output")
	}

	// Dispatch through the plan executor
	for _, e := range output.Effects {
		e.FillDefaultsAndValidate()
		h.pe.StartPlan(nil, e.(*EffectBatch))
	}

	// "before return" dialogue should be dispatched (it has a completionID, so it blocks)
	labels := h.dispatchedLabels()
	foundDialogue := false
	for _, d := range h.dispatched {
		if de, ok := d.Effect.(*EffectDialogue); ok && de.Text == "before return" {
			foundDialogue = true
		}
	}
	if !foundDialogue {
		t.Errorf("dialogue 'before return' should be dispatched, got labels: %v", labels)
	}

	// Complete the dialogue to continue
	for _, d := range h.dispatched {
		if cid := d.Effect.CompletionID(); cid != "" {
			h.pe.MarkComplete(cid)
		}
	}
	h.pe.Update()

	// Return effect should fire and abort the scope.
	// "after return" should NOT be dispatched.
	for _, d := range h.dispatched {
		if de, ok := d.Effect.(*EffectDialogue); ok && de.Text == "after return" {
			t.Error("dialogue 'after return' should NOT be dispatched")
		}
	}
}

func TestReturn_InCustomAction_CallerContinues(t *testing.T) {
	h := newPlanTestHarness()

	oldActions := scriptCustomActions
	scriptCustomActions = map[string]*CustomActionDef{
		"check": {
			Steps: []*StepNode{
				{Kind: "dialogue", Params: "inside action"},
				{Kind: "return", Params: true},
				{Kind: "dialogue", Params: "should not show"},
			},
		},
	}
	defer func() { scriptCustomActions = oldActions }()

	globals := newTestGlobals(nil)

	def := &HandlerDef{
		OnInteractSelf: []*RuleDef{
			{
				Steps: []*StepNode{
					{Kind: "custom_action", Params: map[string]any{"name": "check"}},
					{Kind: "dialogue", Params: "after action"},
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
		t.Fatal("expected output")
	}

	for _, e := range output.Effects {
		e.FillDefaultsAndValidate()
		h.pe.StartPlan(nil, e.(*EffectBatch))
	}

	// Walk through execution: complete each blocking effect until done
	for iterations := 0; iterations < 20; iterations++ {
		var toComplete []string
		for _, d := range h.dispatched {
			if cid := d.Effect.CompletionID(); cid != "" {
				toComplete = append(toComplete, cid)
			}
		}
		if len(toComplete) == 0 {
			break
		}
		h.dispatched = nil
		for _, cid := range toComplete {
			h.pe.MarkComplete(cid)
		}
		h.pe.Update()
	}

	// Collect all dispatched dialogue texts
	var dialogues []string
	for _, d := range h.dispatched {
		if de, ok := d.Effect.(*EffectDialogue); ok {
			dialogues = append(dialogues, de.Text)
		}
	}

	// "inside action" should appear, "should not show" should not, "after action" should appear
	found := map[string]bool{}
	// Re-scan all dispatches from scratch (dispatched slice was reset during iteration)
	// Let's use a different approach - re-run and collect everything

	// Actually, let's reset and do a cleaner trace
	t.Log("Re-running with full trace")

	h2 := newPlanTestHarness()
	handler2 := NewScriptHandler("npc2", def, nil)
	handler2.Init(&NullEntity{id: "npc2"}, globals, nil)

	output2 := handler2.HandleEvent(
		&NullEntity{id: "npc2"}, globals, nil,
		&EventOnInteract{TargetId: "npc2", SourceId: "player"},
	)

	for _, e := range output2.Effects {
		e.FillDefaultsAndValidate()
		h2.pe.StartPlan(nil, e.(*EffectBatch))
	}

	allDialogues := collectDialogues(h2)

	for _, text := range allDialogues {
		found[text] = true
	}

	if !found["inside action"] {
		t.Error("'inside action' should be dispatched")
	}
	if found["should not show"] {
		t.Error("'should not show' should NOT be dispatched")
	}
	if !found["after action"] {
		t.Error("'after action' should be dispatched (caller continues after custom action return)")
	}
}

func TestReturn_InIfBranch_InsideCustomAction(t *testing.T) {
	h := newPlanTestHarness()

	oldActions := scriptCustomActions
	scriptCustomActions = map[string]*CustomActionDef{
		"check_door": {
			Steps: []*StepNode{
				{Kind: "if", Params: map[string]any{
					"when": "var.door_open == true",
					"then": []any{
						map[string]any{"dialogue": "already open"},
						map[string]any{"return": true},
					},
				}},
				{Kind: "dialogue", Params: "door is locked"},
			},
		},
	}
	defer func() { scriptCustomActions = oldActions }()

	globals := newTestGlobals(nil)
	state := map[string]any{"door_open": true}

	def := &HandlerDef{
		Var: map[string]any{"door_open": true},
		OnInteractSelf: []*RuleDef{
			{
				Steps: []*StepNode{
					{Kind: "custom_action", Params: map[string]any{"name": "check_door"}},
					{Kind: "dialogue", Params: "after check"},
				},
			},
		},
	}

	handler := NewScriptHandler("door", def, nil)
	handler.Init(&NullEntity{id: "door"}, globals, state)

	output := handler.HandleEvent(
		&NullEntity{id: "door"}, globals, state,
		&EventOnInteract{TargetId: "door", SourceId: "player"},
	)

	if output == nil {
		t.Fatal("expected output")
	}

	for _, e := range output.Effects {
		e.FillDefaultsAndValidate()
		h.pe.StartPlan(nil, e.(*EffectBatch))
	}

	found := collectDialogues(h)

	has := func(text string) bool {
		for _, d := range found {
			if d == text {
				return true
			}
		}
		return false
	}

	if !has("already open") {
		t.Error("'already open' should be dispatched (if-then branch)")
	}
	if has("door is locked") {
		t.Error("'door is locked' should NOT be dispatched (return in if-then skips rest of custom action)")
	}
	if !has("after check") {
		t.Error("'after check' should be dispatched (caller continues after custom action)")
	}
}

func TestReturn_InWhileLoop(t *testing.T) {
	state := map[string]any{"count": 0}
	tc := newTestTC(nil, state)

	scope := &ReturnScope{Id: "test-while-scope"}
	tc.returnScope = scope

	effects := convertWhileStep(map[string]any{
		"when": "true",
		"max":  100,
		"steps": []any{
			map[string]any{"set_var": map[string]any{
				"key":   "count",
				"value": "int(var.count) + 1",
			}},
			map[string]any{"return": true},
			map[string]any{"dialogue": "should not appear"},
		},
	}, tc, nil)

	deferred := effects[0].(*EffectDeferredBatch)
	children := deferred.BuildEffects(nil, nil)

	// The while should run one iteration: set_var (eager, increments count),
	// then return (eagerly marks scope returned), then breaks.
	// "dialogue" after return should not be in children.
	count := state["count"]
	if count != 1 {
		t.Errorf("count = %v, want 1 (one iteration before return)", count)
	}

	if !scope.returned {
		t.Error("scope should be marked as returned")
	}

	// Children: set_var (EffectFunction) + return (EffectReturn).
	// The return is included so the PlanExecutor can dispatch it for AbortScope.
	if len(children) != 2 {
		t.Errorf("got %d children, want 2 (set_var + return)", len(children))
	}
}

func TestReturn_NilScope(t *testing.T) {
	h := newPlanTestHarness()

	// A return with nil scope should not panic
	e := newReturnEffect(nil)
	e.FillDefaultsAndValidate()
	e.Process(nil, h.state)
}

// collectDialogues drives the plan executor through all blocking effects
// and returns all dialogue texts that were dispatched.
func collectDialogues(h *planTestHarness) []string {
	var allDialogues []string

	for iterations := 0; iterations < 50; iterations++ {
		for _, d := range h.dispatched {
			if de, ok := d.Effect.(*EffectDialogue); ok {
				allDialogues = append(allDialogues, de.Text)
			}
		}

		var toComplete []string
		for _, d := range h.dispatched {
			if _, ok := d.Effect.(*EffectDialogue); ok {
				if cid := d.Effect.CompletionID(); cid != "" {
					toComplete = append(toComplete, cid)
				}
			}
		}
		h.dispatched = nil

		for _, cid := range toComplete {
			h.pe.MarkComplete(cid)
		}
		h.pe.Update()

		if len(h.dispatched) == 0 {
			break
		}
	}

	// Collect any remaining
	for _, d := range h.dispatched {
		if de, ok := d.Effect.(*EffectDialogue); ok {
			allDialogues = append(allDialogues, de.Text)
		}
	}

	return allDialogues
}
