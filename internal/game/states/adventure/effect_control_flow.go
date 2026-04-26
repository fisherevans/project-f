package adventure

import (
	"github.com/rs/zerolog/log"
)

func init() {
	registerStepConverter("if", func(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
		return convertIfStep(step.Params, tc, sequences)
	})
	registerStepConverter("switch", func(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
		return convertSwitchStep(step.Params, tc, sequences)
	})
	registerStepConverter("while", func(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
		return convertWhileStep(step.Params, tc, sequences)
	})
	registerStepConverter("set_var", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		key := mapStr(m, "key")
		valueExpr := mapStr(m, "value")
		if key == "" || valueExpr == "" {
			log.Warn().Msg("set_var requires 'key' and 'value'")
			return nil
		}
		handlerState := tc.handlerState
		return []Effect{NewFunctionEffect(func(_ EntityReader, s *State) {
			prog, err := CompileExpr(valueExpr)
			if err != nil {
				log.Warn().Err(err).Str("expr", valueExpr).Msg("set_var: failed to compile value expression")
				return
			}
			env := tc.getExprEnv()
			env.Var = handlerState
			result, err := EvalExpr(prog, env)
			if err != nil {
				log.Warn().Err(err).Str("expr", valueExpr).Msg("set_var: failed to evaluate value expression")
				return
			}
			handlerState[key] = result
		})}
	})
}

func convertIfStep(params any, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
	m, ok := params.(map[string]any)
	if !ok {
		log.Warn().Msg("if step requires a map with 'when', 'then', and optional 'else'")
		return nil
	}
	whenExpr, _ := m["when"].(string)
	if whenExpr == "" {
		log.Warn().Msg("if step requires a 'when' expression")
		return nil
	}

	thenSteps := extractSubSteps(m["then"])
	elseSteps := extractSubSteps(m["else"])

	thenEffects := convertSteps(thenSteps, tc, sequences)
	elseEffects := convertSteps(elseSteps, tc, sequences)

	return []Effect{&EffectDeferredBatch{
		BuildEffects: func(source EntityReader, s *State) []Effect {
			if tc.returnScope.isReturned() {
				return nil
			}
			prog, err := CompileExpr(whenExpr)
			if err != nil {
				log.Warn().Err(err).Str("expr", whenExpr).Msg("if: failed to compile condition")
				return nil
			}
			env := tc.getExprEnv()
			if EvalExprBool(prog, env) {
				return thenEffects
			}
			return elseEffects
		},
	}}
}

func convertSwitchStep(params any, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
	m, ok := params.(map[string]any)
	if !ok {
		log.Warn().Msg("switch step requires a map with 'on' and 'cases'")
		return nil
	}
	onExpr, _ := m["on"].(string)
	if onExpr == "" {
		log.Warn().Msg("switch step requires an 'on' expression")
		return nil
	}

	type switchCase struct {
		valueExpr string
		effects   []Effect
	}

	var cases []switchCase
	if rawCases, ok := m["cases"].([]any); ok {
		for _, rawCase := range rawCases {
			caseMap, ok := rawCase.(map[string]any)
			if !ok {
				continue
			}
			valueExpr, _ := caseMap["value"].(string)
			steps := extractSubSteps(caseMap["steps"])
			effects := convertSteps(steps, tc, sequences)
			cases = append(cases, switchCase{valueExpr: valueExpr, effects: effects})
		}
	}

	defaultSteps := extractSubSteps(m["default"])
	defaultEffects := convertSteps(defaultSteps, tc, sequences)

	return []Effect{&EffectDeferredBatch{
		BuildEffects: func(source EntityReader, s *State) []Effect {
			if tc.returnScope.isReturned() {
				return nil
			}
			onProg, err := CompileExpr(onExpr)
			if err != nil {
				log.Warn().Err(err).Str("expr", onExpr).Msg("switch: failed to compile 'on' expression")
				return nil
			}
			env := tc.getExprEnv()
			onResult, err := EvalExpr(onProg, env)
			if err != nil {
				log.Warn().Err(err).Str("expr", onExpr).Msg("switch: failed to evaluate 'on' expression")
				return nil
			}
			onStr := ""
			if onResult != nil {
				onStr = EvalExprString(onProg, env)
			}

			for _, c := range cases {
				if c.valueExpr == "" {
					continue
				}
				valueProg, err := CompileExpr(c.valueExpr)
				if err != nil {
					log.Warn().Err(err).Str("expr", c.valueExpr).Msg("switch case: failed to compile value")
					continue
				}
				caseStr := EvalExprString(valueProg, env)
				if onStr == caseStr {
					return c.effects
				}
			}
			return defaultEffects
		},
	}}
}

func convertWhileStep(params any, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
	m, ok := params.(map[string]any)
	if !ok {
		log.Warn().Msg("while step requires a map with 'when' and 'steps'")
		return nil
	}
	whenExpr, _ := m["when"].(string)
	if whenExpr == "" {
		log.Warn().Msg("while step requires a 'when' expression")
		return nil
	}

	maxIterations := mapInt(m, "max", 100)
	if maxIterations > 1000 {
		maxIterations = 1000
	}

	bodySteps := extractSubSteps(m["steps"])

	return []Effect{&EffectDeferredBatch{
		BuildEffects: func(source EntityReader, s *State) []Effect {
			if tc.returnScope.isReturned() {
				return nil
			}
			prog, err := CompileExpr(whenExpr)
			if err != nil {
				log.Warn().Err(err).Str("expr", whenExpr).Msg("while: failed to compile condition")
				return nil
			}
			env := tc.getExprEnv()

			var allEffects []Effect
			for i := 0; i < maxIterations; i++ {
				if tc.returnScope.isReturned() {
					break
				}
				if !EvalExprBool(prog, env) {
					break
				}
				bodyEffects := convertSteps(bodySteps, tc, sequences)
				for _, effect := range bodyEffects {
					if fn, ok := effect.(*EffectFunction); ok {
						fn.Process(source, s)
						if tc.returnScope.isReturned() {
							break
						}
					} else if ret, ok := effect.(*EffectReturn); ok {
						allEffects = append(allEffects, ret)
						if ret.scope != nil {
							ret.scope.returned = true
						}
						break
					}
					allEffects = append(allEffects, effect)
				}
				env = tc.getExprEnv()
			}
			return allEffects
		},
	}}
}

func extractSubSteps(raw any) []*StepNode {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	var steps []*StepNode
	for _, item := range items {
		stepMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for k, v := range stepMap {
			steps = append(steps, &StepNode{Kind: k, Params: v})
		}
	}
	return steps
}
