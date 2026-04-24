import type { SpriteTilesheetAnimation } from "@/types/sprites"

export interface AnimFrame {
    row: number
    col: number
    weight: number
}

export function resolveFrames(
    anim: SpriteTilesheetAnimation,
    totalCols: number,
    totalRows: number,
): AnimFrame[] {
    let frames: AnimFrame[] = []

    if (anim.h_sequence) {
        const from = anim.h_sequence.fromColumn ?? 1
        const to = anim.h_sequence.toColumn ?? totalCols
        const row = anim.h_sequence.row
        if (from <= to) {
            for (let c = from; c <= to; c++) {
                const w = anim.h_sequence.frameWeights?.[c - from] ?? 1
                frames.push({ row, col: c, weight: w })
            }
        } else {
            for (let c = from; c >= to; c--) {
                const w = anim.h_sequence.frameWeights?.[from - c] ?? 1
                frames.push({ row, col: c, weight: w })
            }
        }
    } else if (anim.v_sequence) {
        const from = anim.v_sequence.fromRow ?? 1
        const to = anim.v_sequence.toRow ?? totalRows
        const col = anim.v_sequence.column
        if (from <= to) {
            for (let r = from; r <= to; r++) {
                const w = anim.v_sequence.frameWeights?.[r - from] ?? 1
                frames.push({ row: r, col, weight: w })
            }
        } else {
            for (let r = from; r >= to; r--) {
                const w = anim.v_sequence.frameWeights?.[from - r] ?? 1
                frames.push({ row: r, col, weight: w })
            }
        }
    } else if (anim.tiles) {
        frames = anim.tiles.map((t) => ({
            row: t.row,
            col: t.column,
            weight: t.weight ?? 1,
        }))
    }

    if (anim.reverse) {
        frames.reverse()
    }

    if (anim.pingPong && frames.length > 2) {
        for (let i = frames.length - 2; i > 0; i--) {
            frames.push({ ...frames[i] })
        }
    }

    return frames
}

export class AnimationPlayer {
    frames: AnimFrame[]
    framesPerSecond: number
    jitterPercent: number
    randomized: boolean
    repeat: boolean

    currentFrame = 0
    progression = 0
    complete = false
    playing = true
    private nextJitter = 0

    constructor(anim: SpriteTilesheetAnimation, totalCols: number, totalRows: number) {
        this.frames = resolveFrames(anim, totalCols, totalRows)
        this.framesPerSecond = anim.framesPerSecond || 1
        this.jitterPercent = anim.jitterPercent ?? 0
        this.randomized = anim.randomize ?? false
        this.repeat = anim.repeat ?? true
    }

    update(dt: number) {
        if (!this.playing || this.frames.length === 0) return
        this.progression += dt
        this.progress()
    }

    private progress() {
        for (;;) {
            const f = this.frames[this.currentFrame]
            let w = f.weight
            if (w <= 0) w = 1
            const dur = (w + this.nextJitter) / this.framesPerSecond
            if (this.progression < dur) break
            this.complete = false
            let nextFrame = this.currentFrame + 1
            this.progression -= dur
            if (this.randomized) {
                nextFrame = Math.floor(Math.random() * this.frames.length)
            } else if (nextFrame >= this.frames.length) {
                if (this.repeat) {
                    nextFrame = 0
                } else {
                    nextFrame = this.frames.length - 1
                    this.complete = true
                }
            }
            this.currentFrame = nextFrame
            const maxJitter = this.frames[this.currentFrame].weight * this.jitterPercent
            this.nextJitter = maxJitter * Math.random() - maxJitter / 2
        }
    }

    stepForward() {
        if (this.frames.length === 0) return
        this.currentFrame = (this.currentFrame + 1) % this.frames.length
        this.progression = 0
        this.nextJitter = 0
    }

    stepBackward() {
        if (this.frames.length === 0) return
        this.currentFrame = (this.currentFrame - 1 + this.frames.length) % this.frames.length
        this.progression = 0
        this.nextJitter = 0
    }

    reset() {
        this.currentFrame = 0
        this.progression = 0
        this.complete = false
        this.nextJitter = 0
    }

    get currentTile(): AnimFrame | null {
        if (this.frames.length === 0) return null
        return this.frames[this.currentFrame]
    }
}
