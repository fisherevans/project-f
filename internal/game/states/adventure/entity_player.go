package adventure

import (
	"fisherevans.com/project/f/internal/game"
	anim2 "fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"github.com/gopxl/pixel/v2"
)

var (
	charAnimation = map[input.Direction]*anim2.AnimatedSprite{
		input.Down:  anim2.PigDown,
		input.Up:    anim2.PigUp,
		input.Right: anim2.PigRight,
		input.Left:  anim2.PigLeft,
	}
)

type Player struct {
	MoveableEntity
}

func (c *Player) Update(ctx *game.Context, adv *State, timeDelta float64) {
	if c.Moving {
		charAnimation[c.LastDirection].Update(timeDelta)
		return
	}
	if ctx.Controls.DPad().IsPressed() {
		dx, dy := ctx.Controls.DPad().GetDirection().GetVector()
		c.TriggerMovement(adv, dx, dy)
	}
	currentDirection := ctx.Controls.DPad().GetDirection()
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
