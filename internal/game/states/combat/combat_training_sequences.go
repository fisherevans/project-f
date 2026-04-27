package combat

import (
	"strings"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/overlays"
	"fisherevans.com/project/f/internal/util/highlighter"
	"github.com/rs/zerolog/log"
)

// trainingTargets loads the highlighter targets for a training overlay flow.
// Fatal on missing/invalid flow - these are developer-authored assets loaded
// at startup, so a failure here is a build error, not a runtime condition.
func trainingTargets(flowName string) []highlighter.Target {
	flow, ok := overlays.GetFlow(flowName)
	if !ok {
		log.Fatal().Str("flow", flowName).Msg("combat training: missing overlay flow")
	}
	targets, err := overlays.ResolveTargets(flow)
	if err != nil {
		log.Fatal().Err(err).Str("flow", flowName).Msg("combat training: resolve failed")
	}
	return targets
}

func (s *State) loadTrainingSequence(sequence string) {
	if sequence == "" {
		sequence = "default"
	}
	switch sequence {
	case "default":
		s.loadTrainingSequenceDefault()
	case "training.1":
		s.loadTrainingSequenceTraining1()
	case "training.2":
		s.loadTrainingSequenceTraining2()
	case "none":

	default:
		log.Fatal().Msgf("invalid training sequence %s", sequence)
	}
}

func (s *State) loadTrainingSequenceDefault() {
	s.training.WithOrdered(false)

	standardDelay := 12

	triggerOnce := func(subKey string, sequence *TrainingSequence) {
		key := strings.ToLower("combat.training." + subKey)
		if game.CurrentSave().Globals.Get(key).AsBool(false) {
			return
		}
		original := sequence.OnSequenceComplete
		sequence.OnSequenceComplete = func(state *State) {
			game.CurrentSave().Globals.Set(key, true)
			if original != nil {
				original(state)
			}
		}
		s.training.Add(sequence)
	}

	// status hints

	addPlayerStatusTraining := func(statusType rpg.StatusType, flowName string) {
		triggerOnce(strings.ToLower("player_status."+string(statusType)),
			NewTrainingSequence().
				WithReadyToQueue(WaitSomeTicks(standardDelay, func(state *State) bool {
					return state.Player.GetStatuses().HasStatus(statusType)
				})).
				WithTargets(trainingTargets(flowName)...))
	}
	addPlayerStatusTraining(rpg.StatusPoisoned, "combat_training/status_poisoned")
	addPlayerStatusTraining(rpg.StatusBurning, "combat_training/status_burning")
	addPlayerStatusTraining(rpg.StatusMending, "combat_training/status_mending")
	addPlayerStatusTraining(rpg.StatusIonized, "combat_training/status_ionized")
	addPlayerStatusTraining(rpg.StatusWarded, "combat_training/status_warded")

	// opponent-in-stance hints

	triggerOnce("player_attacks_while_opponent_defends", NewTrainingSequence().
		WithReadyToQueue(WaitSomeTicks(standardDelay, NewSkillCondition().
			WithOnPlayerTick(func(t rpg.SkillTick, a CheckAgainst) bool {
				for _, e := range t.Effects {
					if e.Self == false && e.Damage != nil && a.Stance == rpg.TickStanceDefending {
						return true
					}
				}
				return false
			}).Check)).
		WithTargets(trainingTargets("combat_training/attack_while_opponent_defends")...))

	// player-in-stance hints

	triggerOnce("opponent_attacks_while_player_defends", NewTrainingSequence().
		WithReadyToQueue(WaitSomeTicks(standardDelay, NewSkillCondition().
			WithOnOpponentTick(func(t rpg.SkillTick, a CheckAgainst) bool {
				for _, e := range t.Effects {
					if e.Self == false && e.Damage != nil && a.Stance != rpg.TickStanceDefending {
						return true
					}
				}
				return false
			}).Check)).
		WithTargets(trainingTargets("combat_training/defend_against_opponent_attack")...))
}

func (s *State) loadTrainingSequenceTraining1() {
	s.training.WithOrdered(true)
	s.training.Add(NewTrainingSequence().
		WithTargets(trainingTargets("combat_training/training1_layout")...))
	s.training.Add(NewTrainingSequence().
		WithReadyToQueue(func(s *State) bool {
			return s.Player.NextSkill != nil
		}).
		WithTargets(trainingTargets("combat_training/training1_select_skill")...))
	s.training.Add(NewTrainingSequence().
		WithReadyToQueue(func(s *State) bool {
			return s.Player.NextSkillCommitted
		}).
		WithPauseCombat(false).
		WithTargets(trainingTargets("combat_training/training1_skills_active")...))
	stanceCondition := NewSkillCondition().
		WithOnPlayerTick(func(t rpg.SkillTick, a CheckAgainst) bool {
			for _, e := range t.Effects {
				if e.Self == false && e.Damage != nil && a.Stance != rpg.TickStanceNone {
					return true
				}
			}
			return false
		})
	s.training.Add(NewTrainingSequence().
		WithReadyToQueue(WaitSomeTicks(10, stanceCondition.Check)).
		WithTargets(trainingTargets("combat_training/training1_stances")...))
}

func (s *State) loadTrainingSequenceTraining2() {
	s.training.WithOrdered(false)
	s.training.Add(NewTrainingSequence().
		WithTargets(trainingTargets("combat_training/training2_intro")...))
	s.loadTrainingSequenceDefault()
}
