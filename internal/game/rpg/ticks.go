package rpg

type TickDisplayType int

const (
	TickDisplayNone TickDisplayType = iota
	TickDisplayDamage
)

type CombatStance int

const (
	TickStanceNone       CombatStance = iota
	TickStanceDefending               // reduce damage
	TickStanceReflecting              // reflect some damage back
	TickStanceVulnerable              // take extra damage
	TickStanceExposed                 // stunned if hit
)

type SkillTickDamage struct {
	Amount         int
	RandomVariance int
}

type SkillTickEffect struct {
	Damage *SkillTickDamage
}

type SkillTick struct {
	Effects     []SkillTickEffect
	DisplayType TickDisplayType
	StanceType  CombatStance
}

type SkillTicks []SkillTick

func skillTicks() SkillTicks {
	return []SkillTick{}
}

func (sts SkillTicks) tick(ts ...SkillTick) SkillTicks {
	sts = append(sts, ts...)
	return sts
}

func damageTick() SkillTick {
	return SkillTick{
		DisplayType: TickDisplayDamage,
	}
}

func nothingTick() SkillTick {
	return SkillTick{
		DisplayType: TickDisplayNone,
	}
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
		tick(damageTick().damage(damage, 0)).
		tick(nothingTick().repeat(duration - 1)...)
}
