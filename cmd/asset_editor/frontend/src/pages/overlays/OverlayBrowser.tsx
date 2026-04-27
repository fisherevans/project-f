import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { useOverlays, useDeleteOverlay } from "@/api/overlays"
import { usePageTitle } from "@/hooks/usePageTitle"
import { NamedRectEditor } from "@/components/overlays/NamedRectEditor"
import { Layers, Plus, Trash2, ChevronRight } from "lucide-react"

export function OverlayBrowser() {
    const navigate = useNavigate()
    const { data: flows, isLoading, error } = useOverlays()
    const deleteMutation = useDeleteOverlay()
    usePageTitle("Overlays")

    const [showRects, setShowRects] = useState(true)

    const handleCreate = () => {
        const name = prompt("Flow name (use / for folders, e.g. combat_training/intro):")
        if (!name) return
        const sanitized = name
            .split("/")
            .map(seg => seg.replace(/[^a-z0-9_-]/gi, "_").toLowerCase())
            .filter(Boolean)
            .join("/")
        if (!sanitized) return
        navigate(`/overlays/${sanitized}?new=1`)
    }

    const handleDelete = (name: string, e: React.MouseEvent) => {
        e.stopPropagation()
        if (!confirm(`Delete overlay flow "${name}"?`)) return
        deleteMutation.mutate(name)
    }

    if (isLoading) return <div className="p-4 text-sm text-muted-foreground">Loading overlays...</div>
    if (error) return <div className="p-4 text-sm text-destructive">Error: {error.message}</div>

    return (
        <div className="h-full overflow-auto p-4">
            <div className="max-w-2xl space-y-6">
                <div>
                    <div className="flex items-center justify-between mb-1">
                        <h1 className="text-lg font-semibold">Overlay Flows</h1>
                        <Button variant="outline" size="sm" className="h-7" onClick={handleCreate}>
                            <Plus className="h-3.5 w-3.5 mr-1" /> New Flow
                        </Button>
                    </div>
                    <p className="text-xs text-muted-foreground mb-3">
                        Highlight overlay sequences for tutorials and UI callouts. Each flow defines a step-through sequence of screen regions with messages and badges. Referenced in scripts via <code className="font-mono">highlight_flow</code> or loaded directly in Go code.
                    </p>
                </div>

                {flows && flows.length === 0 && (
                    <p className="text-sm text-muted-foreground">No overlay flows found in assets/overlays/</p>
                )}

                {flows && flows.length > 0 && (
                    <div className="space-y-4">
                        {groupFlowsByFolder(flows).map(([folder, items]) => (
                            <div key={folder} className="space-y-1">
                                {folder && (
                                    <div className="text-[11px] font-mono text-muted-foreground uppercase tracking-wide px-1">
                                        {folder}/
                                    </div>
                                )}
                                {items.map(flow => {
                                    const leaf = flow.name.includes("/") ? flow.name.split("/").pop()! : flow.name
                                    return (
                                        <button
                                            key={flow.name}
                                            onClick={() => navigate(`/overlays/${flow.name}`)}
                                            className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm transition-colors hover:bg-accent"
                                        >
                                            <Layers className="h-4 w-4 shrink-0 text-muted-foreground" />
                                            <div className="min-w-0 flex-1">
                                                <div className="font-medium font-mono">{leaf}</div>
                                                <div className="text-xs text-muted-foreground">
                                                    {flow.targetCount} target{flow.targetCount !== 1 ? "s" : ""}
                                                    {flow.description && ` - ${flow.description}`}
                                                </div>
                                            </div>
                                            <div className="flex items-center gap-1" onClick={e => e.stopPropagation()}>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    className="h-7 w-7 p-0 text-destructive"
                                                    onClick={e => handleDelete(flow.name, e)}
                                                    disabled={deleteMutation.isPending}
                                                >
                                                    <Trash2 className="h-3.5 w-3.5" />
                                                </Button>
                                            </div>
                                            <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
                                        </button>
                                    )
                                })}
                            </div>
                        ))}
                    </div>
                )}

                <div className="border-t border-border pt-4">
                    <button
                        onClick={() => setShowRects(!showRects)}
                        className="text-sm font-semibold text-foreground mb-2 hover:text-accent-foreground"
                    >
                        {showRects ? "▾" : "▸"} Named Rects Library
                    </button>
                    {showRects && <NamedRectEditor />}
                </div>
            </div>
        </div>
    )
}

// Groups flows by their parent folder. Root-level flows go under "" (rendered without a header).
// Returns folder entries sorted alphabetically with root first, items within each folder sorted by name.
function groupFlowsByFolder(flows: { name: string; description?: string; targetCount: number }[]):
    [string, typeof flows][] {
    const groups = new Map<string, typeof flows>()
    for (const flow of flows) {
        const idx = flow.name.lastIndexOf("/")
        const folder = idx >= 0 ? flow.name.slice(0, idx) : ""
        if (!groups.has(folder)) groups.set(folder, [])
        groups.get(folder)!.push(flow)
    }
    const sorted = Array.from(groups.entries()).sort(([a], [b]) => {
        if (a === "") return -1
        if (b === "") return 1
        return a.localeCompare(b)
    })
    for (const [, items] of sorted) {
        items.sort((a, b) => a.name.localeCompare(b.name))
    }
    return sorted
}
