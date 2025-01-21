package map_editor

import (
	"fisherevans.com/project/f/game/states/map_editor/multi_select"
	"fisherevans.com/project/f/game/states/map_editor/sprite_selector"
	"fisherevans.com/project/f/game/states/map_editor/text_entry"
	"fisherevans.com/project/f/resources"
	"fisherevans.com/project/f/util"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"slices"
)

var swatchKeys = []pixel.Button{
	pixel.Key1,
	pixel.Key2,
	pixel.Key3,
	pixel.Key4,
	pixel.Key5,
	pixel.Key6,
	pixel.Key7,
	pixel.Key8,
	pixel.Key9,
	pixel.Key0,
}

func (m *MapEditor) readInputs(ctx *Context, win *opengl.Window, timeDelta float64) {
	m.inputsMouse(ctx, win)
	if win.Pressed(pixel.KeyLeftControl) {
		m.inputsControlCommands(ctx, win)
	} else {
		ctx.DebugTR("hold ctrl: additional commands")
		m.inputsCamera(ctx, win, timeDelta)
		m.inputsRendering(ctx, win)
		m.inputsChangeSwatch(ctx, win)
	}
}

func (m *MapEditor) inputsMouse(ctx *Context, win *opengl.Window) {
	ctx.DebugTR("left mouse: draw")
	ctx.DebugTR("left mouse + shift: delete")
	ctx.DebugTR("right mouse: edit sample")
	layer, exists := resources.Maps[m.selectedMap].Layers[m.selectedLayer]
	if !exists {
		resources.Maps[m.selectedMap].Layers[m.selectedLayer] = &resources.Layer{}
	}
	mouseTile, mouseTileIndex, mouseTileFound := m.mouseTile(ctx, m.selectedMap, m.selectedLayer)
	if win.JustPressed(pixel.MouseButtonRight) && mouseTileFound {
		updated := resources.SwatchSample{
			SpriteId: mouseTile.SpriteId,
		}
		resources.Swatches[m.swatch.SelectedSwatch].Samples[m.swatch.SelectedSample] = updated
		fmt.Printf("change active tile to (%s)\n", updated.SpriteId)
		return
	}
	activeTile, isSet := resources.Swatches[m.swatch.SelectedSwatch].Samples[m.swatch.SelectedSample]
	if !isSet {
		return
	}
	if !ctx.MouseInCanvas || !win.Pressed(pixel.MouseButtonLeft) {
		return
	}
	if win.Pressed(pixel.KeyLeftShift) {
		if mouseTileFound {
			layer.Tiles = append(layer.Tiles[:mouseTileIndex], layer.Tiles[mouseTileIndex+1:]...)
		}
		return
	}
	if mouseTileFound {
		if mouseTile.SpriteId != activeTile.SpriteId {
			mouseTile.SpriteId.Tilesheet = activeTile.SpriteId.Tilesheet
			mouseTile.SpriteId.Column = activeTile.SpriteId.Column
			mouseTile.SpriteId.Row = activeTile.SpriteId.Row
		}
		return
	}
	newTile := &resources.Tile{
		X: ctx.MouseMapLocation.X,
		Y: ctx.MouseMapLocation.Y,
		SpriteId: resources.SpriteId{
			Tilesheet: activeTile.SpriteId.Tilesheet,
			Column:    activeTile.SpriteId.Column,
			Row:       activeTile.SpriteId.Row,
		},
	}
	layer.Tiles = append(layer.Tiles, newTile)
}

func (m *MapEditor) inputsCamera(ctx *Context, win *opengl.Window, timeDelta float64) {
	ctx.DebugTR("wasd [+ shift]: move camera")
	ctx.DebugTR("scroll [+ shift]: move camera")
	moveAmount := keyboardMoveRate
	if win.Pressed(pixel.KeyLeftShift) {
		moveAmount *= 5
	}
	if win.Pressed(pixel.KeyW) {
		m.cameraLocation.Y += moveAmount * timeDelta
	}
	if win.Pressed(pixel.KeyS) {
		m.cameraLocation.Y -= moveAmount * timeDelta
	}
	if win.Pressed(pixel.KeyA) {
		m.cameraLocation.X -= moveAmount * timeDelta
	}
	if win.Pressed(pixel.KeyD) {
		m.cameraLocation.X += moveAmount * timeDelta
	}
	m.cameraLocation = m.cameraLocation.Add(win.MouseScroll().ScaledXY(pixel.V(-mouseScrollRate, mouseScrollRate)))
}

func (m *MapEditor) inputsChangeSwatch(ctx *Context, win *opengl.Window) {
	ctx.DebugTR("tab: edit sample")
	if win.JustPressed(pixel.KeyTab) {
		swapTile := m.swatch.SelectedSample
		ctx.SwapActiveState(sprite_selector.New(resources.Swatches[m.swatch.SelectedSwatch].Samples[swapTile], m, func(newSample resources.SwatchSample) {
			resources.Swatches[m.swatch.SelectedSwatch].Samples[swapTile] = newSample
		}))
		return
	}

	ctx.DebugTR("0-9: select sample")
	for _, key := range swatchKeys {
		if win.JustPressed(key) {
			m.swatch.SelectedSample = key
		}
	}

	ctx.DebugTR("left/right: select sample")
	ctx.DebugTR("shift + left/right: swap sample")
	var selectedIndex int
	for index, key := range swatchKeys {
		if key == m.swatch.SelectedSample {
			selectedIndex = index
			break
		}
	}
	newIndex := selectedIndex
	if win.JustPressed(pixel.KeyLeft) || win.Repeated(pixel.KeyLeft) {
		newIndex = selectedIndex - 1
		if newIndex < 0 {
			newIndex += len(swatchKeySprites)
		}
	} else if win.JustPressed(pixel.KeyRight) || win.Repeated(pixel.KeyRight) {
		newIndex = selectedIndex + 1
		if newIndex >= len(swatchKeySprites) {
			newIndex -= len(swatchKeySprites)
		}
	}
	if newIndex != selectedIndex {
		if win.Pressed(pixel.KeyLeftShift) {
			swap := resources.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[selectedIndex]]
			resources.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[selectedIndex]] = resources.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[newIndex]]
			resources.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[newIndex]] = swap
		}
		m.swatch.SelectedSample = swatchKeys[newIndex]
	}

	ctx.DebugTR("up/down: change swatch")
	if win.JustPressed(pixel.KeyUp) {
		m.swatch.changeSwatch(-1)
	}
	if win.JustPressed(pixel.KeyDown) {
		m.swatch.changeSwatch(1)
	}
}

var layerRenderModeOrder = []layerRenderMode{
	layerRenderAll,
	layerRenderMix,
	layerRenderTransparent,
	layerRenderBelow,
	layerRenderSelected,
}

func (m *MapEditor) inputsRendering(ctx *Context, win *opengl.Window) {
	ctx.DebugTR("h: toggle render mode")

	if win.JustPressed(pixel.KeyH) {
		current := 0
		for index, mode := range layerRenderModeOrder {
			if mode == m.layerRenderMode {
				current = index
			}
		}
		next := current + 1
		if next >= len(layerRenderModeOrder) {
			next = 0
		}
		m.layerRenderMode = layerRenderModeOrder[next]
	}
	ctx.DebugTL("render mode: %s", m.layerRenderMode)
}

func (m *MapEditor) inputsControlCommands(ctx *Context, win *opengl.Window) {
	ctx.DebugTR("ctrl + s: save current swatch and map")
	ctx.DebugTR("ctrl + shift + s: save all swatches and maps")
	if win.JustPressed(pixel.KeyS) {
		if win.Pressed(pixel.KeyLeftShift) {
			resources.SaveAllMaps()
			resources.SaveAllSwatches()
		} else {
			resources.SaveMap(m.selectedMap)
			resources.SaveSwatch(m.swatch.SelectedSwatch)
		}
	}

	ctx.DebugTR("ctrl + b: select another swatch")
	if win.JustPressed(pixel.KeyB) {
		swatches := util.SortedKeys(resources.Swatches)
		selected := slices.Index(swatches, m.swatch.SelectedSwatch)
		ctx.SwapActiveState(multi_select.New[string]("Select Swatch", selected, swatches, m, func(newSwatch string) {
			m.swatch.SelectedSwatch = newSwatch
		}))
	}

	ctx.DebugTR("ctrl + n: make a new swatch")
	if win.JustPressed(pixel.KeyN) {
		initialSamples := resources.Swatches[m.swatch.SelectedSwatch].Copy()
		ctx.SwapActiveState(text_entry.New("New Swatch Name", m.swatch.SelectedSwatch, m, func(newName string) bool {
			if _, ok := resources.Swatches[newName]; ok {
				return false
			}
			resources.Swatches[newName] = initialSamples
			m.swatch.SelectedSwatch = newName
			return true
		}))
	}

	ctx.DebugTR("ctrl + l: change layer")
	if win.JustPressed(pixel.KeyL) {
		selected := 0
		for index, layerName := range resources.MapLayerOrder {
			if layerName == m.selectedLayer {
				selected = index
				break
			}
		}
		ctx.SwapActiveState(multi_select.New("Select Layer", selected, resources.MapLayerOrder, m, func(newName resources.MapLayerName) {
			if _, ok := resources.Maps[m.selectedMap].Layers[newName]; !ok {
				resources.Maps[m.selectedMap].Layers[newName] = &resources.Layer{}
			}
			m.selectedLayer = newName
		}))
	}
}
