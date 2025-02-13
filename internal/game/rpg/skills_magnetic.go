package rpg

var Skill_MagneticPull = Skill{
	Id:          "magnetic_pull",
	Name:        "Magnetic Pull",
	Description: "Manipulate magnetic forces to drag enemies closer.",
	Type:        SkillTypeMagnetic,
}.register()

var Skill_RailgunShot = Skill{
	Id:          "railgun_shot",
	Name:        "Railgun Shot",
	Description: "Launch a hyper-accelerated projectile using electromagnetic force.",
	Type:        SkillTypeMagnetic,
}.register()
