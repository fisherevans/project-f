package game

import (
	"fisherevans.com/project/f/internal/game/rpg"
)

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
	PlayerWon bool
}

type CombatIntentComplete func(combatState State, r CombatIntentResult)

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
	HasTempo      bool
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

type CombatReward struct {
	ExperiencePoints int
	ResearchPoints   int
	ResearchType     rpg.PrimortalType
}

type CombatIntent struct {
	Player           CombatPlayer
	Opponent         CombatOpponent
	Reward           CombatReward
	Background       string
	TrainingSequence string
	OnComplete       CombatIntentComplete
}

type AdventureIntent struct {
	MapName  string
	Waypoint string
}

type SwapStateIntent struct {
	State State
}

type XenologIntent struct {
	Background        State
	PrimortalsEnabled bool
}

type ComputerIntent struct {
	Background State
}

type ComputerReturnData struct {
	PlanetName       string
	Waypoint         string
	PlanetSpriteName string
}

type TravelIntent struct {
	ToIntent         any
	PlanetSpriteName string
}

type BaseTransitionIntent struct {
	From     State
	ToState  State
	ToIntent any
}

type TransitionFadeIntent struct {
	BaseTransitionIntent
	Duration float64
}

type TransitionSwirlIntent struct {
	BaseTransitionIntent
	Duration float64
}

type TransitionGlitchIntent struct {
	BaseTransitionIntent
	Duration float64
}

type TransitionFlushIntent struct {
	BaseTransitionIntent
	Duration float64
}
