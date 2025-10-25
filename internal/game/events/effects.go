package events

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
)

type RunnableFunction func()

func (RunnableFunction) String() string {
	return "<inline function>"
}

type EffectFunction struct {
	Fn RunnableFunction
}

type EffectDialogue struct {
	DialogueId string `auto_generate:"true"`
	Text       string
}

type EffectChatter struct {
	ChatterId       string `auto_generate:"true"`
	EntityId        string
	DurationSeconds float64
	Message         string
}

type EffectYieldElythium struct {
	Amount int
}

type EffectTimer struct {
	TimerId         string `auto_generate:"true"`
	DurationSeconds float64
}

type EffectMutateModeBasedEntity struct {
	EntityId   string
	Mode       *string
	Animations *map[string][]AnimationReference
	Lights     *map[string][]LightConfig
}

type AnimationReference struct {
	Tilesheet string
	Name      string
}

type LightConfig struct {
	Color    string
	Size     float64
	Modifier string `optional:"true"`
}

type EffectMutateBlockingPresence struct {
	EntityId          string
	IsBlockingIngress *bool
}

type EffectSetWorldState struct {
	Key   string
	Value any
}

type EffectSetEntityLocation struct {
	EntityId    string
	ToReference *string   `one_of:"destination"`
	ToLocation  *Location `one_of:"destination"`
	ToEntityId  *string   `one_of:"destination"`
}

type EffectTeleportPlayer struct {
	ToReference     *string   `one_of:"destination"`
	ToLocation      *Location `one_of:"destination"`
	ToEntityId      *string   `one_of:"destination"`
	ExitDirection   *input.Direction
	TransitionStyle *string
}

// Plan execution: serial and parallel effect coordination
type EffectPlan struct {
	PlanId string `auto_generate:"true"`
	Steps  []PlanStep
}

func NewSerialPlan(effects ...Effect) *EffectPlan {
	return &EffectPlan{
		Steps: []PlanStep{
			{
				Serial: effects,
			},
		},
	}
}

func NewParallelPlan(effects ...Effect) *EffectPlan {
	return &EffectPlan{
		Steps: []PlanStep{
			{
				Parallel: effects,
			},
		},
	}
}

func (p *EffectPlan) WithId(id string) *EffectPlan {
	p.PlanId = id
	return p
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
	FadeId          string `auto_generate:"true"`
	DurationSeconds float64
	Transitions     int // number of transitions (1 = fade out, 2 = fade out+in, etc.)
	AutoDeactivate  *bool
	FromColor       *string
	ToColor         *string
}

type EffectDeactivateFade struct {
	FadeId string
}

type EffectTriggerMovement struct {
	EntityId  string
	Direction input.Direction
	MoveState *types.MoveState
}

type EffectSetFollowCamera struct {
	EntityId      *string `one_of:"target"`
	ResetPosition bool
}

type EffectMutateNPC struct {
	EntityId          string
	TalkingAtEntityId *string
}

type EffectTriggerCombat struct {
	CombatId   string `auto_generate:"true"`
	Opponent   *rpg.PrimortalType
	Background string
}
