import { useState, useEffect, useCallback } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useNamedRects, useSaveNamedRects, usePreviewHighlight, useDismissHighlight } from "@/api/overlays"
import { Plus, Trash2, Save, Play, X } from "lucide-react"
import { RectCanvasPreview } from "./RectCanvasPreview"
import YAML from "yaml"
import type { OverlayRect } from "@/types/overlays"

const CANVAS_W = 240
const CANVAS_H = 160

interface Props {
    onRectsChange?: (rects: Record<string, OverlayRect>) => void
}

interface RectRow {
    name: string
    x: number
    y: number
    w: number
    h: number
}

export function NamedRectEditor({ onRectsChange }: Props) {
    const { data, isLoading } = useNamedRects()
    const saveMutation = useSaveNamedRects()
    const previewMutation = usePreviewHighlight()
    const dismissMutation = useDismissHighlight()
    const [rows, setRows] = useState<RectRow[]>([])
    const [selected, setSelected] = useState<number | null>(null)
    const [dirty, setDirty] = useState(false)
    const [previewError, setPreviewError] = useState<string | null>(null)
    const [previewDuration, setPreviewDuration] = useState(() => {
        const stored = localStorage.getItem("overlay-preview-duration")
        return stored ? parseFloat(stored) : 5
    })

    useEffect(() => {
        if (data) {
            const sorted = Object.entries(data.rects)
                .sort(([a], [b]) => a.localeCompare(b))
                .map(([name, r]) => ({ name, ...r }))
            setRows(sorted)
            setDirty(false)
        }
    }, [data])

    useEffect(() => {
        if (onRectsChange) {
            const rects: Record<string, OverlayRect> = {}
            for (const row of rows) {
                if (row.name.trim()) {
                    rects[row.name.trim()] = { x: row.x, y: row.y, w: row.w, h: row.h }
                }
            }
            onRectsChange(rects)
        }
    }, [rows, onRectsChange])

    const updateRow = useCallback((i: number, field: keyof RectRow, value: string | number) => {
        setRows(prev => {
            const next = [...prev]
            next[i] = { ...next[i], [field]: value }
            return next
        })
        setDirty(true)
    }, [])

    const addRow = useCallback(() => {
        setRows(prev => [...prev, { name: "", x: 0, y: 0, w: 32, h: 32 }])
        setSelected(rows.length)
        setDirty(true)
    }, [rows.length])

    const removeRow = useCallback((i: number) => {
        setRows(prev => prev.filter((_, idx) => idx !== i))
        setSelected(prev => {
            if (prev === i) return null
            if (prev !== null && prev > i) return prev - 1
            return prev
        })
        setDirty(true)
    }, [])

    const handleSave = () => {
        const obj: Record<string, OverlayRect> = {}
        for (const row of rows) {
            const name = row.name.trim()
            if (!name) continue
            obj[name] = { x: row.x, y: row.y, w: row.w, h: row.h }
        }
        const yaml = YAML.stringify(obj, { indent: 2 })
        saveMutation.mutate(yaml, { onSuccess: () => setDirty(false) })
    }

    const handleTest = (row: RectRow) => {
        setPreviewError(null)
        previewMutation.mutate(
            {
                flow: { targets: [{ region: { x: row.x, y: row.y, w: row.w, h: row.h } }] },
                duration: previewDuration,
            },
            { onError: err => setPreviewError(err instanceof Error ? err.message : "Preview failed") },
        )
    }

    const handleDismiss = () => {
        setPreviewError(null)
        dismissMutation.mutate()
    }

    const rectsForPreview: Record<string, OverlayRect> = {}
    for (const row of rows) {
        if (row.name.trim()) rectsForPreview[row.name.trim()] = { x: row.x, y: row.y, w: row.w, h: row.h }
    }

    if (isLoading) return <div className="text-xs text-muted-foreground">Loading rects...</div>

    const sel = selected !== null ? rows[selected] : null

    return (
        <div className="space-y-3">
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold">Named Rects</h3>
                <div className="flex gap-1">
                    <Button variant="ghost" size="sm" className="h-6 px-2 text-xs" onClick={addRow}>
                        <Plus className="h-3 w-3 mr-1" /> Add
                    </Button>
                    {dirty && (
                        <Button variant="outline" size="sm" className="h-6 px-2 text-xs" onClick={handleSave} disabled={saveMutation.isPending}>
                            <Save className="h-3 w-3 mr-1" /> Save
                        </Button>
                    )}
                </div>
            </div>
            <p className="text-[11px] text-muted-foreground">
                Shared screen regions (240x160 canvas coords, y=0 at bottom) referenced by name in overlay flows.
            </p>

            {previewError && (
                <div className="text-[11px] text-destructive bg-destructive/10 border border-destructive/20 rounded px-2 py-1">
                    {previewError}
                </div>
            )}

            {/* Side-by-side: list left, detail right */}
            <div className="grid grid-cols-[180px_1fr] gap-3">
                {/* Left: rect list */}
                <div className="space-y-0.5 border border-border rounded-md p-1 max-h-[420px] overflow-auto">
                    {rows.length === 0 && (
                        <p className="text-xs text-muted-foreground p-2">No named rects yet.</p>
                    )}
                    {rows.map((row, i) => (
                        <div
                            key={i}
                            className={`flex flex-col gap-0.5 rounded px-2 py-1 cursor-pointer transition-colors ${
                                i === selected ? "bg-primary text-primary-foreground" : "hover:bg-accent"
                            }`}
                            onClick={() => setSelected(i === selected ? null : i)}
                        >
                            <span className="text-xs font-mono truncate">{row.name || "(unnamed)"}</span>
                            <span className="text-[10px] opacity-70 font-mono">
                                {row.x},{row.y} {row.w}x{row.h}
                            </span>
                        </div>
                    ))}
                </div>

                {/* Right: detail panel */}
                <div className="space-y-3">
                    {sel === null || selected === null ? (
                        <div className="border border-dashed border-border rounded-md p-6 text-center text-xs text-muted-foreground">
                            Select a rect on the left to edit, or click Add to create a new one.
                        </div>
                    ) : (
                        <>
                            {/* Compact canvas preview */}
                            <div className="bg-canvas rounded-md overflow-hidden max-w-[320px]" style={{ aspectRatio: "240/160" }}>
                                <RectCanvasPreview
                                    rects={rectsForPreview}
                                    selectedName={sel.name.trim() || null}
                                    onSelectName={name => {
                                        const idx = rows.findIndex(r => r.name.trim() === name)
                                        setSelected(idx >= 0 ? idx : null)
                                    }}
                                    onRectChange={(name, rect) => {
                                        const idx = rows.findIndex(r => r.name.trim() === name)
                                        if (idx < 0) return
                                        setRows(prev => {
                                            const next = [...prev]
                                            next[idx] = { ...next[idx], ...rect }
                                            return next
                                        })
                                        setDirty(true)
                                    }}
                                />
                            </div>

                            {/* Name */}
                            <div className="space-y-1">
                                <label className="text-[11px] font-medium text-muted-foreground">Name</label>
                                <Input
                                    className="h-7 text-xs font-mono"
                                    value={sel.name}
                                    onChange={e => updateRow(selected, "name", e.target.value)}
                                    placeholder="rect_name"
                                />
                            </div>

                            {/* X/Y/W/H */}
                            <div className="grid grid-cols-4 gap-2">
                                <RectField label="X" value={sel.x} onChange={v => updateRow(selected, "x", v)} />
                                <RectField label="Y" value={sel.y} onChange={v => updateRow(selected, "y", v)} />
                                <RectField label="W" value={sel.w} onChange={v => updateRow(selected, "w", v)} />
                                <RectField label="H" value={sel.h} onChange={v => updateRow(selected, "h", v)} />
                            </div>

                            {/* Quick-set helpers */}
                            <div className="flex flex-wrap gap-1">
                                <QuickBtn label="Origin" onClick={() => {
                                    updateRow(selected, "x", 0)
                                    updateRow(selected, "y", 0)
                                }} />
                                <QuickBtn label="Full" onClick={() => {
                                    updateRow(selected, "x", 0)
                                    updateRow(selected, "y", 0)
                                    updateRow(selected, "w", CANVAS_W)
                                    updateRow(selected, "h", CANVAS_H)
                                }} />
                                <QuickBtn label="X=0" onClick={() => updateRow(selected, "x", 0)} />
                                <QuickBtn label="Y=0" onClick={() => updateRow(selected, "y", 0)} />
                                <QuickBtn label={`W=${CANVAS_W}`} onClick={() => updateRow(selected, "w", CANVAS_W)} />
                                <QuickBtn label={`H=${CANVAS_H}`} onClick={() => updateRow(selected, "h", CANVAS_H)} />
                                <QuickBtn label="Full W" title="X=0, W=240" onClick={() => {
                                    updateRow(selected, "x", 0)
                                    updateRow(selected, "w", CANVAS_W)
                                }} />
                                <QuickBtn label="Full H" title="Y=0, H=160" onClick={() => {
                                    updateRow(selected, "y", 0)
                                    updateRow(selected, "h", CANVAS_H)
                                }} />
                                <QuickBtn label="Center" title="Center on screen" onClick={() => {
                                    updateRow(selected, "x", Math.floor((CANVAS_W - sel.w) / 2))
                                    updateRow(selected, "y", Math.floor((CANVAS_H - sel.h) / 2))
                                }} />
                            </div>

                            {/* Action row: test in game + remove */}
                            <div className="flex items-center justify-between pt-1 border-t border-border">
                                <div className="flex items-center gap-1">
                                    <Input
                                        className="h-7 w-14 text-xs font-mono text-center"
                                        type="number"
                                        min={1}
                                        max={30}
                                        step={1}
                                        value={previewDuration}
                                        onChange={e => {
                                            const v = parseFloat(e.target.value) || 5
                                            setPreviewDuration(v)
                                            localStorage.setItem("overlay-preview-duration", String(v))
                                        }}
                                        title="Seconds to show"
                                    />
                                    <span className="text-[10px] text-muted-foreground mr-1">s</span>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="h-7 text-xs"
                                        onClick={() => handleTest(sel)}
                                        disabled={previewMutation.isPending}
                                        title="Show this rect as a highlight in the running game"
                                    >
                                        <Play className="h-3 w-3 mr-1" /> Test
                                    </Button>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="h-7 text-xs"
                                        onClick={handleDismiss}
                                        disabled={dismissMutation.isPending}
                                    >
                                        <X className="h-3 w-3 mr-1" /> Dismiss
                                    </Button>
                                </div>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-7 px-2 text-xs text-destructive"
                                    onClick={() => removeRow(selected)}
                                >
                                    <Trash2 className="h-3 w-3 mr-1" /> Remove
                                </Button>
                            </div>
                        </>
                    )}
                </div>
            </div>
        </div>
    )
}

function RectField({ label, value, onChange }: { label: string; value: number; onChange: (v: number) => void }) {
    return (
        <div className="space-y-1">
            <label className="text-[10px] font-medium text-muted-foreground uppercase">{label}</label>
            <Input
                className="h-7 text-xs font-mono text-center"
                type="number"
                value={value}
                onChange={e => onChange(parseInt(e.target.value) || 0)}
            />
        </div>
    )
}

function QuickBtn({ label, title, onClick }: { label: string; title?: string; onClick: () => void }) {
    return (
        <button
            className="px-1.5 py-0.5 rounded text-[10px] font-mono bg-accent-blue-tint text-accent-blue hover:bg-accent-blue-edge/30 transition-colors"
            title={title ?? label}
            onClick={onClick}
        >
            {label}
        </button>
    )
}
