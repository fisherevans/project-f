package adventure

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func init() {
	registerStepConverter("custom_action", func(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
		return convertCustomActionStep(step.Params, tc, sequences)
	})
}

const maxCustomActionDepth = 10

func convertCustomActionStep(params any, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
	m, ok := params.(map[string]any)
	if !ok {
		name, ok := params.(string)
		if !ok {
			log.Warn().Msg("custom_action requires a map with 'name' or a string name")
			return nil
		}
		m = map[string]any{"name": name}
	}
	name, _ := m["name"].(string)
	if name == "" {
		log.Warn().Msg("custom_action requires a 'name'")
		return nil
	}

	actionDef, ok := scriptCustomActions[name]
	if !ok {
		log.Warn().Str("name", name).Msg("unknown custom action")
		return nil
	}

	withMap, _ := m["params"].(map[string]any)

	scope := &ReturnScope{Id: nextScopeId("custom_action")}

	parentScope := tc.returnScope

	return []Effect{&EffectDeferredBatch{
		ScopeId: scope.Id,
		BuildEffects: func(source EntityReader, s *State) []Effect {
			if parentScope.isReturned() {
				return nil
			}
			depth := tc.getCustomActionDepth()
			if depth >= maxCustomActionDepth {
				log.Warn().Str("name", name).Int("depth", depth).Msg("custom action recursion limit reached")
				return nil
			}

			evaluatedParams := make(map[string]any)
			for _, paramDef := range actionDef.Params {
				if paramDef.Default != nil {
					evaluatedParams[paramDef.Name] = paramDef.Default
				}
			}

			env := tc.getExprEnv()
			for k, v := range withMap {
				exprStr := fmt.Sprintf("%v", v)
				prog, err := CompileExpr(exprStr)
				if err != nil {
					log.Warn().Err(err).Str("param", k).Str("expr", exprStr).Msg("custom_action: failed to compile param expression")
					evaluatedParams[k] = v
					continue
				}
				result, err := EvalExpr(prog, env)
				if err != nil {
					log.Warn().Err(err).Str("param", k).Str("expr", exprStr).Msg("custom_action: failed to evaluate param expression")
					evaluatedParams[k] = v
					continue
				}
				evaluatedParams[k] = result
			}

			childTC := tc.withCustomActionParams(evaluatedParams)
			childTC.returnScope = scope
			return convertSteps(actionDef.Steps, childTC, sequences)
		},
	}}
}
