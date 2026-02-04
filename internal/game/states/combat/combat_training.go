package combat

import (
	"fisherevans.com/project/f/internal/util/highlighter"
)

type TrainingListener struct {
	activeSequence *TrainingSequence
	sequences      []*TrainingSequence
	ordered        bool

	tickCountAtLastCompletion int
}

func NewTrainingListener(sequences ...*TrainingSequence) *TrainingListener {
	return &TrainingListener{
		activeSequence: nil,
		sequences:      sequences,
		ordered:        true,
	}
}

func (l *TrainingListener) WithOrdered(ordered bool) *TrainingListener {
	l.ordered = ordered
	return l
}

func (l *TrainingListener) Add(target ...*TrainingSequence) {
	l.sequences = append(l.sequences, target...)
}

func (l *TrainingListener) TicksSinceLastSequence(s *State) int {
	return s.Battle.TicksTriggered - l.tickCountAtLastCompletion
}

func (l *TrainingListener) OnTick(s *State) {
	// Completion logic: if we have an active sequence but the highlighter is no longer active, it just finished
	if l.activeSequence != nil && !s.highlighter.IsActive() {
		if l.activeSequence.OnSequenceComplete != nil {
			l.activeSequence.OnSequenceComplete(s)
		}
		l.activeSequence = nil
		l.tickCountAtLastCompletion = s.Battle.TicksTriggered
	}

	if len(l.sequences) == 0 || s.highlighter.IsActive() {
		return
	}

	// Queueing logic: find a sequence that is ready to be shown
	if l.ordered {
		if l.sequences[0].ReadyToQueue(s) {
			l.activeSequence = l.sequences[0]
			l.sequences = l.sequences[1:]
			s.highlighter.AppendTargets(l.activeSequence.Targets...)
		}
	} else {
		for i, seq := range l.sequences {
			if seq.ReadyToQueue(s) {
				l.activeSequence = seq
				l.sequences = append(l.sequences[:i], l.sequences[i+1:]...)
				s.highlighter.AppendTargets(l.activeSequence.Targets...)
				return
			}
		}
	}
}

func (l *TrainingListener) ShouldPauseCombat() bool {
	if l.activeSequence == nil {
		return false
	}
	return l.activeSequence.PauseCombat
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
		PauseCombat:  true,
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
