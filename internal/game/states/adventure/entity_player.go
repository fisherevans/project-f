package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
)
import (
	"fisherevans.com/project/f/internal/game/anim"
	"github.com/gopxl/pixel/v2"
)

type Player struct {
	BaseEntity
	Passable
	Animations             map[MoveState]map[input.Direction]*anim.AnimatedSprite
	Lights                 map[MoveState]*Light
	ColorMask              pixel.RGBA
	movementState          *MovementState
	lastAnimationDirection input.Direction
	lastAnimationState     MoveState
}

func (p *Player) SetMovementState(state *MovementState) {
	p.movementState = state
}

func (p *Player) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	if p.movementState == nil {
		return
	}

	animations, ok := p.Animations[p.movementState.MoveState()]
	if !ok {
		return
	}

	animation, ok := animations[p.movementState.FacingDirection()]
	if !ok {
		return
	}

	sprite := animation.Sprite()
	if p.ColorMask.A > 0 {
		sprite.DrawColorMask(target, matrix, p.ColorMask)
	} else {
		sprite.Draw(target, matrix)
	}
}

func (p *Player) RenderLight(target pixel.Target, matrix pixel.Matrix) {
	if p.Lights == nil || p.movementState == nil {
		return
	}
	light, ok := p.Lights[p.movementState.MoveState()]
	if !ok {
		light, ok = p.Lights[MoveStateIdle]
	}
	if light != nil {
		light.Render(target, matrix)
	}
}

func (p *Player) Location() MapLocation {
	if p.movementState == nil {
		return MapLocation{}
	}
	return p.movementState.Location()
}

func (p *Player) PreciseMapLocation() pixel.Vec {
	if p.movementState == nil {
		return pixel.Vec{}
	}
	return p.movementState.PreciseLocation()
}

func (p *Player) RenderMapLocation() pixel.Vec {
	location := p.PreciseMapLocation()
	// Add Y offset for character rendering (feet position)
	return location.Add(pixel.V(0, 0.25))
}

func (p *Player) IsMoving() bool {
	if p.movementState == nil {
		return false
	}
	return p.movementState.IsMoving()
}

func (p *Player) TeleportTo(s *State, location MapLocation) bool {
	return s.movementController.TeleportTo(s, p.GetEntityId(), location)
}

func (p *Player) DashTowards(s *State, direction input.Direction, location MapLocation) {
	if behavior, ok := s.behaviors[p.GetEntityId()].(*PlayerBehavior); ok {
		behavior.DashTowards(s, direction, location)
	}
}
