package rpg

var Skill_ArcBlast = Skill{
	Id:          "arc_blast",
	Name:        "Arc Blast",
	Description: "Release a burst of electricity, shocking the target.",
	Type:        SkillTypeVoltaic,
}.register()

var Skill_PlasmaBolt = Skill{
	Id:          "plasma_bolt",
	Name:        "Plasma Bolt",
	Description: "Fire a concentrated plasma shot that burns on impact.",
	Type:        SkillTypeVoltaic,
}.register()
