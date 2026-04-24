package adventure

import "testing"

func TestTemplateContext_Resolve(t *testing.T) {
	tc := &TemplateContext{
		SelfId:   "npc_1",
		PlayerId: "player_0",
		SourceId: "player_0",
		Params: map[string]string{
			"target": "world",
		},
	}
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"self", "{{self}}", "npc_1"},
		{"player", "{{player}}", "player_0"},
		{"source", "{{source}}", "player_0"},
		{"param", "{{target}}", "world"},
		{"mixed", "Hello {{target}} from {{self}}", "Hello world from npc_1"},
		{"no_match", "plain text", "plain text"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tc.Resolve(tt.input)
			if got != tt.want {
				t.Errorf("Resolve(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTemplateContext_ResolveAny(t *testing.T) {
	tc := &TemplateContext{
		SelfId:   "npc",
		PlayerId: "player",
	}

	t.Run("string", func(t *testing.T) {
		result := tc.ResolveAny("entity: {{self}}")
		if result != "entity: npc" {
			t.Errorf("got %v", result)
		}
	})

	t.Run("map", func(t *testing.T) {
		input := map[string]any{
			"entity": "{{player}}",
			"amount": 5,
		}
		result := tc.ResolveAny(input).(map[string]any)
		if result["entity"] != "player" {
			t.Errorf("entity = %v", result["entity"])
		}
		if result["amount"] != 5 {
			t.Errorf("amount = %v", result["amount"])
		}
	})

	t.Run("slice", func(t *testing.T) {
		input := []any{"{{self}}", "literal", 42}
		result := tc.ResolveAny(input).([]any)
		if result[0] != "npc" {
			t.Errorf("[0] = %v", result[0])
		}
		if result[1] != "literal" {
			t.Errorf("[1] = %v", result[1])
		}
		if result[2] != 42 {
			t.Errorf("[2] = %v", result[2])
		}
	})

	t.Run("non_string", func(t *testing.T) {
		if tc.ResolveAny(42) != 42 {
			t.Error("non-string should pass through")
		}
		if tc.ResolveAny(true) != true {
			t.Error("bool should pass through")
		}
	})
}
