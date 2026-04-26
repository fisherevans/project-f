package adventure

import (
	"fmt"
	"strings"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util"
)

type TemplateContext struct {
	SelfId     string
	PlayerId   string
	SourceId   string
	Params     map[string]string
	Properties map[string]any
	Globals    StateGlobalsReader

	handlerState       map[string]any
	customActionDepth  int
	customActionParams map[string]any
	returnScope        *ReturnScope
}

func templateContextFromProps(props *util.Properties) map[string]string {
	if props == nil {
		return nil
	}
	params := make(map[string]string)
	for k, v := range props.All() {
		params["prop."+k] = fmt.Sprintf("%v", v)
	}
	return params
}

func (tc *TemplateContext) getExprEnv() *ExprEnv {
	// Rebuilt each call - TemplateContext fields (SourceId) can change between uses
	return NewExprEnv(tc, tc.handlerState, scriptConsts)
}

func (tc *TemplateContext) Resolve(s string) string {
	if !strings.Contains(s, "{{") {
		return s
	}
	// First pass: legacy substitution for simple {{name}} patterns
	s = strings.ReplaceAll(s, "{{self}}", tc.SelfId)
	s = strings.ReplaceAll(s, "{{player}}", tc.PlayerId)
	s = strings.ReplaceAll(s, "{{source}}", tc.SourceId)
	s = strings.ReplaceAll(s, "{{instance_id}}", game.InstanceId)
	for k, v := range tc.Params {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	// If any {{...}} remain after legacy substitution, evaluate with expr engine
	if strings.Contains(s, "{{") {
		s = InterpolateString(s, tc.getExprEnv())
	}
	return s
}

func (tc *TemplateContext) getCustomActionDepth() int {
	return tc.customActionDepth
}

func (tc *TemplateContext) withCustomActionParams(params map[string]any) *TemplateContext {
	child := *tc
	child.customActionDepth = tc.customActionDepth + 1
	child.customActionParams = params
	child.returnScope = nil
	return &child
}

func (tc *TemplateContext) ResolveAny(v any) any {
	switch val := v.(type) {
	case string:
		return tc.Resolve(val)
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v := range val {
			out[tc.Resolve(k)] = tc.ResolveAny(v)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v := range val {
			out[i] = tc.ResolveAny(v)
		}
		return out
	default:
		return v
	}
}
