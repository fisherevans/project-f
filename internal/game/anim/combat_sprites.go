package anim

import "fisherevans.com/project/f/internal/resources"

func IdleRobot(atlas *resources.Atlas) *AnimatedSprite {
	return FromTilesheetRow(atlas, "combat/combatants/destroyer_robot_idle", 1, 4)
}

func IdlePlent(atlas *resources.Atlas) *AnimatedSprite {
	return FromTilesheetRow(atlas, "combat/combatants/plent_idle", 1, 5)
}
