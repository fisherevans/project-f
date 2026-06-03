package game

import (
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/shaders"
)

var controls *input.Controls
var ControlsNoop *input.Controls

func Controls[T State]() *input.Controls {
	if ctx == nil {
		return controls
	}
	// when console it active, prevent all state input
	if Console().IsActive() {
		return ControlsNoop
	}
	// Compare the requested state type T against the runtime active state's type
	if _, ok := any(ctx.activeState).(T); ok {
		return controls
	}
	// Non-active states get a controls instance that is never updated (reads as neutral)
	return ControlsNoop
}

func SetVirtualControls(v input.VirtualState) {
	controls.SetVirtual(v)
}

func UpdateControls(window *opengl.Window) {
	controls.Update(window)
}

var ctx *context

type context struct {
	DebugInfo

	activeState State

	stateIntent any
	stateData   any

	save *rpg.GameSave

	elapsed      float64
	utils        ContextUtils
	debugToggles *DebugToggleSystem
	flags        *FlagRegister
	console      *CommandConsole
}

func Initialize(saveId string, intent any) {
	saves, err := rpg.LoadGameSaves()
	if err != nil {
		panic(err)
	}
	save, ok := saves[saveId]
	if !ok {
		save = &rpg.GameSave{
			SaveId:        saveId,
			CharacterName: "Todd",
		}
		save.FillDefaults()
		log.Info().Str("saveId", saveId).Msg("new save initialized")
	}
	ctx = &context{
		stateIntent:  intent,
		save:         save,
		debugToggles: newToggles(),
		flags:        newFlags(),
	}
	ctx.utils = ContextUtils{}

	controls = input.NewControls()
	ControlsNoop = input.NewControls()

	ctx.console = newConsole(func(input string) {
		if ctx.activeState == nil {
			Console().WriteLines("No active state")
			return
		}
		ctx.activeState.HandleConsoleInput(input)
	})
}

func Update(window *opengl.Window, timeDelta float64) {
	ctx.debugToggles.update(window)
	ctx.elapsed += timeDelta
}

func GetActiveState() State {
	return ctx.activeState
}

func SetActiveStateIntent(intent any) {
	SetActiveStateIntentWithData(intent, nil)
}

func SetActiveStateIntentWithData(intent, data any) {
	if ctx.stateIntent != nil {
		log.Warn().Msgf("something is overriding an existing state intent")
	}
	ctx.stateIntent = intent
	ctx.stateData = data
}

func ApplyIntent() {
	if ctx.stateIntent == nil {
		return
	}
	s, err := CreateStateFromIntent(ctx.stateIntent)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create new state from intent")
	}
	if ctx.activeState != nil {
		ctx.activeState.OnExit()
	}
	ctx.activeState = s
	ctx.stateIntent = nil
	data := ctx.stateData
	ctx.stateData = nil
	ctx.activeState.OnEnter(data)
}

type AppliedShader interface {
	Apply(shaderOptions shaders.Options, timeDelta float64)
}

func DebugNotificationf(format string, args ...any) {
	ctx.Notify(format, args...)
}

func DebugTLf(format string, a ...any) {
	ctx.Debug(AreaTopLeft, format, a...)
}

func DebugBLf(format string, a ...any) {
	ctx.Debug(AreaBottomLeft, format, a...)
}

func DebugTRf(format string, a ...any) {
	ctx.Debug(AreaTopRight, format, a...)
}

func DebugBRf(format string, a ...any) {
	ctx.Debug(AreaBottomRight, format, a...)
}

func DebugToggles() *DebugToggleSystem {
	return ctx.debugToggles
}

func CurrentSave() *rpg.GameSave {
	return ctx.save
}

// SwapSave replaces the active save with the one identified by saveId, loading
// it from disk (or creating a fresh default if it does not exist). The active
// state still references the previous save's data, so callers must trigger a
// state reset afterward for the swap to take full effect. Intended for the dev
// harness; not safe to call mid-frame from gameplay code.
func SwapSave(saveId string) error {
	saves, err := rpg.LoadGameSaves()
	if err != nil {
		return err
	}
	save, ok := saves[saveId]
	if !ok {
		save = &rpg.GameSave{
			SaveId:        saveId,
			CharacterName: "Todd",
		}
		save.FillDefaults()
		log.Info().Str("saveId", saveId).Msg("swap: new save initialized")
	}
	ctx.save = save
	return nil
}

func CurrentSaveOrNil() *rpg.GameSave {
	if ctx == nil {
		return nil
	}
	return ctx.save
}

func Utils() ContextUtils {
	return ctx.utils
}

func PopDebugLines() map[DebugArea][]string {
	return ctx.PopDebugLines()
}

func PopNotifications(time float64) []string {
	return ctx.PopNotifications(time)
}

func TimeElapsed() float64 {
	return ctx.elapsed
}

func Flags() *FlagRegister {
	return ctx.flags
}

func Console() *CommandConsole {
	return ctx.console
}
