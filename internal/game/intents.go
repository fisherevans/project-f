package game

import "fisherevans.com/project/f/internal/game/rpg"

type StartupDeviceIntent struct{}
type StartupCopyrightsIntent struct{}
type StartupDeveloperIntent struct{}

type StartupControlsIntent struct {
	ExitStateIntent any
}
type TitleIntent struct{}

type SelectIntent struct {
	Destinations []SelectIntentDestination
}

func (s SelectIntent) With(name string, intent func() any) SelectIntent {
	s.Destinations = append(s.Destinations, SelectIntentDestination{
		Name:   name,
		Intent: intent,
	})
	return s
}

type SelectIntentDestination struct {
	Name   string
	Intent func() any
}

type MenuIntent struct {
	Background State
}

type CombatIntentResult struct {
	PlayerWon      bool
	ResearchPoints int
}

type CombatIntentComplete func(r CombatIntentResult)

type CombatOpponent struct {
	Type      rpg.PrimortalType
	Archetype string
}

func NewCombatOpponent(t rpg.PrimortalType, archetype string) CombatOpponent {
	return CombatOpponent{
		Type:      t,
		Archetype: archetype,
	}
}

type CombatPlayer struct {
	SkillSet      *rpg.SkillSet
	InitialSync   int
	MaxSync       int
	InitialShield int
	MaxShield     int
}

func NewCombatPlayer(animech *rpg.Animech) CombatPlayer {
	return CombatPlayer{
		SkillSet:      animech.SkillSet,
		InitialSync:   animech.GetMaxSync(),
		MaxSync:       animech.GetMaxSync(),
		InitialShield: animech.GetMaxShield(),
		MaxShield:     animech.GetMaxShield(),
	}
}

type CombatIntent struct {
	Opponent         CombatOpponent
	Player           CombatPlayer
	OnComplete       CombatIntentComplete
	Background       string
	TrainingSequence string
}

type AdventureIntent struct {
	MapName string
}

type SwapStateIntent struct {
	State State
}

type XenologIntent struct {
	Background        State
	PrimortalsEnabled bool
}
