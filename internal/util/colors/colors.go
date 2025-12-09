package colors

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/rpg"
)

var (
	Warm9 = registerNamedColor(HexString("#fcef8d"), "warm_9")
	Warm8 = registerNamedColor(HexString("#ffb879"), "warm_8")
	Warm7 = registerNamedColor(HexString("#ea6262"), "warm_7")
	Warm5 = registerNamedColor(HexString("#cc425e"), "warm_5")
	Warm3 = registerNamedColor(HexString("#a32858"), "warm_3")
	Warm2 = registerNamedColor(HexString("#751756"), "warm_2")
	Warm1 = registerNamedColor(HexString("#611851"), "warm_1")

	Brown9 = registerNamedColor(HexString("#f2ae99"), "brown_9")
	Brown7 = registerNamedColor(HexString("#c97373"), "brown_7")
	Brown5 = registerNamedColor(HexString("#a6555f"), "brown_5")
	Brown3 = registerNamedColor(HexString("#873555"), "brown_3")
	Brown1 = registerNamedColor(Warm1.RGBA, "brown_1")

	Green9 = registerNamedColor(Warm9.RGBA, "green_9")
	Green6 = registerNamedColor(HexString("#abdd64"), "green_6")
	Green4 = registerNamedColor(HexString("#6bc96c"), "green_4")
	Green1 = registerNamedColor(HexString("#5ba675"), "green_1")

	Blurple9 = registerNamedColor(HexString("#aee2ff"), "blurple_9")
	Blurple8 = registerNamedColor(HexString("#8db7ff"), "blurple_8")
	Blurple7 = registerNamedColor(HexString("#6d80fa"), "blurple_7")
	Blurple5 = registerNamedColor(HexString("#8465ec"), "blurple_5")
	Blurple3 = registerNamedColor(HexString("#834dc4"), "blurple_3")
	Blurple2 = registerNamedColor(HexString("#7d2da0"), "blurple_2")
	Blurple1 = registerNamedColor(HexString("#4e187c"), "blurple_1")

	Grey9 = registerNamedColor(HexString("#d9bdc8"), "grey_9")
	Grey6 = registerNamedColor(HexString("#a6859f"), "grey_6")
	Grey4 = registerNamedColor(HexString("#7b5480"), "grey_4")
	Grey1 = registerNamedColor(HexString("#4a3052"), "grey_1")

	Pinkish9 = registerNamedColor(HexString("#ffc3f2"), "pinkish_9")
	Pinkish7 = registerNamedColor(HexString("#ee8fcb"), "pinkish_7")
	Pinkish5 = registerNamedColor(HexString("#d46eb3"), "pinkish_5")
	Pinkish3 = registerNamedColor(HexString("#873e84"), "pinkish_3")
	Pinkish1 = registerNamedColor(Grey1.RGBA, "pinkish_1")

	Black = registerNamedColor(HexString("#000000"), "black")
	White = registerNamedColor(HexString("#ffffff"), "white")

	SkillTypeKinetic  = registerNamedColor(HexString("#c5ccdb"), "kinetic")
	SkillTypeVoltaic  = registerNamedColor(HexString("#d1db42"), "voltaic")
	SkillTypeThermal  = registerNamedColor(HexString("#db4f42"), "thermal")
	SkillTypeSonic    = registerNamedColor(HexString("#42dba3"), "sonic")
	SkillTypeMagnetic = registerNamedColor(HexString("#db4278"), "magnetic")
	SkillTypeAcidic   = registerNamedColor(HexString("#63db42"), "acidic")
	SkillTypeGamma    = registerNamedColor(HexString("#9942db"), "gamma")
	SkillTypeAbyssal  = registerNamedColor(HexString("#5942db"), "abyssal")

	ButtonHighlight = registerNamedColor(HexString("#ecd539"), "button_highlight")
)

var (
	XenoLogDark      = registerNamedColor(HexString("#0053ae"), "xenolog_dark")
	XenoLogClear     = registerNamedColor(HexString("#0476d0"), "xenolog_clear")
	XenoLogText      = registerNamedColor(HexString("#a6d8ff"), "xenolog_text")
	XenoLogHighlight = registerNamedColor(HexString("#ffffff"), "xenolog_highlight")
)

var (
	StatusColors = map[rpg.StatusType]pixel.RGBA{
		rpg.StatusWarded:   HexString("#00cfff"),
		rpg.StatusBurning:  HexString("#e6602b"),
		rpg.StatusPoisoned: HexString("#6b2aa6"),
		rpg.StatusIonized:  HexString("#b5e922"),
		rpg.StatusMending:  HexString("#35e922"),
	}
)
