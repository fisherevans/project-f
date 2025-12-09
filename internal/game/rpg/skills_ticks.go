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

func pounce() SkillTickCombatantTransformation {
	return SkillTickCombatantTransformation{
		Type:        SkillTickCombatantTransformationTypePounce,
		Speed:       1.0,
		Repetitions: 1,
	}
}

func recoil() SkillTickCombatantTransformation {
	return SkillTickCombatantTransformation{
		Type:        SkillTickCombatantTransformationTypeRecoil,
		Speed:       1.0,
		Repetitions: 1,
	}
}

func wiggle() SkillTickCombatantTransformation {
	return SkillTickCombatantTransformation{
		Type:        SkillTickCombatantTransformationTypeWiggle,
		Speed:       1.0,
		Repetitions: 1,
	}
}

func hop(count int) SkillTickCombatantTransformation {
	return SkillTickCombatantTransformation{
		Type:        SkillTickCombatantTransformationTypeHop,
		Speed:       1.0,
		Repetitions: count,
	}
}

/*
- transformation
  - style (point, squish, recoil)
  - repetitions
  - speed?
- masking
  - multiple layers?
  - color/opacity, flashing?
*/

type SkillTickAnimationConfig struct {
	SourceTransformations []SkillTickCombatantTransformation
	TargetTransformations []SkillTickCombatantTransformation
}

func newAnimation() SkillTickAnimationConfig {
	return SkillTickAnimationConfig{}
}

func (c SkillTickAnimationConfig) sourceTransformations(ts ...SkillTickCombatantTransformation) SkillTickAnimationConfig {
	c.SourceTransformations = ts
	return c
}

func (c SkillTickAnimationConfig) targetTransformations(ts ...SkillTickCombatantTransformation) SkillTickAnimationConfig {
	c.TargetTransformations = ts
	return c
}

type SkillTick struct {
	Effects         []SkillTickEffect
	StanceType      CombatStance
	AnimationConfig SkillTickAnimationConfig
}

type SkillTicks []SkillTick

func skillTicks() SkillTicks {
	return []SkillTick{}
}

func (sts SkillTicks) add(ts ...SkillTick) SkillTicks {
	sts = append(sts, ts...)
	return sts
}

func tick() SkillTick {
	return SkillTick{}
}

func stanceTick(stance CombatStance) SkillTick {
	return SkillTick{
		StanceType: stance,
	}
}

func (st SkillTick) damageAmount(amount int) SkillTick {
	return st.damageAmountVaried(amount, 0)
}

func (st SkillTick) damageAmountVaried(amount, variance int) SkillTick {
	return st.damage(&SkillTickDamage{
		Amount:         amount,
		RandomVariance: variance,
	})
}

func (st SkillTick) damage(dmg *SkillTickDamage) SkillTick {
	st.Effects = append(st.Effects, SkillTickEffect{
		Damage: dmg,
	})
	return st
}

func (st SkillTick) animate(config SkillTickAnimationConfig) SkillTick {
	st.AnimationConfig = config
	return st
}

func (st SkillTick) statusSelf(status StatusType, stacks float64) SkillTick {
	return st.status(&SkillTickStatus{
		Status: status,
		Stacks: stacks,
	}, true)
}

func (st SkillTick) statusOpponent(status StatusType, stacks float64) SkillTick {
	return st.status(&SkillTickStatus{
		Status: status,
		Stacks: stacks,
	}, false)
}

func (st SkillTick) status(effect *SkillTickStatus, self bool) SkillTick {
	st.Effects = append(st.Effects, SkillTickEffect{
		Status: effect,
		Self:   self,
	})
	return st
}

func (st SkillTick) repeat(n int) []SkillTick {
	var out []SkillTick
	for i := 0; i < n; i++ {
		out = append(out, st)
	}
	return out
}

func (t SkillTick) validate() []string {
	return nil
}

func simpleDamageSkillTicks(damage int, duration int) []SkillTick {
	return skillTicks().
		add(tick().damageAmount(damage)).
		add(tick().repeat(duration - 1)...)
}
