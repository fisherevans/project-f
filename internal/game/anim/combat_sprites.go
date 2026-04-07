package anim

import "fisherevans.com/project/f/internal/resources"

func IdleRobot(atlas *resources.Atlas) *AnimatedSprite {
	return Load(atlas, "combat/combatants/destroyer_robot_idle")
}

func IdlePlent(atlas *resources.Atlas) *AnimatedSprite {
	return Load(atlas, "combat/combatants/plent_idle")
}

func SkillPendingProgress(atlas *resources.Atlas) *AnimatedSprite {
	return Load(atlas, "combat/menu/skill_pending_progress")
}
