package rpg

type TickDisplayType int

const (
	TickDisplayNone TickDisplayType = iota
	TickDisplayNormal
)

type SkillTickDamage struct {
	Medium DamageMedium
	Amount int
}

type SkillTickEffect struct {
	StaticDamage *SkillTickDamage
}

type SkillTick struct {
	Effects     []SkillTickEffect
	DisplayType TickDisplayType
}

func (t SkillTick) validate() []string {
	return nil
}

func doNothingTick() SkillTick {
	return SkillTick{}
}

func staticDamageTick(damage int, medium DamageMedium) SkillTick {
	return SkillTick{
		DisplayType: TickDisplayNormal,
		Effects: []SkillTickEffect{
			{
				StaticDamage: &SkillTickDamage{
					Medium: medium,
					Amount: damage,
				},
			},
		},
	}
}

func simpleDamageSkillTicks(damage int, medium DamageMedium, duration int) []SkillTick {
	if duration <= 0 {
		panic("duration must be greater than 0")
	}
	ticks := []SkillTick{staticDamageTick(damage, medium)}
	for len(ticks) < duration {
		ticks = append(ticks, doNothingTick())
	}
	return ticks
}
