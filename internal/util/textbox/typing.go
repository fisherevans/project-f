package textbox

type TypingController interface {
	TypeSome(elapsed float64, typeFaster bool) int
}
type defaultTypingController struct {
	partialWork      float64
	timePerCharacter float64
	fasterScale      float64
}

func (c *defaultTypingController) TypeSome(timeDelta float64, typeFaster bool) int {
	if typeFaster {
		timeDelta = timeDelta * c.fasterScale
	}
	c.partialWork += timeDelta
	typingDone := 0
	for c.partialWork >= c.timePerCharacter {
		typingDone++
		c.partialWork -= c.timePerCharacter
	}
	return typingDone
}
