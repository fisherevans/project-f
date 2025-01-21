package input

import (
	"fisherevans.com/project/f/game"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"time"
)

type Controls struct {
	primary   *Button
	secondary *Button
	menu      *Button
	option    *Button
	move      *DirectionalButton
}

func NewControls() *Controls {
	return &Controls{
		primary: &Button{
			sourceKeys: []pixel.Button{pixel.KeySpace, pixel.KeyEnter},
		},
		secondary: &Button{
			sourceKeys: []pixel.Button{pixel.KeyLeftShift, pixel.KeyLeftControl},
		},
		menu: &Button{
			sourceKeys: []pixel.Button{pixel.KeyEscape},
		},
		option: &Button{
			sourceKeys: []pixel.Button{pixel.KeyTab},
		},
		move: &DirectionalButton{
			sourceKeys: map[pixel.Button]Direction{
				pixel.KeyUp:    Up,
				pixel.KeyDown:  Down,
				pixel.KeyLeft:  Left,
				pixel.KeyRight: Right,

				pixel.KeyW: Up,
				pixel.KeyA: Left,
				pixel.KeyS: Down,
				pixel.KeyD: Right,
			},
			lastJustPressed: map[pixel.Button]time.Time{},
		},
	}
}

func (c *Controls) Update(ctx *game.Context, win *opengl.Window) {
	c.primary.updateButton(win)
	c.secondary.updateButton(win)
	c.menu.updateButton(win)
	c.option.updateButton(win)
	c.move.updateDirectional(ctx, win)
}

func (c *Controls) Primary() *Button {
	return c.primary
}
func (c *Controls) Secondary() *Button {
	return c.secondary
}
func (c *Controls) Menu() *Button {
	return c.menu
}
func (c *Controls) Option() *Button {
	return c.option
}
func (c *Controls) DPad() *DirectionalButton {
	return c.move
}

type ButtonState struct {
	pressed     bool
	justPressed bool
	repeated    bool
}

func (b *ButtonState) String() string {
	return fmt.Sprintf("pressed: %t, just: %t, repeated: %t", b.pressed, b.justPressed, b.repeated)
}

func (b *ButtonState) updateState(win *opengl.Window, keys ...pixel.Button) {
	b.pressed = false
	b.justPressed = false
	b.repeated = false
	for _, key := range keys {
		if win.Pressed(key) {
			b.pressed = true
		}
		if win.JustPressed(key) {
			b.justPressed = true
		}
		if win.Repeated(key) {
			b.repeated = true
		}
	}
}

func (b *ButtonState) IsPressed() bool {
	return b.pressed
}

func (b *ButtonState) JustPressed() bool {
	return b.justPressed
}

func (b *ButtonState) JustRepeated() bool {
	return b.repeated
}

func (b *ButtonState) JustPressedOrRepeated() bool {
	return b.justPressed || b.repeated
}

type Button struct {
	ButtonState
	sourceKeys []pixel.Button
}

func (b *Button) updateButton(win *opengl.Window) {
	b.ButtonState.updateState(win, b.sourceKeys...)
}

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

func (d *DirectionalButton) updateDirectional(ctx *game.Context, win *opengl.Window) {
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
