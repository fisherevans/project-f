package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"github.com/rs/zerolog/log"
)

func (s *State) processEffects(effects []events.DispatchedEffect) {
	for _, e := range effects {
		s.processEffectDialogue(e.Source, e.Dialogue)
		s.processEffectTimer(e.Source, e.Timer)
		s.processEffectMutateEntity(e.Source, e.MutateEntity)
		s.processEffectYieldElythium(e.Source, e.YieldElythium)
		s.processEffectChatter(e.Source, e.Chatter)
	}
}

func logEffectf(source events.EntityContext, e any, messageFormat string, args ...any) {
	log.Info().Str("caller", source.Id()).Interface("e", e).Msgf(messageFormat, args...)

}

func (s *State) processEffectDialogue(source events.EntityContext, e *events.EffectDialogue) {
	if e == nil {
		return
	}
	s.dialogues.Append(NewBasicDialogue(e.Text, nil, e.DialogueId))
	logEffectf(source, e, "dialogue added")
}

func (s *State) processEffectTimer(source events.EntityContext, e *events.EffectTimer) {
	if e == nil {
		return
	}
	s.timers.AddTimer(source.Id(), e.TimerId, e.DurationSeconds)
	logEffectf(source, e, "timer added")
}

func (s *State) processEffectMutateEntity(source events.EntityContext, e *events.EffectMutateEntity) {
	if e == nil {
		return
	}
	entity, exists := s.entities[EntityId(e.EntityId)]
	if !exists {
		log.Warn().Str("entityId", e.EntityId).Msg("failed to find entity for change entity mode")
		return
	}
	if e.Mode != nil {
		entity.SetMode(*e.Mode)
	}
	if e.IsPassable != nil {
		entity.SetIsPassable(*e.IsPassable)
	}
	dynamic, isDynamic := entity.(*DynamicEntity)
	if isDynamic {
		if e.DynamicAnimations != nil {
			dynamic.SetDynamicAnimations(e.DynamicAnimations)
		}
		if e.DynamicLights != nil {
			dynamic.SetDynamicLights(e.DynamicLights)
		}
	}
	logEffectf(source, e, "entity mode set")
}

func (s *State) processEffectYieldElythium(source events.EntityContext, e *events.EffectYieldElythium) {
	if e == nil {
		return
	}
	s.hud.ElythiumCount += e.Amount
	logEffectf(source, e, "elythium granted")
}

func (s *State) processEffectChatter(source events.EntityContext, e *events.EffectChatter) {
	if e == nil {
		return
	}
	s.chatters.Add(newBasicEntityChatter(EntityId(e.EntityId), e.DurationSeconds, e.Message, e.ChatterId))
	logEffectf(source, e, "chatter added")
}
