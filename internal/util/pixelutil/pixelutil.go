package pixelutil

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/atlas"
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

type drawableTextureId struct {
	textureId atlas.TextureId
}

func (d *drawableTextureId) Draw(target pixel.Target, m pixel.Matrix) {
	d.textureId.Draw(target, m)
}

func (d *drawableTextureId) DrawColorMask(target pixel.Target, m pixel.Matrix, mask color.Color) {
	d.textureId.DrawColorMask(target, m, mask)
}

func (d *drawableTextureId) Bounds() pixel.Rect {
	return d.textureId.Bounds()
}

func DrawableTextureId(textureId atlas.TextureId) BoundedDrawable {
	return &drawableTextureId{textureId}
}
