import { useRef, useCallback } from "react"
import type { OverlayRect } from "@/types/overlays"

const CANVAS_W = 240
const CANVAS_H = 160
const EDGE_THRESHOLD = 6

type DragMode = "move" | "resize-n" | "resize-s" | "resize-e" | "resize-w" | "resize-ne" | "resize-nw" | "resize-se" | "resize-sw"

interface DragState {
    mode: DragMode
    startGameX: number
    startGameY: number
    origRect: OverlayRect
}

function screenToGame(
    sx: number, sy: number,
    offsetX: number, offsetY: number, scale: number,
): { gx: number; gy: number } {
    const gx = (sx - offsetX) / scale
    const gy = CANVAS_H - (sy - offsetY) / scale
    return { gx, gy }
}

function getLayout(canvas: HTMLCanvasElement) {
    const displayW = canvas.clientWidth
    const displayH = canvas.clientHeight
    const scale = Math.min(displayW / CANVAS_W, displayH / CANVAS_H)
    const offsetX = (displayW - CANVAS_W * scale) / 2
    const offsetY = (displayH - CANVAS_H * scale) / 2
    return { scale, offsetX, offsetY }
}

function detectMode(
    gx: number, gy: number,
    rect: OverlayRect, scale: number,
): DragMode | null {
    const threshold = EDGE_THRESHOLD / scale
    const inX = gx >= rect.x - threshold && gx <= rect.x + rect.w + threshold
    const inY = gy >= rect.y - threshold && gy <= rect.y + rect.h + threshold
    if (!inX || !inY) return null

    const nearLeft = Math.abs(gx - rect.x) < threshold
    const nearRight = Math.abs(gx - (rect.x + rect.w)) < threshold
    const nearBottom = Math.abs(gy - rect.y) < threshold
    const nearTop = Math.abs(gy - (rect.y + rect.h)) < threshold

    if (nearLeft && nearTop) return "resize-nw"
    if (nearRight && nearTop) return "resize-ne"
    if (nearLeft && nearBottom) return "resize-sw"
    if (nearRight && nearBottom) return "resize-se"
    if (nearTop) return "resize-n"
    if (nearBottom) return "resize-s"
    if (nearLeft) return "resize-w"
    if (nearRight) return "resize-e"

    if (gx >= rect.x && gx <= rect.x + rect.w && gy >= rect.y && gy <= rect.y + rect.h) {
        return "move"
    }
    return null
}

function applyDrag(orig: OverlayRect, mode: DragMode, dx: number, dy: number): OverlayRect {
    let { x, y, w, h } = orig

    switch (mode) {
        case "move":
            x += dx
            y += dy
            break
        case "resize-n":
            h += dy
            break
        case "resize-s":
            y += dy
            h -= dy
            break
        case "resize-e":
            w += dx
            break
        case "resize-w":
            x += dx
            w -= dx
            break
        case "resize-ne":
            w += dx
            h += dy
            break
        case "resize-nw":
            x += dx
            w -= dx
            h += dy
            break
        case "resize-se":
            w += dx
            y += dy
            h -= dy
            break
        case "resize-sw":
            x += dx
            w -= dx
            y += dy
            h -= dy
            break
    }

    if (w < 1) { w = 1 }
    if (h < 1) { h = 1 }
    w = Math.round(w)
    h = Math.round(h)
    x = Math.max(0, Math.min(CANVAS_W - w, Math.round(x)))
    y = Math.max(0, Math.min(CANVAS_H - h, Math.round(y)))

    return { x, y, w, h }
}

export function getCursorForMode(mode: DragMode | null): string {
    switch (mode) {
        case "move": return "grab"
        case "resize-n": case "resize-s": return "ns-resize"
        case "resize-e": case "resize-w": return "ew-resize"
        case "resize-ne": case "resize-sw": return "nesw-resize"
        case "resize-nw": case "resize-se": return "nwse-resize"
        default: return "default"
    }
}

export function useRectDrag(
    canvasRef: React.RefObject<HTMLCanvasElement | null>,
    getRect: () => OverlayRect | null,
    onRectChange: ((rect: OverlayRect) => void) | undefined,
    redraw: () => void,
) {
    const dragRef = useRef<DragState | null>(null)
    const didDragRef = useRef(false)

    const handleMouseDown = useCallback((e: React.MouseEvent<HTMLCanvasElement>) => {
        if (!onRectChange) return
        const canvas = canvasRef.current
        if (!canvas) return
        const rect = getRect()
        if (!rect) return

        const canvasBounds = canvas.getBoundingClientRect()
        const { scale, offsetX, offsetY } = getLayout(canvas)
        const { gx, gy } = screenToGame(e.clientX - canvasBounds.left, e.clientY - canvasBounds.top, offsetX, offsetY, scale)

        const mode = detectMode(gx, gy, rect, scale)
        if (!mode) return

        e.preventDefault()
        didDragRef.current = false
        dragRef.current = { mode, startGameX: gx, startGameY: gy, origRect: { ...rect } }

        if (mode === "move") {
            canvas.style.cursor = "grabbing"
        }
    }, [canvasRef, getRect, onRectChange])

    const handleMouseMove = useCallback((e: React.MouseEvent<HTMLCanvasElement>) => {
        const canvas = canvasRef.current
        if (!canvas) return
        const canvasBounds = canvas.getBoundingClientRect()
        const { scale, offsetX, offsetY } = getLayout(canvas)
        const { gx, gy } = screenToGame(e.clientX - canvasBounds.left, e.clientY - canvasBounds.top, offsetX, offsetY, scale)

        if (dragRef.current && onRectChange) {
            const { mode, startGameX, startGameY, origRect } = dragRef.current
            const dx = gx - startGameX
            const dy = gy - startGameY
            const newRect = applyDrag(origRect, mode, dx, dy)
            didDragRef.current = true
            onRectChange(newRect)
            return
        }

        const rect = getRect()
        if (!rect || !onRectChange) {
            canvas.style.cursor = "default"
            return
        }
        const mode = detectMode(gx, gy, rect, scale)
        canvas.style.cursor = getCursorForMode(mode)
    }, [canvasRef, getRect, onRectChange, redraw])

    const handleMouseUp = useCallback((): boolean => {
        if (dragRef.current) {
            const canvas = canvasRef.current
            if (canvas) canvas.style.cursor = "default"
            dragRef.current = null
            return didDragRef.current
        }
        return false
    }, [canvasRef])

    const handleMouseLeave = useCallback(() => {
        if (dragRef.current) {
            dragRef.current = null
        }
    }, [])

    return { handleMouseDown, handleMouseMove, handleMouseUp, handleMouseLeave }
}
