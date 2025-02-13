package rpg

var Skill_GammaStrike = Skill{
	Id:          "gamma_strike",
	Name:        "Gamma Strike",
	Description: "Deliver a strike laced with radiation, weakening foes.",
	Type:        SkillTypeGamma,
}.register()

var Skill_FalloutWave = Skill{
	Id:          "fallout_wave",
	Name:        "Fallout Wave",
	Description: "Release a lingering radiation field that saps enemy strength.",
	Type:        SkillTypeGamma,
}.register()
