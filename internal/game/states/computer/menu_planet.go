package computer

import (
	"math"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
)

var (
	arrowUp, arrowRight, arrowDown, arrowLeft pixelutil.BoundedDrawable
	crosshair                                 pixelutil.BoundedDrawable
)

func init() {
	resources.RunOnceInitialized(func() {
		arrowUp = atlas.GetTilesheetSprite("common/arrows_5px", 1, 1)
		arrowRight = atlas.GetTilesheetSprite("common/arrows_5px", 2, 1)
		arrowDown = atlas.GetTilesheetSprite("common/arrows_5px", 3, 1)
		arrowLeft = atlas.GetTilesheetSprite("common/arrows_5px", 4, 1)
		crosshair = atlas.GetSprite("computer/crosshair")
	})
}

type planetMenu struct {
	screen *screen.Instance[*State]
	planet planet

	crosshairX, crosshairY float64

	selectedWaypoint int
}

func newPlanetMenu(screen *screen.Instance[*State], planet planet) screen.Menu {
	return &planetMenu{
		screen:     screen,
		planet:     planet,
		crosshairX: float64(planet.waypoints[0].x),
		crosshairY: float64(planet.waypoints[0].y),
	}
}

func (m *planetMenu) Enter() {
}

func (m *planetMenu) OnTick(target pixel.Target, timeDelta float64) {
	m.handleInputs()

	planetCenter := gfx.Moved(48, (screenHeight/2)-4)
	planetTarget := m.screen.SpriteFilterBuffer().Target()
	atlas.GetSprite(m.planet.spriteName).Draw(planetTarget, planetCenter)
	crosshair.Draw(planetTarget, planetCenter.Moved(gfx.IVec(int(m.crosshairX)-32, int(m.crosshairY)-32)))

	textAdd.NewSimpleContent(m.planet.Name()).Render(target, planetCenter.Moved(gfx.IVec(0, 52)), tbcfg.Foreground(screenColors.Highlight))

	topLeft := []tbcfg.ConfigOpt{
		tbcfg.HAlignedLeft(),
		tbcfg.VAlignedTop(),
		tbcfg.RenderFrom(gfx.TopLeft),
	}

	x, y := 100, screenHeight-20
	textFF.NewSimpleContent("Select a landing zone").Render(target, gfx.Moved(x, y), append(topLeft, tbcfg.Foreground(screenColors.Dark))...)
	y -= 2
	for i, w := range m.planet.waypoints {
		y -= 10
		mask := screenColors.Text
		if i == m.selectedWaypoint {
			mask = screenColors.Highlight
			arrowRight.Draw(target, gfx.Moved(x-8, y-3).Moved(gfx.TopLeft.Align(arrowRight)))
			m.updateCrosshairPosition(w.x, w.y, timeDelta)
		}
		text57.NewSimpleContent(w.displayName, textbox.WithAutoWrapWidth(120)).Render(target, gfx.Moved(x, y), append(topLeft, tbcfg.Foreground(mask))...)
		y -= 12
		textFF.NewSimpleContent(w.detail).Render(target, gfx.Moved(x, y), append(topLeft, tbcfg.Foreground(screenColors.Text))...)
	}

	badgeBExit.Render(target, gfx.Moved(3, 3), gfx.BottomLeft)
	badgeALaunchFlipped.Render(target, gfx.Moved(screenWidth-3, 3), gfx.BottomRight)
}

func (m *planetMenu) handleInputs() {
	if game.Controls[*State]().ButtonB().JustPressed() {
		m.screen.PopMenu(true)
	}
	if game.Controls[*State]().ButtonA().JustPressed() {
		m.screen.PushMenu(newGoodbyeMenu(m.screen, m.planet, m.selectedWaypoint), true)
	}
	switch game.Controls[*State]().DPad().JustPressedDirection() {
	case input.Up:
		m.selectedWaypoint--
	case input.Down:
		m.selectedWaypoint++
	}
	m.selectedWaypoint = max(0, min(len(m.planet.waypoints)-1, m.selectedWaypoint))
}

var crosshairScrollSpeed = 35.0
var crosshairScrollScale = 3.0

func (m *planetMenu) updateCrosshairPosition(x, y int, timeDelta float64) {
	target := pixel.V(float64(x), float64(y))
	current := pixel.V(m.crosshairX, m.crosshairY)
	deltaVec := target.Sub(current)
	distance := deltaVec.Len()

	if distance > 0 {
		speed := math.Sqrt(distance*crosshairScrollSpeed) * crosshairScrollScale
		step := speed * timeDelta
		if step > distance {
			step = distance
		}
		newPos := current.Add(deltaVec.Unit().Scaled(step))
		m.crosshairX = newPos.X
		m.crosshairY = newPos.Y
	}
}
