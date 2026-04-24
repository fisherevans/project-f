package adventure


type ScriptHandler struct {
	entityId     string
	def          *HandlerDef
	sequences    map[string]*SequenceDef
	handlerState map[string]any
}

func NewScriptHandler(entityId string, def *HandlerDef, sequences map[string]*SequenceDef) *ScriptHandler {
	return &ScriptHandler{
		entityId:     entityId,
		def:          def,
		sequences:    sequences,
		handlerState: make(map[string]any),
	}
}

func newScriptHandlerFactory(def *HandlerDef, sequences map[string]*SequenceDef) *ScriptHandler {
	return &ScriptHandler{
		def:          def,
		sequences:    sequences,
		handlerState: make(map[string]any),
	}
}

func (h *ScriptHandler) templateContext(source EntityReader, globals StateGlobalsReader) *TemplateContext {
	playerId := globals.Get(globalVariableNamePlayerId).AsString("")
	sourceId := ""
	if source != nil {
		sourceId = source.GetId()
	}
	return &TemplateContext{
		SelfId:   h.entityId,
		PlayerId: playerId,
		SourceId: sourceId,
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
		if !evaluateCondition(rule.When, globals, h.handlerState) {
			continue
		}

		effects := convertSteps(rule.Steps, tc, h.sequences)
		if len(effects) == 0 && len(rule.SetState) == 0 {
			continue
		}

		if len(rule.SetState) > 0 {
			for k, v := range rule.SetState {
				h.handlerState[k] = v
			}
		}

		output := NewOutput()
		if len(effects) > 0 {
			output = output.WithSerialPlan(effects...)
		}
		output = output.WithState(h.handlerState)
		return output
	}
	return nil
}

