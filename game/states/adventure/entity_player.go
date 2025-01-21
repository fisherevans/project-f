package adventure

import (
	"fisherevans.com/project/f/game"
	"fisherevans.com/project/f/game/anim"
	"fisherevans.com/project/f/game/input"
	"github.com/gopxl/pixel/v2"
)

var (
	charAnimation = map[input.Direction]*anim.AnimatedSprite{
		input.Down:  anim.PigDown,
		input.Up:    anim.PigUp,
		input.Right: anim.PigRight,
		input.Left:  anim.PigLeft,
	}
)

type Player struct {
	MoveableEntity
}

func (c *Player) Update(ctx *game.Context, adv *AdventureState, timeDelta float64) {
	if c.Moving {
		charAnimation[c.LastDirection].Update(timeDelta)
		return
	}
	if adv.controls.DPad().IsPressed() {
		dx, dy := adv.controls.DPad().GetDirection().GetVector()
		c.TriggerMovement(adv, dx, dy)
	}
	currentDirection := adv.controls.DPad().GetDirection()
	if currentDirection != c.LastDirection {
		if currentDirection == input.NotPressed {
			charAnimation[c.LastDirection].Reset()
		} else {
			c.LastDirection = currentDirection
		}
		charAnimation[c.LastDirection].Reset()
	}
}

func (c *Player) Sprite() *pixel.Sprite {
	a, found := charAnimation[c.LastDirection]
	if !found {
		a = charAnimation[input.Down]
	}
	return a.Sprite()
}

func (c *Player) RenderMapLocation() pixel.Vec {
	location := c.MoveableEntity.RenderMapLocation()
	location = location.Add(pixel.V(0, 0.25))
	return location
}
