package rpg

var Skill_VoidLash = Skill{
	Id:          "void_lash",
	Name:        "Void Lash",
	Description: "Strike with tendrils of void energy, ignoring armor.",
	Type:        SkillTypeAbyssal,
}.register()

var Skill_Nullwave = Skill{
	Id:          "nullwave",
	Name:        "Nullwave",
	Description: "Emit a pulse that destabilizes energy-based beings.",
	Type:        SkillTypeAbyssal,
}.register()
