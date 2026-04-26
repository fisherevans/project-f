import { useCallback, useEffect, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { TilesheetCoordinates, SpriteTilesheetAnimation, SpriteTilesheet } from "@/types/sprites"

interface TilesheetViewerProps {
    imageSrc: string
    imageWidth: number
    imageHeight: number
    tileWidth: number
    tileHeight: number
    sprites?: Record<string, TilesheetCoordinates>
    animations?: Record<string, SpriteTilesheetAnimation>
    selectedAnimation?: string
    onTileClick?: (col: number, row: number, shiftKey: boolean) => void
    bgClass?: string
    tilesheet: SpriteTilesheet
    onTilesheetChange: (ts: SpriteTilesheet) => void
}

const ZOOM_LEVELS = [1, 2, 4, 8] as const

export function TilesheetViewer({
    imageSrc,
    imageWidth,
    imageHeight,
    tileWidth,
    tileHeight,
    sprites,
    animations,
    selectedAnimation,
    onTileClick,
    bgClass,
    tilesheet,
    onTilesheetChange,
}: TilesheetViewerProps) {
    const canvasRef = useRef<HTMLCanvasElement>(null)
    const [zoom, setZoom] = useState(2)
    const [hoverTile, setHoverTile] = useState<{ col: number; row: number } | null>(null)
    const imageRef = useRef<HTMLImageElement | null>(null)

    const cols = Math.floor(imageWidth / tileWidth)
    const rows = Math.floor(imageHeight / tileHeight)
    const canvasWidth = imageWidth * zoom
    const canvasHeight = imageHeight * zoom

    const draw = useCallback(() => {
        const canvas = canvasRef.current
        if (!canvas || !imageRef.current) return
        const ctx = canvas.getContext("2d")
        if (!ctx) return

        ctx.imageSmoothingEnabled = false
        ctx.clearRect(0, 0, canvasWidth, canvasHeight)
        ctx.drawImage(imageRef.current, 0, 0, canvasWidth, canvasHeight)

        ctx.strokeStyle = "rgba(255, 255, 255, 0.3)"
        ctx.lineWidth = 1
        ctx.setLineDash([2, 2])
        for (let c = 1; c < cols; c++) {
            const x = c * tileWidth * zoom
            ctx.beginPath()
            ctx.moveTo(x + 0.5, 0)
            ctx.lineTo(x + 0.5, canvasHeight)
            ctx.stroke()
        }
        for (let r = 1; r < rows; r++) {
            const y = r * tileHeight * zoom
            ctx.beginPath()
            ctx.moveTo(0, y + 0.5)
            ctx.lineTo(canvasWidth, y + 0.5)
            ctx.stroke()
        }
        ctx.setLineDash([])

        if (sprites) {
            for (const [name, coords] of Object.entries(sprites)) {
                const x = (coords.column - 1) * tileWidth * zoom
                const y = (coords.row - 1) * tileHeight * zoom
                ctx.fillStyle = "rgba(0, 200, 0, 0.15)"
                ctx.fillRect(x, y, tileWidth * zoom, tileHeight * zoom)
                ctx.strokeStyle = "rgba(34, 197, 94, 0.8)"
                ctx.lineWidth = 2
                ctx.strokeRect(x + 1, y + 1, tileWidth * zoom - 2, tileHeight * zoom - 2)
                ctx.fillStyle = "rgba(0, 0, 0, 0.6)"
                ctx.font = "10px monospace"
                ctx.fillRect(x + 1, y + 1, ctx.measureText(name).width + 4, 12)
                ctx.fillStyle = "rgba(34, 197, 94, 1)"
                ctx.fillText(name, x + 3, y + 11)
            }
        }

        if (selectedAnimation && animations?.[selectedAnimation]) {
            const anim = animations[selectedAnimation]
            const tiles = getAnimationTiles(anim, cols, rows)
            tiles.forEach((t, i) => {
                const x = (t.col - 1) * tileWidth * zoom
                const y = (t.row - 1) * tileHeight * zoom
                ctx.fillStyle = "rgba(59, 130, 246, 0.2)"
                ctx.fillRect(x, y, tileWidth * zoom, tileHeight * zoom)
                ctx.strokeStyle = "rgba(59, 130, 246, 0.8)"
                ctx.lineWidth = 2
                ctx.strokeRect(x + 1, y + 1, tileWidth * zoom - 2, tileHeight * zoom - 2)
                ctx.fillStyle = "rgba(59, 130, 246, 1)"
                ctx.font = "bold 10px monospace"
                ctx.fillText(String(i + 1), x + 3, y + tileHeight * zoom - 4)
            })
        }

        if (hoverTile) {
            const x = (hoverTile.col - 1) * tileWidth * zoom
            const y = (hoverTile.row - 1) * tileHeight * zoom
            ctx.strokeStyle = "rgba(255, 255, 255, 0.8)"
            ctx.lineWidth = 2
            ctx.strokeRect(x + 1, y + 1, tileWidth * zoom - 2, tileHeight * zoom - 2)
        }
    }, [canvasWidth, canvasHeight, cols, rows, tileWidth, tileHeight, zoom, sprites, animations, selectedAnimation, hoverTile])

    useEffect(() => {
        const img = new Image()
        img.onload = () => {
            imageRef.current = img
            draw()
        }
        img.src = imageSrc
    }, [imageSrc, draw])

    useEffect(() => {
        draw()
    }, [draw])

    const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
        const rect = e.currentTarget.getBoundingClientRect()
        const x = e.clientX - rect.left
        const y = e.clientY - rect.top
        const col = Math.floor(x / (tileWidth * zoom)) + 1
        const row = Math.floor(y / (tileHeight * zoom)) + 1
        if (col >= 1 && col <= cols && row >= 1 && row <= rows) {
            setHoverTile({ col, row })
        } else {
            setHoverTile(null)
        }
    }

    const handleClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
        if (!onTileClick) return
        const rect = e.currentTarget.getBoundingClientRect()
        const x = e.clientX - rect.left
        const y = e.clientY - rect.top
        const col = Math.floor(x / (tileWidth * zoom)) + 1
        const row = Math.floor(y / (tileHeight * zoom)) + 1
        if (col >= 1 && col <= cols && row >= 1 && row <= rows) {
            onTileClick(col, row, e.shiftKey)
        }
    }

    return (
        <div>
            <div className="flex items-center gap-2 mb-2 flex-wrap">
                <div className="flex items-center gap-1.5 text-xs">
                    <Label className="text-xs text-muted-foreground">Tile</Label>
                    <Input
                        type="number"
                        className="w-14 h-6 text-xs px-1"
                        min={1}
                        value={tilesheet.tileWidth}
                        onChange={(e) => onTilesheetChange({ ...tilesheet, tileWidth: parseInt(e.target.value) || 1 })}
                    />
                    <span className="text-muted-foreground">x</span>
                    <Input
                        type="number"
                        className="w-14 h-6 text-xs px-1"
                        min={1}
                        value={tilesheet.tileHeight}
                        onChange={(e) => onTilesheetChange({ ...tilesheet, tileHeight: parseInt(e.target.value) || 1 })}
                    />
                    <span className="text-muted-foreground">= {cols}x{rows} grid</span>
                </div>
                <div className="flex gap-1 ml-auto">
                    {ZOOM_LEVELS.map((z) => (
                        <Button
                            key={z}
                            variant={zoom === z ? "default" : "outline"}
                            size="sm"
                            className="h-6 w-7 text-xs p-0"
                            onClick={() => setZoom(z)}
                        >
                            {z}x
                        </Button>
                    ))}
                </div>
                {hoverTile && (
                    <span className="text-xs text-muted-foreground">
                        r{hoverTile.row} c{hoverTile.col}
                    </span>
                )}
            </div>
            <div className={`overflow-auto rounded border border-border ${bgClass ?? "bg-canvas"}`}>
                <canvas
                    ref={canvasRef}
                    width={canvasWidth}
                    height={canvasHeight}
                    style={{ imageRendering: "pixelated" }}
                    className="cursor-crosshair"
                    onMouseMove={handleMouseMove}
                    onMouseLeave={() => setHoverTile(null)}
                    onClick={handleClick}
                />
            </div>
        </div>
    )
}

function getAnimationTiles(
    anim: SpriteTilesheetAnimation,
    totalCols: number,
    totalRows: number,
): { col: number; row: number }[] {
    if (anim.h_sequence) {
        const from = anim.h_sequence.fromColumn ?? 1
        const to = anim.h_sequence.toColumn ?? totalCols
        const tiles: { col: number; row: number }[] = []
        if (from <= to) {
            for (let c = from; c <= to; c++) tiles.push({ col: c, row: anim.h_sequence.row })
        } else {
            for (let c = from; c >= to; c--) tiles.push({ col: c, row: anim.h_sequence.row })
        }
        return tiles
    }
    if (anim.v_sequence) {
        const from = anim.v_sequence.fromRow ?? 1
        const to = anim.v_sequence.toRow ?? totalRows
        const tiles: { col: number; row: number }[] = []
        if (from <= to) {
            for (let r = from; r <= to; r++) tiles.push({ col: anim.v_sequence.column, row: r })
        } else {
            for (let r = from; r >= to; r--) tiles.push({ col: anim.v_sequence.column, row: r })
        }
        return tiles
    }
    if (anim.tiles) {
        return anim.tiles.map((t) => ({ col: t.column, row: t.row }))
    }
    return []
}
