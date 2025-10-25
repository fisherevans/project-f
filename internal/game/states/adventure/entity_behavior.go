package adventure

import "github.com/rs/zerolog/log"

type EntityBehavior interface {
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
	delete(b.disabledBy, source)
	log.Info().Str("id", b.id).Str("by", source).Msg("behavior enabled")
}

func (b *baseEntityBehavior) Disable(source string) {
	b.disabledBy[source] = struct{}{}
	b.system.positions[b.id].CancelMovement()
	log.Info().Str("id", b.id).Str("by", source).Msg("behavior disabled")
}

func (b *baseEntityBehavior) IsEnabled() bool {
	return len(b.disabledBy) == 0
}
