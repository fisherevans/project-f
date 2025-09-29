package rpg

type CombatStance int

const (
	TickStanceNone       CombatStance = iota
	TickStanceDefending               // shield - reduce damage
	TickStanceReflecting              // bouncing arrow - reflect some damage back
	TickStanceVulnerable              // cross our shield - take extra damage
	TickStanceExposed                 // !!! - interrupts following ticks, can be stunned
)

type SkillTickDamage struct {
	Amount         int
	RandomVariance int
}

type SkillTickStatus struct {
	Status StatusType
	Stacks float64
}

type SkillTickEffect struct {
	Damage *SkillTickDamage
	Status *SkillTickStatus
}

type SkillTick struct {
	Effects    []SkillTickEffect
	StanceType CombatStance
}

type SkillTicks []SkillTick

func skillTicks() SkillTicks {
	return []SkillTick{}
}

func (sts SkillTicks) tick(ts ...SkillTick) SkillTicks {
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

func (st SkillTick) damage(amount, variance int) SkillTick {
	st.Effects = append(st.Effects, SkillTickEffect{
		Damage: &SkillTickDamage{
			Amount:         amount,
			RandomVariance: variance,
		},
	})
	return st
}

func (st SkillTick) status(status StatusType, stacks float64) SkillTick {
	st.Effects = append(st.Effects, SkillTickEffect{
		Status: &SkillTickStatus{
			Status: status,
			Stacks: stacks,
		},
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
		tick(tick().damage(damage, 0)).
		tick(tick().repeat(duration - 1)...)
}
