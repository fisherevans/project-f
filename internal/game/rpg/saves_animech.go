package rpg

const BaseAnimechShield = 25
const BaseAnimechSync = 25

type Animech struct {
	SkillSet          SkillSet `yaml:"skill_set"`
	AdditionalShield  int      `yaml:"additional_shield"`
	AdditionalSync    int      `yaml:"additional_sync"`
	AnimechExperience int      `yaml:"animech_experience"`
}

func (a Animech) GetMaxShield() int {
	return BaseAnimechShield + a.AdditionalShield
}

func (a Animech) GetMaxSync() int {
	return BaseAnimechSync + a.AdditionalSync
}
