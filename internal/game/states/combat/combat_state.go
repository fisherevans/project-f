package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/states/combat/tick_bar"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/highlighter"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"
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
var backgroundVignette *pixel.Sprite

func init() {
	resources.RunOnceInitialized(func() {
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

type Phase int

const (
	PhaseIntro Phase = iota
	PhaseBattle
	PhaseEnd
	PhaseReward
	PhaseTerminal
)

type State struct {
	game.BaseState
	Player     *Player
	Opponent   Opponent
	OnComplete game.CombatIntentComplete
	Battle     *Battle

	phase           Phase
	phaseStartTimes map[Phase]float64

	fx []FX

	combatArrowAlpha       float64
	combatArrowColumn      int
	cachedContents         map[string]*textbox.Content
	skillFlashTimeElapsed  float64
	skillFlashAlpha        float64
	skillFlashAlphaInverse float64

	backgroundSprite pixelutil.BoundedDrawable

	highlighter *highlighter.SequencedDrawer

	batch    *pixel.Batch
	training *TrainingListener

	visibilityBackground       *util.Visibility
	visibilityPlayer           *util.Visibility
	visibilityPlayerStats      *util.Visibility
	visibilityOpponentStats    *util.Visibility
	visibilityOpponent         *util.Visibility
	visibilityActiveSkillEater *util.Visibility
	visibilityActiveSkills     *util.Visibility
	visibilitySkillSelection   *util.Visibility
	visibilityTempoBar         *util.Visibility
	visibilityCombatantExit    *util.Visibility

	introTimers     []actionTimer
	battleEndTimers []actionTimer
	rewardTimers    []actionTimer

	endModal *endModal

	currentRewardModal int
	rewardModals       []*rewardModal

	reward         game.CombatReward
	onCompleteSent bool
}

func New(i game.CombatIntent) game.State {
	s := &State{
		Player:     NewPlayer(i.Player),
		Opponent:   NewPrimortalOpponent(i.Opponent),
		OnComplete: i.OnComplete,
		Battle:     &Battle{},

		phase:           PhaseIntro,
		phaseStartTimes: map[Phase]float64{},

		cachedContents: map[string]*textbox.Content{},

		backgroundSprite: atlas.GetSprite(i.Background),

		highlighter: highlighter.NewSequencedDrawer(highlighter.NewDrawer(atlas, resources.FontNameM3x6)),
		training:    NewTrainingListener(),

		batch: atlas.NewBatch(),

		visibilityBackground:       util.NewVisibility(false, 1.5, true),
		visibilityPlayer:           util.NewVisibility(false, 2, true),
		visibilityOpponent:         util.NewVisibility(false, 2, true),
		visibilityPlayerStats:      util.NewVisibility(false, 0.5, true),
		visibilityOpponentStats:    util.NewVisibility(false, 0.5, true),
		visibilityActiveSkillEater: util.NewVisibility(false, 0.25, true),
		visibilityActiveSkills:     util.NewVisibility(false, 0.75, true),
		visibilitySkillSelection:   util.NewVisibility(false, 0.75, true),
		visibilityTempoBar:         util.NewVisibility(false, 3, true),
		visibilityCombatantExit:    util.NewVisibility(false, 1, true),

		endModal:     newEndModal(i.Opponent.Type.Primortal().Name, i.Reward),
		rewardModals: newRewardModals(i.Reward),
		reward:       i.Reward,
	}
	s.introTimers = []actionTimer{
		newSetVisibleTimer(0, s.visibilityBackground, true),
		newSetVisibleTimer(0.1, s.visibilityPlayer, true),
		newSetVisibleTimer(0.2, s.visibilityOpponent, true),
		newSetVisibleTimer(2, s.visibilityOpponentStats, true),
		newSetVisibleTimer(2, s.visibilityPlayerStats, true),
		newSetVisibleTimer(2, s.visibilitySkillSelection, true),
		newSetVisibleTimer(3, s.visibilityTempoBar, true),
		newSetVisibleTimer(2.5, s.visibilityActiveSkillEater, true),
		newSetVisibleTimer(2.5, s.visibilityActiveSkills, true),
		{
			triggerAfter: 3.75,
		},
	}
	s.battleEndTimers = []actionTimer{
		newSetVisibleTimer(0, s.visibilitySkillSelection, false),
		newSetVisibleTimer(0, s.visibilityTempoBar, false),
		newSetVisibleTimer(0, s.visibilityActiveSkills, false),
		newSetVisibleTimer(0.5, s.visibilityActiveSkillEater, false),
		newSetVisibleTimer(0, s.visibilityBackground, false),
	}
	s.rewardTimers = []actionTimer{
		newSetVisibleTimer(0, s.visibilityOpponentStats, false),
		newSetVisibleTimer(0, s.visibilityPlayerStats, false),
		newSetVisibleTimer(0, s.visibilityCombatantExit, true),
	}
	s.loadTrainingSequence(i.TrainingSequence)
	return s
}

func (s *State) timeSincePhase(p Phase) float64 {
	start, ok := s.phaseStartTimes[p]
	if !ok {
		return 0
	}
	return game.TimeElapsed() - start
}

func (s *State) Controls() *input.Controls {
	if s.highlighter.IsActive() {
		return game.ControlsNoop
	}
	return game.Controls[*State]()
}

func (s *State) ClearColor() pixel.RGBA {
	return colors.Black.RGBA
}

func (s *State) OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64) {
	s.batch.Clear()

	s.visibilityBackground.Update(timeDelta)
	s.visibilityPlayer.Update(timeDelta)
	s.visibilityPlayerStats.Update(timeDelta)
	s.visibilityOpponentStats.Update(timeDelta)
	s.visibilityOpponent.Update(timeDelta)
	s.visibilityActiveSkills.Update(timeDelta)
	s.visibilityActiveSkillEater.Update(timeDelta)
	s.visibilitySkillSelection.Update(timeDelta)
	s.visibilityTempoBar.Update(timeDelta)
	s.visibilityCombatantExit.Update(timeDelta)

	if s.phase == PhaseIntro {
		s.introTimers = s.triggerTimers(s.introTimers, PhaseIntro)
		if len(s.introTimers) == 0 {
			s.phase = PhaseBattle
		}
	}
	if s.phase >= PhaseEnd {
		s.battleEndTimers = s.triggerTimers(s.battleEndTimers, PhaseEnd)
	}
	if s.phase >= PhaseReward {
		s.rewardTimers = s.triggerTimers(s.rewardTimers, PhaseReward)
	}

	bgMask := colors.Lerp(colors.White.RGBA, colors.Black.RGBA, 0.5*s.visibilityBackground.GetInvisibleAmount())
	bgMatrix := pixel.IM.Moved(targetBounds.Center())
	s.backgroundSprite.DrawColorMask(target, bgMatrix, bgMask)
	target.SetComposeMethod(pixel.ComposeMultiply)
	backgroundVignette.Draw(target, pixel.IM.Moved(targetBounds.Center()))
	target.SetComposeMethod(pixel.ComposeOver)

	if s.phase <= PhaseBattle {
		if s.visibilityPlayer.IsFullyVisible() {
			game.DebugTRf("sequences!")
			s.training.OnTick(s)
		}

		s.Opponent.GetHealth().Update(timeDelta)
		s.Player.GetCurrentShield().Update(timeDelta)
		s.Player.GetCurrentSync().Update(timeDelta)

		battleTimeDelta := timeDelta
		if s.phase == PhaseIntro { // slower during intro
			battleTimeDelta = battleTimeDelta * min(s.timeSincePhase(PhaseIntro), 1.0)
		}
		if s.training.ShouldPauseCombat() {
			battleTimeDelta = 0
		}
		s.Battle.Update(s, battleTimeDelta)
		s.Player.GetTempo().Update(s.Battle.TickPlayerNext, s.Player, battleTimeDelta)

		if s.Player.GetCurrentSync().GetCurrentInt() <= 0 || s.Opponent.GetHealth().GetCurrentInt() <= 0 {
			s.phase = PhaseEnd
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

	s.Player.GetRenderer().Render(s.batch, timeDelta, s.Player, s.visibilityPlayer, s.visibilityCombatantExit)
	s.Opponent.GetRenderer().Render(s.batch, timeDelta, s.Opponent, s.visibilityOpponent, s.visibilityCombatantExit)

	for _, fx := range s.fx {
		fx.Render(s.batch)
	}
	s.drawActiveSkills(s.batch, targetBounds, pixel.IM.Moved(pixel.V(targetBounds.Center().X, targetBounds.H())))
	s.Player.Tempo.Render(s.batch, gfx.Moved(game.GameWidth/2, 45), s.visibilityTempoBar, timeDelta)
	s.renderSkills(s.batch, targetBounds, timeDelta)

	s.drawPlayerStats(timeDelta)
	s.drawOpponentStats(timeDelta)

	game.DebugBLf("player status: %s", s.Player.GetStatuses().String())
	game.DebugBLf("opponent status: %s", s.Opponent.GetStatuses().String())

	// END MODAL
	if s.phase >= PhaseEnd {
		topMiddle := pixel.IM.Moved(pixel.V(game.GameWidth/2, game.GameHeight/2))
		s.endModal.render(s.batch, topMiddle, timeDelta, s.phase == PhaseEnd)
	}

	// REWARDS
	if s.phase >= PhaseReward && s.timeSincePhase(PhaseReward) > 1.0 {
		topMiddle := pixel.IM.Moved(pixel.V(game.GameWidth/2, game.GameHeight-23))
		dx := 0
		spacing := 100
		if len(s.rewardModals) > 0 {
			dx -= ((len(s.rewardModals) - 1) * spacing) / 2
		}
		for idx := 0; idx < len(s.rewardModals); idx++ {
			if idx <= s.currentRewardModal {
				s.rewardModals[idx].Render(s.batch, topMiddle.Moved(gfx.IVec(dx, 0)), idx == s.currentRewardModal, timeDelta)
			}
			dx += spacing
		}
	}

	// EXIT PHASE INPUTS
	if s.phase == PhaseEnd {
		if s.Controls().ButtonA().JustPressed() {
			s.endModal.exit()
			s.phase = PhaseReward
		}
	} else if s.phase == PhaseReward {
		if s.Controls().ButtonA().JustPressed() {
			if s.currentRewardModal < len(s.rewardModals) {
				s.currentRewardModal++
			}
			if s.currentRewardModal >= len(s.rewardModals) {
				s.phase = PhaseTerminal
				for _, r := range s.rewardModals {
					r.Exit()
				}
			}
		}
	} else if s.phase == PhaseTerminal {
		if !s.onCompleteSent && (s.Controls().ButtonA().JustPressed() || s.timeSincePhase(PhaseTerminal) > 1.0) {
			game.CurrentSave().GrantExperience(s.reward.ExperiencePoints)
			game.CurrentSave().GrantResearchPoints(s.reward.ResearchType, s.reward.ResearchPoints)
			s.OnComplete(s, game.CombatIntentResult{
				PlayerWon: s.Opponent.GetHealth().GetCurrentInt() <= 0,
			})
			s.onCompleteSent = true
		}
	}

	s.highlighter.Render(s.batch, timeDelta, game.Controls[*State]())
	s.batch.Draw(target)

	if _, ok := s.phaseStartTimes[s.phase]; !ok {
		s.phaseStartTimes[s.phase] = game.TimeElapsed()
	}
}
