package anim

func IdleRobot() *AnimatedSprite {
	return FromTilesheetRow("destroyer_robot_idle", 1, 4)
}

func IdlePlent() *AnimatedSprite {
	return FromTilesheetRow("plent_idle", 1, 5)
}
