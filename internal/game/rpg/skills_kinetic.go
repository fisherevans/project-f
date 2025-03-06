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
	Ticks: skillTicks().
		tick(normalTick().damage(6, 4, DamageMediumPhysical)).
		tick(nothingTick().repeat(2)...),
}.register()
