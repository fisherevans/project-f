package adventure

import (
	"fisherevans.com/project/f/internal/game/rpg"
)

type observedGlobals struct {
	baseGlobals rpg.Globals
	dispatcher  *Dispatcher
}

func observeGlobals(baseGlobals rpg.Globals) *observedGlobals {
	return &observedGlobals{
		baseGlobals: baseGlobals,
	}
}

func (s *observedGlobals) withDispatcher(dispatcher *Dispatcher) {
	s.dispatcher = dispatcher
}

func (s *observedGlobals) Set(key string, value any) *rpg.GlobalValue {
	oldValue := s.baseGlobals.Set(key, value)
	newValue := s.baseGlobals.Get(key)
	if s.dispatcher != nil {
		s.dispatcher.Dispatch(NewEventGlobalVariableUpdated(key, newValue, oldValue))
	}
	return oldValue
}

func (s *observedGlobals) Delete(key string) *rpg.GlobalValue {
	oldValue := s.baseGlobals.Delete(key)
	if s.dispatcher != nil {
		s.dispatcher.Dispatch(NewEventGlobalVariableDeleted(key, oldValue))
	}
	return oldValue
}

func (s *observedGlobals) Get(key string) *rpg.GlobalValue {
	return s.baseGlobals.Get(key)
}
