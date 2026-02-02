package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/highlighter"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type TrainingListener struct {
	activeSequence int
	isPending      bool
	chain          []*TrainingSequence
}

func NewTrainingListener(chain ...*TrainingSequence) *TrainingListener {
	return &TrainingListener{
		activeSequence: 0,
		isPending:      true,
		chain:          chain,
	}
}

func (l *TrainingListener) Add(target ...*TrainingSequence) {
	l.chain = append(l.chain, target...)
}

func (l *TrainingListener) OnTick(s *State) {
	if len(l.chain) == 0 || s.highlighter.IsActive() {
		return
	}
	if !l.isPending && l.activeSequence >= 0 && l.activeSequence < len(l.chain) {
		if l.chain[l.activeSequence].OnSequenceComplete != nil {
			l.chain[l.activeSequence].OnSequenceComplete(s)
		}
		l.activeSequence++
		l.isPending = true
	}
	if l.activeSequence >= len(l.chain) {
		l.activeSequence = 0
		l.chain = nil
		l.isPending = false
		return
	}
	if !l.isPending {
		return
	}
	if !l.chain[l.activeSequence].ReadyToQueue(s) {
		return
	}
	s.highlighter.SetSequence(l.chain[l.activeSequence].Targets...)
	l.isPending = false
}

func (l *TrainingListener) ShouldPauseCombat() bool {
	if len(l.chain) == 0 || l.activeSequence >= len(l.chain) || l.isPending {
		return false
	}
	return l.chain[l.activeSequence].PauseCombat
}

type TrainingSequence struct {
	ReadyToQueue       func(state *State) bool
	Targets            []highlighter.Target
	PauseCombat        bool
	OnSequenceComplete func(state *State)
}

func NewTrainingSequence() *TrainingSequence {
	return &TrainingSequence{
		ReadyToQueue: func(state *State) bool { return true },
	}
}

func (s *TrainingSequence) WithReadyToQueue(ready func(state *State) bool) *TrainingSequence {
	s.ReadyToQueue = ready
	return s
}

func (s *TrainingSequence) WithTargets(targets ...highlighter.Target) *TrainingSequence {
	s.Targets = append(s.Targets, targets...)
	return s
}

func (s *TrainingSequence) WithOnSequenceComplete(onComplete func(state *State)) *TrainingSequence {
	s.OnSequenceComplete = onComplete
	return s
}

func (s *TrainingSequence) WithPauseCombat(pauseCombat bool) *TrainingSequence {
	s.PauseCombat = pauseCombat
	return s
}

func (s *State) loadTrainingSequence(sequence string) {
	r := func(x, y, w, h int) pixel.Rect {
		return pixel.R(float64(x), float64(y), float64(x+w), float64(y+h))
	}

	combatant := func(msg string, isPlayer bool) highlighter.Target {
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
	combatantStats := func(msg string, isPlayer bool) highlighter.Target {
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
	skills := func(msg string) highlighter.Target {
		return highlighter.NewTarget(r(23, 0, 194, 42)).
			WithMessage(highlighter.NewMessage(msg, highlighter.MessageOnTop).Wrapped(140)).
			WithBadge(highlighter.NewBadge(highlighter.BadgeInTopRight))
	}
	active := func(msg string) highlighter.Target {
		return highlighter.NewTarget(r(89, 53, 64, 102)).
			WithMessage(highlighter.NewMessage(msg, highlighter.MessageOnBottom).Wrapped(140)).
			WithBadge(highlighter.NewBadge(highlighter.BadgeOnTopMiddle)).
			NoPadding()
	}
	switch sequence {
	case "":
		return
	case "training.1":
		s.training.Add(NewTrainingSequence().
			WithTargets(
				combatant("This is your Animech!", true),
				combatantStats("Your SHIELD protects you and regenerates between combat.", true),
				combatantStats("Your SYNC is how aligned your soul is with your Animech.", true),
				combatantStats("Taking damage depletes your SHIELD, and then your SYNC.", true),
				combatant("This is your opponent!", false),
				combatantStats("Deplete their HEALTH to win.", false),
				skills("These are your combat skills. Use the D-Pad to select one."),
			))
		s.training.Add(NewTrainingSequence().
			WithReadyToQueue(func(s *State) bool {
				return s.Player.NextSkill != nil
			}).
			WithTargets(
				active("You can see your selected skill here. It's just pending. To select it, press A."),
			))
		s.training.Add(NewTrainingSequence().
			WithReadyToQueue(func(s *State) bool {
				return s.Player.NextSkillCommitted
			}).
			WithTargets(
				active("Both your skills and your opponent's show up here."),
				active("Both combatant's active skills trigger in parallel."),
				active("Skills take different amounts of time to execute."),
				active("Each tick does something different."),
				active("Use your skills to damage and defeat your opponent!."),
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
			WithPauseCombat(true).
			WithTargets(
				active("Some skills put a combatant into a STANCE. Stances are a temporary state that can affect the combatant's behavior and skill execution."),
				active("In this case, you are going to damage your opponent while they are in a DEFENDING stance. This will make your attack less effective."),
				active("Battling is all about lining up your attacks and stances to maximize damage and minimize vulnerability."),
			))
	case "training.2":
		s.training.Add(NewTrainingSequence().
			WithTargets(
				combatant("Get ready, this foe will actually attack you!", false),
			))
		s.training.Add(NewTrainingSequence().
			WithReadyToQueue(WaitSomeTicks(6, func(state *State) bool {
				for _, lvl := range state.Player.GetStatuses().GetLevels() {
					if lvl > 1 {
						return true
					}
				}
				return false
			})).
			WithPauseCombat(true).
			WithTargets(
				combatantStats("This foe POISONED you!", true),
				combatantStats("You'll take POISON damage every few ticks as long as this status is applied.", true),
			))
		attackCondition := NewSkillCondition().
			WithOnPlayerTick(func(t rpg.SkillTick, a CheckAgainst) bool {
				for _, e := range t.Effects {
					if e.Self == false && e.Damage != nil && a.Stance == rpg.TickStanceDefending {
						return true
					}
				}
				return false
			})
		s.training.Add(NewTrainingSequence().
			WithReadyToQueue(WaitSomeTicks(6, attackCondition.Check)).
			WithPauseCombat(true).
			WithTargets(
				active("Don't forget to try and time your attacks to hit when your foe is not DEFENDING."),
			))
		defendCondition := NewSkillCondition().
			WithOnOpponentTick(func(t rpg.SkillTick, a CheckAgainst) bool {
				for _, e := range t.Effects {
					if e.Self == false && e.Damage != nil && a.Stance != rpg.TickStanceDefending {
						return true
					}
				}
				return false
			})
		s.training.Add(NewTrainingSequence().
			WithReadyToQueue(WaitSomeTicks(6, defendCondition.Check)).
			WithPauseCombat(true).
			WithTargets(
				active("Defending goes both ways. Try to guard yourself against your opponents attacks."),
			))
	default:
		log.Fatal().Msgf("invalid training sequence %s", sequence)
	}
}
