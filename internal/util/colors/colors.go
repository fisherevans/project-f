package colors

var (
	Warm9 = registerNamedColor(HexColor("#fcef8d"), "warm_9")
	Warm8 = registerNamedColor(HexColor("#ffb879"), "warm_8")
	Warm7 = registerNamedColor(HexColor("#ea6262"), "warm_7")
	Warm5 = registerNamedColor(HexColor("#cc425e"), "warm_5")
	Warm3 = registerNamedColor(HexColor("#a32858"), "warm_3")
	Warm2 = registerNamedColor(HexColor("#751756"), "warm_2")
	Warm1 = registerNamedColor(HexColor("#611851"), "warm_1")

	Brown9 = registerNamedColor(HexColor("#f2ae99"), "brown_9")
	Brown7 = registerNamedColor(HexColor("#c97373"), "brown_7")
	Brown5 = registerNamedColor(HexColor("#a6555f"), "brown_5")
	Brown3 = registerNamedColor(HexColor("#873555"), "brown_3")
	Brown1 = registerNamedColor(Warm1.RGBA, "brown_1")

	Green9 = registerNamedColor(Warm9.RGBA, "green_9")
	Green6 = registerNamedColor(HexColor("#abdd64"), "green_6")
	Green4 = registerNamedColor(HexColor("#6bc96c"), "green_4")
	Green1 = registerNamedColor(HexColor("#5ba675"), "green_1")

	Blurple9 = registerNamedColor(HexColor("#aee2ff"), "blurple_9")
	Blurple8 = registerNamedColor(HexColor("#8db7ff"), "blurple_8")
	Blurple7 = registerNamedColor(HexColor("#6d80fa"), "blurple_7")
	Blurple5 = registerNamedColor(HexColor("#8465ec"), "blurple_5")
	Blurple3 = registerNamedColor(HexColor("#834dc4"), "blurple_3")
	Blurple2 = registerNamedColor(HexColor("#7d2da0"), "blurple_2")
	Blurple1 = registerNamedColor(HexColor("#4e187c"), "blurple_1")

	Grey9 = registerNamedColor(HexColor("#d9bdc8"), "grey_9")
	Grey6 = registerNamedColor(HexColor("#a6859f"), "grey_6")
	Grey4 = registerNamedColor(HexColor("#7b5480"), "grey_4")
	Grey1 = registerNamedColor(HexColor("#4a3052"), "grey_1")

	Pinkish9 = registerNamedColor(HexColor("#ffc3f2"), "pinkish_9")
	Pinkish7 = registerNamedColor(HexColor("#ee8fcb"), "pinkish_7")
	Pinkish5 = registerNamedColor(HexColor("#d46eb3"), "pinkish_5")
	Pinkish3 = registerNamedColor(HexColor("#873e84"), "pinkish_3")
	Pinkish1 = registerNamedColor(Grey1.RGBA, "pinkish_1")

	Black = registerNamedColor(HexColor("#1f102a"), "black")
	White = registerNamedColor(HexColor("#ffffff"), "white")
)
