package rpg

var Skill_FlameSurge = Skill{
	Id:          "flame_surge",
	Name:        "Flame Surge",
	Description: "Unleash a wave of intense heat, scorching foes.",
	Type:        SkillTypeThermal,
	Ticks: []SkillTick{
		staticDamageTick(1, DamageMediumAether),
		staticDamageTick(2, DamageMediumAether),
		staticDamageTick(3, DamageMediumAether),
		staticDamageTick(5, DamageMediumAether),
		staticDamageTick(8, DamageMediumAether),
		staticDamageTick(13, DamageMediumAether),
	},
}.register()

var Skill_HeatCrash = Skill{
	Id:          "heat_crash",
	Name:        "Heat Crash",
	Description: "Superheat your body and slam into an enemy.",
	Type:        SkillTypeThermal,
	Ticks:       simpleDamageSkillTicks(10, DamageMediumAether, 2),
}.register()
