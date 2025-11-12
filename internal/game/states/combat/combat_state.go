package combat

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"

	"image/color"
	"math"
)

// Things to add
// - enhance damage FX for various conditions (i.e. effective, immune, etc.)
// - add skill log
//   - add data structure to hold it
//   - add rendering
// - add arrow + select popup to view skill details
// - add health delay (interp to target) - allow temp death
//   - render target in bar
// - state stuff
//   - add transition state into combat (i.e. fade)
//   - add battle intro that has creature enter
//   - add initiative roll (maybe add ticks of confusion when starting)
//   - add battle xenolog (fight, run, item, etc.)
//   - add battle end (win, lose)
// - generate experience at battle end (include in end summary)
// - combat abilities
//   - blocking mechanism
//   - slow down tempo (make it smooth)
// - add simple skill animations for effects to make it easier to understand what's happening
//   - make sprite hop, like pokemon
//   - add some sprite for fire, etc.
// - add damage animations (i.e. red flash)
// - sound
//   - add damage noise (FIRST ONE)
//   - add background music
//   - add win vs lose chime
// - add more interesting AI (more skills, show next skill after time)

var ticksPerSecond = 2.25

var atlas = resources.DefaultAtlas()

func init() {
	atlas.Dump("temp", "combat")
}

var backgroundVignette = resources.LoadSprite("combat/background_vignette_mask")

type Phase string

const PhaseBattle Phase = "battle"
const PhaseComplete Phase = "complete"

type State struct {
	game.BaseState
	Player     *Player
	Opponent   Opponent
	OnComplete game.CombatIntentComplete
	Battle     *Battle

	phase Phase

	fx []FX

	combatArrowAlpha       float64
	combatArrowColumn      int
	cachedContents         map[string]*textbox.Content
	skillFlashTimeElapsed  float64
	skillFlashAlpha        float64
	skillFlashAlphaInverse float64

	backgroundSprite pixelutil.BoundedDrawable

	batch *pixel.Batch
}

func New(i game.CombatIntent) game.State {
	return &State{
		Player:     NewPlayer(game.CurrentSave().Animech),
		Opponent:   NewPrimortalOpponent(i.Opponent),
		OnComplete: i.OnComplete,
		Battle:     &Battle{},

		phase: PhaseBattle,

		cachedContents: map[string]*textbox.Content{},

		backgroundSprite: atlas.GetSprite(i.Background),

		batch: atlas.NewBatch(),
	}
}

func (s *State) ClearColor() color.Color {
	return color.Black
}

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	s.batch.Clear()

	s.backgroundSprite.Draw(target, pixel.IM.Moved(targetBounds.Center()))
	target.SetComposeMethod(pixel.ComposeMultiply)
	backgroundVignette.Draw(target, pixel.IM.Moved(targetBounds.Center()))
	target.SetComposeMethod(pixel.ComposeOver)

	if s.phase == PhaseBattle {
		s.Opponent.GetHealth().Update(timeDelta)
		s.Player.GetCurrentShield().Update(timeDelta)
		s.Player.GetCurrentSync().Update(timeDelta)

		s.Battle.Update(s, timeDelta)

		if s.Player.GetCurrentSync().GetCurrentInt() <= 0 || s.Opponent.GetHealth().GetCurrentInt() <= 0 {
			s.phase = PhaseComplete
		}
	}

	var remainingFx []FX
	for _, fx := range s.fx {
		if !fx.Update(s, timeDelta) {
			remainingFx = append(remainingFx, fx)
		}
	}
	s.fx = remainingFx

	s.Player.Update(timeDelta)
	s.Opponent.Update(timeDelta)

	s.renderCombatantSprite(s.Player.GetAnimation(), s.Player, true, timeDelta)
	s.renderCombatantSprite(s.Opponent.GetAnimation(), s.Opponent, false, timeDelta)

	s.drawActiveSkills(s.batch, targetBounds, pixel.IM.Moved(pixel.V(targetBounds.Center().X, targetBounds.H())))

	for _, fx := range s.fx {
		fx.Render(s.batch)
	}

	s.renderSkills(s.batch, targetBounds, timeDelta)

	s.drawPlayerStats()
	s.drawOpponentStats()

	s.Player.GetTempo().Render(s.batch, pixel.IM.Moved(pixel.V(8, game.GameHeight*0.6)))
	s.Opponent.GetTempo().Render(s.batch, pixel.IM.Moved(pixel.V(8+game.GameWidth/2, game.GameHeight*0.6)))

	game.DebugBLf("player status: %s", s.Player.GetStatuses().String())
	game.DebugBLf("opponent status: %s", s.Opponent.GetStatuses().String())

	if s.phase == PhaseComplete {
		overlay := "Battle complete!"
		result := game.CombatIntentResult{}
		if s.Player.GetCurrentSync().GetCurrentInt() <= 0 {
			overlay = "{+c:#e64565,+o}YOU DIED!"
		} else if s.Opponent.GetHealth().GetCurrentInt() <= 0 {
			overlay = "{+c:#45e682,+o}YOU WON!"
			result.PlayerWon = true
			result.ResearchPoints = 1
		}
		content := combatantNameText.NewComplexContent(overlay)
		combatantNameText.Render(s.batch, pixel.IM.Moved(pixel.V(game.GameWidth/2, math.Floor(game.GameHeight*0.6))), content, tbcfg.RenderFrom(gfx.Centered))
		if game.Controls[*State]().ButtonA().JustPressed() || game.Controls[*State]().ButtonB().JustPressed() {
			s.OnComplete(result)
		}
	}

	s.batch.Draw(target)
}

func (s *State) renderCombatantSprite(sprite *anim.AnimatedSprite, com Combatant, leftSide bool, timeDelta float64) {
	colorMask := com.GetColorMask()
	if com.IsDead() {
		colorMask = colors.MixColor(colorMask, colors.HexString("#af8686"))
	} else {
		sprite.Update(timeDelta)
	}

	var position pixel.Vec
	rotateDirection := 1.0
	if leftSide {
		position = pixel.V(math.Floor(game.GameWidth*0.2), math.Floor(game.GameHeight*0.4))
	} else {
		position = pixel.V(math.Floor(game.GameWidth*0.8), math.Floor(game.GameHeight*0.4))
		rotateDirection = -1
	}

	m := pixel.IM
	if com.IsDead() {
		h := sprite.Sprite().Bounds().H()
		ry := -h / 3.0
		m = m.Rotated(pixel.V(0, ry), math.Pi/2.0*rotateDirection)
	}
	m = m.Moved(position)

	sprite.Sprite().DrawColorMask(s.batch, m, colorMask)

}
