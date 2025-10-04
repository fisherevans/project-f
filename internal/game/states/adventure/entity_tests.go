package adventure

import (
	"fmt"
	"math/rand"

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
	TimeTillNextQuip float64
}

func NewEntityCombatTest(id EntityId, location MapLocation) Entity {
	return &EntityCombatTest{
		InnateEntity: InnateEntity{
			BaseEntity: BaseEntity{
				Id:       id,
				Passable: false,
			},
			MapLocation: location,
		},
		TimeTillNextQuip: 10,
	}
}

var dummyCombatSprite = atlas.GetSprite("primortals/dummy_entity")
var dummyQuips = []string{
	"Practice those steps - then try me.",
	"Come close - I don't bite... yet.",
	"Stretch first. I hate easy wins.",
	"Make a move, meatbag.",
	"You look fragile. Let's verify.",
	"Come fight me, big guy.",
}

func (e *EntityCombatTest) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	dummyCombatSprite.Draw(target, matrix)
}

func (e *EntityCombatTest) Update(ctx *game.Context, adv *State, timeDelta float64) {
	e.InnateEntity.Update(ctx, adv, timeDelta)
	e.TimeTillNextQuip -= timeDelta
	if e.TimeTillNextQuip <= 0 {
		adv.chatters.Add(newBasicEntityChatter(e.Id, 4, dummyQuips[rand.Intn(len(dummyQuips))]))
		e.TimeTillNextQuip = 10 + rand.Float64()*10
	}
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
		adv.TriggerCombat(&rpg.Primortal_Dummy.Type, "combat/background_space_base", func(ctx *game.Context, s *State) {
			adv.dialogues.Append(NewBasicDialogue("Well, butter my bolts... you actually did it.", nil))
		})
	}))
}
