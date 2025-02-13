package rpg

import (
	"math"
	"slices"
)

type DamageEffectiveness string

const (
	DamageEffectivenessWeak    DamageEffectiveness = "weak"
	DamageEffectivenessNeutral DamageEffectiveness = "neutral"
	DamageEffectivenessStrong  DamageEffectiveness = "strong"
	DamageEffectivenessImmune  DamageEffectiveness = "immune"
)

func (de DamageEffectiveness) Multiplier() float64 {
	switch de {
	case DamageEffectivenessWeak:
		return 0.5
	case DamageEffectivenessStrong:
		return 2.0
	case DamageEffectivenessImmune:
		return 0.0
	}
	return 1.0
}

var damageEffectivenessLookup = map[SkillType]map[BodyType]DamageEffectiveness{
	SkillTypeKinetic: {
		BodyTypeMetal:     DamageEffectivenessWeak,
		BodyTypeRock:      DamageEffectivenessStrong,
		BodyTypeCrystal:   DamageEffectivenessStrong,
		BodyTypeSynthetic: DamageEffectivenessStrong,
		BodyTypeGas:       DamageEffectivenessImmune,
		BodyTypeAbyssal:   DamageEffectivenessWeak,
	},
	SkillTypeVoltaic: {
		BodyTypeOrganic:   DamageEffectivenessStrong,
		BodyTypeMetal:     DamageEffectivenessWeak,
		BodyTypeRock:      DamageEffectivenessImmune,
		BodyTypeSynthetic: DamageEffectivenessStrong,
		BodyTypeLiquid:    DamageEffectivenessStrong,
	},
	SkillTypeThermal: {
		BodyTypeOrganic: DamageEffectivenessStrong,
		BodyTypeMetal:   DamageEffectivenessStrong,
		BodyTypeLiquid:  DamageEffectivenessImmune,
		BodyTypeAbyssal: DamageEffectivenessWeak,
	},
	SkillTypeSonic: {
		BodyTypeOrganic: DamageEffectivenessWeak,
		BodyTypeRock:    DamageEffectivenessStrong,
		BodyTypeCrystal: DamageEffectivenessWeak,
		BodyTypeGas:     DamageEffectivenessStrong,
	},
	SkillTypeMagnetic: {
		BodyTypeOrganic:   DamageEffectivenessWeak,
		BodyTypeMetal:     DamageEffectivenessStrong,
		BodyTypeCrystal:   DamageEffectivenessImmune,
		BodyTypeSynthetic: DamageEffectivenessStrong,
		BodyTypeGas:       DamageEffectivenessWeak,
		BodyTypeAbyssal:   DamageEffectivenessStrong,
	},
	SkillTypeAcidic: {
		BodyTypeOrganic: DamageEffectivenessStrong,
		BodyTypeMetal:   DamageEffectivenessStrong,
		BodyTypeCrystal: DamageEffectivenessWeak,
		BodyTypeLiquid:  DamageEffectivenessWeak,
	},
	SkillTypeGamma: {
		BodyTypeOrganic:   DamageEffectivenessStrong,
		BodyTypeMetal:     DamageEffectivenessImmune,
		BodyTypeRock:      DamageEffectivenessWeak,
		BodyTypeSynthetic: DamageEffectivenessWeak,
		BodyTypeGas:       DamageEffectivenessStrong,
		BodyTypeAbyssal:   DamageEffectivenessStrong,
	},
	SkillTypeAbyssal: {
		// No explicit values were provided in the table for Abyssal interactions
	},
}

func getDamageEffectiveness(skillType SkillType, bodyType BodyType) DamageEffectiveness {
	if byBodyType, exists := damageEffectivenessLookup[skillType]; exists {
		if effectiveness, exists := byBodyType[bodyType]; exists {
			return effectiveness
		}
	}
	return DamageEffectivenessNeutral
}

const (
	affinityMultiplier = 1.5
)

type DamageSource struct {
	BaseDamage     int
	Affinities     []SkillType
	DamageMedium   DamageMedium
	SkillType      SkillType
	PhysicalAttack int
	AetherAttack   int
}

type DamageTarget struct {
	TargetType      BodyType
	Affinities      []SkillType
	PhysicalDefence int
	AetherDefence   int
}

type DamageResult struct {
	TotalDamage       int
	Effectiveness     DamageEffectiveness
	SourceHadAffinity bool
	TargetHadAffinity bool
}

func ComputeDamage(source DamageSource, target DamageTarget) DamageResult {
	effectiveness := getDamageEffectiveness(source.SkillType, target.TargetType)
	damage := float64(source.BaseDamage) * effectiveness.Multiplier()

	switch source.DamageMedium {
	case DamageMediumPhysical:
		damage *= 1.0 + float64(source.PhysicalAttack)/100.0
		damage *= 1.0 - float64(target.PhysicalDefence)/100.0
	case DamageMediumAether:
		damage *= 1.0 + float64(source.AetherAttack)/100.0
		damage *= 1.0 - float64(target.AetherDefence)/100.0
	}

	// if target has affinity: always reduces damage (regardless of source affinity)
	// else, if source does have affinity, damage is increased
	sourceHasAffinity := slices.Contains(source.Affinities, source.SkillType)
	targetHasAffinity := slices.Contains(target.Affinities, source.SkillType)
	if targetHasAffinity {
		damage /= affinityMultiplier
	} else if sourceHasAffinity {
		damage *= affinityMultiplier
	}

	return DamageResult{
		TotalDamage:       int(math.Ceil(damage)), // short of immune, always deal at least 1 damage
		Effectiveness:     effectiveness,
		SourceHadAffinity: sourceHasAffinity,
		TargetHadAffinity: targetHasAffinity,
	}
}
