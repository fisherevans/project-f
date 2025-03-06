package rpg

var Skill_FlameSurge = Skill{
	Id:          "flame_surge",
	Name:        "Flame Surge",
	Description: "Unleash a wave of intense heat, scorching foes.",
	Type:        SkillTypeThermal,
	Ticks: skillTicks().
		tick(normalTick().damage(1, 0, DamageMediumAether)).
		tick(normalTick().damage(2, 0, DamageMediumAether)).
		tick(normalTick().damage(3, 0, DamageMediumAether)).
		tick(normalTick().damage(5, 0, DamageMediumAether)).
		tick(normalTick().damage(8, 0, DamageMediumAether)).
		tick(normalTick().damage(13, 0, DamageMediumAether)),
}.register()

var Skill_HeatCrash = Skill{
	Id:          "heat_crash",
	Name:        "Heat Crash",
	Description: "Superheat your body and slam into an enemy.",
	Type:        SkillTypeThermal,
	Ticks:       simpleDamageSkillTicks(10, DamageMediumAether, 2),
}.register()
