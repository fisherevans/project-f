package map_editor

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/confirm"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/edit_obj"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/multi_select"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/sprite_selector"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor/text_entry"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"gopkg.in/yaml.v3"
	"slices"
	"sort"
	"strings"
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
	if m.win.JustPressed(pixel.KeyM) {
		switch m.editMode {
		case editModeLayers:
			m.editMode = editModeEntities
		default:
			m.editMode = editModeLayers
		}
	}
	switch m.editMode {
	case editModeLayers:
		m.inputsLayersMouse(ctx)
		if m.win.Pressed(pixel.KeyLeftControl) {
			m.inputsControlCommandsLayers(ctx)
		} else {
			ctx.DebugTR("hold ctrl: additional commands")
			m.inputsCamera(ctx, timeDelta)
			m.inputsLayersRenderMode(ctx)
			m.inputsLayersChangeSwatch(ctx)
		}
	case editModeEntities:
		m.inputsEntitiesMouse(ctx)
		if m.win.Pressed(pixel.KeyLeftControl) {
			m.inputsControlCommandsCommon(ctx)
		} else {
			ctx.DebugTR("hold ctrl: additional commands")
			m.inputsCamera(ctx, timeDelta)
		}
	}
}

func (m *MapEditor) inputsLayersMouse(ctx *Context) {
	ctx.DebugTR("left mouse: draw")
	ctx.DebugTR("left mouse + shift: delete")
	ctx.DebugTR("right mouse: edit sample")
	layer, exists := m.getSelectedMap().Layers[m.selectedLayer]
	if !exists {
		m.getSelectedMap().Layers[m.selectedLayer] = &resources.Layer{}
	}
	mouseTile, mouseTileIndex, mouseTileFound := m.mouseTile(ctx, m.selectedLayer)
	if m.win.JustPressed(pixel.MouseButtonRight) && mouseTileFound {
		newSample := &resources.SwatchSample{
			SpriteId: mouseTile.SpriteId,
		}
		resources.Swatches[m.swatch.SelectedSwatch].Samples[m.swatch.SelectedSample] = newSample
		fmt.Printf("change active tile to (%s)\n", newSample.SpriteId)
		return
	}
	activeTile, isSet := resources.Swatches[m.swatch.SelectedSwatch].Samples[m.swatch.SelectedSample]
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

func (m *MapEditor) inputsLayersChangeSwatch(ctx *Context) {
	ctx.DebugTR("tab: edit sample")
	if m.win.JustPressed(pixel.KeyTab) {
		swapTile := m.swatch.SelectedSample
		ctx.SwapActiveState(sprite_selector.New(m.win, resources.Swatches[m.swatch.SelectedSwatch].Samples[swapTile], m, m, func(newSample *resources.SwatchSample) {
			resources.Swatches[m.swatch.SelectedSwatch].Samples[swapTile] = newSample
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
			swap := resources.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[selectedIndex]]
			resources.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[selectedIndex]] = resources.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[newIndex]]
			resources.Swatches[m.swatch.SelectedSwatch].Samples[swatchKeys[newIndex]] = swap
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

func (m *MapEditor) inputsLayersRenderMode(ctx *Context) {
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

func (m *MapEditor) inputsControlCommandsLayers(ctx *Context) {
	m.inputsControlCommandsCommon(ctx)

	ctx.DebugTR("ctrl + s: save current swatch and map")
	ctx.DebugTR("ctrl + shift + s: save all swatches and maps")
	if m.win.JustPressed(pixel.KeyS) {
		if m.win.Pressed(pixel.KeyLeftShift) {
			resources.SaveAllMaps()
			resources.SaveAllSwatches()
		} else {
			resources.SaveMap(m.selectedMap)
			resources.SaveSwatch(m.swatch.SelectedSwatch)
		}
	}

	ctx.DebugTR("ctrl + b: select another swatch")
	if m.win.JustPressed(pixel.KeyB) {
		swatches := util.SortedKeys(resources.Swatches)
		selected := slices.Index(swatches, m.swatch.SelectedSwatch)
		ctx.SwapActiveState(multi_select.New[string](m.win, "Select Swatch", selected, swatches, m, func(ctx *game.Context, newSwatchId int) {
			m.swatch.SelectedSwatch = swatches[newSwatchId]
			ctx.SwapActiveState(m)
		}))
	}

	ctx.DebugTR("ctrl + n: make a new swatch")
	if m.win.JustPressed(pixel.KeyN) {
		initialSamples := resources.Swatches[m.swatch.SelectedSwatch].Copy()
		ctx.SwapActiveState(text_entry.New(m.win, "New Swatch Name", m.swatch.SelectedSwatch, m, func(ctx *game.Context, newName string) {
			if _, ok := resources.Swatches[newName]; ok {
				fmt.Println("invalid swatch")
				return
			}
			resources.Swatches[newName] = initialSamples
			m.swatch.SelectedSwatch = newName
			ctx.SwapActiveState(m)
		}))
	}

	ctx.DebugTR("ctrl + l: change layer")
	if m.win.JustPressed(pixel.KeyL) {
		selected := 0
		for index, layerName := range resources.MapLayerOrder {
			if layerName == m.selectedLayer {
				selected = index
				break
			}
		}
		ctx.SwapActiveState(multi_select.New(m.win, "Select Layer", selected, resources.MapLayerOrder, m, func(ctx *game.Context, newNameId int) {
			newName := resources.MapLayerOrder[newNameId]
			if _, ok := m.getSelectedMap().Layers[newName]; !ok {
				m.getSelectedMap().Layers[newName] = &resources.Layer{}
			}
			m.selectedLayer = newName
			ctx.SwapActiveState(m)
		}))
	}
}

func (m *MapEditor) inputsControlCommandsCommon(ctx *Context) {
	ctx.DebugTR("ctrl + s: save")
	ctx.DebugTR("ctrl + shift + s: save all swatches and maps")
	if m.win.JustPressed(pixel.KeyS) {
		if m.win.Pressed(pixel.KeyLeftShift) {
			resources.SaveAllMaps()
			resources.SaveAllSwatches()
		} else {
			resources.SaveMap(m.selectedMap)
			resources.SaveSwatch(m.swatch.SelectedSwatch)
		}
	}
}

func (m *MapEditor) inputsEntitiesMouse(ctx *Context) {
	ctx.DebugTR("left mouse: add/edit entity")
	ctx.DebugTR("left mouse + shift: delete entity")
	entityId, entity := m.mouseEntity(ctx)
	entityFound := entityId != ""

	if entityFound {
		ctx.DebugBL("entity: " + entityId)
		if len(entity.Metadata) > 0 {
			ctx.DebugBL("metadata:")
			var keys []string
			for k := range entity.Metadata {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				ctx.DebugBL(fmt.Sprintf("  %s: %s", k, entity.Metadata[k]))
			}
		}
	}

	if !ctx.MouseInCanvas {
		return
	}

	if !m.win.JustPressed(pixel.MouseButtonLeft) {
		return
	}

	if m.win.Pressed(pixel.KeyLeftShift) {
		// paste
		if !entityFound {
			if m.lastDeletedEntity == nil {
				return
			}
			newLocation := ctx.MouseMapLocation
			ctx.SwapActiveState(text_entry.New(m.win, "Entity ID", m.lastDeletedEntity.id, m, func(ctx *game.Context, newEntityId string) {
				if newEntityId == "" {
					fmt.Println("entity id can't be empty")
					return
				}
				if _, exists := m.getSelectedMap().Entities[newEntityId]; exists {
					fmt.Println("entity id already in use")
					return
				}
				newEntity := m.lastDeletedEntity.entity.Copy()
				newEntity.X = newLocation.X
				newEntity.Y = newLocation.Y
				m.getSelectedMap().AddEntity(newEntityId, newEntity)
				ctx.SwapActiveState(m)
			}))
			return
		}
		// delete
		prompt := []string{
			"Delete Entity?",
			"ID: " + entityId,
			fmt.Sprintf("Position: %d, %d", entity.X, entity.Y),
		}
		if len(entity.Metadata) > 0 {
			prompt = append(prompt, "", "Metadata:")
			for k, v := range entity.Metadata {
				prompt = append(prompt, fmt.Sprintf("  %s: %s", k, v))
			}
		}
		ctx.SwapActiveState(confirm.New(m.win, prompt, m, func(ctx *game.Context) {
			m.getSelectedMap().RemoveEntity(entityId)
			y, _ := yaml.Marshal(entity)
			fmt.Println("Deleted entity: \n" + string(y))
			m.lastDeletedEntity = &entityReference{
				id:     entityId,
				entity: entity,
			}
			ctx.SwapActiveState(m)
		}))
		return
	}

	// edit
	if entityFound {
		ctx.SwapActiveState(edit_obj.NewEditEntity(m.win, m, entityId, entity))
		return
	}

	// add
	ctx.SwapActiveState(text_entry.New(m.win, "Entity ID", "", m, func(gameCtx *game.Context, id string) {
		id = strings.Trim(id, " ")
		if id == "" {
			fmt.Println("entity id can't be empty")
			return
		}
		if _, exists := m.getSelectedMap().Entities[id]; exists {
			ctx.Notify("entity id already exists")
			return
		}
		gameCtx.SwapActiveState(text_entry.New(m.win, "Entity Type", "", m, func(gameCtx *game.Context, entityType string) {
			id = strings.Trim(id, " ")
			if id == "" {
				fmt.Println("entity id can't be empty")
				return
			}
			m.getSelectedMap().AddEntity(id, &resources.Entity{
				X:    ctx.MouseMapLocation.X,
				Y:    ctx.MouseMapLocation.Y,
				Type: entityType,
			})
			gameCtx.SwapActiveState(m)
		}))
	}))
}
