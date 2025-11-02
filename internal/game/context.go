package game

import (
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/shaders"
)

var controls *input.Controls
var controlsNoop *input.Controls

func Controls[T State]() *input.Controls {
	if ctx == nil {
		return controls
	}
	// Compare the requested state type T against the runtime active state's type
	if _, ok := any(ctx.activeState).(T); ok {
		return controls
	}
	// Non-active states get a controls instance that is never updated (reads as neutral)
	return controlsNoop
}

func UpdateControls(window *opengl.Window) {
	controls.Update(window)
}

var ctx *context

type context struct {
	DebugInfo

	activeState  State
	customShader AppliedShader

	stateIntent any

	window *opengl.Window

	save *rpg.GameSave

	elapsed      float64
	utils        ContextUtils
	debugToggles *DebugToggleSystem
	flags        *FlagRegister
}

func Initialize(window *opengl.Window, saveId string) {
	saves, err := rpg.LoadGameSaves()
	if err != nil {
		panic(err)
	}
	save, ok := saves[saveId]
	if !ok {
		panic("Save not found: " + saveId)
	}
	ctx = &context{
		stateIntent:  StartupDeviceIntent{},
		window:       window,
		save:         save,
		debugToggles: newToggles(),
		flags:        newFlags(),
	}
	ctx.utils = ContextUtils{}

	controls = input.NewControls()
	controlsNoop = input.NewControls()
}

func Update(window *opengl.Window, timeDelta float64) {
	ctx.debugToggles.update(window)
	ctx.elapsed += timeDelta
}

func GetActiveState() State {
	return ctx.activeState
}

func SetActiveStateIntent(intent any) {
	if ctx.stateIntent != nil {
		log.Warn().Msgf("something is overriding an existing state intent")
	}
	ctx.stateIntent = intent
}

func ApplyIntent() {
	if ctx.stateIntent == nil {
		return
	}
	s, err := createState(ctx.stateIntent)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to ctxreate new state from intent")
	}
	ctx.activeState = s
	ctx.stateIntent = nil
}

func SetCustomShader(shader AppliedShader) {
	ctx.customShader = shader
}

func GetCustomShader() AppliedShader {
	return ctx.customShader
}

func RemoveCustomShader() {
	ctx.customShader = nil
}

type AppliedShader interface {
	Apply(shaderOptions shaders.Options, timeDelta float64)
}

func DebugNotification(format string, args ...any) {
	ctx.Notify(format, args...)
}

func DebugTL(format string, a ...any) {
	ctx.Debug(AreaTopLeft, format, a...)
}

func DebugBL(format string, a ...any) {
	ctx.Debug(AreaBottomLeft, format, a...)
}

func DebugTR(format string, a ...any) {
	ctx.Debug(AreaTopRight, format, a...)
}

func DebugBR(format string, a ...any) {
	ctx.Debug(AreaBottomRight, format, a...)
}

func DebugToggles() *DebugToggleSystem {
	return ctx.debugToggles
}

func CurrentSave() *rpg.GameSave {
	return ctx.save
}

func Window() *opengl.Window {
	return ctx.window
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
