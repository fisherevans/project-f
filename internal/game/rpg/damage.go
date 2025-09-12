package rpg

import (
	"math"

	"github.com/rs/zerolog/log"
)

const (
	stanceDefendingMultiplier  = 0.5
	stanceReflectingMultiplier = 0.75
	stanceExposedMultiplier    = 1.25
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
	TotalDamage int
}

func ComputeDamage(source DamageSource, target DamageTarget) DamageResult {
	damage := float64(source.BaseDamage)

	log.Info().Msgf("base damage: %f", damage)

	tempoMultiplier := 1.0 + float64(source.Tempo)/20 // x2 @ 20
	damage *= tempoMultiplier

	log.Info().Msgf("after tempo: %f", damage)

	switch target.Stance {
	case TickStanceDefending:
		damage *= stanceDefendingMultiplier
	case TickStanceVulnerable:
		damage *= stanceVulnerableMultiplier
	case TickStanceExposed:
		damage *= stanceExposedMultiplier
		// todo apply stun
	case TickStanceReflecting:
		damage *= stanceReflectingMultiplier
		// todo reflect damage in addition to dulling
	}

	log.Info().Msgf("after stance %d: %f", target.Stance, damage)

	return DamageResult{
		TotalDamage: int(math.Ceil(damage)), // short of immune, always deal at least 1 damage
	}
}
