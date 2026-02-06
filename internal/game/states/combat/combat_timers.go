package combat

import "fisherevans.com/project/f/internal/util"

type actionTimer struct {
	triggerAfter float64
	action       func(s *State)
}

func newSetVisibleTimer(after float64, vis *util.Visibility, makeVisible bool) actionTimer {
	return actionTimer{
		triggerAfter: after,
		action:       func(s *State) { vis.SetVisible(makeVisible) },
	}
}

func (s *State) triggerTimers(timers []actionTimer) []actionTimer {
	for idx := 0; idx < len(timers); idx++ {
		timer := timers[idx]
		if s.elapsed >= timer.triggerAfter {
			if timer.action != nil {
				timer.action(s)
			}
			timers = append(timers[:idx], timers[idx+1:]...)
		}
	}
	return timers
}
