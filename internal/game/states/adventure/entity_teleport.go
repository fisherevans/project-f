package adventure

import (
	"fmt"

	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/anim"
)

type EntityTeleport struct {
	InnateEntity
	Passable
	RequiredElythium int
	Destination      TeleportReference
}

func (e *EntityTeleport) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	tiles.Rocket.From(atlas).Draw(target, matrix)
}

func (e *EntityTeleport) Interact(adv *State, source Entity) {
	if adv.hud.ElythiumCount < e.RequiredElythium {
		msg := fmt.Sprintf("You need %d Elythium to travel home!", e.RequiredElythium)
		adv.AddSystemEffect(events.Effect{
			Dialogue: &events.EffectDialogue{
				Text: msg,
			},
		})
		return
	}

	d := string(e.Destination)
	adv.AddSystemEffect(events.Effect{
		Plan: &events.EffectPlan{
			Steps: []events.PlanStep{
				{
					Serial: []events.Effect{
						{
							Dialogue: &events.EffectDialogue{
								Text: "You've managed to escape!",
							},
						},
						{
							YieldElythium: &events.EffectYieldElythium{
								Amount: -e.RequiredElythium,
							},
							TeleportPlayer: &events.EffectTeleportPlayer{
								ToReference: &d,
							},
						},
					},
				},
			},
		},
	})
}

type EntityAnimatedTeleport struct {
	InnateEntity
	OnAnimation  *anim.AnimatedSprite
	OffAnimation *anim.AnimatedSprite
	OnMessage    string
	OffMessage   string

	toggled bool
}

func (e *EntityAnimatedTeleport) Update(adv *State, timeDelta float64) {
	if e.toggled && e.OnAnimation != nil {
		e.OnAnimation.Update(timeDelta)
	}
	if !e.toggled && e.OffAnimation != nil {
		e.OffAnimation.Update(timeDelta)
	}
}

func (e *EntityAnimatedTeleport) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	if e.toggled && e.OnAnimation != nil {
		e.OnAnimation.Sprite().Draw(target, matrix)
	}
	if !e.toggled && e.OffAnimation != nil {
		e.OffAnimation.Sprite().Draw(target, matrix)
	}
}

func (e *EntityAnimatedTeleport) Interact(adv *State, source Entity) {
	msg := e.OffMessage
	if e.toggled {
		msg = e.OnMessage
	}
	if msg == "" {
		return
	}
	e.toggled = true
	adv.AddSerialSystemEffects(
		events.Effect{
			Dialogue: &events.EffectDialogue{
				Text: msg,
			},
		},
		events.Effect{
			Function: &events.EffectFunction{
				Fn: func() {
					e.toggled = false
				},
			},
		})
}
