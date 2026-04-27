import { useState, useEffect, useCallback } from "react"
import { useParams, useNavigate, useSearchParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useOverlay, useSaveOverlay, useDeleteOverlay, useNamedRects, usePreviewHighlight, useDismissHighlight } from "@/api/overlays"
import { usePageTitle } from "@/hooks/usePageTitle"
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges"
import { OverlayCanvasPreview } from "@/components/overlays/OverlayCanvasPreview"
import { ArrowLeft, Save, Trash2, Plus, GripVertical, ChevronUp, ChevronDown, Play, X, Eye } from "lucide-react"
import YAML from "yaml"
import type { OverlayFlow, OverlayTarget, OverlayRect } from "@/types/overlays"

const MESSAGE_PLACEMENTS = ["top", "bottom", "left", "right"]
const BADGE_PLACEMENTS = ["top_right", "bottom_right", "top_middle", "bottom_middle"]

function emptyTarget(): OverlayTarget {
    return {
        region: { x: 60, y: 40, w: 60, h: 40 },
        message: { text: "", placement: "bottom" },
    }
}

function emptyFlow(): OverlayFlow {
    return {
        description: "",
        targets: [emptyTarget()],
    }
}

export function OverlayEditor() {
    const params = useParams()
    const name = params["*"] ?? ""
    const navigate = useNavigate()
    const [searchParams] = useSearchParams()
    const isNew = searchParams.get("new") === "1"

    const { data, isLoading, error } = useOverlay(name)
    const { data: rectsData } = useNamedRects()
    const saveMutation = useSaveOverlay()
    const deleteMutation = useDeleteOverlay()
    const previewMutation = usePreviewHighlight()
    const dismissMutation = useDismissHighlight()

    const [flow, setFlow] = useState<OverlayFlow>(emptyFlow())
    const [activeTarget, setActiveTarget] = useState(0)
    const [dirty, setDirty] = useState(false)
    const [previewError, setPreviewError] = useState<string | null>(null)

    usePageTitle(name ? `${name} - Overlays` : "Overlays")
    useUnsavedChanges(dirty)

    const namedRects = rectsData?.rects ?? {}

    useEffect(() => {
        if (isNew) {
            setFlow(emptyFlow())
            setDirty(true)
        } else if (data) {
            setFlow(data.flow)
            setDirty(false)
        }
    }, [data, isNew])

    const updateFlow = useCallback((updater: (prev: OverlayFlow) => OverlayFlow) => {
        setFlow(prev => {
            const next = updater(prev)
            setDirty(true)
            return next
        })
    }, [])

    const updateTarget = useCallback((index: number, updater: (prev: OverlayTarget) => OverlayTarget) => {
        updateFlow(f => ({
            ...f,
            targets: f.targets.map((t, i) => i === index ? updater(t) : t),
        }))
    }, [updateFlow])

    const handleSave = useCallback(() => {
        const yamlContent = YAML.stringify(flowToYaml(flow), { indent: 2 })
        saveMutation.mutate({ name, content: yamlContent }, {
            onSuccess: () => setDirty(false),
        })
    }, [flow, name, saveMutation])

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if ((e.metaKey || e.ctrlKey) && e.key === "s") {
                e.preventDefault()
                handleSave()
            }
        }
        window.addEventListener("keydown", handleKeyDown)
        return () => window.removeEventListener("keydown", handleKeyDown)
    }, [handleSave])

    const handleDelete = () => {
        if (!confirm(`Delete overlay flow "${name}"?`)) return
        deleteMutation.mutate(name, { onSuccess: () => navigate("/overlays") })
    }

    const handlePreview = () => {
        setPreviewError(null)
        previewMutation.mutate(
            { flow, rects: namedRects },
            { onError: (err) => setPreviewError(err instanceof Error ? err.message : "Preview failed") },
        )
    }

    const handlePreviewStep = (index: number) => {
        setPreviewError(null)
        const singleFlow: OverlayFlow = { targets: [flow.targets[index]] }
        previewMutation.mutate(
            { flow: singleFlow, rects: namedRects },
            { onError: (err) => setPreviewError(err instanceof Error ? err.message : "Preview failed") },
        )
    }

    const handleDismiss = () => {
        setPreviewError(null)
        dismissMutation.mutate()
    }

    const addTarget = () => {
        updateFlow(f => ({ ...f, targets: [...f.targets, emptyTarget()] }))
        setActiveTarget(flow.targets.length)
    }

    const removeTarget = (index: number) => {
        updateFlow(f => ({ ...f, targets: f.targets.filter((_, i) => i !== index) }))
        if (activeTarget >= flow.targets.length - 1) {
            setActiveTarget(Math.max(0, flow.targets.length - 2))
        }
    }

    const moveTarget = (index: number, direction: -1 | 1) => {
        const newIndex = index + direction
        if (newIndex < 0 || newIndex >= flow.targets.length) return
        updateFlow(f => {
            const targets = [...f.targets]
            const temp = targets[index]
            targets[index] = targets[newIndex]
            targets[newIndex] = temp
            return { ...f, targets }
        })
        setActiveTarget(newIndex)
    }

    if (isLoading && !isNew) return <div className="p-4 text-sm text-muted-foreground">Loading...</div>
    if (error && !isNew) return <div className="p-4 text-sm text-destructive">Error: {error.message}</div>

    const current = flow.targets[activeTarget]

    return (
        <div className="flex h-full flex-col">
            {/* Header */}
            <div className="flex items-center gap-3 border-b border-border px-4 py-2 shrink-0">
                <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={() => navigate("/overlays")}>
                    <ArrowLeft className="h-4 w-4" />
                </Button>
                <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium font-mono">{name}</div>
                    <div className="text-xs text-muted-foreground">
                        {flow.targets.length} target{flow.targets.length !== 1 ? "s" : ""}
                        {dirty && <span className="ml-2 text-accent-amber">unsaved</span>}
                    </div>
                </div>
                <div className="flex gap-1">
                    <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handlePreview} disabled={previewMutation.isPending}>
                        <Play className="h-3 w-3 mr-1" /> Preview
                    </Button>
                    <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleDismiss} disabled={dismissMutation.isPending}>
                        <X className="h-3 w-3 mr-1" /> Dismiss
                    </Button>
                    <Button variant="outline" size="sm" className="h-7 text-xs text-destructive" onClick={handleDelete}>
                        <Trash2 className="h-3 w-3 mr-1" /> Delete
                    </Button>
                    <Button size="sm" className="h-7 text-xs" onClick={handleSave} disabled={saveMutation.isPending || !dirty}>
                        <Save className="h-3 w-3 mr-1" /> Save
                    </Button>
                </div>
            </div>

            {previewError && (
                <div className="px-4 py-1 text-xs text-destructive bg-destructive/10 border-b border-border">
                    {previewError}
                </div>
            )}

            {/* Body */}
            <div className="flex flex-1 overflow-hidden">
                {/* Left panel - target list + canvas */}
                <div className="w-72 shrink-0 border-r border-border flex flex-col overflow-hidden">
                    {/* Description */}
                    <div className="p-3 border-b border-border space-y-2">
                        <label className="text-[11px] font-medium text-muted-foreground">Description</label>
                        <Input
                            className="h-7 text-xs"
                            value={flow.description ?? ""}
                            onChange={e => updateFlow(f => ({ ...f, description: e.target.value || undefined }))}
                            placeholder="What this flow explains"
                        />
                    </div>

                    {/* Canvas preview */}
                    <div className="p-3 border-b border-border">
                        <div className="bg-canvas rounded-md overflow-hidden" style={{ aspectRatio: "240/160" }}>
                            <OverlayCanvasPreview
                                targets={flow.targets}
                                namedRects={namedRects}
                                activeIndex={activeTarget}
                                onSelectTarget={setActiveTarget}
                            />
                        </div>
                    </div>

                    {/* Target list */}
                    <div className="flex-1 overflow-auto p-2 space-y-1">
                        <div className="flex items-center justify-between px-1 mb-1">
                            <span className="text-[11px] font-medium text-muted-foreground">Targets</span>
                            <Button variant="ghost" size="sm" className="h-5 px-1 text-[10px]" onClick={addTarget}>
                                <Plus className="h-3 w-3" />
                            </Button>
                        </div>
                        {flow.targets.map((target, i) => (
                            <div
                                key={i}
                                className={`flex items-center gap-1 rounded px-2 py-1.5 text-xs cursor-pointer transition-colors ${
                                    i === activeTarget
                                        ? "bg-primary text-primary-foreground"
                                        : "hover:bg-accent"
                                }`}
                                onClick={() => setActiveTarget(i)}
                            >
                                <GripVertical className="h-3 w-3 shrink-0 opacity-40" />
                                <div className="flex-1 min-w-0 truncate font-mono">
                                    {i + 1}. {target.rect || (target.region ? `(${target.region.x},${target.region.y})` : "no rect")}
                                </div>
                                <div className="flex gap-0.5 shrink-0" onClick={e => e.stopPropagation()}>
                                    <button
                                        className="p-0.5 rounded hover:bg-background/20"
                                        title="Preview this step"
                                        onClick={() => handlePreviewStep(i)}
                                    >
                                        <Eye className="h-3 w-3" />
                                    </button>
                                    <button
                                        className="p-0.5 rounded hover:bg-background/20"
                                        onClick={() => moveTarget(i, -1)}
                                        disabled={i === 0}
                                    >
                                        <ChevronUp className="h-3 w-3" />
                                    </button>
                                    <button
                                        className="p-0.5 rounded hover:bg-background/20"
                                        onClick={() => moveTarget(i, 1)}
                                        disabled={i === flow.targets.length - 1}
                                    >
                                        <ChevronDown className="h-3 w-3" />
                                    </button>
                                    {flow.targets.length > 1 && (
                                        <button
                                            className="p-0.5 rounded hover:bg-background/20 text-destructive"
                                            onClick={() => removeTarget(i)}
                                        >
                                            <Trash2 className="h-3 w-3" />
                                        </button>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>
                </div>

                {/* Right panel - target detail */}
                <div className="flex-1 overflow-auto p-4">
                    {current ? (
                        <TargetEditor
                            target={current}
                            namedRects={namedRects}
                            onChange={updated => updateTarget(activeTarget, () => updated)}
                        />
                    ) : (
                        <p className="text-sm text-muted-foreground">Select a target to edit</p>
                    )}
                </div>
            </div>
        </div>
    )
}

interface TargetEditorProps {
    target: OverlayTarget
    namedRects: Record<string, OverlayRect>
    onChange: (target: OverlayTarget) => void
}

function TargetEditor({ target, namedRects, onChange }: TargetEditorProps) {
    const useNamedRect = !!target.rect || (!target.rect && !target.region)
    const isNamed = useNamedRect && !target.region

    const setRectMode = (named: boolean) => {
        if (named) {
            const firstRect = Object.keys(namedRects)[0]
            onChange({ ...target, rect: firstRect ?? "", region: undefined })
        } else {
            onChange({ ...target, rect: undefined, region: { x: 60, y: 40, w: 60, h: 40 } })
        }
    }

    return (
        <div className="space-y-6 max-w-lg">
            {/* Rect source */}
            <section className="space-y-3">
                <h3 className="text-sm font-semibold border-b border-border pb-1">Region</h3>
                <p className="text-[11px] text-muted-foreground">
                    The screen area to highlight. Use a named rect for reusable regions, or define an inline region for one-off placements.
                </p>

                <div className="flex gap-2">
                    <Button
                        variant={isNamed ? "default" : "outline"}
                        size="sm"
                        className="h-7 text-xs"
                        onClick={() => setRectMode(true)}
                    >
                        Named Rect
                    </Button>
                    <Button
                        variant={!isNamed ? "default" : "outline"}
                        size="sm"
                        className="h-7 text-xs"
                        onClick={() => setRectMode(false)}
                    >
                        Inline Region
                    </Button>
                </div>

                {isNamed ? (
                    <div className="space-y-1">
                        <label className="text-[11px] font-medium text-muted-foreground">Rect name</label>
                        <select
                            className="h-7 w-full rounded-md border border-border bg-background px-2 text-xs font-mono"
                            value={target.rect ?? ""}
                            onChange={e => onChange({ ...target, rect: e.target.value })}
                        >
                            <option value="">Select a rect...</option>
                            {Object.keys(namedRects).sort().map(name => (
                                <option key={name} value={name}>{name}</option>
                            ))}
                        </select>
                        {target.rect && namedRects[target.rect] && (
                            <div className="text-[10px] font-mono text-muted-foreground">
                                {namedRects[target.rect].x}, {namedRects[target.rect].y} -
                                {namedRects[target.rect].w}x{namedRects[target.rect].h}
                            </div>
                        )}
                    </div>
                ) : (
                    <div className="grid grid-cols-4 gap-2">
                        {(["x", "y", "w", "h"] as const).map(field => (
                            <div key={field} className="space-y-1">
                                <label className="text-[10px] font-medium text-muted-foreground uppercase">{field}</label>
                                <Input
                                    className="h-7 text-xs font-mono"
                                    type="number"
                                    value={target.region?.[field] ?? 0}
                                    onChange={e => onChange({
                                        ...target,
                                        region: { ...target.region!, [field]: parseInt(e.target.value) || 0 },
                                    })}
                                />
                            </div>
                        ))}
                    </div>
                )}
            </section>

            {/* Message */}
            <section className="space-y-3">
                <div className="flex items-center justify-between border-b border-border pb-1">
                    <h3 className="text-sm font-semibold">Message</h3>
                    {!target.message ? (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-5 text-[10px]"
                            onClick={() => onChange({ ...target, message: { text: "", placement: "bottom" } })}
                        >
                            <Plus className="h-3 w-3 mr-0.5" /> Add
                        </Button>
                    ) : (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-5 text-[10px] text-destructive"
                            onClick={() => onChange({ ...target, message: undefined })}
                        >
                            Remove
                        </Button>
                    )}
                </div>
                <p className="text-[11px] text-muted-foreground">
                    Text displayed alongside the highlighted region. Placement controls which side of the rect the message appears on.
                </p>

                {target.message && (
                    <div className="space-y-2">
                        <div className="space-y-1">
                            <label className="text-[11px] font-medium text-muted-foreground">Text</label>
                            <textarea
                                className="w-full rounded-md border border-border bg-background px-2 py-1 text-xs min-h-[60px] resize-y"
                                value={target.message.text}
                                onChange={e => onChange({
                                    ...target,
                                    message: { ...target.message!, text: e.target.value },
                                })}
                                placeholder="Message text..."
                            />
                        </div>
                        <div className="grid grid-cols-2 gap-2">
                            <div className="space-y-1">
                                <label className="text-[11px] font-medium text-muted-foreground">Placement</label>
                                <select
                                    className="h-7 w-full rounded-md border border-border bg-background px-2 text-xs"
                                    value={target.message.placement}
                                    onChange={e => onChange({
                                        ...target,
                                        message: { ...target.message!, placement: e.target.value },
                                    })}
                                >
                                    {MESSAGE_PLACEMENTS.map(p => <option key={p} value={p}>{p}</option>)}
                                </select>
                            </div>
                            <div className="space-y-1">
                                <label className="text-[11px] font-medium text-muted-foreground">Wrap width</label>
                                <Input
                                    className="h-7 text-xs font-mono"
                                    type="number"
                                    value={target.message.wrap ?? ""}
                                    onChange={e => onChange({
                                        ...target,
                                        message: {
                                            ...target.message!,
                                            wrap: e.target.value ? parseInt(e.target.value) : undefined,
                                        },
                                    })}
                                    placeholder="auto"
                                />
                            </div>
                        </div>
                    </div>
                )}
            </section>

            {/* Badge */}
            <section className="space-y-3">
                <div className="flex items-center justify-between border-b border-border pb-1">
                    <h3 className="text-sm font-semibold">Badge</h3>
                    {!target.badge ? (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-5 text-[10px]"
                            onClick={() => onChange({ ...target, badge: { placement: "bottom_middle", label: "OK" } })}
                        >
                            <Plus className="h-3 w-3 mr-0.5" /> Add
                        </Button>
                    ) : (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-5 text-[10px] text-destructive"
                            onClick={() => onChange({ ...target, badge: undefined })}
                        >
                            Remove
                        </Button>
                    )}
                </div>
                <p className="text-[11px] text-muted-foreground">
                    Small label shown inside the highlighted region - typically an action prompt like "OK" or "Next".
                </p>

                {target.badge && (
                    <div className="grid grid-cols-2 gap-2">
                        <div className="space-y-1">
                            <label className="text-[11px] font-medium text-muted-foreground">Placement</label>
                            <select
                                className="h-7 w-full rounded-md border border-border bg-background px-2 text-xs"
                                value={target.badge.placement ?? "bottom_middle"}
                                onChange={e => onChange({
                                    ...target,
                                    badge: { ...target.badge!, placement: e.target.value },
                                })}
                            >
                                {BADGE_PLACEMENTS.map(p => <option key={p} value={p}>{p}</option>)}
                            </select>
                        </div>
                        <div className="space-y-1">
                            <label className="text-[11px] font-medium text-muted-foreground">Label</label>
                            <Input
                                className="h-7 text-xs"
                                value={target.badge.label ?? ""}
                                onChange={e => onChange({
                                    ...target,
                                    badge: { ...target.badge!, label: e.target.value || undefined },
                                })}
                                placeholder="OK"
                            />
                        </div>
                    </div>
                )}
            </section>

            {/* No padding */}
            <section className="space-y-2">
                <h3 className="text-sm font-semibold border-b border-border pb-1">Options</h3>
                <label className="flex items-center gap-2 text-xs cursor-pointer">
                    <input
                        type="checkbox"
                        checked={target.noPadding ?? false}
                        onChange={e => onChange({ ...target, noPadding: e.target.checked || undefined })}
                        className="rounded"
                    />
                    <span>No padding</span>
                    <span className="text-[11px] text-muted-foreground">- Highlight region fits exactly to the rect with no extra padding around the cutout</span>
                </label>
            </section>
        </div>
    )
}

function flowToYaml(flow: OverlayFlow): Record<string, unknown> {
    const obj: Record<string, unknown> = {}
    if (flow.description) obj.description = flow.description
    obj.targets = flow.targets.map(t => {
        const target: Record<string, unknown> = {}
        if (t.rect) target.rect = t.rect
        if (t.region) target.region = t.region
        if (t.message) {
            const msg: Record<string, unknown> = { text: t.message.text, placement: t.message.placement }
            if (t.message.wrap) msg.wrap = t.message.wrap
            target.message = msg
        }
        if (t.badge) {
            const badge: Record<string, unknown> = {}
            if (t.badge.placement) badge.placement = t.badge.placement
            if (t.badge.label) badge.label = t.badge.label
            target.badge = badge
        }
        if (t.noPadding) target.no_padding = true
        return target
    })
    return obj
}
