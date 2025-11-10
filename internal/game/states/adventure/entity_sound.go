package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game"
)

type EntitySoundProvider interface {
	Update(timeDelta float64, cameraDistance float64)
}

type StepSoundProvider struct {
	entity        Entity
	soundSupplier SoundSupplier
	falloff       Falloff

	timeUntilNextStep float64
}

func (p *StepSoundProvider) reset() {
	p.timeUntilNextStep = 0
}

const stepSoundFrequency = 0.333
const stepSoundFrequencyJitter = 0.025
const stepSoundGain = -0.0
const stepSoundGainJitter = -.25

func NewStepSoundProvider(entity Entity, supplier SoundSupplier, falloff Falloff) *StepSoundProvider {
	return &StepSoundProvider{
		entity:            entity,
		soundSupplier:     supplier,
		falloff:           falloff,
		timeUntilNextStep: 0,
	}
}

var stepSoundFrequencyScales = map[MoveState]float64{
	MoveStateWalking: 1,
	MoveStateRunning: 0.8,
}

var stepSoundGainAdjustments = map[MoveState]float64{
	MoveStateWalking: -0.5,
	MoveStateRunning: 0,
}

func (p *StepSoundProvider) Update(timeDelta float64, cameraProximityDistance float64) {
	moveState := p.entity.GetMovementState()
	frequencyScale, exists := stepSoundFrequencyScales[moveState]
	if !exists {
		return
	}
	moveStateGainAdjustment := stepSoundGainAdjustments[moveState]
	p.timeUntilNextStep -= timeDelta
	if p.timeUntilNextStep > 0 {
		return
	}
	p.timeUntilNextStep = (stepSoundFrequency + rand.Float64()*stepSoundFrequencyJitter) * frequencyScale
	soundName := p.soundSupplier.Next()
	gain := stepSoundGain + rand.Float64()*stepSoundGainJitter + moveStateGainAdjustment + p.falloff.AttenuationDB(cameraProximityDistance)
	game.GetAudioSystem().PlaySFX(soundName, gain)
}
