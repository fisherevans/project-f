package adventure

import (
	"testing"
)

func TestCompileExpr(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{"simple var", "self", false},
		{"string literal", "'hello'", false},
		{"comparison", "var.count >= 3", false},
		{"array access", "var.messages[0]", false},
		{"function call", "len(var.messages)", false},
		{"nested access", "save.animech.level", false},
		{"math", "var.count + 1", false},
		{"bool", "var.count > 0 && var.count < 10", false},
		{"invalid", "??bad syntax??", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompileExpr(tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompileExpr(%q) error = %v, wantErr %v", tt.src, err, tt.wantErr)
			}
		})
	}
}

func TestEvalExpr(t *testing.T) {
	env := &ExprEnv{
		Var: map[string]any{
			"count":    3,
			"name":     "test",
			"messages": []any{"a", "b", "c"},
		},
		Global: map[string]any{
			"intro_state": "complete",
		},
		Const:  map[string]any{},
		Save:   map[string]any{"character_name": "Fisher"},
		Prop:   map[string]any{"zone_id": "zone_1"},
		Param:  map[string]any{},
		Self:   "npc_1",
		Player: "player_0",
		Source: "player_0",
	}

	tests := []struct {
		name string
		src  string
		want any
	}{
		{"self", "self", "npc_1"},
		{"player", "player", "player_0"},
		{"var access", "var.count", 3},
		{"var string", "var.name", "test"},
		{"array access", "var.messages[0]", "a"},
		{"array len", "len(var.messages)", 3},
		{"comparison true", "var.count >= 3", true},
		{"comparison false", "var.count > 5", false},
		{"math", "var.count + 1", 4},
		{"global access", "global.intro_state", "complete"},
		{"save access", "save.character_name", "Fisher"},
		{"prop access", "prop.zone_id", "zone_1"},
		{"string concat", "self + ':' + player", "npc_1:player_0"},
		{"min", "min(var.count, 1)", 1.0},
		{"max", "max(var.count, 10)", 10.0},
		{"clamp", "clamp(var.count, 0, 2)", 2.0},
		{"str", "str(var.count)", "3"},
		{"contains op", "var.name contains 'es'", true},
		{"startsWith op", "var.name startsWith 'te'", true},
		{"hasKey", "hasKey(var, 'count')", true},
		{"hasKey missing", "hasKey(var, 'missing')", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog, err := CompileExpr(tt.src)
			if err != nil {
				t.Fatalf("CompileExpr(%q) error: %v", tt.src, err)
			}
			got, err := EvalExpr(prog, env)
			if err != nil {
				t.Fatalf("EvalExpr(%q) error: %v", tt.src, err)
			}
			if got != tt.want {
				t.Errorf("EvalExpr(%q) = %v (%T), want %v (%T)", tt.src, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestEvalExprBool(t *testing.T) {
	env := &ExprEnv{
		Var:    map[string]any{"x": 5},
		Global: map[string]any{},
		Const:  map[string]any{},
		Save:   map[string]any{},
		Prop:   map[string]any{},
		Param:  map[string]any{},
	}
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{"true literal", "true", true},
		{"false literal", "false", false},
		{"gt true", "var.x > 3", true},
		{"gt false", "var.x > 10", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog, err := CompileExpr(tt.src)
			if err != nil {
				t.Fatalf("CompileExpr(%q) error: %v", tt.src, err)
			}
			got := EvalExprBool(prog, env)
			if got != tt.want {
				t.Errorf("EvalExprBool(%q) = %v, want %v", tt.src, got, tt.want)
			}
		})
	}
}

func TestInterpolateString(t *testing.T) {
	env := &ExprEnv{
		Var:    map[string]any{"count": 3, "name": "Guard"},
		Global: map[string]any{},
		Const:  map[string]any{},
		Save:   map[string]any{"animech": map[string]any{"level": 5}},
		Prop:   map[string]any{},
		Param:  map[string]any{},
		Self:   "npc_1",
		Player: "player_0",
		Source: "player_0",
	}
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no interp", "plain text", "plain text"},
		{"self", "id: {{self}}", "id: npc_1"},
		{"var access", "count is {{var.count}}", "count is 3"},
		{"math", "next: {{var.count + 1}}", "next: 4"},
		{"nested save", "level {{save.animech.level}}", "level 5"},
		{"multiple", "{{var.name}} ({{self}})", "Guard (npc_1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InterpolateString(tt.input, env)
			if got != tt.want {
				t.Errorf("InterpolateString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConstAccess(t *testing.T) {
	env := &ExprEnv{
		Var:   map[string]any{},
		Global: map[string]any{},
		Const: map[string]any{
			"guard_denials": []any{"No entry.", "I told you.", "Last warning."},
			"max_retries":   5,
		},
		Save:   map[string]any{},
		Prop:   map[string]any{},
		Param:  map[string]any{},
	}
	tests := []struct {
		name string
		src  string
		want any
	}{
		{"const access", "const.max_retries", 5},
		{"const list len", "len(const.guard_denials)", 3},
		{"const list index", "const.guard_denials[0]", "No entry."},
		{"const in expr", "const.max_retries > 3", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog, err := CompileExpr(tt.src)
			if err != nil {
				t.Fatalf("CompileExpr(%q) error: %v", tt.src, err)
			}
			got, err := EvalExpr(prog, env)
			if err != nil {
				t.Fatalf("EvalExpr(%q) error: %v", tt.src, err)
			}
			if got != tt.want {
				t.Errorf("EvalExpr(%q) = %v (%T), want %v (%T)", tt.src, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestCompileInterpolatedString(t *testing.T) {
	env := &ExprEnv{
		Var:    map[string]any{"count": 2},
		Global: map[string]any{},
		Const:  map[string]any{},
		Save:   map[string]any{},
		Prop:   map[string]any{},
		Param:  map[string]any{},
		Self:   "npc",
	}

	t.Run("compile and eval", func(t *testing.T) {
		cis, err := CompileInterpolatedString("Hello {{self}}, count={{var.count}}")
		if err != nil {
			t.Fatal(err)
		}
		got := cis.Eval(env)
		want := "Hello npc, count=2"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("no expressions", func(t *testing.T) {
		cis, err := CompileInterpolatedString("plain text")
		if err != nil {
			t.Fatal(err)
		}
		got := cis.Eval(env)
		if got != "plain text" {
			t.Errorf("got %q, want %q", got, "plain text")
		}
	})

	t.Run("compile error", func(t *testing.T) {
		_, err := CompileInterpolatedString("bad {{??}}")
		if err == nil {
			t.Error("expected error")
		}
	})
}
