package adventure

import (
	"fisherevans.com/project/f/internal/game/rpg"
)

type GameState interface {
	WorldState() rpg.ReadableState
	RunState() rpg.ReadableState
}

type observedMutableState struct {
	dispatcher     *Dispatcher
	state          rpg.MutableState
	createSetEvent func(key string, newValue, oldValue *rpg.StateValue) any
	createDelete   func(key string, oldValue *rpg.StateValue) any
}

func newObservedMutableState(dispatcher *Dispatcher, state rpg.MutableState, createSetEvent func(key string, newValue, oldValue *rpg.StateValue) any, createDelete func(key string, oldValue *rpg.StateValue) any) *observedMutableState {
	return &observedMutableState{
		dispatcher:     dispatcher,
		state:          state,
		createSetEvent: createSetEvent,
		createDelete:   createDelete,
	}
}

func (s *observedMutableState) Set(key string, value any) *rpg.StateValue {
	oldValue := s.state.Set(key, value)
	newValue := s.state.Get(key)
	event := s.createSetEvent(key, newValue, oldValue)
	s.dispatcher.Dispatch(event)
	return oldValue
}

func (s *observedMutableState) Delete(key string) *rpg.StateValue {
	oldValue := s.state.Delete(key)
	event := s.createDelete(key, oldValue)
	s.dispatcher.Dispatch(event)
	return oldValue
}

func (s *observedMutableState) Get(key string) *rpg.StateValue {
	return s.state.Get(key)
}
