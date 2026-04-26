import { useState, useEffect } from "react"
import { useSearchParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Dialog, DialogContent } from "@/components/ui/dialog"
import { usePrimortals, usePrimortal, useSavePrimortal, useDeletePrimortal } from "@/api/rpg"
import { usePageTitle } from "@/hooks/usePageTitle"
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges"
import type { Primortal, CombatArchetype, UnlockableSkill } from "@/types/rpg"
import { Save, Trash2, Plus, X, RotateCcw } from "lucide-react"

export function PrimortalBrowser() {
    const { data: primortals, isLoading, error } = usePrimortals()
    const [searchParams, setSearchParams] = useSearchParams()
    const [selectedType, setSelectedType] = useState(() => searchParams.get("type") ?? "")

    usePageTitle(selectedType ? `${selectedType} - Primortals` : "Primortals")

    useEffect(() => {
        const params = new URLSearchParams(searchParams)
        if (selectedType) {
            if (params.get("type") !== selectedType) { params.set("type", selectedType); setSearchParams(params, { replace: true }) }
        } else {
            if (params.has("type")) { params.delete("type"); setSearchParams(params, { replace: true }) }
        }
    }, [selectedType])

    return (
        <div className="flex h-full overflow-hidden">
            <div className="flex-1 flex flex-col overflow-hidden p-6 gap-4">
                <div className="flex items-center justify-between">
                    <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Primortals</h2>
                    <NewPrimortalButton onCreated={setSelectedType} />
                </div>

                {isLoading && <div className="text-sm text-muted-foreground">Loading...</div>}
                {error && <div className="text-sm text-destructive">Failed to load primortals</div>}

                {primortals && (
                    <div className="flex-1 overflow-auto rounded-md border">
                        <table className="w-full text-sm">
                            <thead className="sticky top-0 bg-muted">
                                <tr>
                                    <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Name</th>
                                    <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Type</th>
                                    <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase w-20">XenoLog</th>
                                    <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase w-20">Sync</th>
                                    <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase w-20">Skills</th>
                                </tr>
                            </thead>
                            <tbody>
                                {primortals.map(p => (
                                    <tr
                                        key={p.type}
                                        className={`border-t cursor-pointer transition-colors ${
                                            p.type === selectedType ? "bg-accent" : "hover:bg-muted/50"
                                        }`}
                                        onClick={() => setSelectedType(p.type)}
                                    >
                                        <td className="px-3 py-1.5 text-xs font-medium">{p.name}</td>
                                        <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">{p.type}</td>
                                        <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">
                                            {p.xeno_log_index || "-"}
                                        </td>
                                        <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">{p.base_sync}</td>
                                        <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">
                                            {p.unlockable_skills ? Object.keys(p.unlockable_skills).length : 0}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>

            <Dialog open={!!selectedType} onOpenChange={open => { if (!open) setSelectedType("") }}>
                <DialogContent className="!max-w-3xl !w-[800px] !max-h-[85vh] !grid-rows-none !flex !flex-col !p-0 !gap-0 overflow-hidden" showCloseButton={false}>
                    {selectedType && (
                        <PrimortalEditor type_={selectedType} onClose={() => setSelectedType("")} onDeleted={() => setSelectedType("")} />
                    )}
                </DialogContent>
            </Dialog>
        </div>
    )
}

function NewPrimortalButton({ onCreated }: { onCreated: (type_: string) => void }) {
    const save = useSavePrimortal()
    const handleCreate = () => {
        const type_ = prompt("Primortal type (snake_case):")
        if (!type_) return
        const p: Primortal = {
            type: type_,
            name: type_.replace(/_/g, " ").replace(/\b\w/g, c => c.toUpperCase()),
            description: "",
            base_sync: 25,
        }
        save.mutate(p, { onSuccess: () => onCreated(type_) })
    }
    return (
        <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleCreate}>
            <Plus className="h-3 w-3 mr-1" /> New Primortal
        </Button>
    )
}

function PrimortalEditor({ type_, onClose, onDeleted }: { type_: string; onClose: () => void; onDeleted: () => void }) {
    const { data, isLoading, error } = usePrimortal(type_)
    const save = useSavePrimortal()
    const del = useDeletePrimortal()
    const [draft, setDraft] = useState<Primortal | null>(null)

    useEffect(() => {
        if (data) setDraft(structuredClone(data))
    }, [data])

    const dirty = !!draft && JSON.stringify(draft) !== JSON.stringify(data)
    useUnsavedChanges(dirty)

    if (isLoading) return <div className="p-4 text-sm text-muted-foreground">Loading...</div>
    if (error) return <div className="p-4 text-sm text-destructive">{error.message}</div>
    if (!draft) return null

    const updateField = <K extends keyof Primortal>(key: K, value: Primortal[K]) => {
        setDraft(d => d ? { ...d, [key]: value } : d)
    }

    return (
        <>
            <div className="flex items-center gap-3 px-4 py-3 border-b shrink-0">
                <div className="flex-1 min-w-0">
                    <Input
                        value={draft.name}
                        onChange={e => updateField("name", e.target.value)}
                        className="h-7 text-sm font-bold"
                    />
                </div>
                <span className="font-mono text-xs text-muted-foreground shrink-0">{draft.type}</span>
                <div className="flex gap-1 shrink-0">
                    <Button
                        variant="outline" size="sm" className="h-7 text-xs gap-1"
                        disabled={!dirty || save.isPending}
                        onClick={() => save.mutate(draft)}
                    >
                        <Save className="h-3 w-3" /> Save
                    </Button>
                    <Button
                        variant="ghost" size="sm" className="h-7 text-xs gap-1"
                        disabled={!dirty}
                        onClick={() => data && setDraft(structuredClone(data))}
                    >
                        <RotateCcw className="h-3 w-3" /> Revert
                    </Button>
                    <Button
                        variant="ghost" size="sm" className="h-7 w-7 p-0 text-destructive"
                        onClick={() => {
                            if (confirm(`Delete primortal "${draft.name}"?`)) {
                                del.mutate(draft.type, { onSuccess: onDeleted })
                            }
                        }}
                    >
                        <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                    <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={onClose}>
                        <X className="h-3.5 w-3.5" />
                    </Button>
                </div>
            </div>

            <div className="flex-1 overflow-auto p-4">
                <textarea
                    value={draft.description}
                    onChange={e => updateField("description", e.target.value)}
                    placeholder="Description shown in the XenoLog entry"
                    className="w-full h-16 text-xs bg-background border rounded px-2 py-1 resize-none mb-3"
                />

                <div className="grid grid-cols-2 gap-2 mb-1">
                    <div>
                        <label className="text-[10px] text-muted-foreground uppercase">Base Sync</label>
                        <Input type="number" className="h-7 text-xs" value={draft.base_sync} onChange={e => updateField("base_sync", +e.target.value)} />
                    </div>
                    <div>
                        <label className="text-[10px] text-muted-foreground uppercase">XenoLog Index</label>
                        <Input type="number" className="h-7 text-xs" value={draft.xeno_log_index || 0} onChange={e => updateField("xeno_log_index", +e.target.value || undefined)} />
                    </div>
                </div>
                <p className="text-[10px] text-muted-foreground mb-4">
                    Base sync is the starting synchronization health for this primortal in combat. XenoLog index determines display order in the creature log (must be unique across all primortals).
                </p>

                <SkillTreeEditor
                    skills={draft.unlockable_skills || {}}
                    onChange={skills => updateField("unlockable_skills", Object.keys(skills).length > 0 ? skills : undefined)}
                />

                <ArchetypesEditor
                    archetypes={draft.combat_archetypes || {}}
                    onChange={archs => updateField("combat_archetypes", Object.keys(archs).length > 0 ? archs : undefined)}
                />
            </div>
        </>
    )
}

function SkillTreeEditor({ skills, onChange }: { skills: Record<string, UnlockableSkill>; onChange: (s: Record<string, UnlockableSkill>) => void }) {
    const entries = Object.entries(skills).sort(([, a], [, b]) => a.cost - b.cost)

    const addSkill = () => {
        const id = prompt("Skill ID:")
        if (!id) return
        onChange({ ...skills, [id]: { cost: 5 } })
    }

    const removeSkill = (id: string) => {
        const next = { ...skills }
        delete next[id]
        onChange(next)
    }

    const updateSkill = (id: string, sk: UnlockableSkill) => {
        onChange({ ...skills, [id]: sk })
    }

    return (
        <div className="mb-4">
            <div className="flex items-center justify-between mb-1">
                <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Unlockable Skills</span>
                <button className="text-[10px] text-muted-foreground hover:text-foreground px-1.5 py-0.5 border rounded" onClick={addSkill}>+ Skill</button>
            </div>
            <p className="text-[10px] text-muted-foreground mb-2">
                Skills the player can unlock by spending research points on this primortal. Cost is in XP. Prerequisites are comma-separated skill IDs that must be unlocked first (from this primortal's tree). Sorted by cost ascending.
            </p>
            <div className="flex flex-col gap-1">
                {entries.map(([id, sk]) => (
                    <div key={id} className="flex items-center gap-1 rounded border px-2 py-1 bg-muted/30">
                        <span className="font-mono text-xs flex-1 min-w-0 truncate">{id}</span>
                        <Input
                            type="number"
                            className="h-5 w-12 text-xs px-1"
                            value={sk.cost}
                            onChange={e => updateSkill(id, { ...sk, cost: +e.target.value })}
                        />
                        <span className="text-[10px] text-muted-foreground">xp</span>
                        <Input
                            className="h-5 w-24 text-xs px-1"
                            value={(sk.prerequisites || []).join(",")}
                            placeholder="prereqs"
                            onChange={e => {
                                const prereqs = e.target.value.split(",").map(s => s.trim()).filter(Boolean)
                                updateSkill(id, { ...sk, prerequisites: prereqs.length > 0 ? prereqs : undefined })
                            }}
                        />
                        <button className="text-muted-foreground hover:text-destructive p-0.5" onClick={() => removeSkill(id)}>
                            <X className="h-3 w-3" />
                        </button>
                    </div>
                ))}
            </div>
        </div>
    )
}

function ArchetypesEditor({ archetypes, onChange }: { archetypes: Record<string, CombatArchetype>; onChange: (a: Record<string, CombatArchetype>) => void }) {
    const entries = Object.entries(archetypes)

    const addArchetype = () => {
        const name = prompt("Archetype name:")
        if (!name) return
        onChange({ ...archetypes, [name]: {} })
    }

    const removeArchetype = (name: string) => {
        const next = { ...archetypes }
        delete next[name]
        onChange(next)
    }

    const updateArchetype = (name: string, arch: CombatArchetype) => {
        onChange({ ...archetypes, [name]: arch })
    }

    return (
        <div>
            <div className="flex items-center justify-between mb-1">
                <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Combat Archetypes</span>
                <button className="text-[10px] text-muted-foreground hover:text-foreground px-1.5 py-0.5 border rounded" onClick={addArchetype}>+ Archetype</button>
            </div>
            <p className="text-[10px] text-muted-foreground mb-2">
                Combat archetypes define how this primortal behaves as an opponent. Each archetype has a sync modifier (±variance) and a skill pool. The opener plays skills in order, then weighted random selection picks subsequent skills. Higher weight = more likely to be chosen.
            </p>
            <div className="flex flex-col gap-2">
                {entries.map(([name, arch]) => (
                    <ArchetypeEditor key={name} name={name} archetype={arch} onChange={a => updateArchetype(name, a)} onRemove={() => removeArchetype(name)} />
                ))}
            </div>
        </div>
    )
}

function ArchetypeEditor({ name, archetype, onChange, onRemove }: { name: string; archetype: CombatArchetype; onChange: (a: CombatArchetype) => void; onRemove: () => void }) {
    const pool = archetype.skill_pool?.random
    const weightedEntries = pool?.weighted_skills ? Object.entries(pool.weighted_skills) : []

    const setWeighted = (entries: [string, number][]) => {
        const ws: Record<string, number> = {}
        for (const [k, v] of entries) ws[k] = v
        const random = {
            ...pool,
            weighted_skills: Object.keys(ws).length > 0 ? ws : undefined,
        }
        onChange({
            ...archetype,
            skill_pool: { random: (random.initial_ordered_skills?.length || random.weighted_skills) ? random : undefined },
        })
    }

    return (
        <div className="rounded border bg-muted/30 p-2">
            <div className="flex items-center justify-between mb-1.5">
                <span className="font-mono text-xs font-medium">{name}</span>
                <button className="text-muted-foreground hover:text-destructive p-0.5" onClick={onRemove}>
                    <X className="h-3 w-3" />
                </button>
            </div>

            <div className="grid grid-cols-2 gap-1 mb-2">
                <div>
                    <label className="text-[10px] text-muted-foreground" title="Added to base_sync for this archetype">+Sync</label>
                    <Input type="number" className="h-5 text-xs px-1" value={archetype.additional_sync || 0}
                        onChange={e => onChange({ ...archetype, additional_sync: +e.target.value || undefined })} />
                </div>
                <div>
                    <label className="text-[10px] text-muted-foreground" title="Random ± added to sync each encounter">±Variance</label>
                    <Input type="number" className="h-5 text-xs px-1" value={archetype.additional_sync_variance || 0}
                        onChange={e => onChange({ ...archetype, additional_sync_variance: +e.target.value || undefined })} />
                </div>
            </div>

            <div>
                <div className="flex items-center justify-between mb-0.5">
                    <label className="text-[10px] text-muted-foreground">Opener</label>
                </div>
                <Input
                    className="h-5 text-xs px-1 mb-1 font-mono"
                    value={(pool?.initial_ordered_skills || []).join(", ")}
                    placeholder="skill_id, skill_id, ... (played in order first)"
                    onChange={e => {
                        const skills = e.target.value.split(",").map(s => s.trim()).filter(Boolean)
                        const random = {
                            ...pool,
                            initial_ordered_skills: skills.length > 0 ? skills : undefined,
                        }
                        onChange({
                            ...archetype,
                            skill_pool: { random: (random.initial_ordered_skills?.length || random.weighted_skills) ? random : undefined },
                        })
                    }}
                />
            </div>

            <div>
                <div className="flex items-center justify-between mb-0.5">
                    <label className="text-[10px] text-muted-foreground">Weighted Skills</label>
                    <button
                        className="text-[10px] text-muted-foreground hover:text-foreground px-1 border rounded"
                        onClick={() => {
                            const id = prompt("Skill ID:")
                            if (!id) return
                            setWeighted([...weightedEntries, [id, 10]])
                        }}
                    >+</button>
                </div>
                {weightedEntries.map(([sid, w], i) => (
                    <div key={i} className="flex items-center gap-1 mb-0.5">
                        <span className="font-mono text-[10px] flex-1 min-w-0 truncate">{sid}</span>
                        <Input
                            type="number"
                            className="h-5 w-14 text-xs px-1"
                            value={w}
                            onChange={e => {
                                const next = [...weightedEntries] as [string, number][]
                                next[i] = [sid, +e.target.value]
                                setWeighted(next)
                            }}
                        />
                        <button
                            className="text-muted-foreground hover:text-destructive p-0.5"
                            onClick={() => setWeighted(weightedEntries.filter((_, idx) => idx !== i) as [string, number][])}
                        >
                            <X className="h-3 w-3" />
                        </button>
                    </div>
                ))}
            </div>
        </div>
    )
}
