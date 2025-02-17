package rpg

var Primortal_Fangroot = Primortal{
	Type:                "fangroot",
	Name:                "Fangroot",
	BaseSync:            12,
	BasePhysicalAttack:  15,
	BasePhysicalDefense: 13,
	BaseAetherAttack:    5,
	BaseAetherDefense:   7,
	BodyType:            BodyTypeOrganic,
	Affinity:            SkillTypeKinetic,
}.register()

var Primortal_Viraglow = Primortal{
	Type:                "viraglow",
	Name:                "Viraglow",
	BaseSync:            14,
	BasePhysicalAttack:  8,
	BasePhysicalDefense: 10,
	BaseAetherAttack:    12,
	BaseAetherDefense:   15,
	BodyType:            BodyTypeOrganic,
	Affinity:            SkillTypeGamma,
}.register()
