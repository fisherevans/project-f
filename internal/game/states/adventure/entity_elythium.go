package adventure

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/util/colors"
)

type ElythiumDepositEntity struct {
	InnateEntity
	unminedlight      *Light
	unminedAnimations []*anim.AnimatedSprite
	minedAnimation    *anim.AnimatedSprite
	mined             bool
	timeMined         float64
}

func NewElythiumDepositEntity(id EntityId, location MapLocation) Entity {
	return &ElythiumDepositEntity{
		mined: false,
		InnateEntity: InnateEntity{
			BaseEntity: BaseEntity{
				Id:       id,
				Passable: false,
			},
			MapLocation: location,
		},
		unminedlight: &Light{
			RenderDetails: LightRenderDetails{
				SizeScale: 1.5,
				ColorMask: colors.HexString("#f06"),
			},
			Modifiers: []LightModifier{
				&LightModifierPulse{
					PeriodSeconds:       4,
					SizeIntensity:       0.1,
					BrightnessIntensity: 0.4,
				},
			},
		},
		unminedAnimations: []*anim.AnimatedSprite{
			anim.Load(atlas, "adventure/entities/elythium/crystals", "default"),
			anim.Load(atlas, "adventure/entities/elythium/crystals_sparkle", "default"),
		},
		minedAnimation: anim.Load(atlas, "adventure/entities/elythium/crystals_rock", "default"),
	}
}

func (e *ElythiumDepositEntity) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	if e.mined {
		e.minedAnimation.Sprite().Draw(target, matrix)
	} else {
		for _, a := range e.unminedAnimations {
			a.Sprite().Draw(target, matrix)
		}
	}
}

func (e *ElythiumDepositEntity) RenderLight(target pixel.Target, matrix pixel.Matrix) {
	if e.mined {
		return
	}
	e.unminedlight.Render(target, matrix)
}

func (e *ElythiumDepositEntity) Update(adv *State, timeDelta float64) {
	if e.mined {
		e.timeMined += timeDelta
		if e.timeMined > 5 {
			e.mined = false
			e.timeMined = 0
		}
		return
	}
	e.unminedlight.Update(timeDelta)
	for _, a := range e.unminedAnimations {
		a.Update(timeDelta)
	}
}

func (e *ElythiumDepositEntity) Interact(adv *State, source Entity) {
	log.Info().Msg("Interacting with deposit")
	if e.mined {
		return
	}
	e.mined = true
	adv.dialogues.Append(NewBasicDialogue("You mine the deposit!", nil))
	adv.hud.ElythiumCount++
}
