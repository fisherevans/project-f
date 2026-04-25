import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ChevronDown, ChevronRight, Trash2, GripVertical } from "lucide-react";
import { getCategoryColor } from "./StepKindPicker";
import { StepParamForm } from "./StepParamForm";
import { StepList } from "./StepList";
import { getStepSubSteps, setStepSubSteps } from "@/lib/scriptUtils";
import type { StepNode, ScriptSchema, StepKindDef } from "@/types/scripts";

interface StepEditorProps {
    step: StepNode;
    schema: ScriptSchema;
    onChange: (step: StepNode) => void;
    onRemove: () => void;
    onMoveUp?: () => void;
    onMoveDown?: () => void;
}

function getStepSummary(step: StepNode): string {
    if (typeof step.params === "string") return step.params;
    if (typeof step.params === "number") return String(step.params);
    if (typeof step.params === "boolean") return String(step.params);
    if (step.kind === "action" && typeof step.params === "object" && step.params !== null) {
        return String((step.params as Record<string, unknown>).name ?? "");
    }
    return "";
}

function hasExpandableContent(step: StepNode, stepDef?: StepKindDef): boolean {
    if (!stepDef) return typeof step.params === "object" && step.params !== null;
    if (stepDef.paramStyle === "map" || stepDef.paramStyle === "string_or_map") {
        if (typeof step.params === "object" && step.params !== null && !Array.isArray(step.params)) return true;
    }
    if (stepDef.acceptsSubSteps) return true;
    return false;
}

export function StepEditor({ step, schema, onChange, onRemove, onMoveUp, onMoveDown }: StepEditorProps) {
    const stepDef = schema.stepKinds[step.kind];
    const category = stepDef?.category ?? "unknown";
    const expandable = hasExpandableContent(step, stepDef);
    const [expanded, setExpanded] = useState(expandable && stepDef?.acceptsSubSteps);

    const summary = getStepSummary(step);
    const subStepGroups = getStepSubSteps(step);
    const subStepKeys = new Set(subStepGroups.map((g) => g.key));

    const isSimpleInline = stepDef && (stepDef.paramStyle === "string" || stepDef.paramStyle === "number" ||
        (stepDef.paramStyle === "string_or_map" && (typeof step.params === "string" || typeof step.params === "number")));

    return (
        <div className="group rounded border border-border/40 bg-background">
            <div className="flex items-center gap-1 px-1 py-0.5">
                <div className="flex items-center gap-0.5 text-muted-foreground">
                    {onMoveUp && (
                        <Button variant="ghost" size="sm" className="h-4 w-4 p-0 opacity-0 group-hover:opacity-100" onClick={onMoveUp}>
                            <GripVertical className="h-3 w-3 rotate-90" />
                        </Button>
                    )}
                </div>
                {expandable && (
                    <button className="shrink-0 text-muted-foreground" onClick={() => setExpanded(!expanded)}>
                        {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                    </button>
                )}
                <Badge variant="outline" className={`shrink-0 text-[10px] px-1.5 py-0 font-mono ${getCategoryColor(category)}`}>
                    {step.kind}
                </Badge>
                {isSimpleInline ? (
                    <div className="flex-1 min-w-0">
                        <StepParamForm stepKind={stepDef} params={step.params} onChange={(params) => onChange({ ...step, params })} />
                    </div>
                ) : (
                    summary && (
                        <span className="flex-1 min-w-0 truncate text-xs text-muted-foreground font-mono">{summary}</span>
                    )
                )}
                <div className="flex items-center gap-0.5 shrink-0">
                    {onMoveUp && (
                        <Button variant="ghost" size="sm" className="h-5 w-5 p-0 opacity-0 group-hover:opacity-100 text-muted-foreground" onClick={onMoveUp} title="Move up">
                            <span className="text-[10px]">↑</span>
                        </Button>
                    )}
                    {onMoveDown && (
                        <Button variant="ghost" size="sm" className="h-5 w-5 p-0 opacity-0 group-hover:opacity-100 text-muted-foreground" onClick={onMoveDown} title="Move down">
                            <span className="text-[10px]">↓</span>
                        </Button>
                    )}
                    <Button variant="ghost" size="sm" className="h-5 w-5 p-0 opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive" onClick={onRemove}>
                        <Trash2 className="h-3 w-3" />
                    </Button>
                </div>
            </div>
            {expanded && !isSimpleInline && stepDef && (
                <div className="border-t border-border/30 px-2 py-1.5 space-y-2">
                    <StepParamForm
                        stepKind={stepDef}
                        params={step.params}
                        onChange={(params) => onChange({ ...step, params })}
                        excludeKeys={subStepKeys}
                    />
                    {subStepGroups.map((group) => (
                        <div key={group.key}>
                            <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60 mb-1">
                                {group.key}
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
