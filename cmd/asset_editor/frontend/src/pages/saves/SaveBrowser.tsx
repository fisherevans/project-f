import { useState, useEffect } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { useSaves, useDeleteSave, useCloneSave } from "@/api/saves"
import { usePageTitle } from "@/hooks/usePageTitle"
import { Copy, Trash2 } from "lucide-react"

export function SaveBrowser() {
    const { data: saves, isLoading, error } = useSaves()
    const deleteSave = useDeleteSave()
    const cloneSave = useCloneSave()
    const navigate = useNavigate()
    const [searchParams] = useSearchParams()
    const [selectedId, setSelectedId] = useState(() => searchParams.get("save") ?? "")

    usePageTitle("Saves")

    useEffect(() => {
        if (saves && selectedId && !saves.find(s => s.save_id === selectedId)) {
            setSelectedId("")
        }
    }, [saves, selectedId])

    const handleClone = (srcId: string) => {
        const newId = prompt("New save ID:")
        if (!newId) return
        cloneSave.mutate({ srcId, newId })
    }

    const handleDelete = (id: string) => {
        if (!confirm(`Delete save "${id}"? This cannot be undone.`)) return
        deleteSave.mutate(id, { onSuccess: () => { if (selectedId === id) setSelectedId("") } })
    }

    if (isLoading) return <div className="p-6 text-sm text-muted-foreground">Loading saves...</div>
    if (error) return <div className="p-6 text-sm text-destructive">Failed to load saves: {error.message}</div>

    return (
        <div className="h-full overflow-auto p-6">
            <div className="max-w-2xl space-y-4">
                <div>
                    <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-1">Game Saves</h2>
                    <p className="text-xs text-muted-foreground mb-3">
                        Save files from game_data/saves/. Each file is a complete game state snapshot including character progress, inventory, globals, and system settings. Editing a save here writes directly to the YAML file - changes take effect next time the game loads that save.
                    </p>
                </div>

                {saves && saves.length === 0 && (
                    <p className="text-sm text-muted-foreground">No save files found in game_data/saves/</p>
                )}

                <div className="space-y-2">
                    {saves?.map(save => (
                        <div
                            key={save.save_id}
                            className="flex items-center gap-3 rounded-md border p-3 hover:bg-accent/50 transition-colors cursor-pointer"
                            onClick={() => navigate(`/saves/${save.save_id}`)}
                        >
                            <div className="flex-1 min-w-0">
                                <div className="text-sm font-medium">{save.character_name || "(unnamed)"}</div>
                                <div className="text-xs font-mono text-muted-foreground">{save.save_id}</div>
                            </div>
                            <div className="flex gap-1" onClick={e => e.stopPropagation()}>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-7 w-7 p-0"
                                    title="Clone save"
                                    onClick={() => handleClone(save.save_id)}
                                    disabled={cloneSave.isPending}
                                >
                                    <Copy className="h-3.5 w-3.5" />
                                </Button>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-7 w-7 p-0 text-destructive"
                                    title="Delete save"
                                    onClick={() => handleDelete(save.save_id)}
                                    disabled={deleteSave.isPending}
                                >
                                    <Trash2 className="h-3.5 w-3.5" />
                                </Button>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    )
}
