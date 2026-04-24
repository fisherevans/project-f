package adventure

import (
	"fmt"
	"strings"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util"
)

type TemplateContext struct {
	SelfId   string
	PlayerId string
	SourceId string
	Params   map[string]string
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

func (tc *TemplateContext) Resolve(s string) string {
	s = strings.ReplaceAll(s, "{{self}}", tc.SelfId)
	s = strings.ReplaceAll(s, "{{player}}", tc.PlayerId)
	s = strings.ReplaceAll(s, "{{source}}", tc.SourceId)
	s = strings.ReplaceAll(s, "{{instance_id}}", game.InstanceId)
	for k, v := range tc.Params {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
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
