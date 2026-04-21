//go:build !js

package runtime

func signalReady()                                        {}
func emitProgress(stage string, current, total int)       {}
