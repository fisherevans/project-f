import { useState, useEffect, useMemo } from "react"
import { useLocation, useNavigate } from "react-router-dom"
import { useScript, useSaveScript, useScriptSchema } from "@/api/scripts"
import { Button } from "@/components/ui/button"
import { ArrowLeft, Save, ChevronDown, ChevronRight } from "lucide-react"
import type { ScriptSchema, StepKindDef, CallableDef, ConditionDef, EventHookDef, ParamDef } from "@/types/scripts"

function ParamList({ params }: { params?: ParamDef[] }) {
    if (!params || params.length === 0) return <span className="text-xs text-muted-foreground italic">no params</span>
    return (
        <div className="mt-1 space-y-0.5">
            {params.map((p) => (
                <div key={p.name} className="flex items-baseline gap-1.5 text-xs">
                    <code className="rounded bg-muted px-1 font-mono text-[11px]">{p.name}</code>
                    <span className="text-muted-foreground">{p.type}</span>
                    {p.required && <span className="text-destructive">required</span>}
                    {p.default !== undefined && (
                        <span className="text-muted-foreground">= {JSON.stringify(p.default)}</span>
                    )}
                    {p.enum && (
                        <span className="text-muted-foreground">[{p.enum.join(", ")}]</span>
                    )}
                    {p.description && (
                        <span className="text-muted-foreground/70">- {p.description}</span>
                    )}
                </div>
            ))}
        </div>
    )
}

function SchemaSection({ title, children, defaultOpen = false }: { title: string; children: React.ReactNode; defaultOpen?: boolean }) {
    const [open, setOpen] = useState(defaultOpen)
    return (
        <div className="border-b border-border">
            <button
                onClick={() => setOpen(!open)}
                className="flex w-full items-center gap-1.5 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wider text-muted-foreground hover:bg-accent/50"
            >
                {open ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                {title}
            </button>
            {open && <div className="px-3 pb-3">{children}</div>}
        </div>
    )
}

function StepKindEntry({ def }: { def: StepKindDef }) {
    return (
        <div className="mb-2">
            <div className="flex items-baseline gap-2">
                <code className="font-mono text-xs font-semibold">{def.name}</code>
                <span className="rounded bg-muted px-1 py-0.5 text-[10px] text-muted-foreground">{def.category}</span>
                <span className="rounded bg-muted px-1 py-0.5 text-[10px] text-muted-foreground">{def.paramStyle}</span>
            </div>
            <p className="text-xs text-muted-foreground">{def.description}</p>
            <ParamList params={def.params} />
        </div>
    )
}

function CallableEntry({ def }: { def: CallableDef | ConditionDef }) {
    return (
        <div className="mb-2">
            <code className="font-mono text-xs font-semibold">{def.name}</code>
            <p className="text-xs text-muted-foreground">{def.description}</p>
            <ParamList params={def.params} />
        </div>
    )
}

function EventHookEntry({ def }: { def: EventHookDef }) {
    return (
        <div className="mb-2">
            <div className="flex items-baseline gap-2">
                <code className="font-mono text-xs font-semibold">{def.yamlKey}</code>
                <span className="text-[10px] text-muted-foreground">{def.eventType}</span>
            </div>
            <p className="text-xs text-muted-foreground">{def.description}</p>
            {def.filterFields && def.filterFields.length > 0 && (
                <div className="mt-1">
                    <span className="text-[10px] font-medium text-muted-foreground">Filter fields:</span>
                    <ParamList params={def.filterFields} />
                </div>
            )}
        </div>
    )
}

function SchemaReference({ schema }: { schema: ScriptSchema }) {
    const stepsByCategory = useMemo(() => {
        const map = new Map<string, StepKindDef[]>()
        for (const def of Object.values(schema.stepKinds)) {
            const cat = def.category
            if (!map.has(cat)) map.set(cat, [])
            map.get(cat)!.push(def)
        }
        for (const defs of map.values()) {
            defs.sort((a, b) => a.name.localeCompare(b.name))
        }
        return map
    }, [schema])

    return (
        <div className="h-full overflow-auto text-sm">
            <SchemaSection title="Step Kinds" defaultOpen={true}>
                {Array.from(stepsByCategory.entries())
                    .sort(([a], [b]) => a.localeCompare(b))
                    .map(([category, defs]) => (
                        <div key={category} className="mb-3">
                            <div className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60">
                                {category}
                            </div>
                            {defs.map((def) => (
                                <StepKindEntry key={def.name} def={def} />
                            ))}
                        </div>
                    ))}
            </SchemaSection>

            <SchemaSection title="Named Actions">
                {Object.values(schema.actions)
                    .sort((a, b) => a.name.localeCompare(b.name))
                    .map((def) => (
                        <CallableEntry key={def.name} def={def} />
                    ))}
            </SchemaSection>

            <SchemaSection title="Named Conditions">
                {Object.values(schema.conditions)
                    .sort((a, b) => a.name.localeCompare(b.name))
                    .map((def) => (
                        <CallableEntry key={def.name} def={def} />
                    ))}
            </SchemaSection>

            <SchemaSection title="Built-in Conditions">
                {Object.values(schema.builtinConditions)
                    .sort((a, b) => a.name.localeCompare(b.name))
                    .map((def) => (
                        <CallableEntry key={def.name} def={def} />
                    ))}
            </SchemaSection>

            <SchemaSection title="Event Hooks">
                {Object.values(schema.eventHooks)
                    .sort((a, b) => a.yamlKey.localeCompare(b.yamlKey))
                    .map((def) => (
                        <EventHookEntry key={def.yamlKey} def={def} />
                    ))}
            </SchemaSection>

            <SchemaSection title="Template Variables">
                <div className="space-y-1">
                    {schema.templateVars.map((v) => (
                        <div key={v.pattern}>
                            <code className="font-mono text-xs font-semibold">{v.pattern}</code>
                            <p className="text-xs text-muted-foreground">{v.description}</p>
                        </div>
                    ))}
                </div>
            </SchemaSection>
        </div>
    )
}

export function ScriptEditor() {
    const location = useLocation()
    const navigate = useNavigate()
    const path = location.pathname.replace(/^\/scripts\//, "")

    const { data: script, isLoading, error } = useScript(path)
    const { data: schema } = useScriptSchema()
    const saveScript = useSaveScript()

    const [content, setContent] = useState("")
    const [dirty, setDirty] = useState(false)

    useEffect(() => {
        if (script) {
            setContent(script.rawYaml)
            setDirty(false)
        }
    }, [script])

    const handleSave = () => {
        saveScript.mutate(
            { path, content },
            { onSuccess: () => setDirty(false) },
        )
    }

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if ((e.metaKey || e.ctrlKey) && e.key === "s") {
            e.preventDefault()
            if (dirty) handleSave()
        }
    }

    if (isLoading) return <div className="p-4 text-muted-foreground">Loading...</div>
    if (error) return <div className="p-4 text-destructive">Error: {error.message}</div>

    return (
        <div className="flex h-full flex-col">
            <div className="flex items-center gap-2 border-b border-border px-3 py-2">
                <Button variant="ghost" size="sm" onClick={() => navigate("/scripts")}>
                    <ArrowLeft className="mr-1 h-3 w-3" />
                    Back
                </Button>
                <div className="flex-1">
                    <span className="text-sm font-medium">{path}</span>
                    {script && (
                        <span className="ml-2 text-xs text-muted-foreground">
                            {script.handlerCount} handler{script.handlerCount !== 1 ? "s" : ""}
                        </span>
                    )}
                </div>
                <Button
                    size="sm"
                    disabled={!dirty || saveScript.isPending}
                    onClick={handleSave}
                >
                    <Save className="mr-1 h-3 w-3" />
                    {saveScript.isPending ? "Saving..." : "Save"}
                </Button>
            </div>
            <div className="flex flex-1 overflow-hidden">
                <div className="flex-1 overflow-hidden">
                    <textarea
                        className="h-full w-full resize-none bg-background p-4 font-mono text-xs leading-relaxed outline-none"
                        value={content}
                        onChange={(e) => {
                            setContent(e.target.value)
                            setDirty(true)
                        }}
                        onKeyDown={handleKeyDown}
                        spellCheck={false}
                    />
                </div>
                {schema && (
                    <div className="w-80 shrink-0 border-l border-border overflow-hidden">
                        <SchemaReference schema={schema} />
                    </div>
                )}
            </div>
        </div>
    )
}
