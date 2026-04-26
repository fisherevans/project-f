package adventure

import "fisherevans.com/project/f/internal/game/input"

func init() {
	registerStepConverter("push_behavior", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		e := NewPushEntityBehaviorEffect(mapStr(m, "entity"))
		if mapBool(m, "scripted_motion") {
			e = e.WithScriptedMotion(EntityBehaviorScriptedMotion{ActivePlayerZone: mapStr(m, "active_player_zone")})
		}
		if facingEntity := mapStr(m, "facing_entity"); facingEntity != "" {
			e = e.WithFacingEntityId(facingEntity)
		}
		return []Effect{e}
	})
	registerStepConverter("pop_behavior", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		return []Effect{NewPopEntityBehaviorEffect(resolveString(step.Params, tc))}
	})
	registerStepConverter("disable_behavior", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		return []Effect{NewMutateEntityBehaviorEffect(mapStr(m, "entity")).WithDisableBy(mapStr(m, "by"))}
	})
	registerStepConverter("enable_behavior", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		return []Effect{NewMutateEntityBehaviorEffect(mapStr(m, "entity")).WithEnableBy(mapStr(m, "by"))}
	})
}

type EffectEntityFaceDirection struct {
	instantEffect
	EntityId     string
	Direction    *input.Direction `one_of:"dir"`
	TargetEntity *string          `one_of:"dir"`
}

func (e *EffectEntityFaceDirection) Process(source EntityReader, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find movement for entity")
	}
	if e.Direction != nil {
		entity.SetFacingDirection(*e.Direction)
	}
	if e.TargetEntity != nil {
		targetEntity, ok := s.entities.GetEntity(*e.TargetEntity)
		if !ok {
			logEffectWarnf(source, e, "failed to find target entity")
			return false
		}
		dir := entity.GetLocation().DirectionTowards(targetEntity.GetLocation())
		entity.SetFacingDirection(dir)
	}
	return true
}

type EffectStartScriptedMotion struct {
	MotionId   string `auto_generate:"true"`
	EntityId   string
	Location   *MapLocation      `one_of:"target"`
	ToEntityId *string           `one_of:"target"`
	Relative   *RelativeLocation `one_of:"target"`
}

func (e *EffectStartScriptedMotion) CompletionID() string {
	if e.MotionId == "" {
		return ""
	}
	return "motion:" + e.MotionId
}

func (e *EffectStartScriptedMotion) Process(source EntityReader, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity")
		return false
	}
	behavior, ok := entity.GetBehavior()
	if !ok {
		logEffectWarnf(source, e, "failed to find behavior for entity")
		return false
	}
	scripted, ok := behavior.(*ScriptedMotionBehavior)
	if !ok {
		logEffectWarnf(source, e, "behavior is not a scripted motion behavior")
		return false
	}
	var target MotionTarget
	if e.Location != nil {
		target = NewPathfindingMotion(e.MotionId, e.CompletionID(), entity, *e.Location)
	} else if e.Relative != nil {
		target = NewRelativeMotion(e.MotionId, e.CompletionID(), e.Relative.Direction, e.Relative.Steps)
	} else if e.ToEntityId != nil {
		toEntity, ok := s.entities.GetEntity(*e.ToEntityId)
		if !ok {
			logEffectWarnf(source, e, "failed to find movement for target entity")
			return false
		}
		target = NewPathfindingMotion(e.MotionId, e.CompletionID(), entity, toEntity.GetLocation())
	}
	scripted.SetTarget(target)
	return true
}

type RelativeLocation struct {
	Direction input.Direction
	Steps     int
}

type EffectPushEntityBehavior struct {
	instantEffect
	EntityId       string
	ScriptedMotion *EntityBehaviorScriptedMotion `one_of:"type"`
	FacingEntity   *EntityBehaviorFacingEntity   `one_of:"type"`
}

func (e *EffectPushEntityBehavior) WithFacingEntityId(id string) *EffectPushEntityBehavior {
	e.FacingEntity = &EntityBehaviorFacingEntity{
		EntityId: &id,
	}
	return e
}

func (e *EffectPushEntityBehavior) Process(source EntityReader, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to override behavior")
		return false
	}
	if e.ScriptedMotion != nil {
		AttachScriptedMotionBehavior(entity, e.ScriptedMotion.ActivePlayerZone)
	} else if e.FacingEntity != nil {
		facing := ""
		if e.FacingEntity.EntityId != nil {
			facing = *e.FacingEntity.EntityId
		}
		AttachFaceEntityBehavior(entity, facing)
	}
	return true
}

type EntityBehaviorPlayer struct {
}

type EntityBehaviorNPC struct {
}

type EntityBehaviorScriptedMotion struct {
	ActivePlayerZone string
}

type EntityBehaviorFacingEntity struct {
	EntityId *string
}

type EffectPopEntityBehavior struct {
	instantEffect
	EntityId string
}

func (e *EffectPopEntityBehavior) Process(source EntityReader, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to pop behavior")
		return false
	}
	entity.PopBehavior()
	return true
}

type EffectMutateEntityBehavior struct {
	instantEffect
	EntityId  string
	DisableBy *string `one_of:"enablement"`
	EnableBy  *string `one_of:"enablement"`
	Reset     *bool
}

func (e *EffectMutateEntityBehavior) Process(source EntityReader, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to mutate")
		return false
	}
	if e.DisableBy != nil {
		entity.DisableBehavior(*e.DisableBy)
	}
	if e.EnableBy != nil {
		entity.EnableBehavior(*e.EnableBy)
	}
	if e.Reset != nil && *e.Reset {
		b, ok := entity.GetBehavior()
		if !ok {
			logEffectWarnf(source, e, "failed to find behavior to mutate")
			return false
		}
		b.Reset()
	}
	return true
}

type EffectTriggerMovement struct {
	instantEffect
	EntityId  string
	Direction *input.Direction `one_of:"to"`
	Location  *MapLocation     `one_of:"to"`
	MoveState *MoveState
}

func (e *EffectTriggerMovement) Process(source EntityReader, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to move")
		return false
	}
	moveState := MoveStateWalking
	if e.MoveState != nil {
		moveState = *e.MoveState
	}
	var location MapLocation
	if e.Direction != nil {
		location = entity.GetLocation().Moved(*e.Direction)
	} else if e.Location != nil {
		location = MapLocation{
			X: e.Location.X,
			Y: e.Location.Y,
		}
	}
	entity.AttemptMovement(location, moveState)
	return true
}

type EffectResetMovement struct {
	instantEffect
	EntityId string
}

func (e *EffectResetMovement) Process(source EntityReader, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity")
		return false
	}
	entity.CancelMovement()
	b, ok := entity.GetBehavior()
	if ok {
		b.Reset()
	}
	return true
}
