package anim

func IdleRobot() *AnimatedSprite {
	return FromTilesheetRow("combat/combatants/destroyer_robot_idle", 1, 4)
}

func IdlePlent() *AnimatedSprite {
	return FromTilesheetRow("combat/combatants/plent_idle", 1, 5)
}
