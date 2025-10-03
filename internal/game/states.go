package game

import (
	"fmt"
	"reflect"

	"fisherevans.com/project/f/internal/game/rpg"
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

	fight := func(p rpg.PrimortalType) {
		i = i.With("Fight "+rpg.Primortals[p].Name, func(ctx *Context) any {
			return CombatIntent{
				Run:        &rpg.Run{},
				Opponent:   p,
				Background: "combat/background_sylvoria",
				OnComplete: func(ctx *Context, r CombatIntentResult) {
					ctx.Notify("Combat complete!")
					ctx.SetActiveStateIntent(InitialState())
				},
			}
		})
	}
	fight(rpg.Primortal_Pumbl.Type)
	fight(rpg.Primortal_Myceli.Type)
	fight(rpg.Primortal_Scintail.Type)
	fight(rpg.Primortal_Toxmidge.Type)
	fight(rpg.Primortal_Volteel.Type)

	return i
}
