import { useState, useEffect } from "react"
import { useSearchParams } from "react-router-dom"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Dialog, DialogContent } from "@/components/ui/dialog"
import { useSkills, useSkill, useSaveSkill, useDeleteSkill } from "@/api/rpg"
import { usePageTitle } from "@/hooks/usePageTitle"
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges"
import type { Skill, SkillTick, TickEffect, DamageScaling } from "@/types/rpg"
import { Save, Trash2, Plus, X, GripVertical, ChevronUp, ChevronDown, Swords, Shield, RotateCcw } from "lucide-react"

const stanceColors: Record<string, string> = {
    defending: "bg-accent-blue-tint border-accent-blue-edge",
    reflecting: "bg-accent-violet-tint border-accent-violet-edge",
    vulnerable: "bg-accent-orange-tint border-accent-orange-edge",
    exposed: "bg-accent-red-tint border-accent-red-edge",
}

const stanceBadgeVariants: Record<string, string> = {
    defending: "bg-accent-blue-tint text-accent-blue border-accent-blue-edge",
    reflecting: "bg-accent-violet-tint text-accent-violet border-accent-violet-edge",
    vulnerable: "bg-accent-orange-tint text-accent-orange border-accent-orange-edge",
    exposed: "bg-accent-red-tint text-accent-red border-accent-red-edge",
}

const statusColors: Record<string, string> = {
    warded: "text-accent-blue",
    burning: "text-accent-orange",
    poisoned: "text-accent-green",
    ionized: "text-accent-yellow",
    mending: "text-accent-teal",
}

const STANCES = ["", "defending", "reflecting", "vulnerable", "exposed"]
const STATUS_TYPES = ["warded", "burning", "poisoned", "ionized", "mending"]
const ANIM_TYPES = ["pounce", "recoil", "wiggle", "hop"]

const ANIM_TYPE_INFO: Record<string, { label: string; motion: string; description: string; baseDuration: string; icon: string }> = {
    pounce: {
        label: "Pounce",
        motion: "→ lunge →",
        description: "Lunges toward the opponent with a forward arc, then returns. Classic melee attack feel.",
        baseDuration: "1.0s",
        icon: "⟶",
    },
    recoil: {
        label: "Recoil",
        motion: "← knockback",
        description: "Snaps backward briefly as if struck, then settles. Used on targets taking a hit.",
        baseDuration: "1.0s",
        icon: "⟵",
    },
    wiggle: {
        label: "Wiggle",
        motion: "↔ shake",
        description: "Rapid horizontal shake in place. Channeling, charging, or status application.",
        baseDuration: "0.5s",
        icon: "↔",
    },
    hop: {
        label: "Hop",
        motion: "↑ bounce",
        description: "Quick vertical bounce. Pairs well with repetitions for multi-hit or projectile skills.",
        baseDuration: "0.5s",
        icon: "↕",
    },
}

export function SkillBrowser() {
    const { data: skills, isLoading, error } = useSkills()
    const [searchParams, setSearchParams] = useSearchParams()
    const [selectedId, setSelectedId] = useState(() => searchParams.get("skill") ?? "")
    const [filter, setFilter] = useState("")

    usePageTitle(selectedId ? `${selectedId} - Skills` : "Skills")

    useEffect(() => {
        const params = new URLSearchParams(searchParams)
        if (selectedId) {
            if (params.get("skill") !== selectedId) { params.set("skill", selectedId); setSearchParams(params, { replace: true }) }
        } else {
            if (params.has("skill")) { params.delete("skill"); setSearchParams(params, { replace: true }) }
        }
    }, [selectedId])

    const filtered = skills?.filter(sk => {
        if (!filter) return true
        const q = filter.toLowerCase()
        return sk.id.toLowerCase().includes(q) || sk.name.toLowerCase().includes(q)
    })

    return (
        <div className="flex h-full overflow-hidden">
            <div className="flex-1 flex flex-col overflow-hidden p-6 gap-4">
                <div className="flex items-center justify-between">
                    <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Skills</h2>
                    <div className="flex items-center gap-2">
                        <Input
                            placeholder="Filter by name or ID..."
                            value={filter}
                            onChange={e => setFilter(e.target.value)}
                            className="h-7 w-48 text-xs"
                        />
                        <NewSkillButton onCreated={setSelectedId} />
                    </div>
                </div>

                {isLoading && <div className="text-sm text-muted-foreground">Loading...</div>}
                {error && <div className="text-sm text-destructive">Failed to load skills</div>}

                {filtered && (
                    <div className="flex-1 overflow-auto rounded-md border">
                        <table className="w-full text-sm">
                            <thead className="sticky top-0 bg-muted">
                                <tr>
                                    <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Name</th>
                                    <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">ID</th>
                                    <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Description</th>
                                    <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase w-20">Ticks</th>
                                </tr>
                            </thead>
                            <tbody>
                                {filtered.map(sk => (
                                    <tr
                                        key={sk.id}
                                        className={`border-t cursor-pointer transition-colors ${
                                            sk.id === selectedId ? "bg-accent" : "hover:bg-muted/50"
                                        }`}
                                        onClick={() => setSelectedId(sk.id)}
                                    >
                                        <td className="px-3 py-1.5 text-xs font-medium">{sk.name}</td>
                                        <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">{sk.id}</td>
                                        <td className="px-3 py-1.5 text-xs text-muted-foreground">{sk.description}</td>
                                        <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">{sk.ticks.length}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>

            <Dialog open={!!selectedId} onOpenChange={open => { if (!open) setSelectedId("") }}>
                <DialogContent className="!max-w-3xl !w-[800px] !max-h-[85vh] !grid-rows-none !flex !flex-col !p-0 !gap-0 overflow-hidden" showCloseButton={false}>
                    {selectedId && (
                        <SkillEditor id={selectedId} onClose={() => setSelectedId("")} onDeleted={() => setSelectedId("")} />
                    )}
                </DialogContent>
            </Dialog>
        </div>
    )
}

function NewSkillButton({ onCreated }: { onCreated: (id: string) => void }) {
    const save = useSaveSkill()
    const handleCreate = () => {
        const id = prompt("Skill ID (snake_case):")
        if (!id) return
        const newSkill: Skill = {
            id,
            name: id.replace(/_/g, " ").replace(/\b\w/g, c => c.toUpperCase()),
            description: "",
            ticks: [{}],
        }
        save.mutate(newSkill, { onSuccess: () => onCreated(id) })
    }
    return (
        <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleCreate}>
            <Plus className="h-3 w-3 mr-1" /> New Skill
        </Button>
    )
}

function SkillEditor({ id, onClose, onDeleted }: { id: string; onClose: () => void; onDeleted: () => void }) {
    const { data, isLoading, error } = useSkill(id)
    const save = useSaveSkill()
    const del = useDeleteSkill()
    const [draft, setDraft] = useState<Skill | null>(null)

    useEffect(() => {
        if (data) setDraft(structuredClone(data))
    }, [data])

    const dirty = !!draft && JSON.stringify(draft) !== JSON.stringify(data)
    useUnsavedChanges(dirty)

    if (isLoading) return <div className="p-4 text-sm text-muted-foreground">Loading...</div>
    if (error) return <div className="p-4 text-sm text-destructive">{error.message}</div>
    if (!draft) return null

    const updateField = <K extends keyof Skill>(key: K, value: Skill[K]) => {
        setDraft(d => d ? { ...d, [key]: value } : d)
    }

    const updateTick = (idx: number, tick: SkillTick) => {
        const ticks = [...draft.ticks]
        ticks[idx] = tick
        updateField("ticks", ticks)
    }

    const addTick = () => {
        updateField("ticks", [...draft.ticks, {}])
    }

    const removeTick = (idx: number) => {
        updateField("ticks", draft.ticks.filter((_, i) => i !== idx))
    }

    const moveTick = (from: number, to: number) => {
        if (to < 0 || to >= draft.ticks.length) return
        const ticks = [...draft.ticks]
        const [moved] = ticks.splice(from, 1)
        ticks.splice(to, 0, moved)
        updateField("ticks", ticks)
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
                <span className="font-mono text-xs text-muted-foreground shrink-0">{draft.id}</span>
                <div className="flex gap-1 shrink-0">
                    <Button
                        variant="outline"
                        size="sm"
                        className="h-7 text-xs gap-1"
                        disabled={!dirty || save.isPending}
                        onClick={() => save.mutate(draft)}
                    >
                        <Save className="h-3 w-3" /> Save
                    </Button>
                    <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 text-xs gap-1"
                        disabled={!dirty}
                        onClick={() => data && setDraft(structuredClone(data))}
                    >
                        <RotateCcw className="h-3 w-3" /> Revert
                    </Button>
                    <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 w-7 p-0 text-destructive"
                        onClick={() => {
                            if (confirm(`Delete skill "${draft.name}"?`)) {
                                del.mutate(draft.id, { onSuccess: onDeleted })
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
                <Input
                    value={draft.description}
                    onChange={e => updateField("description", e.target.value)}
                    placeholder="Short description shown in skill selection UI"
                    className="h-7 text-xs mb-4"
                />

                <div className="flex items-center justify-between mb-1">
                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                        Ticks ({draft.ticks.length})
                    </span>
                    <Button variant="outline" size="sm" className="h-6 text-[10px]" onClick={addTick}>
                        <Plus className="h-3 w-3 mr-0.5" /> Tick
                    </Button>
                </div>
                <p className="text-[10px] text-muted-foreground mb-2">
                    Each tick is one beat of the skill's combat animation. Ticks play in sequence - stance changes, damage, status effects, and animations all happen per-tick.
                </p>

                <div className="flex flex-col gap-0.5">
                    {draft.ticks.map((tick, i) => (
                        <TickEditor
                            key={i}
                            index={i}
                            tick={tick}
                            onChange={t => updateTick(i, t)}
                            onRemove={() => removeTick(i)}
                            onMoveUp={() => moveTick(i, i - 1)}
                            onMoveDown={() => moveTick(i, i + 1)}
                            isFirst={i === 0}
                            isLast={i === draft.ticks.length - 1}
                        />
                    ))}
                </div>
            </div>
        </>
    )
}

function TickEditor({
    index, tick, onChange, onRemove, onMoveUp, onMoveDown, isFirst, isLast,
}: {
    index: number
    tick: SkillTick
    onChange: (t: SkillTick) => void
    onRemove: () => void
    onMoveUp: () => void
    onMoveDown: () => void
    isFirst: boolean
    isLast: boolean
}) {
    const [expanded, setExpanded] = useState(false)
    const stance = tick.stance || ""
    const bgColor = stanceColors[stance] || "bg-muted/30 border-border"

    return (
        <div className={`rounded border ${bgColor}`}>
            <div className="flex items-center gap-1 px-1.5 py-1 cursor-pointer" onClick={() => setExpanded(!expanded)}>
                <span className="font-mono text-[10px] text-muted-foreground w-4 text-right shrink-0">{index}</span>
                <div className="flex-1 flex items-center gap-1 min-w-0 flex-wrap">
                    {stance && (
                        <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium border ${stanceBadgeVariants[stance] ?? ""}`}>
                            {stance}
                        </span>
                    )}
                    {(tick.effects || []).map((eff, i) => (
                        <EffectBadge key={i} effect={eff} />
                    ))}
                    {(tick.animation?.source || []).map((a, i) => (
                        <AnimBadge key={`s${i}`} anim={a} role="source" />
                    ))}
                    {(tick.animation?.target || []).map((a, i) => (
                        <AnimBadge key={`t${i}`} anim={a} role="target" />
                    ))}
                    {!stance && (!tick.effects || tick.effects.length === 0) && !tick.animation && (
                        <span className="text-[10px] text-muted-foreground/50 italic">idle</span>
                    )}
                </div>
                <div className="flex gap-0.5 shrink-0" onClick={e => e.stopPropagation()}>
                    <button className="text-muted-foreground hover:text-foreground disabled:opacity-30 p-0.5" disabled={isFirst} onClick={onMoveUp}>
                        <GripVertical className="h-3 w-3 rotate-180" />
                    </button>
                    <button className="text-muted-foreground hover:text-foreground disabled:opacity-30 p-0.5" disabled={isLast} onClick={onMoveDown}>
                        <GripVertical className="h-3 w-3" />
                    </button>
                    <button className="text-muted-foreground hover:text-destructive p-0.5" onClick={onRemove}>
                        <X className="h-3 w-3" />
                    </button>
                </div>
            </div>

            {expanded && (
                <div className="px-2 pb-2 pt-1 border-t border-border/50 space-y-2">
                    <div>
                        <label className="text-[10px] text-muted-foreground uppercase">Stance</label>
                        <select
                            className="w-full h-6 text-xs bg-background border rounded px-1"
                            value={stance}
                            onChange={e => onChange({ ...tick, stance: e.target.value || undefined })}
                        >
                            {STANCES.map(s => <option key={s} value={s}>{s || "(none)"}</option>)}
                        </select>
                        <p className="text-[10px] text-muted-foreground mt-0.5">
                            {stance === "defending" && "Reduces incoming damage. Must last 2+ ticks."}
                            {stance === "reflecting" && "Returns a portion of incoming damage. Must last 2+ ticks."}
                            {stance === "vulnerable" && "Takes increased damage. Must last 2+ ticks."}
                            {stance === "exposed" && "Critically vulnerable - heavy damage increase. Must last 2+ ticks."}
                            {!stance && "No stance change this tick."}
                        </p>
                    </div>

                    <EffectsEditor
                        effects={tick.effects || []}
                        onChange={effects => onChange({ ...tick, effects: effects.length > 0 ? effects : undefined })}
                    />

                    <AnimationEditor
                        sourceAnims={tick.animation?.source || []}
                        targetAnims={tick.animation?.target || []}
                        onChange={(source, target) => {
                            const anim = (source.length > 0 || target.length > 0)
                                ? { source: source.length > 0 ? source : undefined, target: target.length > 0 ? target : undefined }
                                : undefined
                            onChange({ ...tick, animation: anim })
                        }}
                    />
                </div>
            )}
        </div>
    )
}

function EffectsEditor({ effects, onChange }: { effects: TickEffect[]; onChange: (e: TickEffect[]) => void }) {
    const addDamage = () => onChange([...effects, { damage: { amount: 5 } }])
    const addStatus = () => onChange([...effects, { status: { type: "burning", stacks: 5, target: "opponent" } }])
    const remove = (i: number) => onChange(effects.filter((_, idx) => idx !== i))
    const update = (i: number, eff: TickEffect) => {
        const next = [...effects]
        next[i] = eff
        onChange(next)
    }

    return (
        <div>
            <div className="flex items-center justify-between mb-1">
                <label className="text-[10px] text-muted-foreground uppercase">Effects</label>
                <div className="flex gap-1">
                    <button className="text-[10px] text-muted-foreground hover:text-foreground px-1 border rounded" onClick={addDamage}>+dmg</button>
                    <button className="text-[10px] text-muted-foreground hover:text-foreground px-1 border rounded" onClick={addStatus}>+status</button>
                </div>
            </div>
            {effects.length === 0 && (
                <p className="text-[10px] text-muted-foreground/50 italic mb-1">No effects. Add damage or status effects above.</p>
            )}
            {effects.map((eff, i) => (
                <div key={i} className="flex items-start gap-1 mb-1">
                    <div className="flex-1 bg-background rounded border p-1.5 text-xs space-y-1">
                        {eff.damage && (
                            <>
                                <div className="flex items-center gap-1 flex-wrap">
                                    <span className="text-[10px] text-muted-foreground w-12" title="Base damage dealt">dmg</span>
                                    <Input
                                        type="number"
                                        className="h-5 w-14 text-xs px-1"
                                        value={eff.damage.amount}
                                        onChange={e => update(i, { ...eff, damage: { ...eff.damage!, amount: +e.target.value } })}
                                    />
                                    <span className="text-[10px] text-muted-foreground" title="Random variance added/subtracted from base">±</span>
                                    <Input
                                        type="number"
                                        className="h-5 w-12 text-xs px-1"
                                        placeholder="0"
                                        value={eff.damage.variance || ""}
                                        onChange={e => update(i, { ...eff, damage: { ...eff.damage!, variance: +e.target.value || undefined } })}
                                    />
                                    <span className="text-[10px] text-muted-foreground" title="Chance to miss (0.0 - 1.0)">miss</span>
                                    <Input
                                        type="number"
                                        className="h-5 w-14 text-xs px-1"
                                        placeholder="0"
                                        step="0.05"
                                        min="0"
                                        max="1"
                                        value={eff.damage.miss_rate || ""}
                                        onChange={e => update(i, { ...eff, damage: { ...eff.damage!, miss_rate: +e.target.value || undefined } })}
                                    />
                                </div>
                                <DamageScalingEditor
                                    scaling={eff.damage.scaled_by}
                                    onChange={scaled_by => update(i, { ...eff, damage: { ...eff.damage!, scaled_by: scaled_by || undefined } })}
                                />
                            </>
                        )}
                        {eff.status && (
                            <>
                                <div className="flex items-center gap-1 flex-wrap">
                                    <select
                                        className="h-5 text-xs bg-background border rounded px-1"
                                        value={eff.status.type}
                                        onChange={e => update(i, { ...eff, status: { ...eff.status!, type: e.target.value } })}
                                    >
                                        {STATUS_TYPES.map(s => <option key={s} value={s}>{s}</option>)}
                                    </select>
                                    <Input
                                        type="number"
                                        className="h-5 w-12 text-xs px-1"
                                        title="Number of stacks to apply"
                                        value={eff.status.stacks}
                                        onChange={e => update(i, { ...eff, status: { ...eff.status!, stacks: +e.target.value } })}
                                    />
                                    <select
                                        className="h-5 text-xs bg-background border rounded px-1"
                                        value={eff.status.target}
                                        onChange={e => update(i, { ...eff, status: { ...eff.status!, target: e.target.value } })}
                                    >
                                        <option value="opponent">opponent</option>
                                        <option value="self">self</option>
                                    </select>
                                    <select
                                        className="h-5 text-xs bg-background border rounded px-1"
                                        value={eff.status.require || ""}
                                        onChange={e => update(i, { ...eff, status: { ...eff.status!, require: e.target.value || undefined } })}
                                    >
                                        <option value="">(no requirement)</option>
                                        <option value="existing_stacks">existing stacks</option>
                                        <option value="no_stacks">no stacks</option>
                                    </select>
                                </div>
                                <p className="text-[10px] text-muted-foreground">
                                    {eff.status.require === "existing_stacks" && "Only applies if target already has stacks of this status."}
                                    {eff.status.require === "no_stacks" && "Only applies if target has zero stacks of this status."}
                                </p>
                            </>
                        )}
                    </div>
                    <button className="text-muted-foreground hover:text-destructive p-0.5 mt-1" onClick={() => remove(i)}>
                        <X className="h-3 w-3" />
                    </button>
                </div>
            ))}
        </div>
    )
}

function DamageScalingEditor({ scaling, onChange }: { scaling?: DamageScaling; onChange: (s: DamageScaling | null) => void }) {
    const [open, setOpen] = useState(false)
    const hasScaling = scaling && (
        (scaling.target_status && Object.keys(scaling.target_status).length > 0) ||
        (scaling.source_status && Object.keys(scaling.source_status).length > 0)
    )

    if (!open && !hasScaling) {
        return (
            <button
                className="text-[10px] text-muted-foreground hover:text-foreground"
                onClick={() => setOpen(true)}
            >
                + damage scaling
            </button>
        )
    }

    const renderScalingGroup = (
        label: string,
        description: string,
        data: Record<string, Record<string, number>> | undefined,
        setData: (d: Record<string, Record<string, number>> | undefined) => void,
    ) => {
        const entries = data ? Object.entries(data) : []
        return (
            <div>
                <div className="flex items-center justify-between">
                    <span className="text-[10px] text-muted-foreground">{label}</span>
                    <button
                        className="text-[10px] text-muted-foreground hover:text-foreground px-1 border rounded"
                        onClick={() => {
                            const next = { ...(data || {}) }
                            const status = STATUS_TYPES.find(s => !(s in next)) || STATUS_TYPES[0]
                            next[status] = { "1": 1.5 }
                            setData(next)
                        }}
                    >+</button>
                </div>
                <p className="text-[10px] text-muted-foreground/70 mb-0.5">{description}</p>
                {entries.map(([statusType, stackMap]) => (
                    <div key={statusType} className="flex items-start gap-1 mb-0.5">
                        <select
                            className="h-5 text-xs bg-background border rounded px-1 shrink-0"
                            value={statusType}
                            onChange={e => {
                                const next = { ...data! }
                                const val = next[statusType]
                                delete next[statusType]
                                next[e.target.value] = val
                                setData(next)
                            }}
                        >
                            {STATUS_TYPES.map(s => <option key={s} value={s}>{s}</option>)}
                        </select>
                        <div className="flex-1">
                            {Object.entries(stackMap).map(([stacks, mult]) => (
                                <div key={stacks} className="flex items-center gap-1 mb-0.5">
                                    <Input
                                        type="number"
                                        className="h-5 w-10 text-xs px-1"
                                        title="Stack count threshold"
                                        value={stacks}
                                        onChange={e => {
                                            const newMap = { ...stackMap }
                                            delete newMap[stacks]
                                            newMap[e.target.value] = mult
                                            const next = { ...data! }
                                            next[statusType] = newMap
                                            setData(next)
                                        }}
                                    />
                                    <span className="text-[10px] text-muted-foreground">x</span>
                                    <Input
                                        type="number"
                                        className="h-5 w-14 text-xs px-1"
                                        title="Damage multiplier"
                                        step="0.1"
                                        value={mult}
                                        onChange={e => {
                                            const newMap = { ...stackMap, [stacks]: +e.target.value }
                                            const next = { ...data! }
                                            next[statusType] = newMap
                                            setData(next)
                                        }}
                                    />
                                    <button
                                        className="text-muted-foreground hover:text-destructive p-0.5"
                                        onClick={() => {
                                            const newMap = { ...stackMap }
                                            delete newMap[stacks]
                                            const next = { ...data! }
                                            if (Object.keys(newMap).length === 0) {
                                                delete next[statusType]
                                            } else {
                                                next[statusType] = newMap
                                            }
                                            setData(Object.keys(next).length > 0 ? next : undefined)
                                        }}
                                    >
                                        <X className="h-3 w-3" />
                                    </button>
                                </div>
                            ))}
                            <button
                                className="text-[10px] text-muted-foreground hover:text-foreground"
                                onClick={() => {
                                    const maxStack = Math.max(0, ...Object.keys(stackMap).map(Number))
                                    const newMap = { ...stackMap, [String(maxStack + 1)]: 1.5 }
                                    const next = { ...data! }
                                    next[statusType] = newMap
                                    setData(next)
                                }}
                            >
                                + stack threshold
                            </button>
                        </div>
                    </div>
                ))}
            </div>
        )
    }

    const current = scaling || {}

    return (
        <div className="border border-border rounded p-1.5 space-y-1">
            <div className="flex items-center justify-between">
                <span className="text-[10px] text-muted-foreground uppercase font-medium">Damage Scaling</span>
                <button
                    className="text-[10px] text-muted-foreground hover:text-destructive"
                    onClick={() => { onChange(null); setOpen(false) }}
                >
                    remove
                </button>
            </div>
            <p className="text-[10px] text-muted-foreground/70">
                Multiply damage based on status stacks. At each stack threshold, the highest matching multiplier applies.
            </p>
            {renderScalingGroup(
                "Target status",
                "Scale based on the opponent's status stacks",
                current.target_status,
                d => onChange({ ...current, target_status: d }),
            )}
            {renderScalingGroup(
                "Source status",
                "Scale based on your own status stacks",
                current.source_status,
                d => onChange({ ...current, source_status: d }),
            )}
        </div>
    )
}

function AnimBadge({ anim, role }: { anim: { type: string; speed?: number; repetitions?: number }; role: "source" | "target" }) {
    const info = ANIM_TYPE_INFO[anim.type]
    const label = info?.label || anim.type
    const icon = role === "source" ? <Swords className="h-2.5 w-2.5" /> : <Shield className="h-2.5 w-2.5" />
    const reps = anim.repetitions && anim.repetitions > 1 ? `x${anim.repetitions}` : ""
    const spd = anim.speed && anim.speed !== 1 ? `@${anim.speed}x` : ""
    const suffix = [reps, spd].filter(Boolean).join(" ")
    return (
        <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-[10px] font-medium bg-accent-teal-tint text-accent-teal border border-accent-teal-edge">
            {icon} {label}{suffix && ` ${suffix}`}
        </span>
    )
}

function AnimationEditor({
    sourceAnims, targetAnims, onChange,
}: {
    sourceAnims: { type: string; speed?: number; repetitions?: number }[]
    targetAnims: { type: string; speed?: number; repetitions?: number }[]
    onChange: (source: typeof sourceAnims, target: typeof targetAnims) => void
}) {
    const updateEntry = (list: typeof sourceAnims, idx: number, patch: Partial<typeof sourceAnims[0]>) => {
        const next = [...list]
        next[idx] = { ...next[idx], ...patch }
        return next
    }

    const moveEntry = (list: typeof sourceAnims, from: number, to: number) => {
        if (to < 0 || to >= list.length) return list
        const next = [...list]
        const [moved] = next.splice(from, 1)
        next.splice(to, 0, moved)
        return next
    }

    const renderEntry = (entry: typeof sourceAnims[0], idx: number, list: typeof sourceAnims, setList: (l: typeof sourceAnims) => void) => {
        const info = ANIM_TYPE_INFO[entry.type]
        const effectiveDuration = info
            ? `${(parseFloat(info.baseDuration) / (entry.speed || 1)).toFixed(2)}s`
            : null

        return (
            <div key={idx} className="rounded border border-border bg-background">
                <div className="flex items-center gap-1 px-1.5 py-1">
                    <span className="font-mono text-[10px] text-muted-foreground w-3 text-right shrink-0">{idx}</span>
                    <select
                        className="h-5 text-xs bg-background border rounded px-1 flex-1 min-w-0"
                        value={ANIM_TYPES.includes(entry.type) ? entry.type : "__custom"}
                        onChange={e => {
                            if (e.target.value === "__custom") return
                            setList(updateEntry(list, idx, { type: e.target.value }))
                        }}
                    >
                        {ANIM_TYPES.map(a => (
                            <option key={a} value={a}>{ANIM_TYPE_INFO[a].label} - {ANIM_TYPE_INFO[a].motion}</option>
                        ))}
                        {!ANIM_TYPES.includes(entry.type) && (
                            <option value="__custom">{entry.type} (custom)</option>
                        )}
                    </select>
                    <div className="flex gap-0.5 shrink-0">
                        <button
                            className="text-muted-foreground hover:text-foreground disabled:opacity-30 p-0.5"
                            disabled={idx === 0}
                            onClick={() => setList(moveEntry(list, idx, idx - 1))}
                        >
                            <ChevronUp className="h-3 w-3" />
                        </button>
                        <button
                            className="text-muted-foreground hover:text-foreground disabled:opacity-30 p-0.5"
                            disabled={idx === list.length - 1}
                            onClick={() => setList(moveEntry(list, idx, idx + 1))}
                        >
                            <ChevronDown className="h-3 w-3" />
                        </button>
                        <button
                            className="text-muted-foreground hover:text-destructive p-0.5"
                            onClick={() => setList(list.filter((_, i) => i !== idx))}
                        >
                            <X className="h-3 w-3" />
                        </button>
                    </div>
                </div>
                {info && (
                    <p className="text-[10px] text-muted-foreground/60 px-1.5 pb-1 -mt-0.5">{info.description}</p>
                )}
                {!ANIM_TYPES.includes(entry.type) && (
                    <div className="px-1.5 pb-1">
                        <input
                            className="h-5 text-xs bg-background border rounded px-1 w-full"
                            value={entry.type}
                            placeholder="custom type name"
                            onChange={e => setList(updateEntry(list, idx, { type: e.target.value }))}
                        />
                        <p className="text-[10px] text-muted-foreground/60 mt-0.5">Custom animation type. Must match a registered transform in the combat renderer.</p>
                    </div>
                )}
                <div className="flex items-center gap-2 px-1.5 pb-1.5">
                    <div className="flex-1">
                        <label className="text-[10px] text-muted-foreground">Speed</label>
                        <div className="flex items-center gap-1">
                            <Input
                                type="number"
                                className="h-5 text-xs px-1 flex-1"
                                placeholder="1.0"
                                step="0.1"
                                min="0.1"
                                value={entry.speed || ""}
                                onChange={e => setList(updateEntry(list, idx, { speed: +e.target.value || undefined }))}
                            />
                            {effectiveDuration && (
                                <span className="text-[10px] text-muted-foreground font-mono whitespace-nowrap">{effectiveDuration}</span>
                            )}
                        </div>
                        <p className="text-[10px] text-muted-foreground/60">
                            Multiplier on playback rate. 2 = twice as fast, 0.5 = half speed.{info ? ` Base: ${info.baseDuration}.` : ""}
                        </p>
                    </div>
                    <div className="w-16">
                        <label className="text-[10px] text-muted-foreground">Reps</label>
                        <Input
                            type="number"
                            className="h-5 text-xs px-1"
                            placeholder="1"
                            min="1"
                            value={entry.repetitions && entry.repetitions > 1 ? entry.repetitions : ""}
                            onChange={e => {
                                const reps = +e.target.value
                                setList(updateEntry(list, idx, { repetitions: reps > 1 ? reps : undefined }))
                            }}
                        />
                        <p className="text-[10px] text-muted-foreground/60">Chain N times</p>
                    </div>
                </div>
            </div>
        )
    }

    const renderList = (
        role: "source" | "target",
        label: string,
        description: string,
        emptyHint: string,
        list: typeof sourceAnims,
        setList: (l: typeof sourceAnims) => void,
    ) => (
        <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-1">
                    {role === "source"
                        ? <Swords className="h-3 w-3 text-accent-teal" />
                        : <Shield className="h-3 w-3 text-accent-teal" />
                    }
                    <span className="text-[10px] font-medium text-muted-foreground">{label}</span>
                </div>
                <button
                    className="text-[10px] text-muted-foreground hover:text-foreground px-1.5 py-0.5 border rounded flex items-center gap-0.5"
                    onClick={() => setList([...list, { type: role === "source" ? "pounce" : "recoil" }])}
                >
                    <Plus className="h-2.5 w-2.5" /> Add
                </button>
            </div>
            <p className="text-[10px] text-muted-foreground/60 mb-1">{description}</p>
            {list.length === 0 ? (
                <p className="text-[10px] text-muted-foreground/40 italic py-2 text-center border border-dashed border-border rounded">{emptyHint}</p>
            ) : (
                <div className="space-y-1">
                    {list.map((entry, i) => renderEntry(entry, i, list, setList))}
                </div>
            )}
        </div>
    )

    return (
        <div>
            <label className="text-[10px] text-muted-foreground uppercase">Animation</label>
            <p className="text-[10px] text-muted-foreground/70 mt-0.5 mb-1.5">
                Sprite transforms that play during this tick. Source is the attacker, target is the defender. Transforms within each list play sequentially (chained when repetitions &gt; 1).
            </p>
            <div className="flex gap-2">
                {renderList(
                    "source", "Source (attacker)",
                    "Transforms applied to the skill user's sprite",
                    "No source animation - attacker stays still",
                    sourceAnims,
                    s => onChange(s, targetAnims),
                )}
                {renderList(
                    "target", "Target (defender)",
                    "Transforms applied to the opponent's sprite",
                    "No target animation - defender stays still",
                    targetAnims,
                    t => onChange(sourceAnims, t),
                )}
            </div>
        </div>
    )
}

function EffectBadge({ effect }: { effect: TickEffect }) {
    if (effect.damage) {
        const d = effect.damage
        const label = d.variance ? `${d.amount}±${d.variance} dmg` : `${d.amount} dmg`
        const scaled = d.scaled_by ? " *" : ""
        return (
            <Badge variant="destructive" className="text-[10px] px-1.5 py-0">
                {label}{scaled}
            </Badge>
        )
    }
    if (effect.status) {
        const s = effect.status
        const colorClass = statusColors[s.type] ?? "text-foreground"
        return (
            <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium bg-muted border border-border ${colorClass}`}>
                {s.type} +{s.stacks} ({s.target})
                {s.require === "existing_stacks" && " [req]"}
            </span>
        )
    }
    return null
}
