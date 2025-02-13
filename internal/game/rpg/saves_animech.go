package rpg

const BaseAnimechIntegrity = 25
const BaseAnimechShield = 25
const BaseShieldRegen = 3
const BasePrimortalSlots = 1

type Animech struct {
	AdditionalIntegrity      int `yaml:"additional_integrity"`
	AdditionalShield         int `yaml:"additional_shield"`
	AdditionalShieldRegen    int `yaml:"additional_shield_regen"`
	AdditionalPrimortalSlots int `yaml:"additional_primortal_slots"`
	UpgradePoints            int `yaml:"upgrade_points"`
}
