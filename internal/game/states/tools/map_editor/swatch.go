package map_editor

import (
	resources2 "fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"image/color"
	"slices"
)

var swatchKeySprites []*pixel.Sprite

func init() {
	for index, _ := range swatchKeys {
		swatchKeySprites = append(swatchKeySprites, resources2.GetSprite("ui", index+1, 2).Sprite)
	}
}

const swatchKeyPadding = 0.25

type Swatch struct {
	SelectedSwatch string
	SelectedSample pixel.Button
	Canvas         *opengl.Canvas
}

func newSwatch() *Swatch {
	s := &Swatch{
		SelectedSample: swatchKeys[0],
		SelectedSwatch: "default",
	}
	s.Canvas = opengl.NewCanvas(pixel.R(0, 0, float64(resources2.TileSize*(1+swatchKeyPadding)*len(swatchKeys)), resources2.TileSizeF64*2))
	return s
}

var swatchSpriteDelete = resources2.GetSprite("ui", 1, 1).Sprite
var swatchSpriteDraw = resources2.GetSprite("ui", 2, 1).Sprite
var swatchSpriteSelected = resources2.GetSprite("ui", 3, 1).Sprite
var swatchSpriteSelectedNumber = resources2.GetSprite("ui", 1, 3).Sprite

func (s *Swatch) DrawCanvasOverlay(ctx *Context, win *opengl.Window, target pixel.Target, cameraMatrix pixel.Matrix) {
	// draw brush
	if ctx.MouseInCanvas {
		drawMatrix := cameraMatrix.Moved(pixel.V(float64(ctx.MouseMapLocation.X*resources2.TileSize), float64(ctx.MouseMapLocation.Y*resources2.TileSize)))
		if win.Pressed(pixel.KeyLeftShift) {
			swatchSpriteDelete.Draw(target, drawMatrix)
		} else {
			spriteId := resources2.Swatches[s.SelectedSwatch].Samples[s.SelectedSample].SpriteId
			sprite := resources2.Sprites[spriteId].Sprite
			sprite.Draw(target, drawMatrix)
			swatchSpriteDraw.Draw(target, drawMatrix)
		}
	}
}

func (s *Swatch) DrawSwatch(ctx *Context, win *opengl.Window) {
	s.Canvas.Clear(color.Transparent)
	for index, key := range swatchKeys {
		tile := resources2.Swatches[s.SelectedSwatch].Samples[key]
		if tile.SpriteId.Tilesheet == "" {
			continue
		}
		tileMatrix := pixel.IM.
			Moved(pixel.V((resources2.TileSizeF64*(1+swatchKeyPadding))*float64(index), 0)).
			Moved(pixel.V(resources2.TileSizeF64/2, resources2.TileSizeF64/2))
		numberMatrix := tileMatrix.Moved(pixel.V(0, resources2.TileSizeF64))
		sprite := resources2.Sprites[tile.SpriteId].Sprite
		sprite.Draw(s.Canvas, tileMatrix)
		swatchKeySprites[index].Draw(s.Canvas, numberMatrix)
		if key == s.SelectedSample {
			swatchSpriteSelected.Draw(s.Canvas, tileMatrix)
			swatchSpriteSelectedNumber.Draw(s.Canvas, numberMatrix)
		}
	}
	s.Canvas.Draw(win, pixel.IM.
		Moved(s.Canvas.Bounds().Center()).
		Moved(pixel.V(resources2.TileSizeF64*swatchKeyPadding, resources2.TileSizeF64*swatchKeyPadding)).
		Scaled(pixel.ZV, ctx.CanvasScale))
}

func (s *Swatch) changeSwatch(amount int) {
	swatches := util.SortedKeys(resources2.Swatches)
	next := slices.Index(swatches, s.SelectedSwatch) + amount
	if next < 0 {
		next += len(swatches)
	}
	if next >= len(swatches) {
		next = 0
	}
	s.SelectedSwatch = swatches[next]
}
