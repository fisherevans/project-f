package rpg

import (
    "fmt"
)

type CombatStance int

const (
    TickStanceNone       CombatStance = iota
    TickStanceDefending               // shield - reduce damage
    TickStanceReflecting              // bouncing arrow - reflect some damage back
    TickStanceVulnerable              // cross our shield - take extra damage
    TickStanceExposed                 // !!! - interrupts following ticks, can be stunned
)

func (cs CombatStance) String() string {
    switch cs {
    case TickStanceNone:
        return "None"
    case TickStanceDefending:
        return "Defending"
    case TickStanceReflecting:
        return "Reflecting"
    case TickStanceVulnerable:
        return "Vulnerable"
    case TickStanceExposed:
        return "Exposed"
    default:
        return "?????"
    }
}

func (cs CombatStance) Description() string {
    switch cs {
    case TickStanceNone:
        return "not doing anything special"
    case TickStanceDefending:
        return fmt.Sprintf("Reduce incoming damage by %.f%%", (1.0-stanceDefendingMultiplier)*100)
    case TickStanceReflecting:
        return fmt.Sprintf("Reflect %.f%% of damage back with a %.f%% boost",
            (1-stanceReflectingMultiplierIncoming)*100, stanceReflectingMultiplierOutgoing/stanceReflectingMultiplierIncoming*100)
    case TickStanceVulnerable:
        return fmt.Sprintf("Receive %.f%% more damage", stanceVulnerableMultiplier*100)
    case TickStanceExposed:
        return "Nullify remaining skill ticks if hit"
    }
    return "?????"
}

type SkillTickDamageScalers struct {
    TargetStatus map[StatusType]map[StatusLevel]float64
    SourceStatus map[StatusType]map[StatusLevel]float64
}

type SkillTickDamage struct {
    Amount         int
    RandomVariance int
    MissRate       float64
    ScaledBy       SkillTickDamageScalers
}

type SkillTickStatusRequirement int

const (
    SkillTickStatusRequireNothing = iota
    SkillTickStatusRequireExistingStacks
    SkillTickStatusRequireNoStacks
)

type SkillTickStatus struct {
    Status                StatusType
    Stacks                float64
    RequireExistingStacks SkillTickStatusRequirement
}

type SkillTickEffect struct {
    Self   bool
    Damage *SkillTickDamage
    Status *SkillTickStatus
}

type SkillTickCombatantTransformationType int

const (
    SkillTickCombatantTransformationTypeNone SkillTickCombatantTransformationType = iota
    SkillTickCombatantTransformationTypePounce
    SkillTickCombatantTransformationTypeRecoil
    SkillTickCombatantTransformationTypeWiggle
    SkillTickCombatantTransformationTypeHop
)

type SkillTickCombatantTransformation struct {
    Type        SkillTickCombatantTransformationType
    Speed       float64
    Repetitions int
}

type SkillTickAnimationConfig struct {
    SourceTransformations []SkillTickCombatantTransformation
    TargetTransformations []SkillTickCombatantTransformation
}

type SkillTick struct {
    Effects         []SkillTickEffect
    StanceType      CombatStance
    AnimationConfig SkillTickAnimationConfig
}

func (t SkillTick) validate() []string {
    return nil
}
