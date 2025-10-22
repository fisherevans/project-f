package events

import (
	"fmt"
	"strings"
)

type Effect struct {
	Dialogue      *EffectDialogue
	Timer         *EffectTimer
	YieldElythium *EffectYieldElythium
	MutateEntity  *EffectMutateEntity
	Chatter       *EffectChatter
}

func (e Effect) Validate() error {
	issues := &issues{}
	effectCount := 0
	if e.Dialogue != nil {
		effectCount++
		issues.sub("dialogue").
			requireString("text", e.Dialogue.Text)
	}
	if e.Chatter != nil {
		effectCount++
		issues.sub("chatter").
			requireString("entityId", e.Chatter.EntityId).
			requireString("message", e.Chatter.Message).
			requirePositive("durationSeconds", e.Chatter.DurationSeconds)
	}
	if e.Timer != nil {
		effectCount++
		issues.sub("timer").
			requireString("timerId", e.Timer.TimerId).
			requirePositive("durationSeconds", e.Timer.DurationSeconds)
	}
	if e.YieldElythium != nil {
		effectCount++
		issues.sub("yieldElythium").
			requirePositive("amount", float64(e.YieldElythium.Amount))
	}
	if e.MutateEntity != nil {
		effectCount++
		issues.sub("mutateEntity").
			requireString("entityId", e.MutateEntity.EntityId)
	}
	issues.requirePositive("effectCount", float64(effectCount))
	return issues.toError()
}

type EffectDialogue struct {
	DialogueId string
	Text       string
}

type EffectChatter struct {
	ChatterId       string
	EntityId        string
	DurationSeconds float64
	Message         string
}

type EffectYieldElythium struct {
	Amount int
}

type EffectTimer struct {
	TimerId         string
	DurationSeconds float64
}

type EffectMutateEntity struct {
	EntityId   string
	Mode       *string
	IsPassable *bool

	// DynamicEntity changes
	DynamicAnimations map[string][]DynamicAnimationReference
	DynamicLights     map[string][]LightConfig
}

type DynamicAnimationReference struct {
	Tilesheet string
	Name      string
}

type LightConfig struct {
	Color    string
	Size     float64
	Modifier string
}

// UTILITIES

type issues struct {
	issueList  []string
	namePrefix string
}

func (i *issues) addf(name string, format string, args ...any) {
	i.issueList = append(i.issueList, fmt.Sprintf("%s: %s", i.namePrefix+name, fmt.Sprintf(format, args...)))
}

func (i *issues) toError() error {
	if len(i.issueList) == 0 {
		return nil
	}
	return fmt.Errorf("invalid effect: %v", strings.Join(i.issueList, ", "))
}

func (i *issues) sub(name string) *issues {
	return &issues{
		issueList:  i.issueList,
		namePrefix: i.namePrefix + name,
	}
}

func (i *issues) requireString(name string, value string) *issues {
	if value != "" {
		return i
	}
	i.addf(name, "is required")
	return i
}

func (i *issues) requirePositive(name string, value float64) *issues {
	if value > 0 {
		return i
	}
	i.addf(name, "must be positive, got %f", value)
	return i
}
