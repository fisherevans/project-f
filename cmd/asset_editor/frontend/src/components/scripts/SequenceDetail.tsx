import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Plus, Trash2 } from "lucide-react";
import { StepList } from "./StepList";
import type { SequenceDef, ScriptSchema } from "@/types/scripts";

interface SequenceDetailProps {
    name: string;
    sequence: SequenceDef;
    schema: ScriptSchema;
    onChange: (sequence: SequenceDef) => void;
}

export function SequenceDetail({ name, sequence, schema, onChange }: SequenceDetailProps) {
    const params = sequence.params ?? [];

    const addParam = () => {
        onChange({ ...sequence, params: [...params, `param_${params.length}`] });
    };

    const removeParam = (idx: number) => {
        const next = params.filter((_, i) => i !== idx);
        onChange({ ...sequence, params: next.length > 0 ? next : undefined });
    };

    const updateParam = (idx: number, value: string) => {
        const next = [...params];
        next[idx] = value;
        onChange({ ...sequence, params: next });
    };

    return (
        <div className="flex h-full flex-col overflow-hidden">
            <div className="flex items-center gap-2 border-b border-border px-3 py-1.5">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-accent-teal px-1.5 py-0.5 rounded bg-accent-teal-tint">sequence</span>
                <code className="text-sm font-mono font-semibold">{name}</code>
            </div>
            <ScrollArea className="flex-1 overflow-hidden">
                <div className="space-y-3 p-3">
                    <div className="rounded border border-border/60">
                        <div className="flex items-center gap-1.5 px-2 py-1">
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
                        {params.length > 0 && (
                            <div className="border-t border-border/30 px-2 py-1.5 space-y-1">
                                {params.map((p, i) => (
                                    <div key={i} className="flex items-center gap-1 group/param">
                                        <Input
                                            className="h-6 flex-1 text-xs font-mono"
                                            value={p}
                                            onChange={(e) => updateParam(i, e.target.value)}
                                            placeholder="param name"
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
                    <div className="space-y-1">
                        <div className="text-xs font-semibold text-muted-foreground">Steps</div>
                        <StepList
                            steps={sequence.steps}
                            schema={schema}
                            onChange={(steps) => onChange({ ...sequence, steps })}
                        />
                    </div>
                </div>
            </ScrollArea>
        </div>
    );
}
