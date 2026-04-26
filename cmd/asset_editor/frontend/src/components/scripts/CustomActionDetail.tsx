import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Plus, Trash2, ChevronDown, ChevronRight } from "lucide-react";
import { StepList } from "./StepList";
import type { CustomActionDef, CustomActionParam, ScriptSchema } from "@/types/scripts";

interface CustomActionDetailProps {
    name: string;
    action: CustomActionDef;
    schema: ScriptSchema;
    onChange: (action: CustomActionDef) => void;
}

function ParamEditor({ params, onChange }: {
    params: CustomActionParam[];
    onChange: (params: CustomActionParam[]) => void;
}) {
    const [expanded, setExpanded] = useState(params.length > 0);

    const addParam = () => {
        onChange([...params, { name: `param_${params.length}` }]);
        if (!expanded) setExpanded(true);
    };

    const removeParam = (idx: number) => {
        onChange(params.filter((_, i) => i !== idx));
    };

    const updateParam = (idx: number, updated: CustomActionParam) => {
        const next = [...params];
        next[idx] = updated;
        onChange(next);
    };

    return (
        <div className="rounded border border-border/60">
            <div className="flex items-center gap-1.5 px-2 py-1">
                <button className="shrink-0 text-muted-foreground hover:text-foreground" onClick={() => setExpanded(!expanded)}>
                    {expanded ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </button>
                <span className="text-xs font-semibold text-muted-foreground">Parameters</span>
                <span className="text-[10px] text-muted-foreground/50">
                    {params.length > 0 ? `(${params.length})` : ""}
                </span>
                <div className="flex-1" />
                <Button variant="ghost" size="sm" className="h-5 text-[10px] text-muted-foreground" onClick={addParam}>
                    <Plus className="mr-1 h-2.5 w-2.5" />
                    Add
                </Button>
            </div>
            {expanded && params.length > 0 && (
                <div className="border-t border-border/30 px-2 py-1.5 space-y-1">
                    <div className="text-[10px] text-muted-foreground/50 mb-1">
                        Parameters available as param.* in expressions within this action's steps.
                    </div>
                    {params.map((p, i) => (
                        <div key={i} className="flex items-center gap-1 group/param">
                            <Input
                                className="h-6 w-32 text-xs font-mono"
                                value={p.name}
                                onChange={(e) => updateParam(i, { ...p, name: e.target.value })}
                                placeholder="name"
                            />
                            <Input
                                className="h-6 flex-1 text-xs"
                                value={p.description ?? ""}
                                onChange={(e) => updateParam(i, { ...p, description: e.target.value || undefined })}
                                placeholder="description"
                            />
                            <Input
                                className="h-6 w-24 text-xs font-mono"
                                value={p.default !== undefined && p.default !== null ? String(p.default) : ""}
                                onChange={(e) => {
                                    let val: unknown = e.target.value;
                                    if (val === "true") val = true;
                                    else if (val === "false") val = false;
                                    else if (val !== "" && !isNaN(Number(val))) val = Number(val);
                                    updateParam(i, { ...p, default: val === "" ? undefined : val });
                                }}
                                placeholder="default"
                            />
                            <button
                                className="text-muted-foreground hover:text-destructive opacity-0 group-hover/param:opacity-100 shrink-0"
                                onClick={() => removeParam(i)}
                            >
                                <Trash2 className="h-3 w-3" />
                            </button>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

export function CustomActionDetail({ name, action, schema, onChange }: CustomActionDetailProps) {
    return (
        <div className="flex h-full flex-col overflow-hidden">
            <div className="flex items-center gap-2 border-b border-border px-3 py-1.5">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-accent-violet px-1.5 py-0.5 rounded bg-accent-violet-tint">custom action</span>
                <code className="text-sm font-mono font-semibold">{name}</code>
            </div>
            <ScrollArea className="flex-1 overflow-hidden">
                <div className="space-y-3 p-3">
                    <div className="flex items-center gap-1.5">
                        <span className="text-xs text-muted-foreground w-20 shrink-0">Description</span>
                        <Input
                            className="h-7 flex-1 text-xs"
                            value={action.description ?? ""}
                            onChange={(e) => onChange({ ...action, description: e.target.value || undefined })}
                            placeholder="What this action does"
                        />
                    </div>
                    <ParamEditor
                        params={action.params ?? []}
                        onChange={(params) => onChange({ ...action, params: params.length > 0 ? params : undefined })}
                    />
                    <div className="space-y-1">
                        <div className="text-xs font-semibold text-muted-foreground">Steps</div>
                        <StepList
                            steps={action.steps}
                            schema={schema}
                            onChange={(steps) => onChange({ ...action, steps })}
                        />
                    </div>
                </div>
            </ScrollArea>
        </div>
    );
}
