package input

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"time"
)

type Direction int

const (
	NotPressed Direction = iota
	Up
	Right
	Down
	Left
)

var Directions = []Direction{
	Up,
	Right,
	Down,
	Left,
}

func (d Direction) String() string {
	switch d {
	case Up:
		return "Up"
	case Right:
		return "Right"
	case Down:
		return "Down"
	case Left:
		return "Left"
	default:
		return "NotPressed"
	}
}

func (d Direction) GetVector() (int, int) {
	switch d {
	case Left:
		return -1, 0
	case Right:
		return 1, 0
	case Down:
		return 0, -1
	case Up:
		return 0, 1
	default:
		return 0, 0
	}
}

func (d Direction) Opposite() Direction {
	switch d {
	case Up:
		return Down
	case Right:
		return Left
	case Down:
		return Up
	case Left:
		return Right
	default:
		return NotPressed
	}
}

type DirectionalButton struct {
	ButtonState
	sourceKeys      map[pixel.Button]Direction
	lastJustPressed map[pixel.Button]time.Time
	direction       Direction
}

func (d *DirectionalButton) updateDirectional(win *opengl.Window) {
	var mostRecentPressDirection Direction
	var mostRecentPressKey pixel.Button
	var mostRecentPressTime time.Time
	now := time.Now()
	for key, direction := range d.sourceKeys {
		if win.JustPressed(key) {
			d.lastJustPressed[key] = now
		} else if !win.Pressed(key) {
			delete(d.lastJustPressed, key)
		}
		lastPressed, isPressed := d.lastJustPressed[key]
		if !isPressed {
			continue
		}
		if mostRecentPressTime.Before(lastPressed) {
			mostRecentPressTime = lastPressed
			mostRecentPressKey = key
			mostRecentPressDirection = direction
		}
	}
	if mostRecentPressDirection == NotPressed {
		d.direction = NotPressed
		d.ButtonState.updateState(win)
		return
	}
	d.direction = mostRecentPressDirection
	d.ButtonState.updateState(win, mostRecentPressKey)
}

func (d *DirectionalButton) GetDirection() Direction {
	return d.direction
}
