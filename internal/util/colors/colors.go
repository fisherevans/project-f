package colors

var (
	Warm9 = NamedColor(HexColor("#fcef8d")).register("warm_9")
	Warm8 = NamedColor(HexColor("#ffb879")).register("warm_8")
	Warm7 = NamedColor(HexColor("#ea6262")).register("warm_7")
	Warm5 = NamedColor(HexColor("#cc425e")).register("warm_5")
	Warm3 = NamedColor(HexColor("#a32858")).register("warm_3")
	Warm2 = NamedColor(HexColor("#751756")).register("warm_2")
	Warm1 = NamedColor(HexColor("#611851")).register("warm_1")

	Brown9 = NamedColor(HexColor("#f2ae99")).register("brown_9")
	Brown7 = NamedColor(HexColor("#c97373")).register("brown_7")
	Brown5 = NamedColor(HexColor("#a6555f")).register("brown_5")
	Brown3 = NamedColor(HexColor("#873555")).register("brown_3")
	Brown1 = NamedColor(Warm1).register("brown_1")

	Green9 = NamedColor(Warm9).register("green_9")
	Green6 = NamedColor(HexColor("#abdd64")).register("green_6")
	Green4 = NamedColor(HexColor("#6bc96c")).register("green_4")
	Green1 = NamedColor(HexColor("#5ba675")).register("green_1")

	Blurple9 = NamedColor(HexColor("#aee2ff")).register("blurple_9")
	Blurple8 = NamedColor(HexColor("#8db7ff")).register("blurple_8")
	Blurple7 = NamedColor(HexColor("#6d80fa")).register("blurple_7")
	Blurple5 = NamedColor(HexColor("#8465ec")).register("blurple_5")
	Blurple3 = NamedColor(HexColor("#834dc4")).register("blurple_3")
	Blurple2 = NamedColor(HexColor("#7d2da0")).register("blurple_2")
	Blurple1 = NamedColor(HexColor("#4e187c")).register("blurple_1")

	Grey9 = NamedColor(HexColor("#d9bdc8")).register("grey_9")
	Grey6 = NamedColor(HexColor("#a6859f")).register("grey_6")
	Grey4 = NamedColor(HexColor("#7b5480")).register("grey_4")
	Grey1 = NamedColor(HexColor("#4a3052")).register("grey_1")

	Pinkish9 = NamedColor(HexColor("#ffc3f2")).register("pinkish_9")
	Pinkish7 = NamedColor(HexColor("#ee8fcb")).register("pinkish_7")
	Pinkish5 = NamedColor(HexColor("#d46eb3")).register("pinkish_5")
	Pinkish3 = NamedColor(HexColor("#873e84")).register("pinkish_3")
	Pinkish1 = NamedColor(Grey1).register("pinkish_1")

	Black = NamedColor(HexColor("#1f102a")).register("black")
	White = NamedColor(HexColor("#ffffff")).register("white")
)
