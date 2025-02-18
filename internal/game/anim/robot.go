package anim

var robotFramesPerSecond = 10.0

func RobotDown() *AnimatedSprite {
	return FromTilesheetRow("robot", 2, pigFramesPerSecond)
}

func RobotUp() *AnimatedSprite {
	return FromTilesheetRow("robot", 3, pigFramesPerSecond)
}

func RobotRight() *AnimatedSprite {
	return FromTilesheetRow("robot", 4, pigFramesPerSecond)
}

func RobotLeft() *AnimatedSprite {
	return FromTilesheetRow("robot", 5, pigFramesPerSecond)
}
