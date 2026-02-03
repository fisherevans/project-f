package combat

import (
	"strings"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/highlighter"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

func r(x, y, w, h int) pixel.Rect {
	return pixel.R(float64(x), float64(y), float64(x+w), float64(y+h))
}

func highlightTargetCombatant(msg string, isPlayer bool) highlighter.Target {
	x := 18
	w := 60
	if !isPlayer {
		x = game.GameWidth - x - w
	}
	return highlighter.NewTarget(r(x, 32, w, 84)).
		WithMessage(highlighter.NewMessage(msg, highlighter.MessageOnBottom).Wrapped(80)).
		WithBadge(highlighter.NewBadge(highlighter.BadgeInTopRight)).
		NoPadding()
}
func highlightTargetCombatantStats(msg string, isPlayer bool) highlighter.Target {
	x := 0
	w := 84
	messagePlacement := highlighter.MessageOnRight
	if !isPlayer {
		x = game.GameWidth - w
		messagePlacement = highlighter.MessageOnLeft
	}
	return highlighter.NewTarget(r(x, 108, w, 44)).
		WithBadge(highlighter.NewBadge(highlighter.BadgeOnBottomMiddle)).
		WithMessage(highlighter.NewMessage(msg, messagePlacement).Wrapped(100)).
		NoPadding()
}
func highlightTargetSkills(msg string) highlighter.Target {
	return highlighter.NewTarget(r(23, 0, 194, 42)).
		WithMessage(highlighter.NewMessage(msg, highlighter.MessageOnTop).Wrapped(140)).
		WithBadge(highlighter.NewBadge(highlighter.BadgeInTopRight))
}
func highlightTargetActive(msg string) highlighter.Target {
	return highlighter.NewTarget(r(89, 53, 64, 102)).
		WithMessage(highlighter.NewMessage(msg, highlighter.MessageOnBottom).Wrapped(140)).
		WithBadge(highlighter.NewBadge(highlighter.BadgeOnTopMiddle)).
		NoPadding()
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

	addPlayerStatusTraining := func(statusType rpg.StatusType, targets ...highlighter.Target) {
		triggerOnce(strings.ToLower("player_status."+string(statusType)),
			NewTrainingSequence().
				WithReadyToQueue(WaitSomeTicks(standardDelay, func(state *State) bool {
					return state.Player.GetStatuses().HasStatus(statusType)
				})).
				WithTargets(targets...))
	}
	addPlayerStatusTraining(rpg.StatusPoisoned,
		highlightTargetCombatantStats("You've been POISONED!", true),
		highlightTargetCombatantStats("You'll take damage every few ticks as long as this status is applied.", true))
	addPlayerStatusTraining(rpg.StatusBurning,
		highlightTargetCombatantStats("You've been BURNED!", true),
		highlightTargetCombatantStats("You'll take damage every few ticks as long as this status is applied.", true))
	addPlayerStatusTraining(rpg.StatusMending,
		highlightTargetCombatantStats("You are now MENDING!", true),
		highlightTargetCombatantStats("You'll gain some health every few ticks as long as this status is applied.", true))
	addPlayerStatusTraining(rpg.StatusIonized,
		highlightTargetCombatantStats("You are now IONIZED!", true),
		highlightTargetCombatantStats("Your next attack will deal extra damage.", true))
	addPlayerStatusTraining(rpg.StatusWarded,
		highlightTargetCombatantStats("You are now WARDED!", true),
		highlightTargetCombatantStats("You'll take less damage while this status is applied.", true))

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
		WithTargets(
			highlightTargetActive("Don't forget to try and time your attacks to hit when your foe is not DEFENDING."),
		))

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
		WithTargets(
			highlightTargetActive("Defending goes both ways. Try to guard yourself against your opponents attacks."),
		))
}

func (s *State) loadTrainingSequenceTraining1() {
	s.training.WithOrdered(true)
	s.training.Add(NewTrainingSequence().
		WithTargets(
			highlightTargetCombatant("This is your Animech!", true),
			highlightTargetCombatantStats("Your SHIELD protects you and regenerates between combat.", true),
			highlightTargetCombatantStats("Your SYNC is how aligned your soul is with your Animech.", true),
			highlightTargetCombatantStats("Taking damage depletes your SHIELD, and then your SYNC.", true),
			highlightTargetCombatant("This is your opponent!", false),
			highlightTargetCombatantStats("Deplete their HEALTH to win.", false),
			highlightTargetSkills("These are your combat skills. Use the D-Pad to select one."),
		))
	s.training.Add(NewTrainingSequence().
		WithReadyToQueue(func(s *State) bool {
			return s.Player.NextSkill != nil
		}).
		WithTargets(
			highlightTargetActive("You can see your selected skill here. It's just pending. To select it, press A."),
		))
	s.training.Add(NewTrainingSequence().
		WithReadyToQueue(func(s *State) bool {
			return s.Player.NextSkillCommitted
		}).
		WithTargets(
			highlightTargetActive("Both your skills and your opponent's show up here."),
			highlightTargetActive("Both combatant's active skills trigger in parallel."),
			highlightTargetActive("Skills take different amounts of time to execute."),
			highlightTargetActive("Each tick does something different."),
			highlightTargetActive("Use your skills to damage and defeat your opponent!."),
		))
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
		WithTargets(
			highlightTargetActive("Some skills put a combatant into a STANCE. Stances are a temporary state that can affect the combatant's behavior and skill execution."),
			highlightTargetActive("In this case, you are going to damage your opponent while they are in a DEFENDING stance. This will make your attack less effective."),
			highlightTargetActive("Battling is all about lining up your attacks and stances to maximize damage and minimize vulnerability."),
		))
}

func (s *State) loadTrainingSequenceTraining2() {
	s.training.WithOrdered(false)
	s.training.Add(NewTrainingSequence().
		WithTargets(
			highlightTargetCombatant("Get ready, this foe will actually attack you!", false),
		))
	s.loadTrainingSequenceDefault()
}
