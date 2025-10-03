package adventure

import (
	"fmt"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
)

type EntityMenuTest struct {
	InnateEntity
}

func (e *EntityMenuTest) RenderScene(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityMenuTest) Interact(ctx *game.Context, adv *State, source Entity) {
	ctx.SetActiveStateIntent(game.MenuIntent{
		Background: adv,
	})
}

type EntityCombatTest struct {
	InnateEntity
}

var dummyCombatSprite = atlas.GetSprite("primortals/dummy_entity")

func (e *EntityCombatTest) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	dummyCombatSprite.Draw(target, matrix)
}

func (e *EntityCombatTest) Interact(ctx *game.Context, adv *State, source Entity) {
	stutters := []string{"Beep.", "Boop.", "Die.", "Die!", "{+u}DIE!!!{-u"}
	var message string
	for id, s := range stutters {
		if id > 0 {
			message += fmt.Sprintf("{+w:%d} {-w}", id*3)
		}
		message += s
	}
	log.Info().Msgf("message: %s", message)
	adv.dialogues.Append(NewBasicDialogue(message, func(ctx *game.Context, s *State) {
		adv.TriggerCombat(&rpg.Primortal_Dummy.Type, "combat/background_space_base", nil)
	}))
}
