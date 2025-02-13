package rpg

var Skill_FlameSurge = Skill{
	Id:          "flame_surge",
	Name:        "Flame Surge",
	Description: "Unleash a wave of intense heat, scorching foes.",
	Type:        SkillTypeThermal,
}.register()

var Skill_HeatCrash = Skill{
	Id:          "heat_crash",
	Name:        "Heat Crash",
	Description: "Superheat your body and slam into an enemy.",
	Type:        SkillTypeThermal,
}.register()
