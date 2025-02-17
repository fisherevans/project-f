package rpg

var Skill_SonicBoom = Skill{
	Id:          "sonic_boom",
	Name:        "Sonic Boom",
	Description: "Generate a high-frequency blast that ruptures defenses.",
	Type:        SkillTypeSonic,
	Ticks:       simpleDamageSkillTicks(5, DamageMediumPhysical, 3),
}.register()

var Skill_EchoWave = Skill{
	Id:          "echo_wave",
	Name:        "Echo Wave",
	Description: "Emit a pulse of sound that damages and disorients foes.",
	Type:        SkillTypeSonic,
	Ticks:       simpleDamageSkillTicks(3, DamageMediumPhysical, 5),
}.register()
