package adventure

import (
	"fmt"

	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/rs/zerolog/log"
)

func evaluateCondition(node *ConditionNode, globals StateGlobalsReader, handlerState map[string]any) bool {
	if node == nil {
		return true
	}
	switch node.Kind {
	case "global_eq":
		return evalGlobalComparison(node.Params, globals, func(gv any, expected any) bool {
			return fmt.Sprintf("%v", gv) == fmt.Sprintf("%v", expected)
		})
	case "global_ne":
		return evalGlobalComparison(node.Params, globals, func(gv any, expected any) bool {
			return fmt.Sprintf("%v", gv) != fmt.Sprintf("%v", expected)
		})
	case "global_gt":
		return evalGlobalNumericComparison(node.Params, globals, func(a, b float64) bool { return a > b })
	case "global_gte":
		return evalGlobalNumericComparison(node.Params, globals, func(a, b float64) bool { return a >= b })
	case "global_lt":
		return evalGlobalNumericComparison(node.Params, globals, func(a, b float64) bool { return a < b })
	case "global_lte":
		return evalGlobalNumericComparison(node.Params, globals, func(a, b float64) bool { return a <= b })
	case "global_exists":
		key, _ := node.Params.(string)
		return globals.Get(key).Exists()
	case "global_not_exists":
		key, _ := node.Params.(string)
		return !globals.Get(key).Exists()
	case "handler_state_eq":
		return evalHandlerStateComparison(node.Params, handlerState, func(gv any, expected any) bool {
			return fmt.Sprintf("%v", gv) == fmt.Sprintf("%v", expected)
		})
	case "handler_state_ne":
		return evalHandlerStateComparison(node.Params, handlerState, func(gv any, expected any) bool {
			return fmt.Sprintf("%v", gv) != fmt.Sprintf("%v", expected)
		})
	case "all":
		items, ok := node.Params.([]any)
		if !ok {
			return false
		}
		for _, item := range items {
			sub, err := decodeConditionFromAny(item)
			if err != nil {
				log.Warn().Err(err).Msg("failed to decode 'all' sub-condition")
				return false
			}
			if !evaluateCondition(sub, globals, handlerState) {
				return false
			}
		}
		return true
	case "any":
		items, ok := node.Params.([]any)
		if !ok {
			return false
		}
		for _, item := range items {
			sub, err := decodeConditionFromAny(item)
			if err != nil {
				log.Warn().Err(err).Msg("failed to decode 'any' sub-condition")
				continue
			}
			if evaluateCondition(sub, globals, handlerState) {
				return true
			}
		}
		return false
	case "not":
		sub, err := decodeConditionFromAny(node.Params)
		if err != nil {
			log.Warn().Err(err).Msg("failed to decode 'not' sub-condition")
			return false
		}
		return !evaluateCondition(sub, globals, handlerState)
	default:
		fn, ok := getScriptConditionFactory(node.Kind)
		if !ok {
			log.Warn().Str("kind", node.Kind).Msg("unknown condition kind")
			return false
		}
		params := anyToStringMap(node.Params)
		check := fn(params)
		if check == nil {
			return false
		}
		sg, ok := globals.(*stateGlobals)
		if !ok {
			log.Warn().Str("kind", node.Kind).Msg("cannot evaluate registered condition without state access")
			return false
		}
		return check(sg.state, 0)
	}
}

func decodeConditionFromAny(v any) (*ConditionNode, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("condition must be a map, got %T", v)
	}
	if len(m) != 1 {
		return nil, fmt.Errorf("condition must have exactly one key, got %d", len(m))
	}
	for k, val := range m {
		return &ConditionNode{Kind: k, Params: val}, nil
	}
	return nil, fmt.Errorf("unreachable")
}

func evalGlobalComparison(params any, globals StateGlobalsReader, cmp func(any, any) bool) bool {
	m, ok := params.(map[string]any)
	if !ok {
		return false
	}
	key, hasKey := m["key"]
	value, hasValue := m["value"]
	if hasKey && hasValue {
		keyStr, _ := key.(string)
		gv := globals.Get(keyStr)
		if !gv.Exists() {
			return cmp(nil, value)
		}
		return cmp(gv.Value(), value)
	}
	for k, expected := range m {
		gv := globals.Get(k)
		if !gv.Exists() {
			return cmp(nil, expected)
		}
		if !cmp(gv.Value(), expected) {
			return false
		}
	}
	return true
}

func evalGlobalNumericComparison(params any, globals StateGlobalsReader, cmp func(float64, float64) bool) bool {
	m, ok := params.(map[string]any)
	if !ok {
		return false
	}
	for k, expected := range m {
		if k == "key" || k == "value" {
			continue
		}
		gv := globals.Get(k)
		actual := globalValueToFloat(gv)
		expectedFloat := toFloat64(expected)
		if !cmp(actual, expectedFloat) {
			return false
		}
	}
	key, hasKey := m["key"]
	value, hasValue := m["value"]
	if hasKey && hasValue {
		keyStr, _ := key.(string)
		actual := globalValueToFloat(globals.Get(keyStr))
		expectedFloat := toFloat64(value)
		return cmp(actual, expectedFloat)
	}
	return true
}

func globalValueToFloat(gv *rpg.GlobalValue) float64 {
	if !gv.Exists() {
		return 0
	}
	return toFloat64(gv.Value())
}

func evalHandlerStateComparison(params any, handlerState map[string]any, cmp func(any, any) bool) bool {
	m, ok := params.(map[string]any)
	if !ok {
		return false
	}
	for k, expected := range m {
		actual, exists := handlerState[k]
		if !exists {
			if !cmp(nil, expected) {
				return false
			}
			continue
		}
		if !cmp(actual, expected) {
			return false
		}
	}
	return true
}

func toFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return 0
	}
}

func anyToStringMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func evaluateFilter(filter map[string]any, event any, tc *TemplateContext) bool {
	if len(filter) == 0 {
		return true
	}
	switch e := event.(type) {
	case *EventEntityZoneActivity:
		if zone, ok := filter["zone"]; ok {
			expected := tc.Resolve(fmt.Sprintf("%v", zone))
			if e.ZoneId != expected {
				return false
			}
		}
		if entity, ok := filter["entity"]; ok {
			expected := tc.Resolve(fmt.Sprintf("%v", entity))
			if e.EntityId != expected {
				return false
			}
		}
		if entering, ok := filter["entering"]; ok {
			if b, ok := entering.(bool); ok && b != e.IsEntering {
				return false
			}
		}
		return true
	case *EventBroadcast:
		if id, ok := filter["id"]; ok {
			expected := tc.Resolve(fmt.Sprintf("%v", id))
			if e.Id != expected {
				return false
			}
		}
		return true
	case *EventTimerComplete:
		if createdBy, ok := filter["created_by"]; ok {
			expected := tc.Resolve(fmt.Sprintf("%v", createdBy))
			if e.CreatedBy != expected {
				return false
			}
		}
		if timerId, ok := filter["timer_id"]; ok {
			expected := tc.Resolve(fmt.Sprintf("%v", timerId))
			if e.TimerId != expected {
				return false
			}
		}
		return true
	case *EventGlobalVariableUpdated:
		if key, ok := filter["key"]; ok {
			expected := tc.Resolve(fmt.Sprintf("%v", key))
			if e.Key != expected {
				return false
			}
		}
		return true
	case *EventCombatComplete:
		if combatId, ok := filter["combat_id"]; ok {
			expected := tc.Resolve(fmt.Sprintf("%v", combatId))
			if e.CombatId != expected {
				return false
			}
		}
		return true
	case *EventScriptedMotionComplete:
		if entityId, ok := filter["entity"]; ok {
			expected := tc.Resolve(fmt.Sprintf("%v", entityId))
			if e.EntityId != expected {
				return false
			}
		}
		if motionId, ok := filter["motion_id"]; ok {
			expected := tc.Resolve(fmt.Sprintf("%v", motionId))
			if e.MotionId != expected {
				return false
			}
		}
		if wasCanceled, ok := filter["was_canceled"]; ok {
			if b, ok := wasCanceled.(bool); ok && b != e.WasCanceled {
				return false
			}
		}
		return true
	}
	return true
}
