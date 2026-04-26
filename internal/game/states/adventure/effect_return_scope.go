package adventure

import (
	"fmt"
	"sync/atomic"
)

func init() {
	registerStepConverter("return", func(_ *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		return []Effect{newReturnEffect(tc.returnScope)}
	})
}

var scopeIdCounter uint64

func nextScopeId(prefix string) string {
	return fmt.Sprintf("%s-scope-%d", prefix, atomic.AddUint64(&scopeIdCounter, 1))
}

type ReturnScope struct {
	Id       string
	returned bool
}

func (s *ReturnScope) isReturned() bool {
	return s != nil && s.returned
}

var returnIdCounter uint64

type EffectReturn struct {
	ReturnId string
	scope    *ReturnScope
}

func newReturnEffect(scope *ReturnScope) *EffectReturn {
	return &EffectReturn{
		ReturnId: fmt.Sprintf("return-%d", atomic.AddUint64(&returnIdCounter, 1)),
		scope:    scope,
	}
}

func (e *EffectReturn) FillDefaultsAndValidate() error { return nil }

func (e *EffectReturn) CompletionID() string {
	return "return:" + e.ReturnId
}

func (e *EffectReturn) Process(source EntityReader, s *State) bool {
	if e.scope != nil {
		e.scope.returned = true
		s.planExecutor.AbortScope(e.scope.Id)
	}
	s.planExecutor.MarkComplete(e.CompletionID())
	return true
}
