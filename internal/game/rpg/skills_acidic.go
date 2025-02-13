package rpg

var Skill_AcidSpit = Skill{
	Id:          "acid_spit",
	Name:        "Acid Spit",
	Description: "Expel a corrosive substance that melts defenses.",
	Type:        SkillTypeAcidic,
}.register()

var Skill_Corrode = Skill{
	Id:          "corrode",
	Name:        "Corrode",
	Description: "Slowly break down an enemy's armor with acid.",
	Type:        SkillTypeAcidic,
}.register()
