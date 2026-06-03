package adventure

import (
	"fisherevans.com/project/f/internal/util/rng"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/audio"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/rs/zerolog/log"
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
const stepSoundFrequencyJitter = 0.125
const stepSoundMaxVolume = 1.0
const stepSoundMinVolume = 0.9

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

var stepSoundVolumeMultipliers = map[MoveState]float64{
	MoveStateWalking: 0.9,
	MoveStateRunning: 1,
}

func (p *StepSoundProvider) Update(timeDelta float64, cameraProximityDistance float64) {
	moveState := p.entity.GetMovementState()
	frequencyScale, exists := stepSoundFrequencyScales[moveState]
	if !exists {
		return
	}
	stepSoundVolumeMultiplier := stepSoundVolumeMultipliers[moveState]
	p.timeUntilNextStep -= timeDelta
	if p.timeUntilNextStep > 0 {
		return
	}
	p.timeUntilNextStep = (stepSoundFrequency + rng.Float64()*stepSoundFrequencyJitter) * frequencyScale
	soundName := p.soundSupplier.Next()
	volume := stepSoundVolumeMultiplier
	volume *= interp.Lerp(stepSoundMinVolume, stepSoundMaxVolume, rng.Float64())
	volume *= p.falloff.AttenuationVolume(cameraProximityDistance)
	if volume <= 0 {
		return
	}
	game.GetAudioSystem().PlaySFX(soundName, volume)
}

func (p *StepSoundProvider) Pause(cameraProximityDistance float64) {
}

func (p *StepSoundProvider) Resume(cameraProximityDistance float64) {
}

type SoundEffect struct {
	Name          string
	Loop          bool
	Falloff       Falloff
	Volume        float64
	FadeInSeconds float64
}

func (e SoundEffect) computeVolume(distance float64) float64 {
	falloffVolume := e.Falloff.AttenuationVolume(distance)
	return falloffVolume * e.Volume
}

type ModeBaseSoundProvider struct {
	entity       Entity
	playOnEnter  map[string][]SoundEffect
	lastMode     string
	activeSounds []activeSound
}

type activeSound struct {
	config  SoundEffect
	control *audio.PlaybackControl
}

func (as activeSound) updateVolume(distance float64) {
	as.control.SetVolume(as.config.computeVolume(distance))
}

func NewModeBaseSoundProvider(entity Entity) *ModeBaseSoundProvider {
	return &ModeBaseSoundProvider{
		entity:      entity,
		lastMode:    "~~~first_run",
		playOnEnter: make(map[string][]SoundEffect),
	}
}

func (p *ModeBaseSoundProvider) WithSoundOnEnter(mode string, sound SoundEffect) *ModeBaseSoundProvider {
	p.playOnEnter[mode] = append(p.playOnEnter[mode], sound)
	return p
}

func (p *ModeBaseSoundProvider) Update(timeDelta float64, cameraProximityDistance float64) {
	mode := ModeMetadataKey.Get(p.entity)
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
			volume := sound.computeVolume(cameraProximityDistance)
			log.Info().Str("mode", mode).Str("entity", p.entity.GetId()).Any("sound", sound).Float64("volume", volume).Msg("playing sound on mode change")
			bus := game.GetAudioSystem().Buses.SFX
			control := game.GetAudioSystem().PlaySoundOnBus(sound.Name, bus, volume, &audio.PlaybackOptions{
				Loop:          sound.Loop,
				FadeInSeconds: sound.FadeInSeconds,
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
			sound.updateVolume(cameraProximityDistance)
		}
	}
}

func (p *ModeBaseSoundProvider) Pause(cameraProximityDistance float64) {
	for _, sound := range p.activeSounds {
		sound.control.Pause()
	}
}

func (p *ModeBaseSoundProvider) Resume(cameraProximityDistance float64) {
	for _, sound := range p.activeSounds {
		sound.updateVolume(cameraProximityDistance)
		sound.control.Resume()
	}
}
