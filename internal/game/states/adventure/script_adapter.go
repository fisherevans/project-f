package adventure

import "fmt"

type ScriptHandler struct {
	entityId     string
	def          *HandlerDef
	sequences    map[string]*SequenceDef
	handlerState map[string]any
	propParams   map[string]string
	properties   map[string]any
}

func NewScriptHandler(entityId string, def *HandlerDef, sequences map[string]*SequenceDef) *ScriptHandler {
	return &ScriptHandler{
		entityId:     entityId,
		def:          def,
		sequences:    sequences,
		handlerState: make(map[string]any),
	}
}

func newScriptHandlerFactory(def *HandlerDef, sequences map[string]*SequenceDef, propParams map[string]string, properties map[string]any) *ScriptHandler {
	mergedProps := applyPropDefaults(def.Props, properties)
	mergedPropParams := applyPropDefaultsToParams(def.Props, propParams)
	return &ScriptHandler{
		def:          def,
		sequences:    sequences,
		handlerState: make(map[string]any),
		propParams:   mergedPropParams,
		properties:   mergedProps,
	}
}

func applyPropDefaults(propDefs []*HandlerPropDef, properties map[string]any) map[string]any {
	if len(propDefs) == 0 {
		return properties
	}
	merged := make(map[string]any)
	for _, pd := range propDefs {
		if pd.Default != nil {
			merged[pd.Name] = pd.Default
		}
	}
	for k, v := range properties {
		merged[k] = v
	}
	return merged
}

func applyPropDefaultsToParams(propDefs []*HandlerPropDef, params map[string]string) map[string]string {
	if len(propDefs) == 0 {
		return params
	}
	merged := make(map[string]string)
	for _, pd := range propDefs {
		if pd.Default != nil {
			merged["prop."+pd.Name] = fmt.Sprintf("%v", pd.Default)
		}
	}
	for k, v := range params {
		merged[k] = v
	}
	return merged
}

func (h *ScriptHandler) templateContext(source EntityReader, globals StateGlobalsReader) *TemplateContext {
	playerId := globals.Get(globalVariableNamePlayerId).AsString("")
	sourceId := ""
	if source != nil {
		sourceId = source.GetId()
	}
	return &TemplateContext{
		SelfId:       h.entityId,
		PlayerId:     playerId,
		SourceId:     sourceId,
		Params:       h.propParams,
		Properties:   h.properties,
		Globals:      globals,
		handlerState: h.handlerState,
	}
}

func (h *ScriptHandler) Init(thisEntity EntityReader, globals StateGlobalsReader, state any) *HandlerOutput {
	if h.entityId == "" && thisEntity != nil {
		h.entityId = thisEntity.GetId()
	}
	if state != nil {
		if s, ok := state.(map[string]any); ok {
			h.handlerState = s
		}
	}
	// Initialize var defaults if handler defines them and state is fresh
	if h.def.Var != nil && len(h.handlerState) == 0 {
		for k, v := range h.def.Var {
			h.handlerState[k] = deepCloneAny(v)
		}
	}
	if len(h.def.OnInit) == 0 {
		return nil
	}
	tc := h.templateContext(thisEntity, globals)
	return h.processRules(h.def.OnInit, nil, tc, globals)
}

func (h *ScriptHandler) HandleEvent(thisEntity EntityReader, globals StateGlobalsReader, state any, event any) *HandlerOutput {
	if h.entityId == "" && thisEntity != nil {
		h.entityId = thisEntity.GetId()
	}
	if state != nil {
		if s, ok := state.(map[string]any); ok {
			h.handlerState = s
		}
	}

	var rules []*RuleDef
	tc := h.templateContext(thisEntity, globals)

	switch e := event.(type) {
	case *EventOnInteract:
		if e.TargetId == h.entityId && len(h.def.OnInteractSelf) > 0 {
			tc.SourceId = e.SourceId
			if out := h.processRules(h.def.OnInteractSelf, event, tc, globals); out != nil {
				return out
			}
		}
		if len(h.def.OnInteract) > 0 {
			tc.SourceId = e.SourceId
			rules = h.def.OnInteract
		}
	case *EventEntityZoneActivity:
		rules = h.def.OnZoneActivity
	case *EventBroadcast:
		rules = h.def.OnBroadcast
	case *EventGlobalVariableUpdated:
		rules = h.def.OnGlobalUpdated
	case *EventOnStateEnter:
		rules = h.def.OnStateEnter
	case *EventCombatComplete:
		rules = h.def.OnCombatComplete
	case *EventTimerComplete:
		rules = h.def.OnTimerComplete
	case *EventScriptedMotionComplete:
		if e.EntityId == h.entityId && len(h.def.OnMotionCompleteSelf) > 0 {
			if out := h.processRules(h.def.OnMotionCompleteSelf, event, tc, globals); out != nil {
				return out
			}
		}
		if len(h.def.OnMotionComplete) > 0 {
			rules = h.def.OnMotionComplete
		}
	}

	if len(rules) == 0 {
		return nil
	}
	return h.processRules(rules, event, tc, globals)
}

func (h *ScriptHandler) processRules(rules []*RuleDef, event any, tc *TemplateContext, globals StateGlobalsReader) *HandlerOutput {
	for _, rule := range rules {
		if !evaluateFilter(rule.Filter, event, tc) {
			continue
		}
		if !evaluateCondition(rule.When, globals, h.handlerState, tc) {
			continue
		}

		scope := &ReturnScope{Id: nextScopeId("rule")}
		tc.returnScope = scope
		effects := convertSteps(rule.Steps, tc, h.sequences)
		if len(effects) == 0 {
			continue
		}

		output := NewOutput()
		if len(effects) > 0 {
			plan := NewSerialPlan(effects...)
			plan.ScopeId = scope.Id
			output = output.WithEffects(plan)
		}
		output = output.WithState(h.handlerState)
		return output
	}
	return nil
}

func deepCloneAny(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v := range val {
			out[k] = deepCloneAny(v)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v := range val {
			out[i] = deepCloneAny(v)
		}
		return out
	default:
		return v
	}
}

