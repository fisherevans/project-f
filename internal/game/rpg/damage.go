package rpg

import (
	"math"

	"github.com/rs/zerolog/log"
)

const (
	// todo make defending absolute, reflective relative
	stanceDefendingMultiplier          = 0.25
	stanceReflectingMultiplierIncoming = 0.5
	stanceReflectingMultiplierOutgoing = 0.75
	stanceVulnerableMultiplier         = 1.5
)

type ionizedModifier struct {
	multiplier     float64
	stackReduction float64
}

var (
	statusFortifiedMultipliersIncoming = map[StatusLevel]float64{
		StatusLevel1: 0.75,
		StatusLevel2: 0.5,
		StatusLevel3: 0.25,
	}
	statusIonizedModifiers = map[StatusLevel]ionizedModifier{
		StatusLevel1: {
			multiplier:     1.5,
			stackReduction: 3,
		},
		StatusLevel2: {
			multiplier:     2,
			stackReduction: 5,
		},
		StatusLevel3: {
			multiplier:     2.5,
			stackReduction: 7,
		},
	}
)

type DamageSource struct {
	BaseDamage   int
	Tempo        int
	Stance       CombatStance
	StatusLevels map[StatusType]StatusLevel
}

type DamageTarget struct {
	Stance       CombatStance
	StatusLevels map[StatusType]StatusLevel
}

type DamageResult struct {
	TargetDamage int
	SourceDamage int

	InterruptSkill bool

	TargetStatusStackReductions map[StatusType]float64
	SourceStatusStackReductions map[StatusType]float64
}

func ComputeDamage(source DamageSource, target DamageTarget) DamageResult {
	targetDamage := float64(source.BaseDamage)
	sourceDamage := 0.0

	log.Info().Msgf("base damage: %f", targetDamage)

	tempoMultiplier := 1.0 + float64(source.Tempo)/20 // x2 @ 20
	targetDamage *= tempoMultiplier

	log.Info().Msgf("after tempo: %f", targetDamage)

	if fortifiedLevel, isFortified := target.StatusLevels[StatusFortified]; isFortified {
		if mult, ok := statusFortifiedMultipliersIncoming[fortifiedLevel]; ok {
			targetDamage *= mult
		}
	}

	targetStackReductions := map[StatusType]float64{}
	if ionizedLevel, isIonized := target.StatusLevels[StatusIonized]; isIonized {
		if modifier, ok := statusIonizedModifiers[ionizedLevel]; ok {
			targetDamage *= modifier.multiplier
			targetStackReductions[StatusIonized] = modifier.stackReduction
		}
	}

	interruptSkill := false

	switch target.Stance {
	case TickStanceDefending:
		targetDamage *= stanceDefendingMultiplier
	case TickStanceVulnerable:
		targetDamage *= stanceVulnerableMultiplier
	case TickStanceExposed:
		interruptSkill = true
	case TickStanceReflecting:
		sourceDamage = targetDamage * stanceReflectingMultiplierOutgoing
		targetDamage *= stanceReflectingMultiplierIncoming
	}

	log.Info().Msgf("after stance %d - target: %f - source: %f", target.Stance, targetDamage, sourceDamage)

	return DamageResult{
		TargetDamage:                int(math.Ceil(targetDamage)), // short of immune, always deal at least 1 damage
		SourceDamage:                int(math.Ceil(sourceDamage)),
		InterruptSkill:              interruptSkill,
		TargetStatusStackReductions: targetStackReductions,
	}
}
