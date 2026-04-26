package adventure

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func evaluateCondition(node *ConditionNode, globals StateGlobalsReader, handlerState map[string]any, tc ...*TemplateContext) bool {
	if node == nil {
		return true
	}
	var templateCtx *TemplateContext
	if len(tc) > 0 {
		templateCtx = tc[0]
	}
	switch node.Kind {
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
			if !evaluateCondition(sub, globals, handlerState, templateCtx) {
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
				return false
			}
			if evaluateCondition(sub, globals, handlerState, templateCtx) {
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
		return !evaluateCondition(sub, globals, handlerState, templateCtx)
	case "prop_exists":
		m := anyToStringMap(node.Params)
		key, _ := m["key"].(string)
		if key == "" {
			key, _ = m["value"].(string)
		}
		if templateCtx == nil || templateCtx.Properties == nil {
			return false
		}
		_, exists := templateCtx.Properties[key]
		return exists
	case "prop_eq":
		m := anyToStringMap(node.Params)
		key, _ := m["key"].(string)
		expected := m["value"]
		if templateCtx == nil || templateCtx.Properties == nil {
			return false
		}
		actual, exists := templateCtx.Properties[key]
		if !exists {
			return expected == nil
		}
		return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
	case "entity_idle":
		if templateCtx == nil {
			return false
		}
		entity, ok := globals.GetEntityReader(templateCtx.SelfId)
		if !ok {
			return false
		}
		return !entity.IsMoving() && !entity.HasPushedBehavior() && entity.IsBehaviorEnabled()
	case "expr":
		exprStr, ok := node.Params.(string)
		if !ok {
			log.Warn().Msg("expr condition requires a string expression")
			return false
		}
		if templateCtx == nil {
			return false
		}
		prog, err := CompileExpr(exprStr)
		if err != nil {
			log.Warn().Err(err).Str("expr", exprStr).Msg("failed to compile expr condition")
			return false
		}
		return EvalExprBool(prog, templateCtx.getExprEnv())

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
