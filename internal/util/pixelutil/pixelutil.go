package pixelutil

import (
	"github.com/gopxl/pixel/v2"
	"image/color"
)

type BoundedDrawable interface {
	Draw(target pixel.Target, m pixel.Matrix)
	DrawColorMask(target pixel.Target, m pixel.Matrix, mask color.Color)
	Bounds() pixel.Rect
}

type drawableSprite struct {
	sprite *pixel.Sprite
}

func (d *drawableSprite) Draw(target pixel.Target, m pixel.Matrix) {
	d.sprite.Draw(target, m)
}

func (d *drawableSprite) DrawColorMask(target pixel.Target, m pixel.Matrix, mask color.Color) {
	d.sprite.DrawColorMask(target, m, mask)
}

func (d *drawableSprite) Bounds() pixel.Rect {
	return pixel.R(0, 0, d.sprite.Frame().W(), d.sprite.Frame().H())
}

func DrawableSprite(sprite *pixel.Sprite) BoundedDrawable {
	return &drawableSprite{sprite}
}
