package adventure

import "testing"

func TestSelectFromList(t *testing.T) {
	list := []string{"a", "b", "c"}

	t.Run("cycling", func(t *testing.T) {
		tests := []struct {
			name    string
			counter int
			want    string
		}{
			{"0", 0, "a"},
			{"1", 1, "b"},
			{"2", 2, "c"},
			{"wraps", 3, "a"},
			{"wraps again", 5, "c"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := selectFromList(list, "cycling", tt.counter)
				if got != tt.want {
					t.Errorf("cycling counter=%d: got %q, want %q", tt.counter, got, tt.want)
				}
			})
		}
	})

	t.Run("capped", func(t *testing.T) {
		tests := []struct {
			name    string
			counter int
			want    string
		}{
			{"0", 0, "a"},
			{"1", 1, "b"},
			{"2", 2, "c"},
			{"beyond", 3, "c"},
			{"way beyond", 100, "c"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := selectFromList(list, "capped", tt.counter)
				if got != tt.want {
					t.Errorf("capped counter=%d: got %q, want %q", tt.counter, got, tt.want)
				}
			})
		}
	})

	t.Run("random", func(t *testing.T) {
		got := selectFromList(list, "random", 0)
		if got != "a" && got != "b" && got != "c" {
			t.Errorf("random: got %q, not in list", got)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		got := selectFromList(nil, "cycling", 0)
		if got != "" {
			t.Errorf("empty list: got %q, want empty", got)
		}
	})
}

func TestConvertSteps_PickDialogue(t *testing.T) {
	oldConsts := scriptConsts
	scriptConsts = map[string]any{
		"greetings": []any{"Hello!", "Hi there!", "Hey!"},
	}
	defer func() { scriptConsts = oldConsts }()

	globals := newTestGlobals(map[string]any{
		"greet_counter": 1,
	})

	tc := &TemplateContext{
		SelfId:  "npc",
		Globals: globals,
	}

	t.Run("cycling with counter", func(t *testing.T) {
		steps := []*StepNode{{Kind: "pick_dialogue", Params: map[string]any{
			"list":        "greetings",
			"select":      "cycling",
			"counter_key": "greet_counter",
		}}}
		effects := convertSteps(steps, tc, nil)
		if len(effects) != 2 {
			t.Fatalf("got %d effects, want 2", len(effects))
		}
		d, ok := effects[0].(*EffectDialogue)
		if !ok {
			t.Fatalf("effect[0] = %T, want *EffectDialogue", effects[0])
		}
		if d.Text != "Hi there!" {
			t.Errorf("text = %q, want %q", d.Text, "Hi there!")
		}
		rs, ok := effects[1].(*EffectSetGlobal)
		if !ok {
			t.Fatalf("effect[1] = %T, want *EffectSetGlobal", effects[1])
		}
		if rs.Key != "greet_counter" {
			t.Errorf("key = %q", rs.Key)
		}
		if rs.Value != 2 {
			t.Errorf("value = %v, want 2", rs.Value)
		}
	})

	t.Run("no increment", func(t *testing.T) {
		steps := []*StepNode{{Kind: "pick_dialogue", Params: map[string]any{
			"list":        "greetings",
			"counter_key": "greet_counter",
			"increment":   false,
		}}}
		effects := convertSteps(steps, tc, nil)
		if len(effects) != 1 {
			t.Fatalf("got %d effects, want 1", len(effects))
		}
	})

	t.Run("unknown list", func(t *testing.T) {
		steps := []*StepNode{{Kind: "pick_dialogue", Params: map[string]any{
			"list": "nonexistent",
		}}}
		effects := convertSteps(steps, tc, nil)
		if len(effects) != 0 {
			t.Errorf("got %d effects, want 0 for unknown list", len(effects))
		}
	})
}

func TestConvertSteps_PickSelfDialogue(t *testing.T) {
	oldConsts := scriptConsts
	scriptConsts = map[string]any{
		"thoughts": []any{"Hmm...", "Interesting."},
	}
	defer func() { scriptConsts = oldConsts }()

	tc := &TemplateContext{SelfId: "npc"}

	steps := []*StepNode{{Kind: "pick_self_dialogue", Params: map[string]any{
		"list": "thoughts",
	}}}
	effects := convertSteps(steps, tc, nil)
	if len(effects) != 1 {
		t.Fatalf("got %d effects, want 1", len(effects))
	}
	d, ok := effects[0].(*EffectDialogue)
	if !ok {
		t.Fatalf("effect[0] = %T, want *EffectDialogue", effects[0])
	}
	if d.Style == nil || *d.Style != DialogueStyleSelf {
		t.Error("expected DialogueStyleSelf")
	}
}

func TestConvertSteps_PickChatter(t *testing.T) {
	oldConsts := scriptConsts
	scriptConsts = map[string]any{
		"remarks": []any{"Nice day.", "Cold out."},
	}
	defer func() { scriptConsts = oldConsts }()

	tc := &TemplateContext{SelfId: "npc"}

	steps := []*StepNode{{Kind: "pick_chatter", Params: map[string]any{
		"list":     "remarks",
		"entity":   "{{self}}",
		"duration": 2.5,
	}}}
	effects := convertSteps(steps, tc, nil)
	if len(effects) != 1 {
		t.Fatalf("got %d effects, want 1", len(effects))
	}
	c, ok := effects[0].(*EffectChatter)
	if !ok {
		t.Fatalf("effect[0] = %T, want *EffectChatter", effects[0])
	}
	if c.EntityId != "npc" {
		t.Errorf("entity = %q", c.EntityId)
	}
	if c.DurationSeconds != 2.5 {
		t.Errorf("duration = %v", c.DurationSeconds)
	}
}

func TestConvertSteps_ConfigureModeEntity(t *testing.T) {
	tc := &TemplateContext{SelfId: "door1"}

	t.Run("animations and lights", func(t *testing.T) {
		steps := []*StepNode{{Kind: "configure_mode_entity", Params: map[string]any{
			"entity": "{{self}}",
			"modes": map[string]any{
				"closed": map[string]any{
					"animations": []any{
						map[string]any{"name": "doors/shield:closed"},
					},
					"lights": []any{
						map[string]any{"color": "#127fd7", "size": 1.5},
					},
				},
				"open": map[string]any{
					"animations": []any{
						map[string]any{"name": "doors/shield:open"},
					},
				},
			},
		}}}
		effects := convertSteps(steps, tc, nil)
		if len(effects) != 1 {
			t.Fatalf("got %d effects, want 1", len(effects))
		}
		e, ok := effects[0].(*EffectMutateModeBasedEntity)
		if !ok {
			t.Fatalf("effect[0] = %T, want *EffectMutateModeBasedEntity", effects[0])
		}
		if e.EntityId != "door1" {
			t.Errorf("entity = %q", e.EntityId)
		}
		if len(e.Animations) != 2 {
			t.Errorf("animation modes = %d, want 2", len(e.Animations))
		}
		if len(e.Lights) != 1 {
			t.Errorf("light modes = %d, want 1", len(e.Lights))
		}
	})

	t.Run("with sounds", func(t *testing.T) {
		steps := []*StepNode{{Kind: "configure_mode_entity", Params: map[string]any{
			"entity": "{{self}}",
			"modes": map[string]any{
				"closed": map[string]any{
					"animations": []any{
						map[string]any{"name": "doors/shield:closed"},
					},
					"sounds_on_enter": []any{
						map[string]any{
							"name":    "adventure/props/hum",
							"loop":    true,
							"falloff": "standard",
							"volume":  0.9,
							"fade_in": 1.0,
						},
					},
				},
			},
		}}}
		effects := convertSteps(steps, tc, nil)
		if len(effects) != 2 {
			t.Fatalf("got %d effects, want 2", len(effects))
		}
		_, ok := effects[0].(*EffectMutateModeBasedEntity)
		if !ok {
			t.Fatalf("effect[0] = %T, want *EffectMutateModeBasedEntity", effects[0])
		}
		sp, ok := effects[1].(*EffectAddSoundProvider)
		if !ok {
			t.Fatalf("effect[1] = %T, want *EffectAddSoundProvider", effects[1])
		}
		if sp.EntityId != "door1" {
			t.Errorf("sound provider entity = %q", sp.EntityId)
		}
		if sp.ModeBase == nil {
			t.Fatal("ModeBase is nil")
		}
	})

	t.Run("sounds only", func(t *testing.T) {
		steps := []*StepNode{{Kind: "configure_mode_entity", Params: map[string]any{
			"entity": "{{self}}",
			"modes": map[string]any{
				"active": map[string]any{
					"sounds_on_enter": []any{
						map[string]any{"name": "sfx/buzz", "volume": 0.5, "falloff": "ambient"},
					},
				},
			},
		}}}
		effects := convertSteps(steps, tc, nil)
		if len(effects) != 1 {
			t.Fatalf("got %d effects, want 1", len(effects))
		}
		_, ok := effects[0].(*EffectAddSoundProvider)
		if !ok {
			t.Fatalf("effect[0] = %T, want *EffectAddSoundProvider", effects[0])
		}
	})
}
