package game

import "fisherevans.com/project/f/internal/game/rpg"

type SelectIntent struct {
	Destinations []SelectIntentDestination
}

func (s SelectIntent) With(name string, intent func(ctx *Context) any) SelectIntent {
	s.Destinations = append(s.Destinations, SelectIntentDestination{
		Name:   name,
		Intent: intent,
	})
	return s
}

type SelectIntentDestination struct {
	Name   string
	Intent func(*Context) any
}

type MenuIntent struct {
	Background State
}

type CombatIntentResult struct {
	PlayerWon      bool
	ResearchPoints int
}

type CombatIntentComplete func(ctx *Context, r CombatIntentResult)

type CombatIntent struct {
	Run        *rpg.Run
	Opponent   rpg.PrimortalType
	OnComplete CombatIntentComplete
	Background string
}

type AdventureIntent struct {
	MapName string
	Save    *rpg.GameSave
}

type SwapStateIntent struct {
	State State
}

type XenologIntent struct {
	Background State
}
