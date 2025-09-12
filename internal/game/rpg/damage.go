package rpg

import (
	"math"

	"github.com/rs/zerolog/log"
)

const (
	stanceDefendingMultiplier  = 0.5
	stanceReflectingMultiplier = 0.5
	stanceExposedMultiplier    = 1.5
	stanceVulnerableMultiplier = 1.5
)

type DamageSource struct {
	BaseDamage int
	Tempo      int
	Stance     CombatStance
}

type DamageTarget struct {
	Stance CombatStance
}

type DamageResult struct {
	TargetDamage int
	SourceDamage int
}

func ComputeDamage(source DamageSource, target DamageTarget) DamageResult {
	targetDamage := float64(source.BaseDamage)
	sourceDamage := 0.0

	log.Info().Msgf("base damage: %f", targetDamage)

	tempoMultiplier := 1.0 + float64(source.Tempo)/20 // x2 @ 20
	targetDamage *= tempoMultiplier

	log.Info().Msgf("after tempo: %f", targetDamage)

	switch target.Stance {
	case TickStanceDefending:
		targetDamage *= stanceDefendingMultiplier
	case TickStanceVulnerable:
		targetDamage *= stanceVulnerableMultiplier
	case TickStanceExposed:
		targetDamage *= stanceExposedMultiplier
		// todo apply stun
	case TickStanceReflecting:
		targetDamage *= stanceReflectingMultiplier
		sourceDamage = float64(source.BaseDamage) * (1 - stanceReflectingMultiplier)
	}

	log.Info().Msgf("after stance %d - target: %f - source: %f", target.Stance, targetDamage, sourceDamage)

	return DamageResult{
		TargetDamage: int(math.Ceil(targetDamage)), // short of immune, always deal at least 1 damage
		SourceDamage: int(math.Ceil(sourceDamage)),
	}
}
