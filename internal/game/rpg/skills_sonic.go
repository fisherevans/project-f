package rpg

var Skill_SonicBoom = Skill{
	Id:          "sonic_boom",
	Name:        "Sonic Boom",
	Description: "Generate a high-frequency blast that ruptures defenses.",
	Type:        SkillTypeSonic,
}.register()

var Skill_EchoWave = Skill{
	Id:          "echo_wave",
	Name:        "Echo Wave",
	Description: "Emit a pulse of sound that damages and disorients foes.",
	Type:        SkillTypeSonic,
}.register()
