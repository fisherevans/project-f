package highlighter

import (
	"fisherevans.com/project/f/internal/game/input"
	"github.com/gopxl/pixel/v2"
)

type SequencedDrawer struct {
	drawer *Drawer

	currentTarget int
	sequence      []Target

	doTransition bool
}

func NewSequencedDrawer(drawer *Drawer) *SequencedDrawer {
	return &SequencedDrawer{
		drawer:       drawer,
		doTransition: true,
	}
}

func (s *SequencedDrawer) Render(target pixel.Target, timeDelta float64, controls *input.Controls) {
	if len(s.sequence) == 0 {
		// do nothing, let it render
	} else if s.drawer.IsDismissed() {
		s.drawer.SetTargetArea(s.sequence[s.currentTarget], s.doTransition)
	} else if s.drawer.IsTransitioned() && controls.ButtonA().JustPressed() {
		s.advance()
	} else if controls.ButtonB().JustPressed() {
		s.stepBack()
	}
	s.drawer.Render(target, timeDelta)
}

func (s *SequencedDrawer) stepBack() {
	if s.currentTarget == 0 {
		return
	}
	s.currentTarget--
	s.drawer.SetTargetArea(s.sequence[s.currentTarget], s.doTransition)
}
func (s *SequencedDrawer) advance() {
	s.currentTarget++
	if s.currentTarget >= len(s.sequence) {
		s.ClearSequence()
		s.drawer.Dismiss(s.doTransition)
		return
	}
	s.drawer.SetTargetArea(s.sequence[s.currentTarget], s.doTransition)
}

func (s *SequencedDrawer) ClearSequence() {
	s.currentTarget = 0
	s.sequence = nil
}

func (s *SequencedDrawer) AppendTargets(targets ...Target) {
	s.sequence = append(targets)
	s.currentTarget = 0
}

func (s *SequencedDrawer) IsActive() bool {
	return len(s.sequence) > 0
}
