package adventure

import (
	"math/rand"
	"time"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/audio"
)

type EntitySoundProvider interface {
	Update(timeDelta float64, cameraDistance float64)
	Pause(cameraProximityDistance float64)
	Resume(cameraProximityDistance float64)
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

func (p *StepSoundProvider) Pause(cameraProximityDistance float64) {
}

func (p *StepSoundProvider) Resume(cameraProximityDistance float64) {
}

type SoundEffect struct {
	name    string
	loop    bool
	falloff Falloff
	gain    float64
	fadeIn  time.Duration
}

type ModeBaseSoundProvider struct {
	entity       Entity
	playOnEnter  map[string][]SoundEffect
	lastMode     string
	activeSounds []activeSound
}

type activeSound struct {
	config  SoundEffect
	control *audio.SoundControl
}

func NewModeBaseSoundProvider(entity Entity) *ModeBaseSoundProvider {
	return &ModeBaseSoundProvider{
		entity:      entity,
		playOnEnter: make(map[string][]SoundEffect),
	}
}

func (p *ModeBaseSoundProvider) WithSoundOnEnter(mode string, sound SoundEffect) *ModeBaseSoundProvider {
	p.playOnEnter[mode] = append(p.playOnEnter[mode], sound)
	return p
}

func (p *ModeBaseSoundProvider) Update(timeDelta float64, cameraProximityDistance float64) {
	// todo move mode to metadata
	renderer, ok := p.entity.GetRenderer()
	if !ok {
		return
	}
	modeBased, ok := renderer.(*ModeBasedEntityRenderer)
	if !ok {
		return
	}
	mode := modeBased.currentMode
	if p.lastMode != mode {
		for _, sound := range p.activeSounds {
			if !sound.control.IsPlaying() {
				continue
			}
			sound.control.Stop()
		}
		p.activeSounds = nil
		p.lastMode = mode
		for _, sound := range p.playOnEnter[mode] {
			bus := game.GetAudioSystem().Buses.SFX
			gain := sound.falloff.AttenuationDB(cameraProximityDistance) + sound.gain
			control := game.GetAudioSystem().PlaySoundOnBus(sound.name, bus, gain, &audio.PlaybackOptions{
				Loop:   sound.loop,
				FadeIn: sound.fadeIn,
			})
			p.activeSounds = append(p.activeSounds, activeSound{
				config:  sound,
				control: control,
			})
		}
	} else {
		for _, sound := range p.activeSounds {
			if !sound.control.IsPlaying() {
				continue
			}
			gain := sound.config.falloff.AttenuationDB(cameraProximityDistance) + sound.config.gain
			sound.control.SetVolume(gain)
		}
	}
}

func (p *ModeBaseSoundProvider) Pause(cameraProximityDistance float64) {
	for _, sound := range p.activeSounds {
		sound.control.PauseWithFade(time.Millisecond * 500)
	}
}

func (p *ModeBaseSoundProvider) Resume(cameraProximityDistance float64) {
	for _, sound := range p.activeSounds {
		gain := sound.config.falloff.AttenuationDB(cameraProximityDistance) + sound.config.gain
		sound.control.ResumeWithFade(gain, time.Millisecond*500)
	}
}
