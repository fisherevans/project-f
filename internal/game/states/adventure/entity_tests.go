package adventure

import (
	"fmt"
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/gopxl/pixel/v2"
)

type EntityMenuTest struct {
	InnateEntity
	Passable
}

func (e *EntityMenuTest) RenderScene(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityMenuTest) Interact(adv *State, source Entity) {
	game.SetActiveStateIntent(game.MenuIntent{
		Background: adv,
	})
}

type EntityCombatTest struct {
	InnateEntity
	TimeTillNextQuip float64
	Passable
}

func NewEntityCombatTest(id EntityId, location MapLocation) Entity {
	return &EntityCombatTest{
		InnateEntity: InnateEntity{
			BaseEntity: BaseEntity{
				Id:           id,
				Interactable: true,
			},
			MapLocation: location,
		},
		TimeTillNextQuip: 10,
		Passable:         newPassablePreventIngress(true),
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

func (e *EntityCombatTest) Update(adv *State, timeDelta float64) {
	e.InnateEntity.Update(adv, timeDelta)
	e.TimeTillNextQuip -= timeDelta
	if e.TimeTillNextQuip <= 0 {
		adv.chatters.Add(newBasicEntityChatter(e.Id, 4, dummyQuips[rand.Intn(len(dummyQuips))], ""))
		e.TimeTillNextQuip = 10 + rand.Float64()*10
	}
}

func (e *EntityCombatTest) Interact(adv *State, source Entity) {
	stutters := []string{"Beep.", "Boop.", "Die.", "Die!", "{+u}DIE!!!{-u"}
	var message string
	for id, s := range stutters {
		if id > 0 {
			message += fmt.Sprintf("{+w:%d} {-w}", id*3)
		}
		message += s
	}
	adv.AddSystemEffect(events.Effect{
		Plan: &events.EffectPlan{
			Steps: []events.PlanStep{
				{
					Serial: []events.Effect{
						{
							Dialogue: &events.EffectDialogue{
								Text: message,
							},
						},
						{
							TriggerCombat: &events.EffectTriggerCombat{
								Opponent:   &rpg.Primortal_Dummy.Type,
								Background: "combat/background_space_base",
							},
						},
						{
							Dialogue: &events.EffectDialogue{
								Text: "Well, butter my bolts... you actually did it.",
							},
						},
					},
				},
			},
		},
	})
}
