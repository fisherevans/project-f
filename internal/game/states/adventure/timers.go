package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"github.com/rs/zerolog/log"
)

type timer struct {
	createdBy    string
	id           string
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

func (t *timers) AddTimer(createdBy string, id string, durationSeconds float64) {
	newTimer := &timer{
		createdBy:   createdBy,
		id:          id,
		duration:    durationSeconds,
		repeatCount: 1,
	}
	log.Info().Msgf("Adding timer: %#v", newTimer)
	t.inProgress = append(t.inProgress, newTimer)
}

func (t *timers) Update(deltaSeconds float64, dispatcher *events.Dispatcher, state *State) {
	for id, timer := range t.inProgress {
		timer.elapsed += deltaSeconds
		if timer.elapsed < timer.duration {
			continue
		}
		log.Info().Msgf("Timer triggered: %#v", timer)
		timer.elapsed = 0
		timer.triggerCount++
		if timer.triggerCount == timer.repeatCount {
			log.Info().Msgf("Removing timer: %s/%s", timer.createdBy, timer.id)
			t.inProgress = append(t.inProgress[:id], t.inProgress[id+1:]...)
		}
		dispatcher.Dispatch(events.EventTimerComplete{
			CreatedBy:       timer.createdBy,
			TimerId:         timer.id,
			DurationSeconds: timer.duration,
			TriggerCount:    timer.triggerCount,
		})
		// Mark timer complete for plan tracking
		if state != nil {
			state.planExecutor.MarkTimerComplete(timer.id)
		}
	}
}
