package adventure

import (
	"testing"

	"fisherevans.com/project/f/internal/schema"
)

var knownBuiltinConditions = []string{
	"all", "any", "not",
	"expr",
}

func loadSchema(t *testing.T) *schema.ScriptSchema {
	t.Helper()
	s, err := schema.LoadScriptSchema()
	if err != nil {
		t.Fatalf("failed to load script schema: %v", err)
	}
	return s
}

func TestAllActionsHaveSchema(t *testing.T) {
	s := loadSchema(t)
	for name := range scriptActions {
		if _, ok := s.Actions[name]; !ok {
			t.Errorf("action %q is registered in Go but missing from schema JSON", name)
		}
	}
	for name := range s.Actions {
		if _, ok := scriptActions[name]; !ok {
			t.Errorf("schema JSON has action %q but it is not registered in Go", name)
		}
	}
}

func TestAllConditionsHaveSchema(t *testing.T) {
	s := loadSchema(t)
	for name := range scriptConditions {
		if _, ok := s.Conditions[name]; !ok {
			t.Errorf("condition %q is registered in Go but missing from schema JSON", name)
		}
	}
	for name := range s.Conditions {
		if _, ok := scriptConditions[name]; !ok {
			t.Errorf("schema JSON has condition %q but it is not registered in Go", name)
		}
	}
}

func TestAllStepKindsHaveSchema(t *testing.T) {
	s := loadSchema(t)
	for name := range stepConverters {
		if _, ok := s.StepKinds[name]; !ok {
			t.Errorf("step kind %q is registered in Go but missing from schema JSON", name)
		}
	}
	for name := range s.StepKinds {
		if _, ok := stepConverters[name]; !ok {
			t.Errorf("schema JSON has step kind %q but it is not registered in Go", name)
		}
	}
}

func TestAllBuiltinConditionsHaveSchema(t *testing.T) {
	s := loadSchema(t)
	for _, name := range knownBuiltinConditions {
		if _, ok := s.BuiltinConditions[name]; !ok {
			t.Errorf("builtin condition %q is in knownBuiltinConditions but missing from schema JSON", name)
		}
	}
	for name := range s.BuiltinConditions {
		found := false
		for _, known := range knownBuiltinConditions {
			if known == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("schema JSON has builtin condition %q but it is not in knownBuiltinConditions list", name)
		}
	}
}

func TestEventHooksHaveSchema(t *testing.T) {
	s := loadSchema(t)
	expectedHooks := []string{
		"on_init", "on_interact_self", "on_interact",
		"on_zone_activity", "on_broadcast", "on_global_updated",
		"on_state_enter", "on_combat_complete", "on_timer_complete",
		"on_motion_complete_self", "on_motion_complete",
	}
	for _, name := range expectedHooks {
		if _, ok := s.EventHooks[name]; !ok {
			t.Errorf("event hook %q is missing from schema JSON", name)
		}
	}
	for name := range s.EventHooks {
		found := false
		for _, expected := range expectedHooks {
			if expected == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("schema JSON has event hook %q but it is not in expected list", name)
		}
	}
}

func TestSchemaLoadsSuccessfully(t *testing.T) {
	s := loadSchema(t)
	if len(s.StepKinds) == 0 {
		t.Error("schema has no step kinds")
	}
	if len(s.Actions) == 0 {
		t.Error("schema has no actions")
	}
	if len(s.Conditions) == 0 {
		t.Error("schema has no conditions")
	}
	if len(s.BuiltinConditions) == 0 {
		t.Error("schema has no builtin conditions")
	}
	if len(s.EventHooks) == 0 {
		t.Error("schema has no event hooks")
	}
	if len(s.TemplateVars) == 0 {
		t.Error("schema has no template vars")
	}
}
