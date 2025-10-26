package adventure

import "fisherevans.com/project/f/internal/game/events"

type Dispatcher interface {
	ProcessEffects(effects ...events.Effect)
	EmitEvents(events ...any)
}

type stateDispatcher struct {
	s *State
}

func (sd *stateDispatcher) ProcessEffects(effects ...events.Effect) {
	sd.s.ExecuteSystemEffects(effects...)
}

func (sd *stateDispatcher) EmitEvents(events ...any) {
	sd.s.eventDispatcher.Dispatch(events...)
}
