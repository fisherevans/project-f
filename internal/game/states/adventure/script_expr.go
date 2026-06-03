package adventure

import (
	"fisherevans.com/project/f/internal/util/rng"
	"fmt"
	"math"
	"regexp"
	"strings"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/rs/zerolog/log"
)

var interpolationPattern = regexp.MustCompile(`\{\{(.+?)\}\}`)

type ExprEnv struct {
	Var    map[string]any `expr:"var"`
	Global map[string]any `expr:"global"`
	Const  map[string]any `expr:"const"`
	Save   map[string]any `expr:"save"`
	Prop   map[string]any `expr:"prop"`
	Param  map[string]any `expr:"param"`
	Self   string         `expr:"self"`
	Player string         `expr:"player"`
	Source string         `expr:"source"`
}

var exprFunctions = []expr.Option{
	// len() is built-in to expr

	expr.Function("min", func(params ...any) (any, error) {
		if len(params) != 2 {
			return nil, fmt.Errorf("min() requires exactly 2 arguments")
		}
		a, b := toExprFloat(params[0]), toExprFloat(params[1])
		return math.Min(a, b), nil
	}, new(func(any, any) float64)),

	expr.Function("max", func(params ...any) (any, error) {
		if len(params) != 2 {
			return nil, fmt.Errorf("max() requires exactly 2 arguments")
		}
		a, b := toExprFloat(params[0]), toExprFloat(params[1])
		return math.Max(a, b), nil
	}, new(func(any, any) float64)),

	expr.Function("clamp", func(params ...any) (any, error) {
		if len(params) != 3 {
			return nil, fmt.Errorf("clamp() requires exactly 3 arguments")
		}
		v, lo, hi := toExprFloat(params[0]), toExprFloat(params[1]), toExprFloat(params[2])
		return math.Max(lo, math.Min(v, hi)), nil
	}, new(func(any, any, any) float64)),

	expr.Function("str", func(params ...any) (any, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("str() requires exactly 1 argument")
		}
		return fmt.Sprintf("%v", params[0]), nil
	}, new(func(any) string)),

	expr.Function("int", func(params ...any) (any, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("int() requires exactly 1 argument")
		}
		return int(toExprFloat(params[0])), nil
	}, new(func(any) int)),

	expr.Function("float", func(params ...any) (any, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("float() requires exactly 1 argument")
		}
		return toExprFloat(params[0]), nil
	}, new(func(any) float64)),

	expr.Function("rand", func(params ...any) (any, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("rand() requires exactly 1 argument")
		}
		n := int(toExprFloat(params[0]))
		if n <= 0 {
			return 0, nil
		}
		return rng.IntN(n), nil
	}, new(func(any) int)),

	expr.Function("randf", func(params ...any) (any, error) {
		return rng.Float64(), nil
	}, new(func() float64)),

	// Note: contains/startsWith/endsWith are built-in operators in expr:
	//   "hello" contains "ell"
	//   "hello" startsWith "hel"
	//   "hello" endsWith "llo"
	// len() is also built-in.

	expr.Function("keys", func(params ...any) (any, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("keys() requires exactly 1 argument")
		}
		m, ok := params[0].(map[string]any)
		if !ok {
			return []any{}, nil
		}
		result := make([]any, 0, len(m))
		for k := range m {
			result = append(result, k)
		}
		return result, nil
	}, new(func(any) []any)),

	expr.Function("values", func(params ...any) (any, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("values() requires exactly 1 argument")
		}
		m, ok := params[0].(map[string]any)
		if !ok {
			return []any{}, nil
		}
		result := make([]any, 0, len(m))
		for _, v := range m {
			result = append(result, v)
		}
		return result, nil
	}, new(func(any) []any)),

	expr.Function("hasKey", func(params ...any) (any, error) {
		if len(params) != 2 {
			return nil, fmt.Errorf("hasKey() requires exactly 2 arguments")
		}
		m, ok := params[0].(map[string]any)
		if !ok {
			return false, nil
		}
		key, ok := params[1].(string)
		if !ok {
			return false, nil
		}
		_, exists := m[key]
		return exists, nil
	}, new(func(any, any) bool)),
}

func compileExprOptions() []expr.Option {
	opts := make([]expr.Option, 0, len(exprFunctions)+1)
	opts = append(opts, expr.Env(ExprEnv{}))
	opts = append(opts, exprFunctions...)
	return opts
}

var compiledExprOptions = compileExprOptions()

func CompileExpr(src string) (*vm.Program, error) {
	return expr.Compile(src, compiledExprOptions...)
}

func EvalExpr(prog *vm.Program, env *ExprEnv) (any, error) {
	return expr.Run(prog, env)
}

func EvalExprString(prog *vm.Program, env *ExprEnv) string {
	result, err := expr.Run(prog, env)
	if err != nil {
		log.Warn().Err(err).Msg("expr eval failed (string)")
		return ""
	}
	return fmt.Sprintf("%v", result)
}

func EvalExprBool(prog *vm.Program, env *ExprEnv) bool {
	result, err := expr.Run(prog, env)
	if err != nil {
		log.Warn().Err(err).Msg("expr eval failed (bool)")
		return false
	}
	switch v := result.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	case nil:
		return false
	default:
		return true
	}
}

type compiledSegment struct {
	literal string
	program *vm.Program
}

type CompiledInterpolatedString struct {
	segments []compiledSegment
	raw      string
}

func CompileInterpolatedString(s string) (*CompiledInterpolatedString, error) {
	matches := interpolationPattern.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return &CompiledInterpolatedString{
			segments: []compiledSegment{{literal: s}},
			raw:      s,
		}, nil
	}

	var segments []compiledSegment
	lastEnd := 0
	for _, match := range matches {
		if match[0] > lastEnd {
			segments = append(segments, compiledSegment{literal: s[lastEnd:match[0]]})
		}
		exprSrc := s[match[2]:match[3]]
		prog, err := CompileExpr(exprSrc)
		if err != nil {
			return nil, fmt.Errorf("compile expression %q in %q: %w", exprSrc, s, err)
		}
		segments = append(segments, compiledSegment{program: prog})
		lastEnd = match[1]
	}
	if lastEnd < len(s) {
		segments = append(segments, compiledSegment{literal: s[lastEnd:]})
	}
	return &CompiledInterpolatedString{segments: segments, raw: s}, nil
}

func (cis *CompiledInterpolatedString) Eval(env *ExprEnv) string {
	if len(cis.segments) == 1 && cis.segments[0].program == nil {
		return cis.segments[0].literal
	}
	var b strings.Builder
	for _, seg := range cis.segments {
		if seg.program == nil {
			b.WriteString(seg.literal)
		} else {
			b.WriteString(EvalExprString(seg.program, env))
		}
	}
	return b.String()
}

func InterpolateString(s string, env *ExprEnv) string {
	if !strings.Contains(s, "{{") {
		return s
	}
	return interpolationPattern.ReplaceAllStringFunc(s, func(match string) string {
		exprSrc := match[2 : len(match)-2]
		prog, err := CompileExpr(exprSrc)
		if err != nil {
			log.Warn().Err(err).Str("expr", exprSrc).Msg("failed to compile interpolation")
			return match
		}
		return EvalExprString(prog, env)
	})
}

func InterpolateAny(v any, env *ExprEnv) any {
	switch val := v.(type) {
	case string:
		return InterpolateString(val, env)
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v := range val {
			out[InterpolateString(k, env)] = InterpolateAny(v, env)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v := range val {
			out[i] = InterpolateAny(v, env)
		}
		return out
	default:
		return v
	}
}

func NewExprEnv(tc *TemplateContext, handlerState map[string]any, consts map[string]any) *ExprEnv {
	env := &ExprEnv{
		Var:    handlerState,
		Const:  consts,
		Prop:   tc.Properties,
		Self:   tc.SelfId,
		Player: tc.PlayerId,
		Source: tc.SourceId,
	}
	if env.Var == nil {
		env.Var = map[string]any{}
	}
	if env.Const == nil {
		env.Const = map[string]any{}
	}
	if env.Prop == nil {
		env.Prop = map[string]any{}
	}

	// Build global map from globals reader
	env.Global = buildGlobalMap(tc.Globals)

	env.Save = buildSaveMap(game.CurrentSaveOrNil())

	// Build param map - custom action params take priority
	if tc.customActionParams != nil {
		env.Param = tc.customActionParams
	} else if tc.Params != nil {
		env.Param = make(map[string]any, len(tc.Params))
		for k, v := range tc.Params {
			cleanKey := strings.TrimPrefix(k, "prop.")
			env.Param[cleanKey] = v
		}
	}
	if env.Param == nil {
		env.Param = map[string]any{}
	}

	return env
}

func buildGlobalMap(globals StateGlobalsReader) map[string]any {
	if globals == nil {
		return map[string]any{}
	}
	result := map[string]any{}
	for _, key := range globals.KeysWithPrefix("") {
		gv := globals.Get(key)
		if gv.Exists() {
			result[key] = gv.Value()
		}
	}
	return result
}

func buildSaveMap(save *rpg.GameSave) map[string]any {
	if save == nil {
		return map[string]any{}
	}
	m := map[string]any{
		"character_name": save.CharacterName,
		"save_id":        save.SaveId,
	}
	if save.Animech != nil {
		animech := map[string]any{
			"experience":     save.Animech.Experience,
			"current_sync":   save.Animech.CurrentSync,
			"current_shield": save.Animech.CurrentShield,
		}
		if save.Animech.Upgrades != nil {
			animech["level"] = save.Animech.Upgrades.GetLevel()
			animech["shield_level"] = save.Animech.Upgrades.ShieldLevel
			animech["sync_level"] = save.Animech.Upgrades.SyncLevel
			animech["max_shield"] = save.Animech.GetMaxShield()
			animech["max_sync"] = save.Animech.GetMaxSync()
		}
		if save.Animech.SkillSet != nil {
			skills := map[string]any{
				"up":    string(save.Animech.SkillSet.Skill1),
				"down":  string(save.Animech.SkillSet.Skill2),
				"left":  string(save.Animech.SkillSet.Skill3),
				"right": string(save.Animech.SkillSet.Skill4),
			}
			animech["skills"] = skills
		}
		m["animech"] = animech
	}
	unlockedSkills := make([]any, 0)
	for id := range save.ControlledUnlockedSkills {
		unlockedSkills = append(unlockedSkills, string(id))
	}
	m["unlocked_skills"] = unlockedSkills
	return m
}

func toExprFloat(v any) float64 {
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
	case string:
		return 0
	case bool:
		if val {
			return 1
		}
		return 0
	default:
		return 0
	}
}
