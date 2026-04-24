import { useCallback, useEffect, useRef } from "react"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import type { SpriteFrame, FrameMode, FrameSide } from "@/types/sprites"

const CARDINAL_SIDES: FrameSide[] = ["top", "right", "bottom", "left"]

interface FrameEditorProps {
    frame: SpriteFrame
    onChange: (frame: SpriteFrame) => void
    imageSrc: string
    imageWidth: number
    imageHeight: number
}

export function FrameEditor({ frame, onChange, imageSrc, imageWidth, imageHeight }: FrameEditorProps) {
    const canvasRef = useRef<HTMLCanvasElement>(null)
    const imageRef = useRef<HTMLImageElement | null>(null)

    const updateDefaults = (patch: Partial<SpriteFrame["defaults"]>) => {
        onChange({ ...frame, defaults: { ...frame.defaults, ...patch } })
    }

    const updateSideValue = (
        field: "cutMargin" | "padding",
        side: FrameSide,
        value: number,
    ) => {
        const current = frame[field] ?? {}
        onChange({ ...frame, [field]: { ...current, [side]: value } })
    }

    const updateSideMode = (side: FrameSide, mode: FrameMode) => {
        const current = frame.frameModes ?? {}
        onChange({ ...frame, frameModes: { ...current, [side]: mode } })
    }

    const getSideValue = (field: "cutMargin" | "padding", side: FrameSide): number => {
        return frame[field]?.[side] ?? frame.defaults[field === "cutMargin" ? "cutMargin" : "padding"]
    }

    const getSideMode = (side: FrameSide): FrameMode => {
        return frame.frameModes?.[side] ?? frame.defaults.frameMode
    }

    const previewScale = Math.max(1, Math.floor(200 / Math.max(imageWidth, imageHeight)))

    const drawPreview = useCallback(() => {
        const canvas = canvasRef.current
        const img = imageRef.current
        if (!canvas || !img) return
        const ctx = canvas.getContext("2d")
        if (!ctx) return

        const w = imageWidth * previewScale
        const h = imageHeight * previewScale
        canvas.width = w
        canvas.height = h

        ctx.imageSmoothingEnabled = false
        ctx.clearRect(0, 0, w, h)
        ctx.drawImage(img, 0, 0, w, h)

        const top = getSideValue("cutMargin", "top") * previewScale
        const right = getSideValue("cutMargin", "right") * previewScale
        const bottom = getSideValue("cutMargin", "bottom") * previewScale
        const left = getSideValue("cutMargin", "left") * previewScale

        ctx.setLineDash([4, 4])
        ctx.lineWidth = 1

        ctx.strokeStyle = "rgba(239, 68, 68, 0.8)"
        ctx.beginPath()
        ctx.moveTo(0, top + 0.5)
        ctx.lineTo(w, top + 0.5)
        ctx.stroke()

        ctx.beginPath()
        ctx.moveTo(0, h - bottom + 0.5)
        ctx.lineTo(w, h - bottom + 0.5)
        ctx.stroke()

        ctx.strokeStyle = "rgba(59, 130, 246, 0.8)"
        ctx.beginPath()
        ctx.moveTo(left + 0.5, 0)
        ctx.lineTo(left + 0.5, h)
        ctx.stroke()

        ctx.beginPath()
        ctx.moveTo(w - right + 0.5, 0)
        ctx.lineTo(w - right + 0.5, h)
        ctx.stroke()

        ctx.setLineDash([])

        const regions = [
            { x: 0, y: 0, w: left, h: top, color: "rgba(168, 85, 247, 0.15)", label: "TL" },
            { x: left, y: 0, w: w - left - right, h: top, color: "rgba(239, 68, 68, 0.10)", label: "T" },
            { x: w - right, y: 0, w: right, h: top, color: "rgba(168, 85, 247, 0.15)", label: "TR" },
            { x: 0, y: top, w: left, h: h - top - bottom, color: "rgba(59, 130, 246, 0.10)", label: "L" },
            { x: left, y: top, w: w - left - right, h: h - top - bottom, color: "rgba(34, 197, 94, 0.10)", label: "M" },
            { x: w - right, y: top, w: right, h: h - top - bottom, color: "rgba(59, 130, 246, 0.10)", label: "R" },
            { x: 0, y: h - bottom, w: left, h: bottom, color: "rgba(168, 85, 247, 0.15)", label: "BL" },
            { x: left, y: h - bottom, w: w - left - right, h: bottom, color: "rgba(239, 68, 68, 0.10)", label: "B" },
            { x: w - right, y: h - bottom, w: right, h: bottom, color: "rgba(168, 85, 247, 0.15)", label: "BR" },
        ]

        for (const r of regions) {
            ctx.fillStyle = r.color
            ctx.fillRect(r.x, r.y, r.w, r.h)
        }
    }, [imageWidth, imageHeight, previewScale, frame])

    useEffect(() => {
        const img = new Image()
        img.onload = () => {
            imageRef.current = img
            drawPreview()
        }
        img.src = imageSrc
    }, [imageSrc, drawPreview])

    useEffect(() => {
        drawPreview()
    }, [drawPreview])

    return (
        <Card>
            <CardHeader className="py-3 px-4">
                <CardTitle className="text-sm">9-Slice Frame</CardTitle>
            </CardHeader>
            <CardContent className="px-4 pb-4 space-y-4">
                <div className="flex gap-4">
                    <div className="rounded border border-border bg-zinc-900 overflow-hidden inline-block">
                        <canvas
                            ref={canvasRef}
                            style={{ imageRendering: "pixelated" }}
                        />
                    </div>
                    <div className="space-y-3 flex-1">
                        <div>
                            <Label className="text-xs font-medium">Defaults</Label>
                            <div className="grid grid-cols-3 gap-2 mt-1">
                                <div>
                                    <Label className="text-xs text-muted-foreground">Cut Margin</Label>
                                    <Input
                                        type="number"
                                        className="h-7"
                                        min={0}
                                        value={frame.defaults.cutMargin}
                                        onChange={(e) => updateDefaults({ cutMargin: parseInt(e.target.value) || 0 })}
                                    />
                                </div>
                                <div>
                                    <Label className="text-xs text-muted-foreground">Padding</Label>
                                    <Input
                                        type="number"
                                        className="h-7"
                                        min={0}
                                        value={frame.defaults.padding}
                                        onChange={(e) => updateDefaults({ padding: parseInt(e.target.value) || 0 })}
                                    />
                                </div>
                                <div>
                                    <Label className="text-xs text-muted-foreground">Mode</Label>
                                    <Select
                                        value={frame.defaults.frameMode || "stretch"}
                                        onValueChange={(v) => updateDefaults({ frameMode: v as FrameMode })}
                                    >
                                        <SelectTrigger className="h-7">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="stretch">Stretch</SelectItem>
                                            <SelectItem value="repeat">Repeat</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <Separator />

                <div>
                    <Label className="text-xs font-medium mb-2 block">Per-Side Overrides</Label>
                    <div className="grid grid-cols-4 gap-2">
                        {CARDINAL_SIDES.map((side) => (
                            <div key={side} className="space-y-1">
                                <Label className="text-xs text-muted-foreground capitalize">{side}</Label>
                                <div className="space-y-1">
                                    <Input
                                        type="number"
                                        className="h-7"
                                        min={0}
                                        placeholder="margin"
                                        value={getSideValue("cutMargin", side)}
                                        onChange={(e) => updateSideValue("cutMargin", side, parseInt(e.target.value) || 0)}
                                    />
                                    <Input
                                        type="number"
                                        className="h-7"
                                        min={0}
                                        placeholder="padding"
                                        value={getSideValue("padding", side)}
                                        onChange={(e) => updateSideValue("padding", side, parseInt(e.target.value) || 0)}
                                    />
                                    <Select
                                        value={getSideMode(side)}
                                        onValueChange={(v) => updateSideMode(side, v as FrameMode)}
                                    >
                                        <SelectTrigger className="h-7">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="stretch">Stretch</SelectItem>
                                            <SelectItem value="repeat">Repeat</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}
