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
	EntityId string
	types.ModeBaseRenderConfig
}

func (e *EffectMutateModeBasedEntity) WithMode(mode string) *EffectMutateModeBasedEntity {
	e.Mode = &mode
	return e
}

func (e *EffectMutateModeBasedEntity) WithAnimations(anims map[string][]types.AnimationReference) *EffectMutateModeBasedEntity {
	e.Animations = anims
	return e
}

func (e *EffectMutateModeBasedEntity) WithLights(lights map[string][]types.LightConfig) *EffectMutateModeBasedEntity {
	e.Lights = lights
	return e
}

type EffectResetModeBasedEntityAnimation struct {
	EntityId string
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
	Serial   []Effect `one_of:"type"`
	Parallel []Effect `one_of:"type"`
}

type EffectMutateEntityBehavior struct {
	EntityId  string
	DisableBy *string `one_of:"enablement"`
	EnableBy  *string `one_of:"enablement"`
	Reset     *bool
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
	Direction *input.Direction `one_of:"to"`
	Location  *Location        `one_of:"to"`
	MoveState *types.MoveState
}

type EffectResetMovement struct {
	EntityId string
}

type EffectOverrideCamera struct {
	Follow *FollowCamera `one_of:"type"`
}

type EffectPopCameraOverride struct {
	MaintainCurrentLocation bool
}

type EffectMutateFollowCamera struct {
	FollowEntityId *string
	ResetPosition  *bool
}

type FollowCamera struct {
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

type EffectEntityFaceDirection struct {
	EntityId  string
	Direction input.Direction
}

// todo implement handlers for these
// todo register motion id in plan system!
// todo script some motion!

type EffectStartScriptedMotion struct {
	MotionId   string `auto_generate:"true"`
	EntityId   string
	Location   *Location         `one_of:"target"`
	ToEntityId *string           `one_of:"target"`
	Relative   *RelativeLocation `one_of:"target"`
}

type RelativeLocation struct {
	Direction input.Direction
	Steps     int
}

type EffectOverrideEntityBehavior struct {
	EntityId       string
	ScriptedMotion *EntityBehaviorScriptedMotion `one_of:"type"`
}

type EntityBehaviorPlayer struct {
}

type EntityBehaviorNPC struct {
}

type EntityBehaviorScriptedMotion struct {
}

type EffectPopEntityBehaviorOverride struct {
	EntityId string
}
