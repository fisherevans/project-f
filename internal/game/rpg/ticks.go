package rpg

type TickDisplayType int

const (
	TickDisplayNone TickDisplayType = iota
	TickDisplayDamage
)

type TickStanceType int

const (
	TickStanceNone       TickStanceType = iota
	TickStanceDefending                 // reduce damage
	TickStanceReflecting                // reflect some damage back
	TickStanceReposing                  // redirect incoming damage (maybe within threshold)
	TickStanceVulnerable                // take extra damage
	TickStanceExposed                   // stunned if hit
)

type SkillTickDamage struct {
	Medium         DamageMedium
	Amount         int
	RandomVariance int
}

type SkillTickEffect struct {
	Damage *SkillTickDamage
}

type SkillTick struct {
	Effects     []SkillTickEffect
	DisplayType TickDisplayType
	StanceType  TickStanceType
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

func stanceTick(stance TickStanceType) SkillTick {
	return SkillTick{
		StanceType: stance,
	}
}

func (st SkillTick) damage(amount, variance int, medium DamageMedium) SkillTick {
	st.Effects = append(st.Effects, SkillTickEffect{
		Damage: &SkillTickDamage{
			Medium:         medium,
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

func simpleDamageSkillTicks(damage int, medium DamageMedium, duration int) []SkillTick {
	return skillTicks().
		tick(damageTick().damage(damage, 0, medium)).
		tick(nothingTick().repeat(duration - 1)...)
}
