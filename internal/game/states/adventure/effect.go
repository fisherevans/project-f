package adventure

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Effect is the interface that all effect types must implement
type Effect interface {
	FillDefaultsAndValidate() error
	Process(source EntityReader, s *State) bool
	CompletionID() string
}

type instantEffect struct{}

func (_ instantEffect) CompletionID() string {
	return ""
}

func logEffectInfof(source EntityReader, e any, messageFormat string, args ...any) {
	logEffect(zerolog.InfoLevel, source, e, messageFormat, args...)
}

func logEffectWarnf(source EntityReader, e any, messageFormat string, args ...any) {
	logEffect(zerolog.WarnLevel, source, e, messageFormat, args...)
}

func logEffect(level zerolog.Level, source EntityReader, e any, messageFormat string, args ...any) {
	log.WithLevel(level).Str("caller", source.GetId()).Interface("e", e).Msgf(messageFormat, args...)
}

type RunnableFunction func(s *State)

func (RunnableFunction) String() string {
	return "<inline function>"
}

type EffectFunction struct {
	instantEffect
	Fn RunnableFunction
}

func (e *EffectFunction) Process(source EntityReader, s *State) bool {
	e.Fn(s)
	return true
}

type EffectTimer struct {
	TimerId         string `auto_generate:"true"`
	DurationSeconds float64
}

func (e *EffectTimer) CompletionID() string {
	if e.TimerId == "" {
		return ""
	}
	return "timer:" + e.TimerId
}

func (e *EffectTimer) Process(source EntityReader, s *State) bool {
	s.timers.AddTimer(source.GetId(), e.TimerId, e.CompletionID(), e.DurationSeconds)
	return true
}

type EffectSetWorldState struct {
	instantEffect
	Key   string
	Value any
}

func (e *EffectSetWorldState) Process(source EntityReader, s *State) bool {
	s.globals.Set(e.Key, e.Value)
	return true
}

type EffectSetRunState struct {
	instantEffect
	Key   string
	Value any
}

func (e *EffectSetRunState) Process(source EntityReader, s *State) bool {
	s.globals.Set(e.Key, e.Value)
	return true
}

type EffectBatch struct {
	BatchId           string `auto_generate:"true"`
	Effects           []Effect
	ExecuteInParallel *bool // defaults to false (serial)
}

func (e *EffectBatch) CompletionID() string {
	if e.BatchId == "" {
		return ""
	}
	return "batch:" + e.BatchId
}

func (e *EffectBatch) Process(source EntityReader, s *State) bool {
	s.planExecutor.StartPlan(source, e)
	return true
}

func NewSerialPlan(effects ...Effect) *EffectBatch {
	return &EffectBatch{
		Effects: effects,
	}
}

func NewParallelPlan(effects ...Effect) *EffectBatch {
	parallel := true
	return &EffectBatch{
		Effects:           effects,
		ExecuteInParallel: &parallel,
	}
}

func (b *EffectBatch) WithId(id string) *EffectBatch {
	b.BatchId = id
	return b
}

func (b *EffectBatch) IsParallel() bool {
	return b.ExecuteInParallel != nil && *b.ExecuteInParallel
}

type EffectWaitForCondition struct {
	ConditionId string `auto_generate:"true"`
	Check       ConditionCheck
}

func (e *EffectWaitForCondition) CompletionID() string {
	if e.ConditionId == "" {
		return ""
	}
	return "condition:" + e.ConditionId
}

func (e *EffectWaitForCondition) Process(source EntityReader, s *State) bool {
	s.conditions.AddCondition(e.ConditionId, e.CompletionID(), e.Check)
	return true
}

type EffectSendEvent struct {
	instantEffect
	Event any
}

func (e *EffectSendEvent) Process(source EntityReader, s *State) bool {
	if e == nil {
		return false
	}
	s.eventDispatcher.Dispatch(e.Event)
	return true
}

type EffectSendBroadcast struct {
	instantEffect
	BroadcastId string
	Data        any
}

func (e *EffectSendBroadcast) Process(source EntityReader, s *State) bool {
	if e == nil {
		return false
	}
	s.eventDispatcher.Dispatch(NewEventBroadcast(e.BroadcastId, e.Data))
	return true
}
