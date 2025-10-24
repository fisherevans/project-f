package events

import (
	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

type WorldStateReader interface {
	Get(key string) any
	GetAsString(key string) string
	Has(key string) bool
	GetRun() *rpg.Run
	ToGojaValue(vm *goja.Runtime) goja.Value
}

type WorldState interface {
	WorldStateReader
	Set(key string, value any)
	SetFromGojaValue(vm *goja.Runtime, value goja.Value)
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

func (s *mapState) ToGojaValue(vm *goja.Runtime) goja.Value {
	if s == nil {
		return vm.NewObject()
	}
	if len(s.data) == 0 {
		return vm.NewObject()
	}
	// Create a new JS object and copy properties
	// This makes it mutable in JS
	obj := vm.NewObject()
	for k, v := range s.data {
		if err := obj.Set(k, v); err != nil {
			log.Warn().Err(err).Msgf("Failed to set property %s", k)
		}
	}
	return obj
}

func (s *mapState) SetFromGojaValue(vm *goja.Runtime, value goja.Value) {
	if value == nil {
		s.data = nil
		return
	}
	exported := value.Export()
	exportedMap, ok := exported.(map[string]interface{})
	if !ok {
		log.Error().Msgf("received a non-map object value from goja: %#v", exported)
		return
	}
	s.data = exportedMap
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
