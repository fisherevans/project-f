package computer

import (
	"math"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
)

// Variables declared in init_vars.go
var (
	textFF              *textbox.Instance
	text57              *textbox.Instance
	textAdd             *textbox.Instance
	textItalic          *textbox.Instance
	selectedPlanetFrame *frames.Instance
)

func init() {
	resources.RunOnceInitialized(func() {
		textFF = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
			tbcfg.NewConfig(screenWidth, 10,
				tbcfg.Foreground(screenColors.Text),
				tbcfg.RenderFrom(gfx.TopCenter),
				tbcfg.VAligned(tbcfg.VAlignTop),
				tbcfg.HAligned(tbcfg.HAlignCenter),
			))
		text57 = textbox.NewInstance(atlas.GetFont(resources.FontNameM5x7),
			tbcfg.NewConfig(screenWidth, 10,
				tbcfg.Foreground(screenColors.Text),
				tbcfg.RenderFrom(gfx.TopCenter),
				tbcfg.VAligned(tbcfg.VAlignTop),
				tbcfg.HAligned(tbcfg.HAlignCenter),
			))
		textAdd = textbox.NewInstance(atlas.GetFont(resources.FontNameAddStandard),
			tbcfg.NewConfig(screenWidth, 10,
				tbcfg.Foreground(screenColors.Text),
				tbcfg.RenderFrom(gfx.TopCenter),
				tbcfg.VAligned(tbcfg.VAlignTop),
				tbcfg.HAligned(tbcfg.HAlignCenter),
			))
		textItalic = textbox.NewInstance(atlas.GetFont(resources.FontNameFF57i),
			tbcfg.NewConfig(screenWidth, 10,
				tbcfg.Foreground(screenColors.Text),
				tbcfg.RenderFrom(gfx.TopCenter),
				tbcfg.VAligned(tbcfg.VAlignTop),
				tbcfg.HAligned(tbcfg.HAlignCenter),
			))
		selectedPlanetFrame = frames.New("computer/planet_select_frame", atlas)
	})
}

type waypoint struct {
	displayName string
	loadName    string
	detail      string

	x, y int
}

type planet struct {
	displayName string
	mapName     string
	description string
	spriteName  string
	waypoints   []waypoint
}

func (p planet) Name() string {
	if p.displayName == "" {
		return "???????"
	}
	return p.displayName
}

func (p planet) Description() string {
	if p.description == "" {
		return "No details are known about this planet..."
	}
	return p.description
}

var planets = []planet{
	{
		displayName: "Sylvoria",
		mapName:     "sylvoria",
		description: "A lush forest covered planet.",
		spriteName:  "computer/planet_1",
		waypoints: []waypoint{
			{
				displayName: "New Beginnings",
				loadName:    "new_beginnings",
				detail:      "A central landing zone",
				x:           32,
				y:           32,
			},
			{
				displayName: "???????",
				detail:      "??????????",
				x:           27,
				y:           5,
			},
			{
				displayName: "???????",
				detail:      "??????????",
				x:           5,
				y:           32,
			},
			{
				displayName: "Nomad Outpost",
				loadName:    "nomad_outpost",
				detail:      "North west on an island",
				x:           46,
				y:           50,
			},
		},
	},
	{
		spriteName: "computer/planet_2",
	},
	{
		spriteName: "computer/planet_3",
	},
	{
		spriteName: "computer/planet_4",
	},
}

var farScrollSpeed = 15.0

type mapMenu struct {
	screen *screen.Instance[*State]

	selectedPlanet int
	renderedIndex  float64
}

func newMapMenu(screen *screen.Instance[*State]) screen.Menu {
	home := &mapMenu{
		screen: screen,
	}
	return home
}

func (m *mapMenu) Enter() {
}

func (m *mapMenu) OnTick(target pixel.Target, timeDelta float64) {
	m.handleInputs()
	m.updateScrollDelta(timeDelta)

	game.DebugTLf("selected planet: %d", m.selectedPlanet)
	game.DebugTLf("rendered index: %.3f", m.renderedIndex)

	selectPlanetCenterX, selectPlanetCenterY := 40.0, m.screen.Center().Y
	for index, p := range planets {
		renderOffset := m.renderedIndex - float64(index)
		y := math.Floor(selectPlanetCenterY + renderOffset*44)
		x := math.Floor(selectPlanetCenterX - math.Pow(renderOffset*3, 2))
		atlas.GetSprite(p.spriteName).Draw(m.screen.SpriteFilterBuffer().Target(), pixel.IM.Scaled(pixel.ZV, 0.5).Moved(pixel.V(x, y)))
	}

	selectProgressInverse := min(math.Abs(m.renderedIndex-float64(m.selectedPlanet)), 1)
	frameWidth := int(selectProgressInverse*15.0) + 44
	frameHeight := int(selectProgressInverse*5.0) + 38
	frameMask := colors.Lerp(screenColors.Text, screenColors.Highlight, 1.0-selectProgressInverse)
	selectedPlanetFrame.Draw(target, gfx.R(frameWidth, frameHeight),
		pixel.IM.Moved(pixel.V(selectPlanetCenterX, selectPlanetCenterY)),
		frames.WithRenderOrigin(gfx.Centered),
		frames.WithColor(frameMask))

	x, y := 140, 95
	textFF.NewSimpleContent("Travel to...").Render(target, gfx.Moved(x, y), tbcfg.Foreground(screenColors.Dark))
	y -= 7
	textAdd.NewSimpleContent(planets[m.selectedPlanet].Name()).Render(target, gfx.Moved(x, y), tbcfg.Foreground(screenColors.Highlight))
	y -= 14
	textItalic.NewSimpleContent(planets[m.selectedPlanet].Description(), textbox.WithAutoWrapWidth(120)).Render(target, gfx.Moved(x, y), tbcfg.Foreground(screenColors.Text))

	badgeBExit.Render(target, gfx.Moved(3, 3), gfx.BottomLeft)
	badgeASelectFlipped.Render(target, gfx.Moved(screenWidth-3, 3), gfx.BottomRight)
}

func (m *mapMenu) handleInputs() {
	if game.Controls[*State]().ButtonA().JustPressed() {
		m.screen.PushMenu(newPlanetMenu(m.screen, planets[m.selectedPlanet]), true)
	}
	if game.Controls[*State]().ButtonB().JustPressed() {
		m.screen.Holder().ToggleOpenState()
	}
	switch game.Controls[*State]().DPad().JustPressedDirection() {
	case input.Up:
		m.selectedPlanet--
	case input.Down:
		m.selectedPlanet++
	}
	m.selectedPlanet = max(0, min(len(planets)-1, m.selectedPlanet))
}

func (m *mapMenu) updateScrollDelta(timeDelta float64) {
	direction := 1.0
	if float64(m.selectedPlanet) < m.renderedIndex {
		direction = -1.0
	}
	distance := math.Abs(m.renderedIndex - float64(m.selectedPlanet))
	delta := timeDelta * math.Sqrt(distance*farScrollSpeed)
	if delta > distance {
		delta = distance
	}
	m.renderedIndex += delta * direction
}
