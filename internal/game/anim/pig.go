package anim

var pigFramesPerSecond = 2.0

func PigDown() *AnimatedSprite {
	return FromTilesheetRow("animated_pig", 2, pigFramesPerSecond)
}

func PigUp() *AnimatedSprite {
	return FromTilesheetRow("animated_pig", 3, pigFramesPerSecond)
}

func PigRight() *AnimatedSprite {
	return FromTilesheetRow("animated_pig", 4, pigFramesPerSecond)
}

func PigLeft() *AnimatedSprite {
	return FromTilesheetRow("animated_pig", 5, pigFramesPerSecond)
}
