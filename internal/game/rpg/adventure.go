package rpg

type DeployedAnimech struct {
	*Animech
	AnimechExperience   int
	PrimortalExperience map[PrimortalType]int
	CurrentIntegrity    int
	CurrentShield       int
	DeployedPrimortals  []*DeployedPrimortal
}

type DeployedPrimortal struct {
	*CapturedPrimortal
	Experience  int
	CurrentSync int

	PhysicalAttackModifier  int
	PhysicalDefenceModifier int

	AetherAttackModifier  int
	AetherDefenceModifier int
}
