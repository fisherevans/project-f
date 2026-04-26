import { useState, useEffect } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { useSave, useSaveSave, useDeleteSave } from "@/api/saves"
import { usePageTitle } from "@/hooks/usePageTitle"
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges"
import { Save, Trash2, ArrowLeft } from "lucide-react"
import { YamlEditor } from "@/components/ui/yaml-editor"

export function SaveEditor() {
    const params = useParams()
    const id = params.id ?? ""
    const navigate = useNavigate()
    const { data, isLoading, error } = useSave(id)
    const saveMutation = useSaveSave()
    const deleteMutation = useDeleteSave()

    const [yaml, setYaml] = useState("")
    const [dirty, setDirty] = useState(false)

    usePageTitle(id ? `${id} - Saves` : "Saves")
    useUnsavedChanges(dirty)

    useEffect(() => {
        if (data) {
            setYaml(data.raw_yaml)
            setDirty(false)
        }
    }, [data])

    const handleSave = () => {
        saveMutation.mutate({ id, rawYaml: yaml }, {
            onSuccess: () => setDirty(false),
        })
    }

    const handleDelete = () => {
        if (!confirm(`Delete save "${id}"? This cannot be undone.`)) return
        deleteMutation.mutate(id, { onSuccess: () => navigate("/saves") })
    }

    if (isLoading) return <div className="p-6 text-sm text-muted-foreground">Loading save...</div>
    if (error) return <div className="p-6 text-sm text-destructive">Failed to load save: {error.message}</div>
    if (!data) return null

    return (
        <div className="flex h-full flex-col">
            <div className="flex items-center gap-3 border-b border-border px-4 py-2 shrink-0">
                <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={() => navigate("/saves")}>
                    <ArrowLeft className="h-4 w-4" />
                </Button>
                <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium">{data.character_name || "(unnamed)"}</div>
                    <div className="text-xs font-mono text-muted-foreground">{data.save_id}</div>
                </div>
                <Button
                    variant="outline"
                    size="sm"
                    className="h-7 text-destructive"
                    onClick={handleDelete}
                    disabled={deleteMutation.isPending}
                >
                    <Trash2 className="h-3.5 w-3.5 mr-1" /> Delete
                </Button>
                <Button
                    size="sm"
                    className="h-7"
                    onClick={handleSave}
                    disabled={!dirty || saveMutation.isPending}
                >
                    <Save className="h-3.5 w-3.5 mr-1" />
                    {saveMutation.isPending ? "Saving..." : "Save"}
                </Button>
            </div>

            <div className="flex-1 flex flex-col overflow-hidden p-4 gap-3">
                <div>
                    <p className="text-xs text-muted-foreground mb-2">
                        Edit the raw YAML save file. The structure includes character info, animech (skill set, upgrades), loadouts, inventory, primortal progress, unlocked skills, globals (world state variables), and system settings (display, audio, combat speed, lighting). Changes write directly to game_data/saves/{id}.yaml.
                    </p>
                    <div className="text-[10px] text-muted-foreground space-y-0.5 mb-3">
                        <p><span className="font-medium">animech.skill_set</span> - 4 directional combat skills (skill_1 through skill_4)</p>
                        <p><span className="font-medium">loadouts</span> - named preset skill configurations</p>
                        <p><span className="font-medium">primortals</span> - per-creature discovery progress and research points</p>
                        <p><span className="font-medium">unlocked_skills</span> - skill IDs the player can equip</p>
                        <p><span className="font-medium">globals</span> - persistent world state variables (quest flags, counters)</p>
                        <p><span className="font-medium">system_settings</span> - lighting, combat speed, retro frame, audio, display options</p>
                    </div>
                </div>
                <div className="flex-1 w-full rounded-md border bg-card overflow-hidden">
                    <YamlEditor
                        value={yaml}
                        onChange={(v) => { setYaml(v); setDirty(true) }}
                    />
                </div>
            </div>
        </div>
    )
}
