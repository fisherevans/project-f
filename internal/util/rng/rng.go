// Package rng is the game's shared, seedable source of randomness for
// gameplay and visible-state values (animation jitter, effect spawns, damage
// variance, NPC behavior, script rand/randf, ...).
//
// By default it is seeded nondeterministically at startup, matching the old
// behavior of the per-call-site math/rand usage it replaces. Calling SetSeed
// makes the whole stream reproducible, which the dev harness uses for
// byte-stable golden-image regression: with a fixed seed and fixed-dt stepping
// from a reset, a captured frame is identical run to run.
//
// All functions are safe for concurrent use, but determinism only holds when
// the consumers run on a single thread (the game thread). Randomness consumed
// off the game thread - notably the audio package, which runs on the speaker
// goroutine - is intentionally NOT routed through here, because its interleaving
// with game-thread draws would be nondeterministic.
package rng

import (
	mrand "math/rand"
	"math/rand/v2"
	"sync"
)

var (
	mu  sync.Mutex
	src *rand.Rand
)

func init() {
	// Nondeterministic by default (uses the runtime's global entropy).
	src = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
}

// SetSeed reseeds the shared stream, making all subsequent draws reproducible.
func SetSeed(seed uint64) {
	mu.Lock()
	defer mu.Unlock()
	src = rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
}

func Float64() float64 {
	mu.Lock()
	defer mu.Unlock()
	return src.Float64()
}

func Float32() float32 {
	mu.Lock()
	defer mu.Unlock()
	return src.Float32()
}

// IntN returns a non-negative int in [0,n). It panics if n <= 0, matching the
// standard library.
func IntN(n int) int {
	mu.Lock()
	defer mu.Unlock()
	return src.IntN(n)
}

// Intn is an alias for IntN, easing migration from math/rand (v1).
func Intn(n int) int {
	return IntN(n)
}

// Int returns a non-negative pseudo-random int.
func Int() int {
	mu.Lock()
	defer mu.Unlock()
	return int(src.Uint64() >> 1)
}

func NormFloat64() float64 {
	mu.Lock()
	defer mu.Unlock()
	return src.NormFloat64()
}

func Uint64() uint64 {
	mu.Lock()
	defer mu.Unlock()
	return src.Uint64()
}

// NewV1 mints a standard math/rand (v1) generator seeded from the shared
// stream, for APIs that require a *math/rand.Rand. Because its seed is drawn
// from the shared stream, it stays reproducible under SetSeed as long as the
// call order is deterministic.
func NewV1() *mrand.Rand {
	return mrand.New(mrand.NewSource(int64(Uint64())))
}
