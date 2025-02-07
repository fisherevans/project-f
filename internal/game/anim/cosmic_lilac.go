package anim

var tilesheetCosmicChars = "cosmic_lilac_chars"

var astronautFps = 2.0

func AstronautDown() *AnimatedSprite {
	return FromSprites(tilesheetCosmicChars, 4, 2, 6, astronautFps)
}

func AstronautRight() *AnimatedSprite {
	return FromSprites(tilesheetCosmicChars, 5, 2, 6, astronautFps)
}

func AstronautUp() *AnimatedSprite {
	return FromSprites(tilesheetCosmicChars, 6, 2, 6, astronautFps)
}
