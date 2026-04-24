package adventure

import (
	"testing"
)

func TestParseScriptFile_BasicHandler(t *testing.T) {
	yaml := `
handlers:
  test_entity:
    on_interact_self:
      - steps:
          - dialogue: "Hello!"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(sf.Handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(sf.Handlers))
	}
	h := sf.Handlers["test_entity"]
	if h == nil {
		t.Fatal("handler 'test_entity' not found")
	}
	if len(h.OnInteractSelf) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(h.OnInteractSelf))
	}
	rule := h.OnInteractSelf[0]
	if len(rule.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(rule.Steps))
	}
	if rule.Steps[0].Kind != "dialogue" {
		t.Errorf("expected step kind 'dialogue', got %q", rule.Steps[0].Kind)
	}
	if rule.Steps[0].Params != "Hello!" {
		t.Errorf("expected params 'Hello!', got %v", rule.Steps[0].Params)
	}
}

func TestParseScriptFile_Conditions(t *testing.T) {
	yaml := `
handlers:
  test:
    on_interact_self:
      - when:
          global_eq:
            key: "my_key"
            value: true
        steps:
          - dialogue: "Door is open"
      - when:
          not:
            global_eq:
              key: "my_key"
              value: true
        steps:
          - dialogue: "Door is closed"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	h := sf.Handlers["test"]
	if len(h.OnInteractSelf) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(h.OnInteractSelf))
	}
	r1 := h.OnInteractSelf[0]
	if r1.When == nil {
		t.Fatal("expected when condition on rule 1")
	}
	if r1.When.Kind != "global_eq" {
		t.Errorf("expected 'global_eq', got %q", r1.When.Kind)
	}
	r2 := h.OnInteractSelf[1]
	if r2.When == nil {
		t.Fatal("expected when condition on rule 2")
	}
	if r2.When.Kind != "not" {
		t.Errorf("expected 'not', got %q", r2.When.Kind)
	}
}

func TestParseScriptFile_AllCondition(t *testing.T) {
	yaml := `
handlers:
  test:
    on_interact_self:
      - when:
          all:
            - global_eq:
                key1: true
            - global_eq:
                key2: false
        steps:
          - dialogue: "Both match"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	h := sf.Handlers["test"]
	r := h.OnInteractSelf[0]
	if r.When.Kind != "all" {
		t.Errorf("expected 'all', got %q", r.When.Kind)
	}
	items, ok := r.When.Params.([]any)
	if !ok {
		t.Fatalf("expected 'all' params to be a list, got %T", r.When.Params)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 sub-conditions, got %d", len(items))
	}
}

func TestParseScriptFile_Filters(t *testing.T) {
	yaml := `
handlers:
  test:
    on_zone_activity:
      - filter:
          zone: "my_zone"
          entity: "{{player}}"
          entering: true
        steps:
          - dialogue: "Entered zone"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	h := sf.Handlers["test"]
	r := h.OnZoneActivity[0]
	if r.Filter["zone"] != "my_zone" {
		t.Errorf("expected zone 'my_zone', got %v", r.Filter["zone"])
	}
	if r.Filter["entering"] != true {
		t.Errorf("expected entering true, got %v", r.Filter["entering"])
	}
}

func TestParseScriptFile_MultipleStepKinds(t *testing.T) {
	yaml := `
handlers:
  test:
    on_interact_self:
      - steps:
          - dialogue: "Hello"
          - timer: 0.5
          - set_run_state:
              key: "foo"
              value: true
          - play_sound: "beep"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	steps := sf.Handlers["test"].OnInteractSelf[0].Steps
	if len(steps) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(steps))
	}
	tests := []struct {
		kind string
	}{
		{"dialogue"},
		{"timer"},
		{"set_run_state"},
		{"play_sound"},
	}
	for i, tt := range tests {
		if steps[i].Kind != tt.kind {
			t.Errorf("step %d: expected kind %q, got %q", i, tt.kind, steps[i].Kind)
		}
	}
}

func TestParseScriptFile_FocusedSequence(t *testing.T) {
	yaml := `
handlers:
  test:
    on_interact_self:
      - steps:
          - focused_sequence:
              move_camera: true
              effects:
                - dialogue: "Hello"
                - timer: 1.5
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	steps := sf.Handlers["test"].OnInteractSelf[0].Steps
	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}
	if steps[0].Kind != "focused_sequence" {
		t.Errorf("expected 'focused_sequence', got %q", steps[0].Kind)
	}
	m, ok := steps[0].Params.(map[string]any)
	if !ok {
		t.Fatalf("expected map params, got %T", steps[0].Params)
	}
	if m["move_camera"] != true {
		t.Errorf("expected move_camera true, got %v", m["move_camera"])
	}
	effects, ok := m["effects"].([]any)
	if !ok {
		t.Fatalf("expected effects list, got %T", m["effects"])
	}
	if len(effects) != 2 {
		t.Fatalf("expected 2 effects, got %d", len(effects))
	}
}

func TestParseScriptFile_HandlerState(t *testing.T) {
	yaml := `
handlers:
  test:
    on_init:
      - set_state:
          ready: true
        steps: []
    on_zone_activity:
      - when:
          handler_state_eq:
            ready: true
        set_state:
          ready: false
        steps:
          - dialogue: "Triggered"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	h := sf.Handlers["test"]
	if len(h.OnInit) != 1 {
		t.Fatalf("expected 1 on_init rule, got %d", len(h.OnInit))
	}
	initRule := h.OnInit[0]
	if initRule.SetState["ready"] != true {
		t.Errorf("expected set_state ready=true, got %v", initRule.SetState["ready"])
	}
	zoneRule := h.OnZoneActivity[0]
	if zoneRule.SetState["ready"] != false {
		t.Errorf("expected set_state ready=false, got %v", zoneRule.SetState["ready"])
	}
}

func TestParseScriptFile_Sequences(t *testing.T) {
	yaml := `
sequences:
  my_sequence:
    params:
      - target
    steps:
      - dialogue: "Hello {{target}}"
      - timer: 1.0

handlers:
  test:
    on_interact_self:
      - steps:
          - ref:
              name: my_sequence
              with:
                target: "world"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(sf.Sequences) != 1 {
		t.Fatalf("expected 1 sequence, got %d", len(sf.Sequences))
	}
	seq := sf.Sequences["my_sequence"]
	if len(seq.Params) != 1 || seq.Params[0] != "target" {
		t.Errorf("expected params [target], got %v", seq.Params)
	}
	if len(seq.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(seq.Steps))
	}
}
