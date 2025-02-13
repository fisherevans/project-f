package rpg

type CapturedPrimortal struct {
	Nickname string `yaml:"nickname"`

	AdditionalSync            int `yaml:"additional_sync"`
	AdditionalPhysicalAttack  int `yaml:"additional_physical_attack"`
	AdditionalPhysicalDefence int `yaml:"additional_physical_defence"`
	AdditionalAetherAttack    int `yaml:"additional_aether_attack"`
	AdditionalAetherDefence   int `yaml:"additional_aether_defence"`

	ResearchPoints int `yaml:"research_points"`

	SelectedSkills  []SkillId `yaml:"selected_skills"`
	AvailableSkills []SkillId `yaml:"available_skills"`
}
