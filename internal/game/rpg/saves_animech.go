package rpg

const BaseAnimechShield = 25
const BaseShieldRegen = 3
const BasePrimortalSlots = 1

type Animech struct {
	AdditionalShield int `yaml:"additional_shield"`
}

func (a Animech) GetMaxShield() int {
	return BaseAnimechShield + a.AdditionalShield
}
