package rpg

type CapturedPrimortal struct {
	PrimortalType PrimortalType `yaml:"primortal_type"`
	Nickname      string        `yaml:"nickname"`

	AdditionalSync            int `yaml:"additional_sync"`
	AdditionalPhysicalAttack  int `yaml:"additional_physical_attack"`
	AdditionalPhysicalDefense int `yaml:"additional_physical_defence"`
	AdditionalAetherAttack    int `yaml:"additional_aether_attack"`
	AdditionalAetherDefense   int `yaml:"additional_aether_defence"`

	ResearchPoints int `yaml:"research_points"`

	SelectedSkills  []SkillId `yaml:"selected_skills"`
	AvailableSkills []SkillId `yaml:"available_skills"`
}

func (p CapturedPrimortal) Base() Primortal {
	return Primortals[p.PrimortalType]
}
