package adventure

import (
	"testing"

	"fisherevans.com/project/f/internal/game/rpg"
)

type testGlobals struct {
	values map[string]any
}

func newTestGlobals(values map[string]any) *testGlobals {
	return &testGlobals{values: values}
}

func (g *testGlobals) Get(key string) *rpg.GlobalValue {
	v, ok := g.values[key]
	if !ok {
		return rpg.NewTestGlobalValue(key, nil, false)
	}
	return rpg.NewTestGlobalValue(key, v, true)
}

func (g *testGlobals) GetEntityReader(id string) (EntityReader, bool) {
	return nil, false
}

func (g *testGlobals) GetZonesAt(loc MapLocation) Zones {
	return nil
}

func (g *testGlobals) Player() EntityReader {
	return &NullEntity{id: "test_player"}
}

func (g *testGlobals) KeysWithPrefix(prefix string) []string {
	return nil
}

func TestEvaluateCondition_GlobalEq(t *testing.T) {
	globals := newTestGlobals(map[string]any{
		"my_key": true,
	})
	tests := []struct {
		name   string
		cond   *ConditionNode
		expect bool
	}{
		{
			"match_key_value",
			&ConditionNode{Kind: "global_eq", Params: map[string]any{"key": "my_key", "value": true}},
			true,
		},
		{
			"no_match_key_value",
			&ConditionNode{Kind: "global_eq", Params: map[string]any{"key": "my_key", "value": false}},
			false,
		},
		{
			"match_shorthand",
			&ConditionNode{Kind: "global_eq", Params: map[string]any{"my_key": true}},
			true,
		},
		{
			"missing_key",
			&ConditionNode{Kind: "global_eq", Params: map[string]any{"key": "missing", "value": true}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateCondition(tt.cond, globals, nil)
			if got != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, got)
			}
		})
	}
}

func TestEvaluateCondition_NumericComparisons(t *testing.T) {
	globals := newTestGlobals(map[string]any{
		"elythium": 10,
	})
	tests := []struct {
		name   string
		kind   string
		value  int
		expect bool
	}{
		{"gte_pass", "global_gte", 10, true},
		{"gte_fail", "global_gte", 11, false},
		{"gt_fail", "global_gt", 10, false},
		{"gt_pass", "global_gt", 9, true},
		{"lt_pass", "global_lt", 11, true},
		{"lt_fail", "global_lt", 10, false},
		{"lte_pass", "global_lte", 10, true},
		{"lte_fail", "global_lte", 9, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := &ConditionNode{Kind: tt.kind, Params: map[string]any{"elythium": tt.value}}
			got := evaluateCondition(cond, globals, nil)
			if got != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, got)
			}
		})
	}
}

func TestEvaluateCondition_Not(t *testing.T) {
	globals := newTestGlobals(map[string]any{
		"door": true,
	})
	cond := &ConditionNode{
		Kind: "not",
		Params: map[string]any{
			"global_eq": map[string]any{"door": true},
		},
	}
	if evaluateCondition(cond, globals, nil) {
		t.Error("expected false (not of true)")
	}
	cond2 := &ConditionNode{
		Kind: "not",
		Params: map[string]any{
			"global_eq": map[string]any{"door": false},
		},
	}
	if !evaluateCondition(cond2, globals, nil) {
		t.Error("expected true (not of false)")
	}
}

func TestEvaluateCondition_All(t *testing.T) {
	globals := newTestGlobals(map[string]any{
		"a": true,
		"b": true,
	})
	cond := &ConditionNode{
		Kind: "all",
		Params: []any{
			map[string]any{"global_eq": map[string]any{"a": true}},
			map[string]any{"global_eq": map[string]any{"b": true}},
		},
	}
	if !evaluateCondition(cond, globals, nil) {
		t.Error("expected true (all match)")
	}
	cond2 := &ConditionNode{
		Kind: "all",
		Params: []any{
			map[string]any{"global_eq": map[string]any{"a": true}},
			map[string]any{"global_eq": map[string]any{"b": false}},
		},
	}
	if evaluateCondition(cond2, globals, nil) {
		t.Error("expected false (not all match)")
	}
}

func TestEvaluateCondition_HandlerState(t *testing.T) {
	globals := newTestGlobals(nil)
	state := map[string]any{"ready": true}
	cond := &ConditionNode{
		Kind:   "handler_state_eq",
		Params: map[string]any{"ready": true},
	}
	if !evaluateCondition(cond, globals, state) {
		t.Error("expected true")
	}
	cond2 := &ConditionNode{
		Kind:   "handler_state_eq",
		Params: map[string]any{"ready": false},
	}
	if evaluateCondition(cond2, globals, state) {
		t.Error("expected false")
	}
}

func TestEvaluateCondition_Nil(t *testing.T) {
	globals := newTestGlobals(nil)
	if !evaluateCondition(nil, globals, nil) {
		t.Error("nil condition should return true")
	}
}

func TestEvaluateFilter_ZoneActivity(t *testing.T) {
	tc := &TemplateContext{
		SelfId:   "listener",
		PlayerId: "player_0",
	}
	event := &EventEntityZoneActivity{
		EntityId:   "player_0",
		ZoneId:     "my_zone",
		IsEntering: true,
	}
	tests := []struct {
		name   string
		filter map[string]any
		expect bool
	}{
		{"match_all", map[string]any{"zone": "my_zone", "entity": "{{player}}", "entering": true}, true},
		{"wrong_zone", map[string]any{"zone": "other_zone"}, false},
		{"wrong_entity", map[string]any{"entity": "npc_1"}, false},
		{"wrong_entering", map[string]any{"entering": false}, false},
		{"empty_filter", map[string]any{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateFilter(tt.filter, event, tc)
			if got != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, got)
			}
		})
	}
}

func TestEvaluateFilter_Broadcast(t *testing.T) {
	tc := &TemplateContext{SelfId: "listener"}
	event := &EventBroadcast{Id: "combat.trigger"}
	if !evaluateFilter(map[string]any{"id": "combat.trigger"}, event, tc) {
		t.Error("expected match")
	}
	if evaluateFilter(map[string]any{"id": "other"}, event, tc) {
		t.Error("expected no match")
	}
}
