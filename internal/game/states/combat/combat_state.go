package combat

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/game/states/combat/tick_bar"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
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

var atlas *resources.Atlas
var backgroundVignette *pixel.Sprite

func init() {
	resources.RunOnceInitialized(func() {
		atlas = resources.DefaultAtlas()
		backgroundVignette = resources.LoadSprite("combat/background_vignette_mask")
		baseFxText = text.New(pixel.ZV, atlas.GetFont(resources.FontNameM3x6).Atlas).AlignedTo(pixel.Center)
		skillEaterSprite = atlas.GetSprite("combat/tick_bar/skill_eater")
		tickBarRenderer = tick_bar.NewRenderer(atlas)
		statusFrame = frames.New("combat/status_frame", atlas)
		statusBorder = atlas.GetSprite("combat/status_border")
		statusLevelSprites = map[rpg.StatusLevel]pixelutil.BoundedDrawable{
			rpg.StatusLevel0: atlas.GetTilesheetSprite("combat/status_level", 4, 1), // empty
			rpg.StatusLevel1: atlas.GetTilesheetSprite("combat/status_level", 1, 1),
			rpg.StatusLevel2: atlas.GetTilesheetSprite("combat/status_level", 2, 1),
			rpg.StatusLevel3: atlas.GetTilesheetSprite("combat/status_level", 3, 1),
		}
		tempoLevel3Border = anim.Load(atlas, "combat/combatant_stats/tempo:level_3_border")
		tempoLevel2Border = atlas.GetSprite("combat/combatant_stats/tempo:level_2_border")
		tempoLevel1Border = atlas.GetSprite("combat/combatant_stats/tempo:level_1_border")
		tempoBase = atlas.GetSprite("combat/combatant_stats/tempo:base")
		tempoName = atlas.GetSprite("combat/combatant_stats/tempo:name")
		tempoBarGradient = atlas.GetSprite("combat/combatant_stats/tempo_bar:gradient")
		tempoBarTick = atlas.GetSprite("combat/combatant_stats/tempo_bar:tick")
		combatStatFrame = frames.New("combat/combatant_stats/box", atlas)
		statBarFrame = frames.New("combat/combatant_stats/bar", atlas)
		combatantNameText = textbox.NewInstance(atlas.GetFont(resources.FontNameAddStandard), tbcfg.NewConfig(200, 0, tbcfg.WithExpandMode(tbcfg.ExpandFit)))
		combatantStatText = textbox.NewInstance(atlas.GetFont(resources.FontNameFF), tbcfg.NewConfig(200, 0, tbcfg.WithExpandMode(tbcfg.ExpandFit)))
		noneSelectedSprite = atlas.GetSprite("combat/tick_bar/skill_none_selected")
		statNameBoxSprite = atlas.GetTilesheetSprite("combat/combatant_stats/background", 1, 1)
		statRightSprite = atlas.GetTilesheetSprite("combat/combatant_stats/background", 2, 1)
		statBottomSprite = atlas.GetTilesheetSprite("combat/combatant_stats/background", 3, 1)
		skillFrame = frames.New("combat/menu/skill_frame", atlas)
		skillPendingFrame = frames.New("combat/menu/skill_pending_frame", atlas)
		skillText = textbox.NewInstance(atlas.GetFont(resources.FontNameM3x6), tbcfg.NewConfig(skillFrameWidth, skillFrameHeight,
			tbcfg.Foreground(colors.Black.RGBA),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignMiddle),
		))
		skillPendingProgress = anim.SkillPendingProgress(atlas)
		skillStatsBadge = badges.Using(atlas).ButtonAction("select", "stats", badges.ButtonStyleStandard)
		skillPendingCancelBadge = badges.Using(atlas).ButtonAction("a", "commit", badges.ButtonStyleStandard)
		skillCommittedCancelBadge = badges.Using(atlas).ButtonAction("b", "cancel", badges.ButtonStyleStandard)
		skillMenuBadge = badges.Using(atlas).ButtonAction("start", "xenolog", badges.ButtonStyleStandard)
	})
}

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
		s.Player.GetTempo().Update(s.Battle.TickPlayerNext, s.Player, timeDelta)

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

	s.Player.GetRenderer().Render(s.batch, timeDelta, s.Player)
	s.Opponent.GetRenderer().Render(s.batch, timeDelta, s.Opponent)

	if s.phase == PhaseBattle {
		for _, fx := range s.fx {
			fx.Render(s.batch)
		}
		s.drawActiveSkills(s.batch, targetBounds, pixel.IM.Moved(pixel.V(targetBounds.Center().X, targetBounds.H())))
		s.Player.Tempo.Render(s.batch, gfx.Moved(game.GameWidth/2, 45), timeDelta)
		s.renderSkills(s.batch, targetBounds, timeDelta)
	}

	s.drawPlayerStats(timeDelta)
	s.drawOpponentStats(timeDelta)

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
