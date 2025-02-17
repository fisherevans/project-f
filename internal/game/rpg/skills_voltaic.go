package rpg

var Skill_ArcBlast = Skill{
	Id:          "arc_blast",
	Name:        "Arc Blast",
	Description: "Release a burst of electricity, shocking the target.",
	Type:        SkillTypeVoltaic,
	Ticks:       simpleDamageSkillTicks(5, DamageMediumAether, 3),
}.register()

var Skill_PlasmaBolt = Skill{
	Id:          "plasma_bolt",
	Name:        "Plasma Bolt",
	Description: "Fire a concentrated plasma shot that burns on impact.",
	Type:        SkillTypeVoltaic,
	Ticks:       simpleDamageSkillTicks(10, DamageMediumAether, 2),
}.register()
