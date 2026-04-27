import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ChevronDown, ChevronRight, X, ArrowUp, ArrowDown, Plus, BookOpen, Copy, ClipboardPaste } from "lucide-react";
import { getCategoryColor } from "./StepKindPicker";
import { ReferencePopover } from "./ReferencePopover";
import { StepParamForm } from "./StepParamForm";
import { StepList } from "./StepList";
import { ExpressionInput } from "./inputs/ExpressionInput";
import { useExprContext } from "./ExprContext";
import { useScriptClipboard } from "./ScriptClipboard";
import { getStepSubSteps, setStepSubSteps, parseSteps, serializeSteps } from "@/lib/scriptUtils";
import type { StepNode, ScriptSchema, StepKindDef, CustomActionDef } from "@/types/scripts";

interface StepEditorProps {
    step: StepNode;
    stepIndex: number;
    schema: ScriptSchema;
    onChange: (step: StepNode) => void;
    onRemove: () => void;
    onMoveUp?: () => void;
    onMoveDown?: () => void;
    onClone?: () => void;
    onPasteBefore?: () => void;
    onPasteAfter?: () => void;
}

function formatParamValue(v: unknown): string {
    if (v === null || v === undefined) return "";
    if (typeof v === "string" || typeof v === "number" || typeof v === "boolean") return String(v);
    if (Array.isArray(v)) return `[${v.length}]`;
    return "{...}";
}

function mapParamsSummary(m: Record<string, unknown>, skip?: Set<string>): string {
    const parts: string[] = [];
    for (const [k, v] of Object.entries(m)) {
        if (skip?.has(k)) continue;
        if (v === undefined || v === null || v === "" || v === false) continue;
        if (Array.isArray(v) && (k === "steps" || k === "then" || k === "else" || k === "cases")) continue;
        const fv = formatParamValue(v);
        if (v === true) {
            parts.push(k);
        } else {
            parts.push(`${k}: ${fv}`);
        }
    }
    return parts.join(", ");
}

function getStepSummary(step: StepNode): string {
    if (step.kind === "return") return "";
    if (typeof step.params === "string") return `(${step.params})`;
    if (typeof step.params === "number") return `(${step.params})`;
    if (typeof step.params === "boolean") return `(${step.params})`;
    if (typeof step.params !== "object" || step.params === null) return "";
    const m = step.params as Record<string, unknown>;
    switch (step.kind) {
        case "action":
        case "custom_action":
        case "ref": {
            const name = String(m.name ?? "");
            const rest = mapParamsSummary(m, new Set(["name"]));
            return rest ? `${name}(${rest})` : name;
        }
        case "set_var":
        case "set_run_state":
        case "set_world_state":
            return `${m.key ?? "?"} = ${m.value ?? "?"}`;
        case "if":
            return String(m.when ?? "");
        case "while":
            return m.max ? `${m.when ?? ""} (max ${m.max})` : String(m.when ?? "");
        case "switch":
            return `on ${m.on ?? "?"}` + (Array.isArray(m.cases) ? ` (${m.cases.length} cases)` : "");
        default: {
            const s = mapParamsSummary(m);
            return s ? `(${s})` : "";
        }
    }
}

function hasExpandableContent(step: StepNode, stepDef?: StepKindDef): boolean {
    if (!stepDef) return typeof step.params === "object" && step.params !== null;
    if (stepDef.paramStyle === "list") return true;
    if (stepDef.paramStyle === "string_or_map") {
        return (stepDef.params?.length ?? 0) > 1 || (typeof step.params === "object" && step.params !== null);
    }
    if (stepDef.paramStyle === "map") {
        if (typeof step.params === "object" && step.params !== null && !Array.isArray(step.params)) return true;
    }
    if (stepDef.acceptsSubSteps) return true;
    return false;
}

const CATEGORY_BORDER_COLORS: Record<string, string> = {
    text: "border-l-accent-blue-edge",
    flow: "border-l-accent-violet-edge",
    state: "border-l-accent-amber-edge",
    entity: "border-l-accent-teal-edge",
    camera: "border-l-accent-blue-edge",
    audio: "border-l-accent-violet-edge",
    transition: "border-l-accent-orange-edge",
    rpg: "border-l-accent-red-edge",
    combat: "border-l-accent-red-edge",
};

function CustomActionStepsModal({ action, actionName, schema, onClose }: {
    action: CustomActionDef;
    actionName: string;
    schema: ScriptSchema;
    onClose: () => void;
}) {
    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
            <div className="bg-popover border border-border rounded-lg shadow-xl w-[500px] max-h-[70vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
                <div className="flex items-center justify-between border-b border-border px-3 py-2">
                    <div>
                        <code className="text-sm font-mono font-semibold">{actionName}</code>
                        {action.description && (
                            <div className="text-xs text-muted-foreground mt-0.5">{action.description}</div>
                        )}
                    </div>
                    <button className="text-muted-foreground hover:text-foreground" onClick={onClose}>
                        <X className="h-4 w-4" />
                    </button>
                </div>
                <div className="overflow-y-auto p-3">
                    <StepList steps={action.steps} schema={schema} onChange={() => {}} />
                </div>
            </div>
        </div>
    );
}

function CustomActionInlinePreview({ action, actionName, schema }: {
    action: CustomActionDef;
    actionName: string;
    schema: ScriptSchema;
}) {
    const [showSteps, setShowSteps] = useState(false);

    return (
        <div className="rounded border border-accent-violet-edge/30 bg-accent-violet-tint/30 px-2 py-1.5 space-y-1.5">
            <div className="flex items-center gap-2">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-accent-violet/60">Definition</span>
                {action.description && (
                    <span className="text-[10px] text-muted-foreground truncate">{action.description}</span>
                )}
            </div>
            {action.params && action.params.length > 0 && (
                <div className="flex flex-wrap gap-x-3 gap-y-0.5">
                    {action.params.map((p) => (
                        <span key={p.name} className="text-[10px]">
                            <code className="font-mono text-accent-violet">{p.name}</code>
                            {p.default !== undefined ? (
                                <span className="text-muted-foreground/60"> = {String(p.default)}</span>
                            ) : (
                                <span className="text-accent-amber"> *</span>
                            )}
                        </span>
                    ))}
                </div>
            )}
            <button
                className="text-[10px] text-accent-violet hover:text-accent-violet/80 font-medium"
                onClick={() => setShowSteps(true)}
            >
                View {action.steps.length} step{action.steps.length !== 1 ? "s" : ""}
            </button>
            {showSteps && (
                <CustomActionStepsModal
                    action={action}
                    actionName={actionName}
                    schema={schema}
                    onClose={() => setShowSteps(false)}
                />
            )}
        </div>
    );
}

function SwitchCasesEditor({ step, schema, onChange }: { step: StepNode; schema: ScriptSchema; onChange: (step: StepNode) => void }) {
    const exprCtx = useExprContext();
    const params = step.params as Record<string, unknown>;
    const cases = (Array.isArray(params.cases) ? params.cases : []) as Record<string, unknown>[];

    const addCase = () => {
        const next = [...cases, { value: "''", steps: [] }];
        onChange({ ...step, params: { ...params, cases: next } });
    };

    const removeCase = (idx: number) => {
        const next = cases.filter((_, i) => i !== idx);
        onChange({ ...step, params: { ...params, cases: next } });
    };

    const updateCaseValue = (idx: number, value: string) => {
        const next = [...cases];
        next[idx] = { ...next[idx], value };
        onChange({ ...step, params: { ...params, cases: next } });
    };

    const updateCaseSteps = (idx: number, steps: StepNode[]) => {
        const next = [...cases];
        next[idx] = { ...next[idx], steps: serializeSteps(steps) };
        onChange({ ...step, params: { ...params, cases: next } });
    };

    return (
        <div className="space-y-2">
            <div className="flex items-center justify-between">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60">Cases</span>
                <Button variant="ghost" size="sm" className="h-5 text-[10px] text-muted-foreground" onClick={addCase}>
                    <Plus className="mr-1 h-2.5 w-2.5" />
                    Add case
                </Button>
            </div>
            {cases.map((c, i) => (
                <div key={i} className="ml-1 border-l-2 border-accent-violet-edge pl-2 space-y-1">
                    <div className="flex items-center gap-1.5">
                        <span className="text-[10px] text-muted-foreground shrink-0">value</span>
                        <div className="flex-1">
                            <ExpressionInput
                                value={String(c.value ?? "")}
                                onChange={(v) => updateCaseValue(i, v)}
                                placeholder="Expression (e.g. 'intro')"
                                handlerVarKeys={exprCtx.handlerVarKeys}
                                constKeys={exprCtx.constKeys}
                                otherConstKeys={exprCtx.otherConstKeys}
                            />
                        </div>
                        <button className="text-muted-foreground hover:text-destructive shrink-0" onClick={() => removeCase(i)}>
                            <X className="h-3 w-3" />
                        </button>
                    </div>
                    <StepList
                        steps={Array.isArray(c.steps) ? parseSteps(c.steps as unknown[]) : []}
                        schema={schema}
                        onChange={(steps) => updateCaseSteps(i, steps)}
                    />
                </div>
            ))}
        </div>
    );
}

export function StepEditor({ step, stepIndex, schema, onChange, onRemove, onMoveUp, onMoveDown, onClone, onPasteBefore, onPasteAfter }: StepEditorProps) {
    const exprCtx = useExprContext();
    const clipboard = useScriptClipboard();
    const [showActions, setShowActions] = useState(false);
    const stepDef = schema.stepKinds[step.kind];
    const category = stepDef?.category ?? "unknown";
    const expandable = hasExpandableContent(step, stepDef);
    const [expanded, setExpanded] = useState(expandable && (stepDef?.acceptsSubSteps || stepDef?.paramStyle === "list"));

    const summary = getStepSummary(step);
    const subStepGroups = getStepSubSteps(step);
    const subStepKeys = new Set(subStepGroups.map((g) => g.key));
    const isList = stepDef?.paramStyle === "list";

    const isSimpleInline = stepDef && (stepDef.paramStyle === "string" || stepDef.paramStyle === "number");

    const borderColor = CATEGORY_BORDER_COLORS[category] ?? "border-l-border";

    const actionName = step.kind === "custom_action"
        ? typeof step.params === "string" ? step.params : String((step.params as Record<string, unknown>)?.name ?? "")
        : "";
    const actionDef = actionName ? exprCtx.customActions[actionName] : undefined;

    return (
        <div className={`rounded border border-border/40 border-l-2 ${borderColor} bg-background`}>
            <div className="flex items-center gap-1 px-1.5 py-1 min-h-[28px]">
                <span className="text-[10px] text-muted-foreground/50 w-4 shrink-0 text-right tabular-nums">{stepIndex + 1}</span>
                {expandable && (
                    <button className="shrink-0 text-muted-foreground hover:text-foreground" onClick={() => {
                        if (!expanded && stepDef?.paramStyle === "string_or_map" && typeof step.params !== "object") {
                            const mapParams: Record<string, unknown> = {};
                            const firstParam = stepDef.params?.[0];
                            if (firstParam && step.params != null) mapParams[firstParam.name] = step.params;
                            onChange({ ...step, params: mapParams });
                        }
                        setExpanded(!expanded);
                    }}>
                        {expanded ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                    </button>
                )}
                <ReferencePopover target={{ type: "step", name: step.kind }}>
                    <Badge variant="outline" className={`text-[10px] px-1.5 py-0 font-mono hover:ring-1 hover:ring-accent-violet/30 cursor-help ${getCategoryColor(category)}`}>
                        {step.kind}
                        <BookOpen className="ml-0.5 h-2.5 w-2.5 opacity-30" />
                    </Badge>
                </ReferencePopover>
                {isSimpleInline ? (
                    <div className="flex-1 min-w-0">
                        <StepParamForm stepKind={stepDef} params={step.params} onChange={(params) => onChange({ ...step, params })} schema={schema} />
                    </div>
                ) : (
                    summary && !expanded && (
                        <span className="flex-1 min-w-0 truncate text-[11px] text-muted-foreground font-mono">
                            {summary}
                            {(step.kind === "action" || step.kind === "ref") && summary && (
                                <span className="inline-flex ml-1 align-middle" onClick={(e) => e.stopPropagation()}>
                                    <ReferencePopover target={{ type: "action", name: summary }}>
                                        <BookOpen className="h-2.5 w-2.5 text-muted-foreground/30 hover:text-accent-violet cursor-help" />
                                    </ReferencePopover>
                                </span>
                            )}
                        </span>
                    )
                )}
                {!isSimpleInline && !summary && !expandable && (
                    <span className="flex-1" />
                )}
                {(isSimpleInline || summary || !expandable) && <span className="flex-1 min-w-0" />}
                {actionDef && !expanded && (
                    <span className="text-[10px] text-muted-foreground/50 truncate shrink min-w-0">
                        {actionDef.description}
                    </span>
                )}
                <div className="relative flex items-center shrink-0">
                    {onClone && (
                        <Button variant="ghost" size="sm" className="h-5 w-5 p-0 opacity-0 group-hover/step:opacity-100 text-muted-foreground hover:text-foreground" onClick={onClone} title="Clone step">
                            <Copy className="h-3 w-3" />
                        </Button>
                    )}
                    {(onPasteBefore || onPasteAfter) && (
                        <div className="relative">
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-5 w-5 p-0 opacity-0 group-hover/step:opacity-100 text-accent-violet hover:text-accent-violet"
                                onClick={() => setShowActions(!showActions)}
                                title="Paste options"
                            >
                                <ClipboardPaste className="h-3 w-3" />
                            </Button>
                            {showActions && (
                                <>
                                    <div className="fixed inset-0 z-40" onClick={() => setShowActions(false)} />
                                    <div className="absolute right-0 top-full mt-1 z-50 bg-popover border border-border rounded-md shadow-lg py-0.5 min-w-[120px]">
                                        {onPasteBefore && (
                                            <button
                                                className="flex w-full items-center gap-1.5 px-2 py-1 text-xs hover:bg-accent text-left"
                                                onClick={() => { onPasteBefore(); setShowActions(false); }}
                                            >
                                                Paste before
                                            </button>
                                        )}
                                        {onPasteAfter && (
                                            <button
                                                className="flex w-full items-center gap-1.5 px-2 py-1 text-xs hover:bg-accent text-left"
                                                onClick={() => { onPasteAfter(); setShowActions(false); }}
                                            >
                                                Paste after
                                            </button>
                                        )}
                                        <button
                                            className="flex w-full items-center gap-1.5 px-2 py-1 text-xs hover:bg-accent text-left"
                                            onClick={() => { clipboard.copyStep(step); setShowActions(false); }}
                                        >
                                            Copy step
                                        </button>
                                    </div>
                                </>
                            )}
                        </div>
                    )}
                    {!onClone && !onPasteBefore && !onPasteAfter && (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-5 w-5 p-0 opacity-0 group-hover/step:opacity-100 text-muted-foreground hover:text-foreground"
                            onClick={() => clipboard.copyStep(step)}
                            title="Copy step"
                        >
                            <ClipboardPaste className="h-3 w-3" />
                        </Button>
                    )}
                    {onMoveUp && (
                        <Button variant="ghost" size="sm" className="h-5 w-5 p-0 opacity-0 group-hover/step:opacity-100 text-muted-foreground hover:text-foreground" onClick={onMoveUp}>
                            <ArrowUp className="h-3 w-3" />
                        </Button>
                    )}
                    {onMoveDown && (
                        <Button variant="ghost" size="sm" className="h-5 w-5 p-0 opacity-0 group-hover/step:opacity-100 text-muted-foreground hover:text-foreground" onClick={onMoveDown}>
                            <ArrowDown className="h-3 w-3" />
                        </Button>
                    )}
                    <Button variant="ghost" size="sm" className="h-5 w-5 p-0 opacity-0 group-hover/step:opacity-100 text-muted-foreground hover:text-destructive" onClick={onRemove}>
                        <X className="h-3 w-3" />
                    </Button>
                </div>
            </div>
            {expanded && isList && (
                <div className="border-t border-border/30 px-2 py-1.5">
                    <StepList
                        steps={Array.isArray(step.params) ? parseSteps(step.params as unknown[]) : []}
                        schema={schema}
                        onChange={(steps) => onChange({ ...step, params: serializeSteps(steps) })}
                    />
                </div>
            )}
            {expanded && !isList && !isSimpleInline && stepDef && (
                <div className="border-t border-border/30 px-2 py-1.5 space-y-2">
                    <StepParamForm
                        stepKind={stepDef}
                        params={step.params}
                        onChange={(params) => onChange({ ...step, params })}
                        excludeKeys={step.kind === "switch" ? new Set([...subStepKeys, "cases"]) : subStepKeys}
                        schema={schema}
                    />
                    {step.kind === "switch" && (
                        <SwitchCasesEditor step={step} schema={schema} onChange={onChange} />
                    )}
                    {subStepGroups
                        .filter((g) => !g.key.startsWith("cases["))
                        .map((group) => (
                        <div key={group.key} className="ml-1 border-l-2 border-border/40 pl-2">
                            <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60 mb-1">
                                {group.key.replace(/_/g, " ")}
                            </div>
                            <StepList
                                steps={group.steps}
                                schema={schema}
                                onChange={(steps) => onChange(setStepSubSteps(step, group.key, steps))}
                            />
                        </div>
                    ))}
                    {actionDef && (
                        <CustomActionInlinePreview action={actionDef} actionName={actionName} schema={schema} />
                    )}
                </div>
            )}
        </div>
    );
}
