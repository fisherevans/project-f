package adventure

import (
	"fisherevans.com/project/f/internal/game/rpg"
)

type StateGlobalsReader interface {
	rpg.GlobalsReader
	GetEntityReader(id string) (EntityReader, bool)
	GetZonesAt(loc MapLocation) Zones
	Player() EntityReader
	KeysWithPrefix(prefix string) []string
}

type stateGlobals struct {
	baseGlobals rpg.Globals
	state       *State
}

func observeGlobals(baseGlobals rpg.Globals, state *State) *stateGlobals {
	return &stateGlobals{
		baseGlobals: baseGlobals,
		state:       state,
	}
}

func (s *stateGlobals) Set(key string, value any) *rpg.GlobalValue {
	oldValue := s.baseGlobals.Set(key, value)
	newValue := s.baseGlobals.Get(key)
	if s.state.eventDispatcher != nil {
		s.state.eventDispatcher.Dispatch(NewEventGlobalVariableUpdated(key, newValue, oldValue))
	}
	return oldValue
}

func (s *stateGlobals) Delete(key string) *rpg.GlobalValue {
	oldValue := s.baseGlobals.Delete(key)
	if s.state.eventDispatcher != nil {
		s.state.eventDispatcher.Dispatch(NewEventGlobalVariableDeleted(key, oldValue))
	}
	return oldValue
}

func (s *stateGlobals) Get(key string) *rpg.GlobalValue {
	return s.baseGlobals.Get(key)
}

func (s *stateGlobals) GetEntityReader(id string) (EntityReader, bool) {
	var r EntityReader
	var ok bool
	r, ok = s.state.entities.GetEntity(id)
	return r, ok
}

func (s *stateGlobals) GetZonesAt(loc MapLocation) Zones {
	return s.state.zones.ZonesAtSet(loc)
}

func (s *stateGlobals) Player() EntityReader {
	e, _ := s.state.entities.GetEntity(s.state.player)
	return e
}

func (s *stateGlobals) KeysWithPrefix(prefix string) []string {
	return s.baseGlobals.KeysWithPrefix(prefix)
}
