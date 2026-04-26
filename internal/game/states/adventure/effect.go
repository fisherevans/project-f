package adventure

import (
	"fisherevans.com/project/f/internal/util/highlighter"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	setGlobalConverter := func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		return []Effect{NewSetGlobalEffect(mapStr(m, "key"), m["value"])}
	}
	registerStepConverter("set_global", setGlobalConverter)
	registerStepConverter("set_world_state", setGlobalConverter)
	registerStepConverter("set_run_state", setGlobalConverter)
	registerStepConverter("broadcast", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		var data any
		if d, ok := m["data"]; ok {
			data = d
		}
		return []Effect{NewSendBroadcastEffect(mapStr(m, "id"), data)}
	})
	registerStepConverter("wait_for", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		condName := mapStr(m, "condition")
		fn, ok := getScriptConditionFactory(condName)
		if !ok {
			log.Warn().Str("condition", condName).Msg("unknown script condition in wait_for")
			return nil
		}
		return []Effect{NewWaitForConditionEffect(fn(m))}
	})
	registerStepConverter("wait_for_animation", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		return []Effect{NewWaitForAnimationComplete(resolveString(step.Params, tc))}
	})
	registerStepConverter("tooltip", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		return []Effect{NewPushTooltipEffect(resolveString(step.Params, tc))}
	})
	registerStepConverter("highlight_sequence", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		targetsRaw, ok := m["targets"].([]any)
		if !ok {
			log.Warn().Msg("highlight_sequence: targets must be a list")
			return nil
		}
		var targets []highlighter.Target
		for _, raw := range targetsRaw {
			tm, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			regionMap, _ := tm["region"].(map[string]any)
			area := pixel.R(
				toFloat64(regionMap["x"]),
				toFloat64(regionMap["y"]),
				toFloat64(regionMap["x"])+toFloat64(regionMap["w"]),
				toFloat64(regionMap["y"])+toFloat64(regionMap["h"]),
			)
			t := highlighter.NewTarget(area)
			if msgMap, ok := tm["message"].(map[string]any); ok {
				text, _ := msgMap["text"].(string)
				placement := parseMessagePlacement(mapStr(msgMap, "placement"))
				msg := highlighter.NewMessage(tc.Resolve(text), placement)
				if wrap := mapInt(msgMap, "wrap", 0); wrap > 0 {
					msg = msg.Wrapped(wrap)
				}
				t = t.WithMessage(msg)
			}
			if badgeMap, ok := tm["badge"].(map[string]any); ok {
				placement := parseBadgePlacement(mapStr(badgeMap, "placement"))
				badge := highlighter.NewBadge(placement)
				if label := mapStr(badgeMap, "label"); label != "" {
					badge = badge.WithLabel(tc.Resolve(label))
				}
				t = t.WithBadge(badge)
			}
			if noPad, ok := tm["no_padding"].(bool); ok && noPad {
				t = t.NoPadding()
			}
			targets = append(targets, t)
		}
		if len(targets) == 0 {
			return nil
		}
		return []Effect{NewSetHighlightSequenceEffect(targets)}
	})
}

func parseMessagePlacement(s string) highlighter.MessagePlacement {
	switch s {
	case "top":
		return highlighter.MessageOnTop
	case "left":
		return highlighter.MessageOnLeft
	case "right":
		return highlighter.MessageOnRight
	default:
		return highlighter.MessageOnBottom
	}
}

func parseBadgePlacement(s string) highlighter.BadgePlacement {
	switch s {
	case "top_right":
		return highlighter.BadgeInTopRight
	case "bottom_right":
		return highlighter.BadgeInBottomRight
	case "top_middle":
		return highlighter.BadgeOnTopMiddle
	default:
		return highlighter.BadgeOnBottomMiddle
	}
}

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
	callerId := "<system>"
	if source != nil {
		callerId = source.GetId()
	}
	log.WithLevel(level).Str("caller", callerId).Interface("e", e).Msgf(messageFormat, args...)
}

type RunnableFunction func(source EntityReader, s *State)

func (RunnableFunction) String() string {
	return "<inline function>"
}

type EffectFunction struct {
	instantEffect
	Fn RunnableFunction
}

func (e *EffectFunction) Process(source EntityReader, s *State) bool {
	e.Fn(source, s)
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

type EffectSetGlobal struct {
	instantEffect
	Key   string
	Value any
}

func (e *EffectSetGlobal) Process(source EntityReader, s *State) bool {
	s.globals.Set(e.Key, e.Value)
	return true
}

type EffectBatch struct {
	BatchId           string `auto_generate:"true"`
	Effects           []Effect
	ExecuteInParallel *bool // defaults to false (serial)
	ScopeId           string
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

func NewWaitForAnimationComplete(entityId string) *EffectWaitForCondition {
	return NewWaitForConditionEffect(func(s *State, td float64) bool {
		e, r, ok := GetModeBasedRenderer(s, entityId)
		if !ok {
			log.Fatal().Msgf("failed to get renderer for %s", entityId)
		}
		return r.getBasicEntityRenderer(ModeMetadataKey.Get(e)).AreAnimationsComplete()
	})
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

type EffectPushTooltip struct {
	instantEffect
	Message string
}

func (e *EffectPushTooltip) Process(source EntityReader, s *State) bool {
	if e == nil {
		return false
	}
	s.tooltips.Add(e.Message)
	return true
}

type EffectSetHighlightSequence struct {
	instantEffect
	Targets []highlighter.Target
}

func (e *EffectSetHighlightSequence) Process(source EntityReader, s *State) bool {
	if e == nil {
		return false
	}
	s.highlighter.AppendTargets(e.Targets...)
	return true
}
