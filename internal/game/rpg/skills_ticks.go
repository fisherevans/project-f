package rpg

type CombatStance int

const (
	TickStanceNone       CombatStance = iota
	TickStanceDefending               // shield - reduce damage
	TickStanceReflecting              // bouncing arrow - reflect some damage back
	TickStanceVulnerable              // cross our shield - take extra damage
	TickStanceExposed                 // !!! - interrupts following ticks, can be stunned
)

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

type SkillTick struct {
	Effects    []SkillTickEffect
	StanceType CombatStance
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
