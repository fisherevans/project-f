package rpg

var Skill_Tackle = Skill{
	Id:          "tackle",
	Name:        "Tackle",
	Description: "Tackle an enemy, dealing kinetic damage.",
	Type:        SkillTypeKinetic,
	Ticks:       simpleDamageSkillTicks(5, DamageMediumPhysical, 3),
}.register()

var Skill_Crush = Skill{
	Id:          "crush",
	Name:        "Crush",
	Description: "Slam down with immense force, dealing heavy kinetic damage.",
	Type:        SkillTypeKinetic,
	Ticks:       simpleDamageSkillTicks(10, DamageMediumPhysical, 2),
}.register()
