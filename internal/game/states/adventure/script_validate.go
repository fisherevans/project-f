package adventure

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

var validStepKinds = map[string]bool{
	"dialogue":              true,
	"self_dialogue":         true,
	"chatter":               true,
	"timer":                 true,
	"play_sound":            true,
	"set_world_state":       true,
	"set_run_state":         true,
	"delete_entity":         true,
	"face_direction":        true,
	"face_entity":           true,
	"scripted_motion":       true,
	"push_behavior":         true,
	"pop_behavior":          true,
	"disable_behavior":      true,
	"enable_behavior":       true,
	"reset_movement":        true,
	"trigger_movement":      true,
	"set_mode":              true,
	"set_blocking":          true,
	"reset_animation":       true,
	"change_player_renderer": true,
	"override_camera":       true,
	"pop_camera":            true,
	"mutate_camera":         true,
	"teleport_player":       true,
	"load_map":              true,
	"fade":                  true,
	"deactivate_fade":       true,
	"yield_elythium":        true,
	"tooltip":               true,
	"broadcast":             true,
	"wait_for":              true,
	"wait_for_animation":    true,
	"action":                true,
	"parallel":              true,
	"focused_sequence":      true,
	"pick_dialogue":         true,
	"pick_self_dialogue":    true,
	"pick_chatter":          true,
	"configure_mode_entity": true,
	"ref":                   true,
	"set_var":               true,
	"if":                    true,
	"switch":                true,
	"while":                 true,
	"custom_action":         true,
	"return":                true,
}

var builtinConditionKinds = map[string]bool{
	"global_eq":        true,
	"global_ne":        true,
	"global_gt":        true,
	"global_gte":       true,
	"global_lt":        true,
	"global_lte":       true,
	"global_exists":    true,
	"global_not_exists": true,
	"handler_state_eq": true,
	"handler_state_ne": true,
	"all":              true,
	"any":              true,
	"not":              true,
	"prop_exists":      true,
	"prop_eq":          true,
	"entity_idle":      true,
	"expr":             true,
}

func validateScriptFiles() {
	var errors []string

	for name, handler := range scriptHandlerDefs {
		for _, err := range validateHandler(handler) {
			errors = append(errors, fmt.Sprintf("handler %q: %s", name, err))
		}
	}
	for name, seq := range scriptSequences {
		for _, err := range validateSteps(seq.Steps) {
			errors = append(errors, fmt.Sprintf("sequence %q: %s", name, err))
		}
	}
	for name, action := range scriptCustomActions {
		for _, err := range validateSteps(action.Steps) {
			errors = append(errors, fmt.Sprintf("custom_action %q: %s", name, err))
		}
	}

	if len(errors) > 0 {
		log.Fatal().Strs("errors", errors).Int("count", len(errors)).Msg("script validation failed")
	}
}

func validateHandler(handler *HandlerDef) []string {
	var errors []string
	allRules := [][]*RuleDef{
		handler.OnInteractSelf,
		handler.OnInteract,
		handler.OnZoneActivity,
		handler.OnBroadcast,
		handler.OnGlobalUpdated,
		handler.OnInit,
		handler.OnStateEnter,
		handler.OnCombatComplete,
		handler.OnTimerComplete,
		handler.OnMotionCompleteSelf,
		handler.OnMotionComplete,
	}
	for _, rules := range allRules {
		for _, rule := range rules {
			errors = append(errors, validateCondition(rule.When)...)
			errors = append(errors, validateSteps(rule.Steps)...)
		}
	}
	return errors
}

func validateCondition(node *ConditionNode) []string {
	if node == nil {
		return nil
	}
	var errors []string

	if !builtinConditionKinds[node.Kind] {
		if _, ok := scriptConditions[node.Kind]; !ok {
			errors = append(errors, fmt.Sprintf("unknown condition kind %q", node.Kind))
		}
	}

	switch node.Kind {
	case "all", "any":
		items, ok := node.Params.([]any)
		if ok {
			for _, item := range items {
				sub, err := decodeConditionFromAny(item)
				if err == nil {
					errors = append(errors, validateCondition(sub)...)
				}
			}
		}
	case "not":
		sub, err := decodeConditionFromAny(node.Params)
		if err == nil {
			errors = append(errors, validateCondition(sub)...)
		}
	}
	return errors
}

func validateSteps(steps []*StepNode) []string {
	var errors []string
	for _, step := range steps {
		errors = append(errors, validateStep(step)...)
	}
	return errors
}

func validateStep(step *StepNode) []string {
	var errors []string

	if !validStepKinds[step.Kind] {
		errors = append(errors, fmt.Sprintf("unknown step kind %q", step.Kind))
		return errors
	}

	switch step.Kind {
	case "action":
		name := ""
		switch v := step.Params.(type) {
		case string:
			name = v
		default:
			if m, ok := step.Params.(map[string]any); ok {
				name, _ = m["name"].(string)
			}
		}
		if name != "" {
			if _, ok := scriptActions[name]; !ok {
				errors = append(errors, fmt.Sprintf("unknown action %q", name))
			}
		}

	case "custom_action":
		name := ""
		switch v := step.Params.(type) {
		case string:
			name = v
		default:
			if m, ok := step.Params.(map[string]any); ok {
				name, _ = m["name"].(string)
			}
		}
		if name != "" {
			if _, ok := scriptCustomActions[name]; !ok {
				errors = append(errors, fmt.Sprintf("unknown custom action %q", name))
			}
		}

	case "ref":
		if m, ok := step.Params.(map[string]any); ok {
			name, _ := m["name"].(string)
			if name != "" {
				if _, ok := scriptSequences[name]; !ok {
					errors = append(errors, fmt.Sprintf("unknown sequence ref %q", name))
				}
			}
		}

	case "wait_for":
		if m, ok := step.Params.(map[string]any); ok {
			condName, _ := m["condition"].(string)
			if condName != "" {
				if _, ok := getScriptConditionFactory(condName); !ok {
					errors = append(errors, fmt.Sprintf("unknown wait_for condition %q", condName))
				}
			}
		}

	case "pick_dialogue", "pick_self_dialogue", "pick_chatter":
		listName := ""
		if m, ok := step.Params.(map[string]any); ok {
			listName, _ = m["list"].(string)
		}
		if listName != "" && !strings.Contains(listName, "{{") {
			listName = strings.Trim(listName, "'\"")
			if _, ok := getDataList(listName); !ok {
				errors = append(errors, fmt.Sprintf("unknown data list %q in %s", listName, step.Kind))
			}
		}

	case "if":
		if m, ok := step.Params.(map[string]any); ok {
			errors = append(errors, validateSteps(extractSubSteps(m["then"]))...)
			errors = append(errors, validateSteps(extractSubSteps(m["else"]))...)
		}

	case "switch":
		if m, ok := step.Params.(map[string]any); ok {
			if rawCases, ok := m["cases"].([]any); ok {
				for _, rawCase := range rawCases {
					if caseMap, ok := rawCase.(map[string]any); ok {
						errors = append(errors, validateSteps(extractSubSteps(caseMap["steps"]))...)
					}
				}
			}
			errors = append(errors, validateSteps(extractSubSteps(m["default"]))...)
		}

	case "while":
		if m, ok := step.Params.(map[string]any); ok {
			errors = append(errors, validateSteps(extractSubSteps(m["steps"]))...)
		}

	case "parallel":
		if items, ok := step.Params.([]any); ok {
			for _, item := range items {
				if stepMap, ok := item.(map[string]any); ok {
					for k, v := range stepMap {
						errors = append(errors, validateStep(&StepNode{Kind: k, Params: v})...)
					}
				}
			}
		}

	case "focused_sequence":
		if m, ok := step.Params.(map[string]any); ok {
			errors = append(errors, validateSteps(extractSubSteps(m["effects"]))...)
			errors = append(errors, validateSteps(extractSubSteps(m["post_effects"]))...)
		}

	case "teleport_player":
		if m, ok := step.Params.(map[string]any); ok {
			errors = append(errors, validateSteps(extractSubSteps(m["interstitial"]))...)
		}
	}

	return errors
}
