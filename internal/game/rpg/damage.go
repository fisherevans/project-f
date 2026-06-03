package rpg

import (
	"fisherevans.com/project/f/internal/util/rng"
	"fmt"
	"math"

	"github.com/rs/zerolog/log"
)

type TempoLevel int

const (
	TempoLevel0 TempoLevel = 0
	TempoLevel1 TempoLevel = 1
	TempoLevel2 TempoLevel = 2
	TempoLevel3 TempoLevel = 3
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
	tempoMultipliers = map[TempoLevel]float64{
		TempoLevel0: 1,
		TempoLevel1: 1.5,
		TempoLevel2: 2,
		TempoLevel3: 2.5,
	}
)

type CombatantStats struct {
	IsPlayer     bool
	TempoLevel   TempoLevel
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

func ComputeDamage(dmg SkillTickDamage, source CombatantStats, target CombatantStats) (result DamageResult) {
	var calcNotes []string
	defer func() {
		log.Info().Msgf("damage result: %v", result)
		for _, note := range calcNotes {
			log.Info().Msgf(" - %s", note)
		}
	}()
	if dmg.MissRate > 0 && rng.Float64() < dmg.MissRate {
		return DamageResult{
			Missed: true,
		}
	}

	targetDamage := float64(dmg.Amount)
	if dmg.RandomVariance > 0 {
		targetDamage += float64(dmg.RandomVariance) * (2*rng.Float64() - 1)
	}
	sourceDamage := 0.0

	calcNotes = append(calcNotes, fmt.Sprintf("base damage: %f", targetDamage))

	tempoMultiplier, exists := tempoMultipliers[source.TempoLevel]
	if !exists {
		log.Error().Msgf("tempo multiplier not found for tempo level: %d", source.TempoLevel)
		tempoMultiplier = tempoMultipliers[TempoLevel0]
	}
	targetDamage *= tempoMultiplier

	calcNotes = append(calcNotes, fmt.Sprintf("after tempo multiplier %.3f: %f", tempoMultiplier, targetDamage))

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

	calcNotes = append(calcNotes, fmt.Sprintf("after stance %d - target: %f - source: %f", target.Stance, targetDamage, sourceDamage))

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
