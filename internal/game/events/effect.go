package events

import (
	"fmt"
	"strings"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
)

type Effect struct {
	Function          *EffectFunction
	Dialogue          *EffectDialogue
	Timer             *EffectTimer
	YieldElythium     *EffectYieldElythium
	MutateEntity      *EffectMutateEntity
	Chatter           *EffectChatter
	SetWorldState     *EffectSetWorldState
	SetEntityLocation *EffectSetEntityLocation
	TeleportPlayer    *EffectTeleportPlayer
	Plan              *EffectPlan
	BlockInput        *EffectBlockInput
	Fade              *EffectFade
	DeactivateFade    *EffectDeactivateFade
	TriggerMovement   *EffectTriggerMovement
	SetFollowCamera   *EffectSetFollowCamera
	MutateNPC         *EffectMutateNPC
	TriggerCombat     *EffectTriggerCombat
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
	if e.SetWorldState != nil {
		effectCount++
		issues.sub("setWorldState").
			requireString("key", e.SetWorldState.Key)
	}
	if e.SetEntityLocation != nil {
		effectCount++
		issues.sub("setEntityLocation").
			requireString("entityId", e.SetEntityLocation.EntityId)
	}
	if e.TeleportPlayer != nil {
		effectCount++
		// TeleportPlayer validation - at least one destination must be specified
	}
	if e.Plan != nil {
		effectCount++
		// Plan validation handled separately
	}
	if e.BlockInput != nil {
		effectCount++
	}
	if e.Fade != nil {
		effectCount++
		issues.sub("fade").
			requirePositive("durationSeconds", e.Fade.DurationSeconds)
	}
	if e.DeactivateFade != nil {
		effectCount++
		issues.sub("deactivateFade").
			requireString("fadeId", e.DeactivateFade.FadeId)
	}
	if e.TriggerMovement != nil {
		effectCount++
		issues.sub("triggerMovement").
			requireString("entityId", e.TriggerMovement.EntityId).
			requireCondition("direction", e.TriggerMovement.Direction != input.NotPressed)
	}
	issues.requirePositive("effectCount", float64(effectCount))
	return issues.toError()
}

type RunnableFunction func()

func (RunnableFunction) String() string {
	return "<inline function>"
}

type EffectFunction struct {
	Fn RunnableFunction
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

type EffectSetWorldState struct {
	Key   string
	Value any
}

type EffectSetEntityLocation struct {
	EntityId    string
	ToReference *string
	ToLocation  *Location
	ToEntityId  *string
}

type EffectTeleportPlayer struct {
	ToReference     *string
	ToLocation      *Location
	ToEntityId      *string
	ExitDirection   *input.Direction // Direction to face and walk after teleporting
	TransitionStyle *string
}

type Location struct {
	X int
	Y int
}

// Plan execution: serial and parallel effect coordination
type EffectPlan struct {
	PlanId string
	Steps  []PlanStep
}

type PlanStep struct {
	// Serial: effects execute one after another
	// Parallel: all effects in the list run simultaneously, wait for all to complete
	Serial   []Effect
	Parallel []Effect
}

type EffectBlockInput struct {
	Blocked bool // true = block, false = unblock
}

type EffectFade struct {
	FadeId          string
	DurationSeconds float64
	AutoDeactivate  *bool
	FromColor       *string // hex color, e.g. "#000000"
	ToColor         *string // hex color, e.g. "#000000"
	Transitions     int     // number of transitions (1 = fade out, 2 = fade out+in, etc.)
}

type EffectDeactivateFade struct {
	FadeId string
}

type EffectTriggerMovement struct {
	EntityId  string
	Direction input.Direction
}

type EffectSetFollowCamera struct {
	EntityId      *string
	ResetPosition bool
}

type EffectMutateNPC struct {
	EntityId          string
	TalkingAtEntityId *string
	IsTalking         *bool
}

type EffectTriggerCombat struct {
	CombatId   string
	Opponent   *rpg.PrimortalType
	Background string
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

func (i *issues) requireCondition(name string, isValid bool) *issues {
	if isValid {
		return i
	}
	i.addf(name, "is invalid")
	return i
}

func (i *issues) requirePositive(name string, value float64) *issues {
	if value > 0 {
		return i
	}
	i.addf(name, "must be positive, got %f", value)
	return i
}
