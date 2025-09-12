package rpg

type CapturedPrimortal struct {
	PrimortalType  PrimortalType `yaml:"primortal_type"`
	Nickname       string        `yaml:"nickname"`
	AdditionalSync int           `yaml:"additional_sync"`
	SelectedSkills []SkillId     `yaml:"selected_skills"`
}

func (p CapturedPrimortal) GetMaxSync() int {
	return p.Base().BaseSync + p.AdditionalSync
}

func (p CapturedPrimortal) Base() Primortal {
	return Primortals[p.PrimortalType]
}
