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
	var keys []string
	for k := range g.values {
		if prefix == "" || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k)
		}
	}
	return keys
}

func TestEvaluateCondition_Expr(t *testing.T) {
	globals := newTestGlobals(map[string]any{
		"my_key":   true,
		"elythium": 10,
	})
	tc := &TemplateContext{
		SelfId:  "npc",
		Globals: globals,
	}

	tests := []struct {
		name   string
		expr   string
		expect bool
	}{
		{"global eq true", "global.my_key == true", true},
		{"global eq false", "global.my_key == false", false},
		{"global ne", "global.my_key != false", true},
		{"global gt pass", "global.elythium > 9", true},
		{"global gt fail", "global.elythium > 10", false},
		{"global gte pass", "global.elythium >= 10", true},
		{"global gte fail", "global.elythium >= 11", false},
		{"global lt pass", "global.elythium < 11", true},
		{"global lt fail", "global.elythium < 10", false},
		{"global lte pass", "global.elythium <= 10", true},
		{"global lte fail", "global.elythium <= 9", false},
		{"hasKey exists", "hasKey(global, 'my_key')", true},
		{"hasKey not exists", "hasKey(global, 'missing')", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := &ConditionNode{Kind: "expr", Params: tt.expr}
			got := evaluateCondition(cond, globals, nil, tc)
			if got != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, got)
			}
		})
	}
}

func TestEvaluateCondition_ExprHandlerState(t *testing.T) {
	globals := newTestGlobals(nil)
	state := map[string]any{"ready": true, "mood": "neutral"}
	tc := &TemplateContext{
		SelfId:       "npc",
		Globals:      globals,
		handlerState: state,
	}

	tests := []struct {
		name   string
		expr   string
		expect bool
	}{
		{"var eq true", "var.ready == true", true},
		{"var eq false", "var.ready == false", false},
		{"var ne", "var.ready != true", false},
		{"var string eq", "var.mood == 'neutral'", true},
		{"var string ne", "var.mood == 'happy'", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := &ConditionNode{Kind: "expr", Params: tt.expr}
			got := evaluateCondition(cond, globals, state, tc)
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
	tc := &TemplateContext{SelfId: "npc", Globals: globals}

	cond := &ConditionNode{
		Kind: "not",
		Params: map[string]any{
			"expr": "global.door == true",
		},
	}
	if evaluateCondition(cond, globals, nil, tc) {
		t.Error("expected false (not of true)")
	}
	cond2 := &ConditionNode{
		Kind: "not",
		Params: map[string]any{
			"expr": "global.door == false",
		},
	}
	if !evaluateCondition(cond2, globals, nil, tc) {
		t.Error("expected true (not of false)")
	}
}

func TestEvaluateCondition_All(t *testing.T) {
	globals := newTestGlobals(map[string]any{
		"a": true,
		"b": true,
	})
	tc := &TemplateContext{SelfId: "npc", Globals: globals}

	cond := &ConditionNode{
		Kind: "all",
		Params: []any{
			map[string]any{"expr": "global.a == true"},
			map[string]any{"expr": "global.b == true"},
		},
	}
	if !evaluateCondition(cond, globals, nil, tc) {
		t.Error("expected true (all match)")
	}
	cond2 := &ConditionNode{
		Kind: "all",
		Params: []any{
			map[string]any{"expr": "global.a == true"},
			map[string]any{"expr": "global.b == false"},
		},
	}
	if evaluateCondition(cond2, globals, nil, tc) {
		t.Error("expected false (not all match)")
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

type testEntityGlobals struct {
	testGlobals
	entities map[string]EntityReader
}

func newTestEntityGlobals(values map[string]any, entities map[string]EntityReader) *testEntityGlobals {
	return &testEntityGlobals{
		testGlobals: testGlobals{values: values},
		entities:    entities,
	}
}

func (g *testEntityGlobals) GetEntityReader(id string) (EntityReader, bool) {
	e, ok := g.entities[id]
	return e, ok
}

type testEntity struct {
	NullEntity
	moving           bool
	hasPushedBehav   bool
	behaviorDisabled bool
}

func (e *testEntity) IsMoving() bool         { return e.moving }
func (e *testEntity) HasPushedBehavior() bool { return e.hasPushedBehav }
func (e *testEntity) IsBehaviorEnabled() bool { return !e.behaviorDisabled }

func TestEvaluateCondition_PropExists(t *testing.T) {
	globals := newTestGlobals(nil)

	tests := []struct {
		name  string
		props map[string]any
		cond  *ConditionNode
		want  bool
	}{
		{
			"exists with key param",
			map[string]any{"color": "red", "size": 5},
			&ConditionNode{Kind: "prop_exists", Params: map[string]any{"key": "color"}},
			true,
		},
		{
			"not exists with key param",
			map[string]any{"color": "red"},
			&ConditionNode{Kind: "prop_exists", Params: map[string]any{"key": "shape"}},
			false,
		},
		{
			"exists with compact format",
			map[string]any{"color": "red"},
			&ConditionNode{Kind: "prop_exists", Params: map[string]any{"value": "color"}},
			true,
		},
		{
			"nil properties",
			nil,
			&ConditionNode{Kind: "prop_exists", Params: map[string]any{"key": "color"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &TemplateContext{Properties: tt.props}
			got := evaluateCondition(tt.cond, globals, nil, tc)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateCondition_PropEq(t *testing.T) {
	globals := newTestGlobals(nil)
	props := map[string]any{"color": "red", "count": 3}
	tc := &TemplateContext{Properties: props}

	tests := []struct {
		name string
		cond *ConditionNode
		want bool
	}{
		{
			"string match",
			&ConditionNode{Kind: "prop_eq", Params: map[string]any{"key": "color", "value": "red"}},
			true,
		},
		{
			"string mismatch",
			&ConditionNode{Kind: "prop_eq", Params: map[string]any{"key": "color", "value": "blue"}},
			false,
		},
		{
			"int match",
			&ConditionNode{Kind: "prop_eq", Params: map[string]any{"key": "count", "value": 3}},
			true,
		},
		{
			"missing key nil value",
			&ConditionNode{Kind: "prop_eq", Params: map[string]any{"key": "missing", "value": nil}},
			true,
		},
		{
			"missing key non-nil value",
			&ConditionNode{Kind: "prop_eq", Params: map[string]any{"key": "missing", "value": "x"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateCondition(tt.cond, globals, nil, tc)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateCondition_PropEq_NilProperties(t *testing.T) {
	globals := newTestGlobals(nil)
	tc := &TemplateContext{Properties: nil}
	cond := &ConditionNode{Kind: "prop_eq", Params: map[string]any{"key": "color", "value": "red"}}
	if evaluateCondition(cond, globals, nil, tc) {
		t.Error("expected false with nil properties")
	}
}

func TestEvaluateCondition_EntityIdle(t *testing.T) {
	tests := []struct {
		name   string
		entity *testEntity
		want   bool
	}{
		{"idle", &testEntity{NullEntity: NullEntity{id: "npc"}}, true},
		{"moving", &testEntity{NullEntity: NullEntity{id: "npc"}, moving: true}, false},
		{"pushed behavior", &testEntity{NullEntity: NullEntity{id: "npc"}, hasPushedBehav: true}, false},
		{"behavior disabled", &testEntity{NullEntity: NullEntity{id: "npc"}, behaviorDisabled: true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globals := newTestEntityGlobals(nil, map[string]EntityReader{"npc": tt.entity})
			tc := &TemplateContext{SelfId: "npc"}
			cond := &ConditionNode{Kind: "entity_idle", Params: map[string]any{}}
			got := evaluateCondition(cond, globals, nil, tc)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("nil template context", func(t *testing.T) {
		globals := newTestGlobals(nil)
		cond := &ConditionNode{Kind: "entity_idle", Params: map[string]any{}}
		if evaluateCondition(cond, globals, nil) {
			t.Error("expected false with nil template context")
		}
	})
}
