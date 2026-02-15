package computer

import (
	"fmt"

	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
)

var (
	frame3px, frame3pxBorder *frames.Instance
	shipSprite               pixelutil.BoundedDrawable
)

func init() {
	resources.RunOnceInitialized(func() {
		frame3px = frames.New("common/rounded_frame_3px", atlas, frames.WithRenderOrigin(gfx.Centered))
		frame3pxBorder = frames.New("common/rounded_border_frame_3px", atlas, frames.WithRenderOrigin(gfx.Centered))
		shipSprite = atlas.GetSprite("computer/ship")
	})
}

type typeTimer struct {
	delay float64
	text  string
}

type goodbyeMenu struct {
	screen        *screen.Instance[*State]
	skipped       bool
	planet        planet
	waypointId    int
	timerElapsed  float64
	typingElapsed float64
	timers        []typeTimer
	lines         []*textbox.Content
	currentLine   string
	typeQueue     []byte
	exitTime      float64
}

var normalLine = 0.25

func newGoodbyeMenu(screen *screen.Instance[*State], p planet, waypointId int) screen.Menu {
	return &goodbyeMenu{
		screen:     screen,
		planet:     p,
		waypointId: waypointId,
		timers: []typeTimer{
			{delay: normalLine, text: "Initiating launch sequence...\n"},
			{delay: normalLine, text: fmt.Sprintf("Target: %s, %s...\n", p.Name(), p.waypoints[waypointId].displayName)},
			{delay: normalLine, text: "Waypoint lock acquired\n"},
			{delay: normalLine, text: "Primortal sync nominal\n"},
			{delay: normalLine, text: "Bon voyage, captain....."},
		},
	}
}

func (m *goodbyeMenu) Enter() {
}

var (
	terminalFrameW = 200
	terminalFrameH = 90
)

func (m *goodbyeMenu) OnTick(target pixel.Target, timeDelta float64) {
	if m.skipped {
		timeDelta = timeDelta * 10
	}
	m.handleInputs()
	m.typeText(timeDelta)

	frameCenter := m.screen.Center().Add(gfx.IVec(0, 7))

	shipSprite.DrawColorMask(target, pixel.IM.Moved(frameCenter), screenColors.Dark)

	if len(m.typeQueue) == 0 && len(m.timers) == 0 {
		m.exitTime += timeDelta
		exitDelay := 2.0
		if m.exitTime > exitDelay {
			frameCenter.Y += (m.exitTime - exitDelay) * 90
		}
		if m.exitTime > exitDelay+1.5 {
			m.triggerLaunch()
		}
	}

	frame3px.Draw(target,
		gfx.R(terminalFrameW, terminalFrameH),
		pixel.IM.Moved(frameCenter),
		frames.WithColor(screenColors.Dark))

	frame3pxBorder.Draw(target,
		gfx.R(terminalFrameW, terminalFrameH),
		pixel.IM.Moved(frameCenter),
		frames.WithColor(screenColors.Highlight))

	x, y := int(frameCenter.X)-60, int(frameCenter.Y)+terminalFrameH/2-15

	renderOpts := []tbcfg.ConfigOpt{
		tbcfg.RenderFrom(gfx.TopLeft),
		tbcfg.VAlignedTop(),
		tbcfg.HAlignedLeft(),
	}

	lineHeight := 14

	for _, line := range m.lines {
		line.Render(target, gfx.Moved(x, y), append(renderOpts, tbcfg.Foreground(screenColors.Text))...)
		y -= lineHeight
	}

	lastLine := m.currentLine
	if game.Utils().TimeCycleSin(2) > 0.5 {
		lastLine += "_"
	}
	currentContent := goodbyeContent(lastLine)
	currentContent.Render(target, gfx.Moved(x, y), append(renderOpts, tbcfg.Foreground(screenColors.Highlight))...)

	badgeBCancel.Render(target, gfx.Moved(3, 3), gfx.BottomLeft)
	badgeALaunchFlipped.Render(target, gfx.Moved(screenWidth-3, 3), gfx.BottomRight)
}

func (m *goodbyeMenu) handleInputs() {
	if game.Controls[*State]().ButtonB().JustPressed() {
		m.screen.PopMenu(true)
	}
	if game.Controls[*State]().ButtonA().JustPressed() {
		m.skipped = true
	}
}

func (m *goodbyeMenu) triggerLaunch() {
	m.screen.Holder().CloseWithData(game.ComputerReturnData{
		PlanetName:       m.planet.mapName,
		PlanetSpriteName: m.planet.spriteName,
		Waypoint:         m.planet.waypoints[m.waypointId].loadName,
	})
}

var typingSpeed = 0.0125

func (m *goodbyeMenu) typeText(timeDelta float64) {
	if len(m.typeQueue) > 0 {
		m.typingElapsed += timeDelta
		for len(m.typeQueue) > 0 && m.typingElapsed >= typingSpeed {
			m.typingElapsed -= typingSpeed
			nextChar := m.typeQueue[0]
			m.typeQueue = m.typeQueue[1:]
			switch nextChar {
			case '\n':
				m.lines = append(m.lines, goodbyeContent(m.currentLine))
				m.currentLine = ""
			default:
				m.currentLine += string(nextChar)
			}
		}
		return
	}
	m.timerElapsed += timeDelta
	for len(m.timers) > 0 && m.timerElapsed >= m.timers[0].delay {
		m.typeQueue = append(m.typeQueue, m.timers[0].text...)
		m.timerElapsed -= m.timers[0].delay
		m.timers = m.timers[1:]
	}
	if len(m.typeQueue) == 0 {
		m.typingElapsed = 0
	}
}

func goodbyeContent(s string) *textbox.Content {
	return textFF.NewSimpleContent(s)
}
