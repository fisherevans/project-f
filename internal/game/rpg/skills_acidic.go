package rpg

var Skill_AcidSpit = Skill{
	Id:          "acid_spit",
	Name:        "Acid Spit",
	Description: "Expel a corrosive substance that melts defenses.",
	Type:        SkillTypeAcidic,
	Ticks:       simpleDamageSkillTicks(5, DamageMediumPhysical, 3),
}.register()

var Skill_Corrode = Skill{
	Id:          "corrode",
	Name:        "Corrode",
	Description: "Slowly break down an enemy's armor with acid.",
	Type:        SkillTypeAcidic,
	Ticks:       simpleDamageSkillTicks(3, DamageMediumAether, 5),
}.register()
