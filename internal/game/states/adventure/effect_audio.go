package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/audio"
)

type EffectPlaySound struct {
	PlaybackId string `auto_generate:"true"`
	Sound      string
	Volume     *float64
}

func (e *EffectPlaySound) CompletionID() string {
	if e.PlaybackId == "" {
		return ""
	}
	return "sound:" + e.PlaybackId
}

func (e *EffectPlaySound) Process(source EntityReader, s *State) bool {
	if e == nil {
		return false
	}
	volume := 1.0
	if e.Volume != nil {
		volume = *e.Volume
	}
	var onComplete func()
	if e.PlaybackId != "" {
		onComplete = func() {
			s.planExecutor.MarkComplete(e.CompletionID())
		}
	}
	a := game.GetAudioSystem()
	a.PlaySoundOnBus(e.Sound, a.Buses.SFX, volume, &audio.PlaybackOptions{
		OnComplete: onComplete,
	})
	return true
}
