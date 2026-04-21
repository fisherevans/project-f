//go:build js && wasm

package runtime

import "syscall/js"

func signalReady() {
	fn := js.Global().Get("primortalOnReady")
	if fn.Type() == js.TypeFunction {
		fn.Invoke()
	}
	// Block until JS calls primortalStart() after the user's first gesture.
	// This ensures audio context is unlocked before the game loop (and intro
	// music) starts.
	ch := make(chan struct{})
	cb := js.FuncOf(func(this js.Value, args []js.Value) any {
		close(ch)
		return nil
	})
	js.Global().Set("primortalStart", cb)
	<-ch
	cb.Release()
}

func emitProgress(stage string, current, total int) {
	fn := js.Global().Get("primortalOnProgress")
	if fn.Type() == js.TypeFunction {
		fn.Invoke(stage, current, total)
	}
}
