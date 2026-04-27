import { useRef, useEffect, useCallback } from "react"
import type { OverlayRect } from "@/types/overlays"
import { useRectDrag } from "./useRectDrag"

const CANVAS_W = 240
const CANVAS_H = 160

interface Props {
    rects: Record<string, OverlayRect>
    selectedName: string | null
    onSelectName?: (name: string) => void
    onRectChange?: (name: string, rect: OverlayRect) => void
}

function toCanvasY(rect: OverlayRect): number {
    return CANVAS_H - rect.y - rect.h
}

export function RectCanvasPreview({ rects, selectedName, onSelectName, onRectChange }: Props) {
    const canvasRef = useRef<HTMLCanvasElement>(null)

    const getSelectedRect = useCallback((): OverlayRect | null => {
        if (!selectedName || !rects[selectedName]) return null
        return rects[selectedName]
    }, [selectedName, rects])

    const getLayout = useCallback(() => {
        const canvas = canvasRef.current
        if (!canvas) return null
        const displayW = canvas.clientWidth
        const displayH = canvas.clientHeight
        const scale = Math.min(displayW / CANVAS_W, displayH / CANVAS_H)
        const offsetX = (displayW - CANVAS_W * scale) / 2
        const offsetY = (displayH - CANVAS_H * scale) / 2
        return { displayW, displayH, scale, offsetX, offsetY }
    }, [])

    const draw = useCallback(() => {
        const canvas = canvasRef.current
        if (!canvas) return
        const ctx = canvas.getContext("2d")
        if (!ctx) return

        const layout = getLayout()
        if (!layout) return
        const { displayW, displayH, scale, offsetX, offsetY } = layout

        const dpr = window.devicePixelRatio || 1
        canvas.width = displayW * dpr
        canvas.height = displayH * dpr
        ctx.scale(dpr, dpr)

        ctx.clearRect(0, 0, displayW, displayH)

        ctx.fillStyle = "#1a1a2e"
        ctx.fillRect(offsetX, offsetY, CANVAS_W * scale, CANVAS_H * scale)

        // Grid lines every 16px
        ctx.strokeStyle = "rgba(255, 255, 255, 0.04)"
        ctx.lineWidth = 1
        for (let gx = 16; gx < CANVAS_W; gx += 16) {
            ctx.beginPath()
            ctx.moveTo(offsetX + gx * scale, offsetY)
            ctx.lineTo(offsetX + gx * scale, offsetY + CANVAS_H * scale)
            ctx.stroke()
        }
        for (let gy = 16; gy < CANVAS_H; gy += 16) {
            ctx.beginPath()
            ctx.moveTo(offsetX, offsetY + gy * scale)
            ctx.lineTo(offsetX + CANVAS_W * scale, offsetY + gy * scale)
            ctx.stroke()
        }

        const entries = Object.entries(rects)
        for (const [name, rect] of entries) {
            const isSelected = name === selectedName
            const x = offsetX + rect.x * scale
            const y = offsetY + toCanvasY(rect) * scale
            const w = rect.w * scale
            const h = rect.h * scale

            if (isSelected) {
                ctx.fillStyle = "rgba(245, 158, 11, 0.15)"
                ctx.fillRect(x, y, w, h)
                ctx.strokeStyle = "#f59e0b"
                ctx.lineWidth = 2
                ctx.setLineDash([])
            } else {
                ctx.strokeStyle = "rgba(100, 200, 255, 0.35)"
                ctx.lineWidth = 1
                ctx.setLineDash([3, 3])
            }
            ctx.strokeRect(x, y, w, h)

            ctx.setLineDash([])
            ctx.fillStyle = isSelected ? "#f59e0b" : "rgba(100, 200, 255, 0.5)"
            ctx.font = `${Math.max(9, 10 * scale)}px monospace`
            ctx.textAlign = "left"
            ctx.textBaseline = "bottom"
            ctx.fillText(name, x + 2, y - 2)
        }
    }, [rects, selectedName, getLayout])

    useEffect(() => {
        draw()
        window.addEventListener("resize", draw)
        return () => window.removeEventListener("resize", draw)
    }, [draw])

    const handleSelectedRectChange = useCallback((rect: OverlayRect) => {
        if (selectedName && onRectChange) {
            onRectChange(selectedName, rect)
        }
    }, [selectedName, onRectChange])

    const { handleMouseDown, handleMouseMove, handleMouseUp, handleMouseLeave } = useRectDrag(
        canvasRef,
        getSelectedRect,
        selectedName ? handleSelectedRectChange : undefined,
        draw,
    )

    const onMouseDown = useCallback((e: React.MouseEvent<HTMLCanvasElement>) => {
        handleMouseDown(e)
    }, [handleMouseDown])

    const onMouseUp = useCallback((e: React.MouseEvent<HTMLCanvasElement>) => {
        const wasDragging = handleMouseUp()
        if (wasDragging || !onSelectName) return
        const canvas = canvasRef.current
        if (!canvas) return

        const layout = getLayout()
        if (!layout) return
        const { scale, offsetX, offsetY } = layout

        const canvasRect = canvas.getBoundingClientRect()
        const clickX = e.clientX - canvasRect.left
        const clickY = e.clientY - canvasRect.top

        const entries = Object.entries(rects)
        for (let i = entries.length - 1; i >= 0; i--) {
            const [name, rect] = entries[i]
            const x = offsetX + rect.x * scale
            const y = offsetY + toCanvasY(rect) * scale
            const w = rect.w * scale
            const h = rect.h * scale
            if (clickX >= x && clickX <= x + w && clickY >= y && clickY <= y + h) {
                onSelectName(name)
                return
            }
        }
    }, [handleMouseUp, onSelectName, rects, getLayout])

    return (
        <canvas
            ref={canvasRef}
            className="w-full h-full"
            style={{ imageRendering: "pixelated" }}
            onMouseDown={onMouseDown}
            onMouseMove={handleMouseMove}
            onMouseUp={onMouseUp}
            onMouseLeave={handleMouseLeave}
        />
    )
}
