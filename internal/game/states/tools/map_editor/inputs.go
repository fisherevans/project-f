package map_editor

import (
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/multi_select"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/sprite_selector"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/text_entry"
	resources2 "fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fmt"
	"github.com/gopxl/pixel/v2"
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

func (m *MapEditor) readInputs(ctx *Context, timeDelta float64) {
	m.inputsMouse(ctx)
	if m.win.Pressed(pixel.KeyLeftControl) {
		m.inputsControlCommands(ctx)
	} else {
		ctx.DebugTR("hold ctrl: additional commands")
		m.inputsCamera(ctx, timeDelta)
		m.inputsRendering(ctx)
		m.inputsChangeSwatch(ctx)
	}
}

func (m *MapEditor) inputsMouse(ctx *Context) {
	ctx.DebugTR("left mouse: draw")
	ctx.DebugTR("left mouse + shift: delete")
	ctx.DebugTR("right mouse: edit sample")
	layer, exists := resources2.Maps[m.selectedMap].Layers[m.selectedLayer]
	if !exists {
		resources2.Maps[m.selectedMap].Layers[m.selectedLayer] = &resources2.Layer{}
	}
	mouseTile, mouseTileIndex, mouseTileFound := m.mouseTile(ctx, m.selectedMap, m.selectedLayer)
	if m.win.JustPressed(pixel.MouseButtonRight) && mouseTileFound {
		updated := resources2.SwatchSample{
			SpriteId: mouseTile.SpriteId,
		}
		resources2.Swatches[m.swatch.SelectedSwatch].Samples[m.swatch.SelectedSample] = updated
		fmt.Printf("change active tile to (%s)\n", updated.SpriteId)
		return
	}
	activeTile, isSet := resources2.Swatches[m.swatch.SelectedSwatch].Samples[m.swatch.SelectedSample]
	if !isSet {
		return
	}
	if !ctx.MouseInCanvas || !m.win.Pressed(pixel.MouseButtonLeft) {
		return
	}
	if m.win.Pressed(pixel.KeyLeftShift) {
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
	newTile := &resources2.Tile{
		X: ctx.MouseMapLocation.X,
		Y: ctx.MouseMapLocation.Y,
		SpriteId: resources2.SpriteId{
			Tilesheet: activeTile.SpriteId.Tilesheet,
			Column:    activeTile.SpriteId.Column,
			Row:       activeTile.SpriteId.Row,
		},
	}
	layer.Tiles = append(layer.Tiles, newTile)
}

func (m *MapEditor) inputsCamera(ctx *Context, timeDelta float64) {
	ctx.DebugTR("wasd [+ shift]: move camera")
	ctx.DebugTR("scroll [+ shift]: move camera")
	moveAmount := keyboardMoveRate
	if m.win.Pressed(pixel.KeyLeftShift) {
		moveAmount *= 5
	}
	if m.win.Pressed(pixel.KeyW) {
		m.cameraLocation.Y += moveAmount * timeDelta
	}
	if m.win.Pressed(pixel.KeyS) {
		m.cameraLocation.Y -= moveAmount * timeDelta
	}
	if m.win.Pressed(pixel.KeyA) {
		m.cameraLocation.X -= moveAmount * timeDelta
	}
	if m.win.Pressed(pixel.KeyD) {
		m.cameraLocation.X += moveAmount * timeDelta
	}
	m.cameraLocation = m.cameraLocation.Add(m.win.MouseScroll().ScaledXY(pixel.V(-mouseScrollRate, mouseScrollRate)))
}

func (m *MapEditor) inputsChangeSwatch(ctx *Context) {
	ctx.DebugTR("tab: edit sample")
	if m.win.JustPressed(pixel.KeyTab) {
		swapTile := m.swatch.SelectedSample
		ctx.SwapActiveState(sprite_selector.New(m.win, resources2.Swatches[m.swatch.SelectedSwatch].Samples[swapTile], m, func(newSample resources2.SwatchSample) {
			resources2.Swatches[m.swatch.SelectedSwatch].Samples[swapTile] = newSample
		}))
		return
	}

	ctx.DebugTR("0-9: select sample")
	for _, key := range swatchKeys {
		if m.win.JustPressed(key) {
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
	if m.win.JustPressed(pixel.KeyLeft) || m.win.Repeated(pixel.KeyLeft) {
		newIndex = selectedIndex - 1
		if newIndex < 0 {
			newIndex += len(swatchKeySprites)
		}
	} else if m.win.JustPressed(pixel.KeyRight) || m.win.Repeated(pixel.KeyRight) {
		newIndex = selectedIndex + 1
		if newIndex >= len(swatchKeySprites) {
			newIndex -= len(swatchKeySprites)
		}
	}
	if newIndex != selectedIndex {
		if m.win.Pressed(pixel.KeyLeftShift) {
			swap := resources2.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[selectedIndex]]
			resources2.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[selectedIndex]] = resources2.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[newIndex]]
			resources2.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[newIndex]] = swap
		}
		m.swatch.SelectedSample = swatchKeys[newIndex]
	}

	ctx.DebugTR("up/down: change swatch")
	if m.win.JustPressed(pixel.KeyUp) {
		m.swatch.changeSwatch(-1)
	}
	if m.win.JustPressed(pixel.KeyDown) {
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

func (m *MapEditor) inputsRendering(ctx *Context) {
	ctx.DebugTR("h: toggle render mode")

	if m.win.JustPressed(pixel.KeyH) {
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

func (m *MapEditor) inputsControlCommands(ctx *Context) {
	ctx.DebugTR("ctrl + s: save current swatch and map")
	ctx.DebugTR("ctrl + shift + s: save all swatches and maps")
	if m.win.JustPressed(pixel.KeyS) {
		if m.win.Pressed(pixel.KeyLeftShift) {
			resources2.SaveAllMaps()
			resources2.SaveAllSwatches()
		} else {
			resources2.SaveMap(m.selectedMap)
			resources2.SaveSwatch(m.swatch.SelectedSwatch)
		}
	}

	ctx.DebugTR("ctrl + b: select another swatch")
	if m.win.JustPressed(pixel.KeyB) {
		swatches := util.SortedKeys(resources2.Swatches)
		selected := slices.Index(swatches, m.swatch.SelectedSwatch)
		ctx.SwapActiveState(multi_select.New[string](m.win, "Select Swatch", selected, swatches, m, func(newSwatch string) {
			m.swatch.SelectedSwatch = newSwatch
		}))
	}

	ctx.DebugTR("ctrl + n: make a new swatch")
	if m.win.JustPressed(pixel.KeyN) {
		initialSamples := resources2.Swatches[m.swatch.SelectedSwatch].Copy()
		ctx.SwapActiveState(text_entry.New(m.win, "New Swatch Name", m.swatch.SelectedSwatch, m, func(newName string) bool {
			if _, ok := resources2.Swatches[newName]; ok {
				return false
			}
			resources2.Swatches[newName] = initialSamples
			m.swatch.SelectedSwatch = newName
			return true
		}))
	}

	ctx.DebugTR("ctrl + l: change layer")
	if m.win.JustPressed(pixel.KeyL) {
		selected := 0
		for index, layerName := range resources2.MapLayerOrder {
			if layerName == m.selectedLayer {
				selected = index
				break
			}
		}
		ctx.SwapActiveState(multi_select.New(m.win, "Select Layer", selected, resources2.MapLayerOrder, m, func(newName resources2.MapLayerName) {
			if _, ok := resources2.Maps[m.selectedMap].Layers[newName]; !ok {
				resources2.Maps[m.selectedMap].Layers[newName] = &resources2.Layer{}
			}
			m.selectedLayer = newName
		}))
	}
}
