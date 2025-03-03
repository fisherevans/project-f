package typecolors

import (
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
)

func SkillTypeColor(skillType rpg.SkillType) colors.NamedColor {
	switch skillType {
	case rpg.SkillTypeKinetic:
		return colors.Grey4
	case rpg.SkillTypeVoltaic:
		return colors.Green9
	case rpg.SkillTypeThermal:
		return colors.Warm5
	case rpg.SkillTypeSonic:
		return colors.Grey9
	case rpg.SkillTypeMagnetic:
		return colors.Warm9
	case rpg.SkillTypeAcidic:
		return colors.Green6
	case rpg.SkillTypeGamma:
		return colors.Blurple5
	case rpg.SkillTypeAbyssal:
		return colors.Blurple3
	default:
		return colors.White
	}
}
