package adventure

import "github.com/rs/zerolog/log"

type EntityBehavior interface {
	DisabledSources() []string
	Disable(source string)
	Enable(source string)
	IsEnabled() bool
	MovementComplete(dispatcher Dispatcher)
	Update(timeDelta float64, position *EntityPosition, dispatcher Dispatcher)
	Reset()
}

type baseEntityBehavior struct {
	id         string
	system     *EntitySystem
	disabledBy map[string]struct{}
}

func newBaseEntityBehavior(id string, system *EntitySystem) *baseEntityBehavior {
	return &baseEntityBehavior{
		id:         id,
		system:     system,
		disabledBy: make(map[string]struct{}),
	}
}

func (b *baseEntityBehavior) Enable(source string) {
	if _, exists := b.disabledBy[source]; !exists {
		return
	}
	delete(b.disabledBy, source)
	log.Info().Str("id", b.id).Str("by", source).Msg("behavior enabled")
}

func (b *baseEntityBehavior) Disable(source string) {
	if _, exists := b.disabledBy[source]; exists {
		return
	}
	b.disabledBy[source] = struct{}{}
	b.system.positions[b.id].CancelMovement()
	log.Info().Str("id", b.id).Str("by", source).Msg("behavior disabled")
}

func (b *baseEntityBehavior) DisabledSources() []string {
	var l []string
	for k := range b.disabledBy {
		l = append(l, k)
	}
	return l
}

func (b *baseEntityBehavior) IsEnabled() bool {
	return len(b.disabledBy) == 0
}

type EntityBehaviorOverride struct {
	replacedBehavior EntityBehavior
	newBehavior      EntityBehavior
}

func NewEntityBehaviorOverride(replacedBehavior EntityBehavior, newBehavior EntityBehavior) *EntityBehaviorOverride {
	for _, src := range replacedBehavior.DisabledSources() {
		newBehavior.Disable(src)
	}
	replacedBehavior.Reset()
	return &EntityBehaviorOverride{
		replacedBehavior: replacedBehavior,
		newBehavior:      newBehavior,
	}
}

func (e *EntityBehaviorOverride) Disable(source string) {
	e.replacedBehavior.Disable(source)
	e.newBehavior.Disable(source)
}

func (e *EntityBehaviorOverride) Enable(source string) {
	e.replacedBehavior.Enable(source)
	e.newBehavior.Enable(source)
}

func (e *EntityBehaviorOverride) IsEnabled() bool {
	return e.newBehavior.IsEnabled()
}

func (e *EntityBehaviorOverride) DisabledSources() []string {
	return e.newBehavior.DisabledSources()
}

func (e *EntityBehaviorOverride) MovementComplete(dispatcher Dispatcher) {
	e.newBehavior.MovementComplete(dispatcher)
}

func (e *EntityBehaviorOverride) Update(timeDelta float64, position *EntityPosition, dispatcher Dispatcher) {
	e.newBehavior.Update(timeDelta, position, dispatcher)
}

func (e *EntityBehaviorOverride) Reset() {
	e.replacedBehavior.Reset()
	e.newBehavior.Reset()
}
