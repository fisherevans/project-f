package events

import (
	"fisherevans.com/project/f/internal/game/rpg"
)

type WorldStateReader interface {
	Get(key string) any
	GetAsString(key string) string
	Has(key string) bool
	GetRun() *rpg.Run
}

type WorldState interface {
	WorldStateReader
	Set(key string, value any)
	Delete(key string)
}

func NewWorldState(run *rpg.Run) WorldState {
	return &mapState{
		run: run,
	}
}

type mapState struct {
	data map[string]any
	run  *rpg.Run
}

func (s *mapState) GetRun() *rpg.Run {
	return s.run
}

func (s *mapState) Has(key string) bool {
	_, exists := s.getMap()[key]
	return exists
}

func (s *mapState) Get(key string) any {
	val, _ := s.getMap()[key]
	return val
}

func (s *mapState) GetAsString(key string) string {
	val := s.Get(key)
	str, _ := val.(string)
	return str
}

func (s *mapState) Set(key string, value any) {
	s.getMap()[key] = value
}

func (s *mapState) Delete(key string) {
	delete(s.getMap(), key)
}

func (s *mapState) getMap() map[string]any {
	if s.data == nil {
		s.data = make(map[string]any)
	}
	return s.data
}
