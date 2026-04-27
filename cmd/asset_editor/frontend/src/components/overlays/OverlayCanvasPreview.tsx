import { useRef, useEffect, useCallback } from "react"
import type { OverlayTarget, OverlayRect } from "@/types/overlays"

const CANVAS_W = 240
const CANVAS_H = 160

interface Props {
    targets: OverlayTarget[]
    namedRects: Record<string, OverlayRect>
    activeIndex: number
    onSelectTarget?: (index: number) => void
}

function resolveRect(target: OverlayTarget, namedRects: Record<string, OverlayRect>): OverlayRect | null {
    if (target.region) return target.region
    if (target.rect && namedRects[target.rect]) return namedRects[target.rect]
    return null
}

export function OverlayCanvasPreview({ targets, namedRects, activeIndex, onSelectTarget }: Props) {
    const canvasRef = useRef<HTMLCanvasElement>(null)

    const draw = useCallback(() => {
        const canvas = canvasRef.current
        if (!canvas) return
        const ctx = canvas.getContext("2d")
        if (!ctx) return

        const dpr = window.devicePixelRatio || 1
        const displayW = canvas.clientWidth
        const displayH = canvas.clientHeight
        canvas.width = displayW * dpr
        canvas.height = displayH * dpr
        ctx.scale(dpr, dpr)

        const scale = Math.min(displayW / CANVAS_W, displayH / CANVAS_H)
        const offsetX = (displayW - CANVAS_W * scale) / 2
        const offsetY = (displayH - CANVAS_H * scale) / 2

        ctx.clearRect(0, 0, displayW, displayH)

        // Game canvas background
        ctx.fillStyle = "#1a1a2e"
        ctx.fillRect(offsetX, offsetY, CANVAS_W * scale, CANVAS_H * scale)

        // Draw all rects faintly
        targets.forEach((target, i) => {
            const rect = resolveRect(target, namedRects)
            if (!rect) return

            const x = offsetX + rect.x * scale
            const y = offsetY + rect.y * scale
            const w = rect.w * scale
            const h = rect.h * scale

            const isActive = i === activeIndex

            if (isActive) {
                // Highlighted cutout - brighter
                ctx.fillStyle = "rgba(255, 255, 255, 0.12)"
                ctx.fillRect(x, y, w, h)

                ctx.strokeStyle = "#f59e0b"
                ctx.lineWidth = 2
                ctx.setLineDash([])
                ctx.strokeRect(x, y, w, h)

                // Message placement indicator
                if (target.message) {
                    drawMessageIndicator(ctx, x, y, w, h, target.message.placement, scale)
                }

                // Badge placement indicator
                if (target.badge?.placement) {
                    drawBadgeIndicator(ctx, x, y, w, h, target.badge.placement, scale)
                }
            } else {
                ctx.strokeStyle = "rgba(255, 255, 255, 0.25)"
                ctx.lineWidth = 1
                ctx.setLineDash([4, 4])
                ctx.strokeRect(x, y, w, h)
            }

            // Step number
            ctx.setLineDash([])
            ctx.fillStyle = isActive ? "#f59e0b" : "rgba(255, 255, 255, 0.4)"
            ctx.font = `${Math.max(10, 11 * scale)}px monospace`
            ctx.textAlign = "left"
            ctx.textBaseline = "top"
            ctx.fillText(`${i + 1}`, x + 2 * scale, y + 2 * scale)
        })

        // Named rect labels (dimmed, for all named rects)
        ctx.setLineDash([2, 3])
        ctx.strokeStyle = "rgba(100, 200, 255, 0.15)"
        ctx.lineWidth = 1
        ctx.fillStyle = "rgba(100, 200, 255, 0.25)"
        ctx.font = `${Math.max(8, 9 * scale)}px monospace`
        ctx.textBaseline = "bottom"
        for (const [name, rect] of Object.entries(namedRects)) {
            const inUse = targets.some(t => t.rect === name)
            if (inUse) continue
            const x = offsetX + rect.x * scale
            const y = offsetY + rect.y * scale
            const w = rect.w * scale
            const h = rect.h * scale
            ctx.strokeRect(x, y, w, h)
            ctx.fillText(name, x, y - 2)
        }
        ctx.setLineDash([])
    }, [targets, namedRects, activeIndex])

    useEffect(() => {
        draw()
        window.addEventListener("resize", draw)
        return () => window.removeEventListener("resize", draw)
    }, [draw])

    const handleClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
        if (!onSelectTarget) return
        const canvas = canvasRef.current
        if (!canvas) return

        const rect = canvas.getBoundingClientRect()
        const clickX = e.clientX - rect.left
        const clickY = e.clientY - rect.top

        const displayW = canvas.clientWidth
        const displayH = canvas.clientHeight
        const scale = Math.min(displayW / CANVAS_W, displayH / CANVAS_H)
        const offsetX = (displayW - CANVAS_W * scale) / 2
        const offsetY = (displayH - CANVAS_H * scale) / 2

        for (let i = targets.length - 1; i >= 0; i--) {
            const r = resolveRect(targets[i], namedRects)
            if (!r) continue
            const x = offsetX + r.x * scale
            const y = offsetY + r.y * scale
            const w = r.w * scale
            const h = r.h * scale
            if (clickX >= x && clickX <= x + w && clickY >= y && clickY <= y + h) {
                onSelectTarget(i)
                return
            }
        }
    }

    return (
        <canvas
            ref={canvasRef}
            className="w-full h-full cursor-pointer"
            style={{ imageRendering: "pixelated" }}
            onClick={handleClick}
        />
    )
}

function drawMessageIndicator(
    ctx: CanvasRenderingContext2D,
    x: number, y: number, w: number, h: number,
    placement: string, scale: number,
) {
    ctx.fillStyle = "rgba(245, 158, 11, 0.2)"
    ctx.strokeStyle = "rgba(245, 158, 11, 0.5)"
    ctx.lineWidth = 1
    ctx.setLineDash([3, 3])

    const msgW = 60 * scale
    const msgH = 20 * scale
    const gap = 4 * scale

    let mx: number, my: number
    switch (placement) {
        case "left":
            mx = x - msgW - gap
            my = y + (h - msgH) / 2
            break
        case "right":
            mx = x + w + gap
            my = y + (h - msgH) / 2
            break
        case "top":
            mx = x + (w - msgW) / 2
            my = y - msgH - gap
            break
        case "bottom":
        default:
            mx = x + (w - msgW) / 2
            my = y + h + gap
            break
    }

    ctx.fillRect(mx, my, msgW, msgH)
    ctx.strokeRect(mx, my, msgW, msgH)

    ctx.setLineDash([])
    ctx.fillStyle = "rgba(245, 158, 11, 0.7)"
    ctx.font = `${Math.max(8, 9 * scale)}px monospace`
    ctx.textAlign = "center"
    ctx.textBaseline = "middle"
    ctx.fillText("msg", mx + msgW / 2, my + msgH / 2)
}

function drawBadgeIndicator(
    ctx: CanvasRenderingContext2D,
    x: number, y: number, w: number, h: number,
    placement: string, scale: number,
) {
    ctx.fillStyle = "rgba(139, 92, 246, 0.3)"
    const badgeW = 24 * scale
    const badgeH = 10 * scale
    const gap = 2 * scale

    let bx: number, by: number
    switch (placement) {
        case "top_right":
            bx = x + w - badgeW - gap
            by = y + gap
            break
        case "top_middle":
            bx = x + (w - badgeW) / 2
            by = y + gap
            break
        case "bottom_right":
            bx = x + w - badgeW - gap
            by = y + h - badgeH - gap
            break
        case "bottom_middle":
        default:
            bx = x + (w - badgeW) / 2
            by = y + h - badgeH - gap
            break
    }

    ctx.fillRect(bx, by, badgeW, badgeH)
    ctx.fillStyle = "rgba(139, 92, 246, 0.7)"
    ctx.font = `${Math.max(7, 8 * scale)}px monospace`
    ctx.textAlign = "center"
    ctx.textBaseline = "middle"
    ctx.fillText("badge", bx + badgeW / 2, by + badgeH / 2)
}
