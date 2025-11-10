package adventure

import "fisherevans.com/project/f/internal/game"

type EffectPlaySound struct {
	PlaybackId string `auto_generate:"true"`
	Sound      string
	Gain       *float64
}

func (e *EffectPlaySound) CompletionID() string {
	if e.PlaybackId == "" {
		return ""
	}
	return "sound:" + e.PlaybackId
}

func (e *EffectPlaySound) Process(source EntityContext, s *State) bool {
	if e == nil {
		return false
	}
	gain := 0.0
	if e.Gain != nil {
		gain = *e.Gain
	}
	var onComplete func()
	if e.PlaybackId != "" {
		onComplete = func() {
			s.planExecutor.MarkComplete(e.CompletionID())
		}
	}
	a := game.GetAudioSystem()
	a.PlaySoundOnBusWithCallback(e.Sound, a.Buses.SFX, gain, nil, onComplete)
	logEffectInfof(source, e, "sound played")
	return true
}
