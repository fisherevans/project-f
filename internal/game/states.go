package game

import (
	"fmt"
	"reflect"

	"fisherevans.com/project/f/internal/game/rpg"
)

func init() {
	RegisterStateFactory(swapStateIntent)
}

func swapStateIntent(s SwapStateIntent) State {
	return s.State
}

var stateFactories = map[reflect.Type]func(any) State{}

func RegisterStateFactory[T any](fn func(T) State) {
	key := reflect.TypeOf((*T)(nil)).Elem()
	stateFactories[key] = func(cfg any) State {
		c, ok := cfg.(T)
		if !ok {
			panic(fmt.Sprintf("config type %T does not match %v", cfg, key))
		}
		return fn(c)
	}
}

func createState(config any) (State, error) {
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

func DefaultSelectorState() SelectIntent {
	i := SelectIntent{}
	i = i.With("Adventure", func() any {
		return AdventureIntent{
			MapName: "map1",
		}
	})

	fight := func(p rpg.PrimortalType) {
		i = i.With("Fight "+rpg.Primortals[p].Name, func() any {
			return CombatIntent{
				Opponent:   p,
				Background: "combat/background_sylvoria",
				OnComplete: func(r CombatIntentResult) {
					ctx.Notify("Combat complete!")
					SetActiveStateIntent(DefaultSelectorState())
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
