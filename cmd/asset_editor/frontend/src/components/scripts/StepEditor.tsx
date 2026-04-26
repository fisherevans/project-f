import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ChevronDown, ChevronRight, X, ArrowUp, ArrowDown, Plus } from "lucide-react";
import { getCategoryColor } from "./StepKindPicker";
import { StepParamForm } from "./StepParamForm";
import { StepList } from "./StepList";
import { ExpressionInput } from "./inputs/ExpressionInput";
import { useExprContext } from "./ExprContext";
import { getStepSubSteps, setStepSubSteps, parseSteps, serializeSteps } from "@/lib/scriptUtils";
import type { StepNode, ScriptSchema, StepKindDef } from "@/types/scripts";

interface StepEditorProps {
    step: StepNode;
    stepIndex: number;
    schema: ScriptSchema;
    onChange: (step: StepNode) => void;
    onRemove: () => void;
    onMoveUp?: () => void;
    onMoveDown?: () => void;
}

function getStepSummary(step: StepNode): string {
    if (step.kind === "return") return "";
    if (typeof step.params === "string") return step.params;
    if (typeof step.params === "number") return String(step.params);
    if (typeof step.params === "boolean") return String(step.params);
    if (typeof step.params !== "object" || step.params === null) return "";
    const m = step.params as Record<string, unknown>;
    switch (step.kind) {
        case "action":
        case "custom_action":
        case "ref":
            return String(m.name ?? "");
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
        case "focused_sequence":
            return String(m.target ?? m.entity ?? "");
        case "teleport_player":
            return String(m.to_entity ?? m.to_reference ?? "");
        case "pick_dialogue":
        case "pick_chatter":
        case "pick_self_dialogue":
            return String(m.list ?? "");
        default:
            return "";
    }
}

function hasExpandableContent(step: StepNode, stepDef?: StepKindDef): boolean {
    if (!stepDef) return typeof step.params === "object" && step.params !== null;
    if (stepDef.paramStyle === "list") return true;
    if (stepDef.paramStyle === "map" || stepDef.paramStyle === "string_or_map") {
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

export function StepEditor({ step, stepIndex, schema, onChange, onRemove, onMoveUp, onMoveDown }: StepEditorProps) {
    const stepDef = schema.stepKinds[step.kind];
    const category = stepDef?.category ?? "unknown";
    const expandable = hasExpandableContent(step, stepDef);
    const [expanded, setExpanded] = useState(expandable && (stepDef?.acceptsSubSteps || stepDef?.paramStyle === "list"));

    const summary = getStepSummary(step);
    const subStepGroups = getStepSubSteps(step);
    const subStepKeys = new Set(subStepGroups.map((g) => g.key));
    const isList = stepDef?.paramStyle === "list";

    const isSimpleInline = stepDef && (stepDef.paramStyle === "string" || stepDef.paramStyle === "number" ||
        (stepDef.paramStyle === "string_or_map" && (typeof step.params === "string" || typeof step.params === "number")));

    const borderColor = CATEGORY_BORDER_COLORS[category] ?? "border-l-border";

    return (
        <div className={`rounded border border-border/40 border-l-2 ${borderColor} bg-background`}>
            <div className="flex items-center gap-1 px-1.5 py-1 min-h-[28px]">
                <span className="text-[10px] text-muted-foreground/50 w-4 shrink-0 text-right tabular-nums">{stepIndex + 1}</span>
                {expandable && (
                    <button className="shrink-0 text-muted-foreground hover:text-foreground" onClick={() => setExpanded(!expanded)}>
                        {expanded ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                    </button>
                )}
                <Badge variant="outline" className={`shrink-0 text-[10px] px-1.5 py-0 font-mono ${getCategoryColor(category)}`}>
                    {step.kind}
                </Badge>
                {isSimpleInline ? (
                    <div className="flex-1 min-w-0">
                        <StepParamForm stepKind={stepDef} params={step.params} onChange={(params) => onChange({ ...step, params })} schema={schema} />
                    </div>
                ) : (
                    summary && !expanded && (
                        <span className="flex-1 min-w-0 truncate text-xs text-muted-foreground font-mono">{summary}</span>
                    )
                )}
                {!isSimpleInline && !summary && !expandable && (
                    <span className="flex-1" />
                )}
                {(isSimpleInline || summary || !expandable) && <span className="flex-1 min-w-0" />}
                <div className="flex items-center shrink-0">
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
                </div>
            )}
        </div>
    );
}
