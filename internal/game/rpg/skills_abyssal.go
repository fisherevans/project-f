package rpg

var Skill_VoidLash = Skill{
	Id:          "void_lash",
	Name:        "Void Lash",
	Description: "Strike with tendrils of void energy, ignoring armor.",
	Type:        SkillTypeAbyssal,
	Ticks:       simpleDamageSkillTicks(10, DamageMediumAether, 2),
}.register()

var Skill_Nullwave = Skill{
	Id:          "nullwave",
	Name:        "Nullwave",
	Description: "Emit a pulse that destabilizes energy-based beings.",
	Type:        SkillTypeAbyssal,
	Ticks:       simpleDamageSkillTicks(5, DamageMediumAether, 3),
}.register()
