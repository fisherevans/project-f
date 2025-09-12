package rpg

var Skill_Tackle = Skill{
	Id:          "tackle",
	Name:        "Tackle",
	Description: "Tackle an enemy, dealing kinetic damage.",
	Type:        SkillTypeKinetic,
	Ticks:       simpleDamageSkillTicks(3, DamageMediumPhysical, 3),
}.register()

var Skill_Crush = Skill{
	Id:          "crush",
	Name:        "Crush",
	Description: "Slam down with immense force, dealing heavy kinetic damage.",
	Type:        SkillTypeKinetic,
	Ticks: skillTicks().
		tick(damageTick().damage(6, 4, DamageMediumPhysical)).
		tick(nothingTick().repeat(2)...),
}.register()

var Skill_Shunt = Skill{
	Id:          "shunt",
	Name:        "Shunt",
	Description: "Shunt the foe and enter a defencive stance.",
	Type:        SkillTypeKinetic,
	Ticks: skillTicks().
		tick(damageTick().damage(6, 4, DamageMediumPhysical)).
		tick(nothingTick().repeat(2)...).
		tick(stanceTick(TickStanceDefending).repeat(3)...),
}.register()

var Skill_Block = Skill{
	Id:          "block",
	Name:        "Block",
	Description: "Raise your defences briefly",
	Type:        SkillTypeKinetic,
	Ticks: skillTicks().
		tick(stanceTick(TickStanceDefending).repeat(3)...),
}.register()

var Skill_Riposte = Skill{
	Id:          "riposte",
	Name:        "Riposte",
	Description: "Reflect damage after being exposed",
	Type:        SkillTypeKinetic,
	Ticks: skillTicks().
		tick(stanceTick(TickStanceExposed).repeat(3)...).
		tick(stanceTick(TickStanceReflecting).repeat(3)...),
}.register()

var Skill_DrawnBlow = Skill{
	Id:          "drawn_blow",
	Name:        "Drawn Blow",
	Description: "Make a huge hit after being exposed",
	Type:        SkillTypeKinetic,
	Ticks: skillTicks().
		tick(stanceTick(TickStanceExposed).repeat(4)...).
		tick(damageTick().damage(15, 4, DamageMediumPhysical)),
}.register()
