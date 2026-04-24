import { useCallback, useEffect, useRef, useState } from "react"
import { Play, Pause, SkipBack, SkipForward, RotateCcw } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Slider } from "@/components/ui/slider"
import { AnimationPlayer } from "@/lib/animationEngine"
import type { SpriteTilesheetAnimation } from "@/types/sprites"

interface AnimationPreviewProps {
    imageSrc: string
    imageWidth: number
    imageHeight: number
    tileWidth: number
    tileHeight: number
    totalCols: number
    totalRows: number
    animation: SpriteTilesheetAnimation
}

export function AnimationPreview({
    imageSrc,
    imageWidth: _imageWidth,
    imageHeight: _imageHeight,
    tileWidth,
    tileHeight,
    totalCols,
    totalRows,
    animation,
}: AnimationPreviewProps) {
    const canvasRef = useRef<HTMLCanvasElement>(null)
    const stripCanvasRef = useRef<HTMLCanvasElement>(null)
    const imageRef = useRef<HTMLImageElement | null>(null)
    const playerRef = useRef<AnimationPlayer | null>(null)
    const rafRef = useRef<number>(0)
    const lastTimeRef = useRef<number>(0)
    const [playing, setPlaying] = useState(true)
    const [speed, setSpeed] = useState(1)
    const [frameIndex, setFrameIndex] = useState(0)
    const [frameCount, setFrameCount] = useState(0)

    void _imageWidth
    void _imageHeight

    const previewScale = Math.max(1, Math.floor(128 / Math.max(tileWidth, tileHeight)))

    useEffect(() => {
        const player = new AnimationPlayer(animation, totalCols, totalRows)
        player.playing = playing
        playerRef.current = player
        setFrameCount(player.frames.length)
        setFrameIndex(0)
    }, [animation, totalCols, totalRows, playing])

    useEffect(() => {
        const img = new Image()
        img.onload = () => {
            imageRef.current = img
        }
        img.src = imageSrc
    }, [imageSrc])

    const drawFrame = useCallback((player: AnimationPlayer) => {
        const canvas = canvasRef.current
        const img = imageRef.current
        if (!canvas || !img) return
        const ctx = canvas.getContext("2d")
        if (!ctx) return

        ctx.imageSmoothingEnabled = false
        ctx.clearRect(0, 0, canvas.width, canvas.height)

        const tile = player.currentTile
        if (!tile) return

        const sx = (tile.col - 1) * tileWidth
        const sy = (tile.row - 1) * tileHeight
        ctx.drawImage(img, sx, sy, tileWidth, tileHeight, 0, 0, tileWidth * previewScale, tileHeight * previewScale)
    }, [tileWidth, tileHeight, previewScale])

    const drawStrip = useCallback((player: AnimationPlayer) => {
        const canvas = stripCanvasRef.current
        const img = imageRef.current
        if (!canvas || !img || player.frames.length === 0) return
        const ctx = canvas.getContext("2d")
        if (!ctx) return

        const stripSize = 32
        const count = player.frames.length
        canvas.width = count * (stripSize + 2) - 2
        canvas.height = stripSize
        ctx.imageSmoothingEnabled = false
        ctx.clearRect(0, 0, canvas.width, canvas.height)

        for (let i = 0; i < count; i++) {
            const f = player.frames[i]
            const sx = (f.col - 1) * tileWidth
            const sy = (f.row - 1) * tileHeight
            const dx = i * (stripSize + 2)

            if (i === player.currentFrame) {
                ctx.fillStyle = "rgba(59, 130, 246, 0.3)"
                ctx.fillRect(dx - 1, 0, stripSize + 2, stripSize)
                ctx.strokeStyle = "rgba(59, 130, 246, 0.8)"
                ctx.lineWidth = 2
                ctx.strokeRect(dx, 0, stripSize, stripSize)
            }

            ctx.drawImage(img, sx, sy, tileWidth, tileHeight, dx, 0, stripSize, stripSize)
        }
    }, [tileWidth, tileHeight])

    useEffect(() => {
        const animate = (time: number) => {
            const player = playerRef.current
            if (!player) return
            if (lastTimeRef.current > 0) {
                const dt = ((time - lastTimeRef.current) / 1000) * speed
                player.update(dt)
                setFrameIndex(player.currentFrame)
            }
            lastTimeRef.current = time
            drawFrame(player)
            drawStrip(player)
            rafRef.current = requestAnimationFrame(animate)
        }
        if (playing) {
            lastTimeRef.current = 0
            rafRef.current = requestAnimationFrame(animate)
        } else {
            const player = playerRef.current
            if (player) {
                drawFrame(player)
                drawStrip(player)
            }
        }
        return () => cancelAnimationFrame(rafRef.current)
    }, [playing, speed, drawFrame, drawStrip])

    const handleStepForward = () => {
        playerRef.current?.stepForward()
        const player = playerRef.current
        if (player) {
            setFrameIndex(player.currentFrame)
            drawFrame(player)
            drawStrip(player)
        }
    }

    const handleStepBackward = () => {
        playerRef.current?.stepBackward()
        const player = playerRef.current
        if (player) {
            setFrameIndex(player.currentFrame)
            drawFrame(player)
            drawStrip(player)
        }
    }

    const handleReset = () => {
        playerRef.current?.reset()
        const player = playerRef.current
        if (player) {
            setFrameIndex(0)
            drawFrame(player)
            drawStrip(player)
        }
    }

    const handleStripClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
        const player = playerRef.current
        if (!player || player.frames.length === 0) return
        const rect = e.currentTarget.getBoundingClientRect()
        const x = e.clientX - rect.left
        const stripSize = 32
        const idx = Math.floor(x / (stripSize + 2))
        if (idx >= 0 && idx < player.frames.length) {
            player.currentFrame = idx
            player.progression = 0
            setFrameIndex(idx)
            drawFrame(player)
            drawStrip(player)
        }
    }

    const currentTile = playerRef.current?.currentTile

    return (
        <div className="space-y-3">
            <div className="flex items-start gap-4">
                <div className="rounded border border-border bg-zinc-900 overflow-hidden inline-block">
                    <canvas
                        ref={canvasRef}
                        width={tileWidth * previewScale}
                        height={tileHeight * previewScale}
                        style={{ imageRendering: "pixelated" }}
                    />
                </div>
                <div className="space-y-2 text-xs text-muted-foreground">
                    <p>Frame: {frameIndex + 1} / {frameCount}</p>
                    {currentTile && (
                        <>
                            <p>Tile: r{currentTile.row} c{currentTile.col}</p>
                            <p>Weight: {currentTile.weight}</p>
                        </>
                    )}
                    <p>FPS: {animation.framesPerSecond} x{speed.toFixed(1)}</p>
                </div>
            </div>

            <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={handleReset}>
                    <RotateCcw className="h-3.5 w-3.5" />
                </Button>
                <Button variant="outline" size="sm" onClick={handleStepBackward}>
                    <SkipBack className="h-3.5 w-3.5" />
                </Button>
                <Button variant="outline" size="sm" onClick={() => setPlaying(!playing)}>
                    {playing ? <Pause className="h-3.5 w-3.5" /> : <Play className="h-3.5 w-3.5" />}
                </Button>
                <Button variant="outline" size="sm" onClick={handleStepForward}>
                    <SkipForward className="h-3.5 w-3.5" />
                </Button>
                <div className="flex items-center gap-2 ml-4">
                    <span className="text-xs text-muted-foreground w-8">x{speed.toFixed(1)}</span>
                    <Slider
                        className="w-24"
                        min={0.1}
                        max={3}
                        step={0.1}
                        value={[speed]}
                        onValueChange={(v) => setSpeed(Array.isArray(v) ? v[0] : v)}
                    />
                </div>
            </div>

            {frameCount > 0 && (
                <div className="overflow-x-auto">
                    <canvas
                        ref={stripCanvasRef}
                        style={{ imageRendering: "pixelated", cursor: "pointer" }}
                        onClick={handleStripClick}
                    />
                </div>
            )}
        </div>
    )
}
