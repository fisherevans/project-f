import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useNamedRects, useSaveNamedRects } from "@/api/overlays"
import { Plus, Trash2, Save } from "lucide-react"
import YAML from "yaml"
import type { OverlayRect } from "@/types/overlays"

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
    const [rows, setRows] = useState<RectRow[]>([])
    const [dirty, setDirty] = useState(false)

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

    const updateRow = (i: number, field: keyof RectRow, value: string | number) => {
        setRows(prev => {
            const next = [...prev]
            next[i] = { ...next[i], [field]: value }
            return next
        })
        setDirty(true)
    }

    const addRow = () => {
        setRows(prev => [...prev, { name: "", x: 0, y: 0, w: 32, h: 32 }])
        setDirty(true)
    }

    const removeRow = (i: number) => {
        setRows(prev => prev.filter((_, idx) => idx !== i))
        setDirty(true)
    }

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

    if (isLoading) return <div className="text-xs text-muted-foreground">Loading rects...</div>

    return (
        <div className="space-y-2">
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
                Shared screen regions (240x160 canvas coords) referenced by name in overlay flows. Define once, reuse across flows.
            </p>
            {rows.length === 0 ? (
                <p className="text-xs text-muted-foreground">No named rects defined yet.</p>
            ) : (
                <div className="space-y-1">
                    <div className="grid grid-cols-[1fr_50px_50px_50px_50px_28px] gap-1 text-[10px] font-medium text-muted-foreground px-1">
                        <span>Name</span>
                        <span>X</span>
                        <span>Y</span>
                        <span>W</span>
                        <span>H</span>
                        <span />
                    </div>
                    {rows.map((row, i) => (
                        <div key={i} className="grid grid-cols-[1fr_50px_50px_50px_50px_28px] gap-1 items-center">
                            <Input
                                className="h-6 text-xs font-mono"
                                value={row.name}
                                onChange={e => updateRow(i, "name", e.target.value)}
                                placeholder="rect_name"
                            />
                            <Input
                                className="h-6 text-xs font-mono"
                                type="number"
                                value={row.x}
                                onChange={e => updateRow(i, "x", parseInt(e.target.value) || 0)}
                            />
                            <Input
                                className="h-6 text-xs font-mono"
                                type="number"
                                value={row.y}
                                onChange={e => updateRow(i, "y", parseInt(e.target.value) || 0)}
                            />
                            <Input
                                className="h-6 text-xs font-mono"
                                type="number"
                                value={row.w}
                                onChange={e => updateRow(i, "w", parseInt(e.target.value) || 0)}
                            />
                            <Input
                                className="h-6 text-xs font-mono"
                                type="number"
                                value={row.h}
                                onChange={e => updateRow(i, "h", parseInt(e.target.value) || 0)}
                            />
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 w-6 p-0 text-destructive"
                                onClick={() => removeRow(i)}
                            >
                                <Trash2 className="h-3 w-3" />
                            </Button>
                        </div>
                    ))}
                </div>
            )}
        </div>
    )
}
