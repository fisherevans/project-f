package map_editor

import (
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

var entitySpritePaste = resources.GetSprite("ui", 5, 1).Sprite
var entitySpriteExists = resources.GetSprite("ui", 6, 1).Sprite
var entitySpriteEdit = resources.GetSprite("ui", 7, 1).Sprite
var entitySpriteDelete = resources.GetSprite("ui", 8, 1).Sprite
var entitySpriteAdd = resources.GetSprite("ui", 9, 1).Sprite
var entitySpriteOverlay = resources.GetSprite("ui", 10, 1).Sprite

func (m *MapEditor) DrawEntityOverlay(ctx *Context, win *opengl.Window, target pixel.Target, cameraMatrix pixel.Matrix) {
	if !ctx.MouseInCanvas {
		return
	}

	_, selectedEntity := m.mouseEntity(ctx)

	var sprites []*pixel.Sprite
	sprites = append(sprites, entitySpriteOverlay)
	if win.Pressed(pixel.KeyLeftShift) {
		if selectedEntity != nil {
			sprites = append(sprites, entitySpriteDelete)
		} else if m.lastDeletedEntity != nil {
			sprites = append(sprites, entitySpritePaste)
		}
	} else {
		if selectedEntity == nil {
			sprites = append(sprites, entitySpriteAdd)
		} else {
			sprites = append(sprites, entitySpriteEdit)
		}
	}

	drawMatrix := cameraMatrix.Moved(pixel.V(float64(ctx.MouseMapLocation.X*resources.TileSize), float64(ctx.MouseMapLocation.Y*resources.TileSize)))
	for _, sprite := range sprites {
		sprite.Draw(target, drawMatrix)
	}
}
