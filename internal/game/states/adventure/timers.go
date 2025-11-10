package adventure

import (
	"github.com/rs/zerolog/log"
)

type timer struct {
	createdBy    string
	timerId      string
	completionId string
	elapsed      float64
	duration     float64
	repeatCount  int
	triggerCount int
}

type timers struct {
	inProgress []*timer
}

func newTimers() *timers {
	return &timers{}
}

func (t *timers) AddTimer(createdBy, timerId, completionId string, durationSeconds float64) {
	newTimer := &timer{
		createdBy:    createdBy,
		timerId:      timerId,
		completionId: completionId,
		duration:     durationSeconds,
		repeatCount:  1,
	}
	log.Info().Msgf("Adding timer: %#v", newTimer)
	t.inProgress = append(t.inProgress, newTimer)
}

func (t *timers) Update(deltaSeconds float64, dispatcher *Dispatcher, state *State) {
	for i := len(t.inProgress) - 1; i >= 0; i-- {
		timer := t.inProgress[i]
		timer.elapsed += deltaSeconds
		if timer.elapsed < timer.duration {
			continue
		}
		log.Info().Msgf("Timer triggered: %#v", timer)
		timer.elapsed = 0
		timer.triggerCount++
		if timer.triggerCount == timer.repeatCount {
			log.Info().Msgf("Removing timer: %s/%s", timer.createdBy, timer.timerId)
			t.inProgress = append(t.inProgress[:i], t.inProgress[i+1:]...)
		}
		dispatcher.Dispatch(EventTimerComplete{
			CreatedBy:       timer.createdBy,
			TimerId:         timer.timerId,
			DurationSeconds: timer.duration,
			TriggerCount:    timer.triggerCount,
		})
		// Mark timer complete for plan tracking
		state.planExecutor.MarkComplete(timer.completionId)
	}
}
