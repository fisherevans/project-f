//go:build js && wasm

package runtime

import "syscall/js"

func signalReady() {
	fn := js.Global().Get("primortalOnReady")
	if fn.Type() == js.TypeFunction {
		fn.Invoke()
	}
}

func emitProgress(stage string, current, total int) {
	fn := js.Global().Get("primortalOnProgress")
	if fn.Type() == js.TypeFunction {
		fn.Invoke(stage, current, total)
	}
}
