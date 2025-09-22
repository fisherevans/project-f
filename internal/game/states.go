package game

import (
	"fmt"
	"reflect"
)

func init() {
	RegisterStateFactory(swapStateIntent)
}

func swapStateIntent(ctx *Context, s SwapStateIntent) State {
	return s.State
}

var stateFactories = map[reflect.Type]func(*Context, any) State{}

func RegisterStateFactory[T any](fn func(*Context, T) State) {
	key := reflect.TypeOf((*T)(nil)).Elem()
	stateFactories[key] = func(ctx *Context, cfg any) State {
		c, ok := cfg.(T)
		if !ok {
			panic(fmt.Sprintf("config type %T does not match %v", cfg, key))
		}
		return fn(ctx, c)
	}
}

func createState(ctx *Context, config any) (State, error) {
	if config == nil {
		return nil, fmt.Errorf("nil config")
	}
	key := reflect.TypeOf(config)
	fn, ok := stateFactories[key]
	if !ok {
		return nil, fmt.Errorf("no state factory registered for config type %v", key)
	}
	return fn(ctx, config), nil
}

func InitialState() SelectIntent {
	i := SelectIntent{}
	i = i.With("Adventure", func(ctx *Context) any {
		return AdventureIntent{
			MapName: "map1",
			Save:    ctx.GameSave,
		}
	})
	i = i.With("Combat", func(ctx *Context) any {
		return CombatIntent{
			Animech: ctx.GameSave.NewDeployment(),
			OnComplete: func(ctx *Context, r CombatIntentResult) {
				ctx.Notify("Combat complete!")
				ctx.SetActiveStateIntent(InitialState())
			},
		}
	})
	return i
}
