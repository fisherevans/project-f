package rpg

import (
	"fmt"
	"math"
	"math/rand"

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
	statusWardedMultipliersIncoming = map[StatusLevel]float64{
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

type CombatantStats struct {
	Tempo        int
	Stance       CombatStance
	StatusLevels map[StatusType]StatusLevel
}

type Damage struct {
	Damage SkillTickDamage
}

type DamageResult struct {
	TargetDamage int
	SourceDamage int

	Missed         bool
	InterruptSkill bool

	TargetStatusStackReductions map[StatusType]float64
	SourceStatusStackReductions map[StatusType]float64
}

func ComputeDamage(dmg SkillTickDamage, source CombatantStats, target CombatantStats) DamageResult {
	if dmg.MissRate > 0 && rand.Float64() < dmg.MissRate {
		return DamageResult{
			Missed: true,
		}
	}

	if dmg.Amount == 5 {
		fmt.Println("hi")
	}

	targetDamage := float64(dmg.Amount)
	if dmg.RandomVariance > 0 {
		targetDamage += float64(dmg.RandomVariance) * (2*rand.Float64() - 1)
	}
	sourceDamage := 0.0

	log.Info().Msgf("base damage: %f", targetDamage)

	tempoMultiplier := 1.0 + float64(source.Tempo)/20 // x2 @ 20
	targetDamage *= tempoMultiplier

	log.Info().Msgf("after tempo: %f", targetDamage)

	if wardedLevel, isWarded := target.StatusLevels[StatusWarded]; isWarded {
		if mult, ok := statusWardedMultipliersIncoming[wardedLevel]; ok {
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

	targetDamage = ScaleByStatus(targetDamage, target.StatusLevels, dmg.ScaledBy.TargetStatus)
	targetDamage = ScaleByStatus(targetDamage, source.StatusLevels, dmg.ScaledBy.SourceStatus)

	log.Info().Msgf("after stance %d - target: %f - source: %f", target.Stance, targetDamage, sourceDamage)

	return DamageResult{
		TargetDamage:                int(math.Ceil(targetDamage)), // short of immune, always deal at least 1 damage
		SourceDamage:                int(math.Ceil(sourceDamage)),
		InterruptSkill:              interruptSkill,
		TargetStatusStackReductions: targetStackReductions,
	}
}

func ScaleByStatus(amount float64, statusLevels map[StatusType]StatusLevel, statusMultipliers map[StatusType]map[StatusLevel]float64) float64 {
	if statusMultipliers == nil || statusLevels == nil {
		return amount
	}
	for status, level := range statusLevels {
		if multiplier, ok := statusMultipliers[status][level]; ok {
			amount *= multiplier
		}
	}
	return amount
}
