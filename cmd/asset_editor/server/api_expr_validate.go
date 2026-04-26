package server

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/expr-lang/expr"
)

type exprEnv struct {
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

var exprCompileOpts = buildExprCompileOpts()

func buildExprCompileOpts() []expr.Option {
	return []expr.Option{
		expr.Env(exprEnv{}),
		expr.Function("min", func(params ...any) (any, error) {
			return math.Min(toFloat(params[0]), toFloat(params[1])), nil
		}, new(func(any, any) float64)),
		expr.Function("max", func(params ...any) (any, error) {
			return math.Max(toFloat(params[0]), toFloat(params[1])), nil
		}, new(func(any, any) float64)),
		expr.Function("clamp", func(params ...any) (any, error) {
			return math.Max(toFloat(params[1]), math.Min(toFloat(params[0]), toFloat(params[2]))), nil
		}, new(func(any, any, any) float64)),
		expr.Function("str", func(params ...any) (any, error) {
			return fmt.Sprintf("%v", params[0]), nil
		}, new(func(any) string)),
		expr.Function("int", func(params ...any) (any, error) {
			return int(toFloat(params[0])), nil
		}, new(func(any) int)),
		expr.Function("float", func(params ...any) (any, error) {
			return toFloat(params[0]), nil
		}, new(func(any) float64)),
		expr.Function("rand", func(params ...any) (any, error) {
			return 0, nil
		}, new(func(any) int)),
		expr.Function("randf", func(params ...any) (any, error) {
			return 0.0, nil
		}, new(func() float64)),
		expr.Function("keys", func(params ...any) (any, error) {
			return []any{}, nil
		}, new(func(any) []any)),
		expr.Function("values", func(params ...any) (any, error) {
			return []any{}, nil
		}, new(func(any) []any)),
		expr.Function("hasKey", func(params ...any) (any, error) {
			return false, nil
		}, new(func(any, any) bool)),
	}
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case float64:
		return n
	default:
		return 0
	}
}

const maxExprValidateBody = 4096

func (s *Server) handleValidateExpr(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	limited := io.LimitReader(r.Body, maxExprValidateBody)
	if err := json.NewDecoder(limited).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Expression == "" {
		writeJSON(w, map[string]any{"valid": true})
		return
	}
	_, err := expr.Compile(req.Expression, exprCompileOpts...)
	if err != nil {
		writeJSON(w, map[string]any{"valid": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"valid": true})
}
