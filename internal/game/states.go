package game

import (
	"fmt"
	"reflect"
)

func DoSwapStateIntent(s SwapStateIntent) State {
	return s.State
}

var stateFactories = map[reflect.Type]func(any) State{}

func RegisterStateFactory[T any](fn func(T) State) {
	key := reflect.TypeOf((*T)(nil)).Elem()
	if _, ok := stateFactories[key]; ok {
		panic(fmt.Sprintf("state factory already registered for %s", key.String()))
	}
	stateFactories[key] = func(cfg any) State {
		c, ok := cfg.(T)
		if !ok {
			panic(fmt.Sprintf("config type %T does not match %v", cfg, key))
		}
		return fn(c)
	}
}

func CreateStateFromIntent(config any) (State, error) {
	if config == nil {
		return nil, fmt.Errorf("nil config")
	}
	key := reflect.TypeOf(config)
	fn, ok := stateFactories[key]
	if !ok {
		return nil, fmt.Errorf("no state factory registered for config type %v", key)
	}
	return fn(config), nil
}
