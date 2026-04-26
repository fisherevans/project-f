import { useState, useRef, useEffect, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { BookOpen } from "lucide-react";
import { useScriptSchema } from "@/api/scripts";
import { getCategoryColor } from "./StepKindPicker";
import { Badge } from "@/components/ui/badge";
import { generateStepExample, generateActionExample, generateConditionExample, generateHookExample } from "@/lib/yamlExamples";
import type { ScriptSchema, StepKindDef, CallableDef, ConditionDef, EventHookDef, ParamDef } from "@/types/scripts";

type RefTarget =
    | { type: "step"; name: string }
    | { type: "action"; name: string }
    | { type: "condition"; name: string }
    | { type: "hook"; yamlKey: string };

function anchorFor(target: RefTarget): string {
    switch (target.type) {
        case "step": return `step-${target.name}`;
        case "action": return `action-${target.name}`;
        case "condition": return `condition-${target.name}`;
        case "hook": return `hook-${target.yamlKey}`;
    }
}

export function ReferencePopover({ target, children }: { target: RefTarget; children: ReactNode }) {
    const [open, setOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);
    const { data: schema } = useScriptSchema();

    useEffect(() => {
        if (!open) return;
        function handleClickOutside(e: MouseEvent) {
            if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
                setOpen(false);
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, [open]);

    useEffect(() => {
        if (!open) return;
        function handleKeyDown(e: KeyboardEvent) {
            if (e.key === "Escape") setOpen(false);
        }
        document.addEventListener("keydown", handleKeyDown);
        return () => document.removeEventListener("keydown", handleKeyDown);
    }, [open]);

    const content = schema ? renderContent(target, schema) : null;

    return (
        <div ref={containerRef} className="relative inline-flex">
            <button
                type="button"
                className="inline-flex items-center"
                onClick={(e) => { e.preventDefault(); e.stopPropagation(); setOpen(!open); }}
            >
                {children}
            </button>
            {open && content && (
                <div className="absolute left-0 top-full z-50 mt-1 w-[380px] rounded-md border border-border bg-popover shadow-lg overflow-hidden">
                    <div className="max-h-[50vh] overflow-y-auto">
                        {content}
                    </div>
                    <div className="border-t border-border/50 px-3 py-1.5 bg-muted/20">
                        <Link
                            to={`/reference#${anchorFor(target)}`}
                            className="text-[10px] text-accent-violet hover:underline flex items-center gap-1"
                            onClick={() => setOpen(false)}
                        >
                            <BookOpen className="h-2.5 w-2.5" />
                            View full reference page
                        </Link>
                    </div>
                </div>
            )}
        </div>
    );
}

function renderContent(target: RefTarget, schema: ScriptSchema): ReactNode {
    switch (target.type) {
        case "step": {
            const def = schema.stepKinds[target.name];
            if (!def) return <NotFound name={target.name} />;
            return <StepPopover def={def} />;
        }
        case "action": {
            const def = schema.actions[target.name];
            if (!def) return <NotFound name={target.name} />;
            return <ActionPopover def={def} />;
        }
        case "condition": {
            const def = schema.conditions[target.name] ?? schema.builtinConditions[target.name];
            if (!def) return <NotFound name={target.name} />;
            return <ConditionPopover def={def} />;
        }
        case "hook": {
            const def = schema.eventHooks[target.yamlKey];
            if (!def) return <NotFound name={target.yamlKey} />;
            return <HookPopover def={def} />;
        }
    }
}

function NotFound({ name }: { name: string }) {
    return <div className="px-3 py-4 text-xs text-muted-foreground text-center">No reference found for <code className="font-mono">{name}</code></div>;
}

function CompactParamTable({ params }: { params: ParamDef[] }) {
    if (!params || params.length === 0) return null;
    return (
        <div className="space-y-0.5">
            {params.map(p => (
                <div key={p.name} className="flex items-baseline gap-1.5 text-[11px]">
                    <code className="font-mono text-accent-violet shrink-0">{p.name}</code>
                    {p.required && <span className="text-accent-red text-[9px]">*</span>}
                    <span className="text-muted-foreground/50 shrink-0">{p.type}</span>
                    {p.default !== undefined && <span className="text-muted-foreground/40">= {String(p.default)}</span>}
                    {p.description && <span className="text-muted-foreground truncate">{p.description}</span>}
                </div>
            ))}
        </div>
    );
}

function MiniYaml({ code }: { code: string }) {
    return (
        <pre className="bg-muted/30 text-[10px] font-mono leading-relaxed px-2 py-1.5 rounded border border-border/30 overflow-x-auto text-foreground/80">
            {code}
        </pre>
    );
}

function StepPopover({ def }: { def: StepKindDef }) {
    const categoryColor = getCategoryColor(def.category);
    const nonStepParams = def.params?.filter(p => p.type !== "steps") ?? [];
    const subStepParams = def.params?.filter(p => p.type === "steps") ?? [];

    return (
        <div className="px-3 py-2 space-y-2">
            <div className="flex items-center gap-2">
                <code className="text-xs font-mono font-semibold">{def.name}</code>
                <Badge variant="outline" className={`text-[10px] px-1.5 py-0 ${categoryColor}`}>{def.category}</Badge>
                {def.acceptsSubSteps && (
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0 text-muted-foreground">sub-steps</Badge>
                )}
            </div>
            <p className="text-[11px] text-muted-foreground leading-snug">{def.description}</p>
            {nonStepParams.length > 0 && (
                <div>
                    <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/50 mb-0.5">Parameters</div>
                    <CompactParamTable params={nonStepParams} />
                </div>
            )}
            {subStepParams.length > 0 && (
                <div className="text-[10px] text-muted-foreground/50">
                    Sub-step fields: {subStepParams.map(p => <code key={p.name} className="font-mono text-accent-violet mx-0.5">{p.name}</code>)}
                </div>
            )}
            <MiniYaml code={generateStepExample(def)} />
        </div>
    );
}

function ActionPopover({ def }: { def: CallableDef }) {
    return (
        <div className="px-3 py-2 space-y-2">
            <div className="flex items-center gap-2">
                <code className="text-xs font-mono font-semibold">{def.name}</code>
                <Badge variant="outline" className="text-[10px] px-1.5 py-0 bg-accent-teal-tint text-accent-teal">action</Badge>
            </div>
            <p className="text-[11px] text-muted-foreground leading-snug">{def.description}</p>
            {def.params && def.params.length > 0 && (
                <div>
                    <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/50 mb-0.5">Parameters</div>
                    <CompactParamTable params={def.params} />
                </div>
            )}
            <MiniYaml code={generateActionExample(def)} />
        </div>
    );
}

function ConditionPopover({ def }: { def: CallableDef | ConditionDef }) {
    const isBuiltin = "isComposite" in def;
    const isComposite = "isComposite" in def && def.isComposite;

    return (
        <div className="px-3 py-2 space-y-2">
            <div className="flex items-center gap-2">
                <code className="text-xs font-mono font-semibold">{def.name}</code>
                {isBuiltin && <Badge variant="outline" className="text-[10px] px-1.5 py-0 bg-accent-amber-tint text-accent-amber">builtin</Badge>}
                {isComposite && <Badge variant="outline" className="text-[10px] px-1.5 py-0 bg-accent-blue-tint text-accent-blue">composite</Badge>}
                {!isBuiltin && <Badge variant="outline" className="text-[10px] px-1.5 py-0 bg-accent-teal-tint text-accent-teal">named</Badge>}
            </div>
            <p className="text-[11px] text-muted-foreground leading-snug">{def.description}</p>
            {def.params && def.params.length > 0 && (
                <div>
                    <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/50 mb-0.5">Parameters</div>
                    <CompactParamTable params={def.params} />
                </div>
            )}
            <MiniYaml code={generateConditionExample(def)} />
        </div>
    );
}

function HookPopover({ def }: { def: EventHookDef }) {
    return (
        <div className="px-3 py-2 space-y-2">
            <div className="flex items-center gap-2">
                <code className="text-xs font-mono font-semibold">{def.yamlKey}</code>
                <span className="text-[10px] text-muted-foreground">{def.name}</span>
                <span className="text-[10px] text-muted-foreground/50 ml-auto font-mono">{def.eventType}</span>
            </div>
            <p className="text-[11px] text-muted-foreground leading-snug">{def.description}</p>
            {def.filterFields && def.filterFields.length > 0 && (
                <div>
                    <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/50 mb-0.5">Filter Fields</div>
                    <CompactParamTable params={def.filterFields} />
                </div>
            )}
            <MiniYaml code={generateHookExample(def)} />
        </div>
    );
}
