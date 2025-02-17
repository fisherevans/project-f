package rpg

var Skill_FlameSurge = Skill{
	Id:          "flame_surge",
	Name:        "Flame Surge",
	Description: "Unleash a wave of intense heat, scorching foes.",
	Type:        SkillTypeThermal,
	Ticks:       simpleDamageSkillTicks(5, DamageMediumAether, 6),
}.register()

var Skill_HeatCrash = Skill{
	Id:          "heat_crash",
	Name:        "Heat Crash",
	Description: "Superheat your body and slam into an enemy.",
	Type:        SkillTypeThermal,
	Ticks:       simpleDamageSkillTicks(10, DamageMediumAether, 2),
}.register()
