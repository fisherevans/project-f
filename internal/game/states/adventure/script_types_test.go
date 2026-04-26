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
	if len(h.OnInteractSelf.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(h.OnInteractSelf.Rules))
	}
	rule := h.OnInteractSelf.Rules[0]
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
          expr: "global.my_key == true"
        steps:
          - dialogue: "Door is open"
      - when:
          not:
            expr: "global.my_key == true"
        steps:
          - dialogue: "Door is closed"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	h := sf.Handlers["test"]
	if len(h.OnInteractSelf.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(h.OnInteractSelf.Rules))
	}
	r1 := h.OnInteractSelf.Rules[0]
	if r1.When == nil {
		t.Fatal("expected when condition on rule 1")
	}
	if r1.When.Kind != "expr" {
		t.Errorf("expected 'expr', got %q", r1.When.Kind)
	}
	r2 := h.OnInteractSelf.Rules[1]
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
            - expr: "global.key1 == true"
            - expr: "global.key2 == false"
        steps:
          - dialogue: "Both match"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	h := sf.Handlers["test"]
	r := h.OnInteractSelf.Rules[0]
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
	r := h.OnZoneActivity.Rules[0]
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
	steps := sf.Handlers["test"].OnInteractSelf.Rules[0].Steps
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
	steps := sf.Handlers["test"].OnInteractSelf.Rules[0].Steps
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

func TestParseScriptFile_MultipleHooks(t *testing.T) {
	yaml := `
handlers:
  test:
    on_init:
      - steps:
          - set_var:
              key: ready
              value: "true"
    on_zone_activity:
      - when:
          expr: "var.ready == true"
        steps:
          - dialogue: "Triggered"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	h := sf.Handlers["test"]
	if len(h.OnInit.Rules) != 1 {
		t.Fatalf("expected 1 on_init rule, got %d", len(h.OnInit.Rules))
	}
	initRule := h.OnInit.Rules[0]
	if len(initRule.Steps) != 1 {
		t.Fatalf("expected 1 step in on_init, got %d", len(initRule.Steps))
	}
	zoneRule := h.OnZoneActivity.Rules[0]
	if zoneRule.When == nil {
		t.Fatal("expected when condition on zone rule")
	}
}

func TestParseScriptFile_HookWithMode(t *testing.T) {
	yaml := `
handlers:
  test:
    on_broadcast:
      mode: all
      rules:
        - filter:
            id: alert
          steps:
            - dialogue: "handler A"
        - filter:
            id: alert
          steps:
            - dialogue: "handler B"
`
	sf, err := ParseScriptFile([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	h := sf.Handlers["test"]
	if h.OnBroadcast == nil {
		t.Fatal("expected on_broadcast hook")
	}
	if h.OnBroadcast.Mode != HookModeAll {
		t.Errorf("mode = %q, want %q", h.OnBroadcast.Mode, HookModeAll)
	}
	if len(h.OnBroadcast.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(h.OnBroadcast.Rules))
	}
}

func TestHookDef_YAMLRoundTrip_ModePreserved(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantMode string
		wantLen  int
	}{
		{
			name: "list format defaults to first_match",
			input: `
handlers:
  test:
    on_interact_self:
      - steps:
          - dialogue: "hello"
`,
			wantMode: HookModeFirstMatch,
			wantLen:  1,
		},
		{
			name: "map format with mode all",
			input: `
handlers:
  test:
    on_interact_self:
      mode: all
      rules:
        - steps:
            - dialogue: "hello"
        - steps:
            - dialogue: "world"
`,
			wantMode: HookModeAll,
			wantLen:  2,
		},
		{
			name: "map format with explicit first_match",
			input: `
handlers:
  test:
    on_interact_self:
      mode: first_match
      rules:
        - steps:
            - dialogue: "hello"
`,
			wantMode: "first_match",
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf, err := ParseScriptFile([]byte(tt.input))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			hook := sf.Handlers["test"].OnInteractSelf
			if hook == nil {
				t.Fatal("expected on_interact_self hook")
			}
			if hook.Mode != tt.wantMode {
				t.Errorf("mode = %q, want %q", hook.Mode, tt.wantMode)
			}
			if len(hook.Rules) != tt.wantLen {
				t.Errorf("rules count = %d, want %d", len(hook.Rules), tt.wantLen)
			}
		})
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
              params:
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
