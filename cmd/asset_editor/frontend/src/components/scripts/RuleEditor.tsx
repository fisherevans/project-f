import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Trash2, Plus } from "lucide-react";
import { StepList } from "./StepList";
import { ConditionEditor, createEmptyCondition } from "./ConditionEditor";
import { ZoneIdInput } from "./inputs/ZoneIdInput";
import type { RuleDef, ScriptSchema, EventHookDef } from "@/types/scripts";

interface RuleEditorProps {
    rule: RuleDef;
    ruleIndex: number;
    hookDef?: EventHookDef;
    schema: ScriptSchema;
    onChange: (rule: RuleDef) => void;
    onRemove: () => void;
}

export function RuleEditor({ rule, ruleIndex, hookDef, schema, onChange, onRemove }: RuleEditorProps) {
    const hasFilterFields = hookDef?.filterFields && hookDef.filterFields.length > 0;

    return (
        <div className="rounded border border-border/50 bg-muted/10">
            <div className="flex items-center justify-between border-b border-border/30 px-2 py-1">
                <span className="text-xs font-medium text-muted-foreground">Rule {ruleIndex + 1}</span>
                <Button variant="ghost" size="sm" className="h-5 w-5 p-0 text-muted-foreground hover:text-destructive" onClick={onRemove}>
                    <Trash2 className="h-3 w-3" />
                </Button>
            </div>
            <div className="space-y-2 p-2">
                {hasFilterFields && (
                    <FilterSection
                        filter={rule.filter}
                        hookDef={hookDef!}
                        onChange={(filter) => onChange({ ...rule, filter })}
                    />
                )}

                <ConditionSection
                    condition={rule.when}
                    schema={schema}
                    onChange={(when) => onChange({ ...rule, when })}
                    onRemove={() => {
                        const { when: _, ...rest } = rule;
                        onChange(rest as RuleDef);
                    }}
                />

                <SetStateSection
                    setState={rule.set_state}
                    onChange={(set_state) => onChange({ ...rule, set_state })}
                />

                <div>
                    <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60 mb-1">
                        Steps
                    </div>
                    <StepList
                        steps={rule.steps}
                        schema={schema}
                        onChange={(steps) => onChange({ ...rule, steps })}
                    />
                </div>
            </div>
        </div>
    );
}

function FilterSection({ filter, hookDef, onChange }: {
    filter?: Record<string, unknown>;
    hookDef: EventHookDef;
    onChange: (filter?: Record<string, unknown>) => void;
}) {
    const hasFilter = filter && Object.keys(filter).length > 0;

    if (!hasFilter) {
        return (
            <Button
                variant="ghost"
                size="sm"
                className="h-6 text-xs text-muted-foreground"
                onClick={() => onChange({})}
            >
                <Plus className="mr-1 h-3 w-3" />
                Add filter
            </Button>
        );
    }

    return (
        <div>
            <div className="flex items-center justify-between">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60">Filter</span>
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-4 w-4 p-0 text-muted-foreground hover:text-destructive"
                    onClick={() => onChange(undefined)}
                >
                    <Trash2 className="h-2.5 w-2.5" />
                </Button>
            </div>
            <div className="space-y-1 mt-1">
                {hookDef.filterFields!.map((field) => (
                    <div key={field.name} className="flex items-center gap-1.5">
                        <span className="text-xs text-muted-foreground w-24 shrink-0">{field.name}</span>
                        {field.name === "zone" ? (
                            <div className="flex-1">
                                <ZoneIdInput
                                    value={String(filter![field.name] ?? "")}
                                    placeholder={field.description}
                                    onChange={(v) => {
                                        const next = { ...filter! };
                                        if (!v) { delete next[field.name]; } else { next[field.name] = v; }
                                        onChange(Object.keys(next).length > 0 ? next : undefined);
                                    }}
                                />
                            </div>
                        ) : (
                            <Input
                                className="h-6 flex-1 text-xs font-mono"
                                value={String(filter![field.name] ?? "")}
                                placeholder={field.description}
                                onChange={(e) => {
                                    const next = { ...filter! };
                                    if (e.target.value === "") {
                                        delete next[field.name];
                                    } else {
                                        let v: unknown = e.target.value;
                                        if (field.type === "bool") v = v === "true";
                                        next[field.name] = v;
                                    }
                                    onChange(Object.keys(next).length > 0 ? next : undefined);
                                }}
                            />
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
}

function ConditionSection({ condition, schema, onChange, onRemove }: {
    condition?: import("@/types/scripts").ConditionNode;
    schema: ScriptSchema;
    onChange: (condition: import("@/types/scripts").ConditionNode) => void;
    onRemove: () => void;
}) {
    if (!condition) {
        return (
            <Button
                variant="ghost"
                size="sm"
                className="h-6 text-xs text-muted-foreground"
                onClick={() => onChange(createEmptyCondition())}
            >
                <Plus className="mr-1 h-3 w-3" />
                Add condition
            </Button>
        );
    }

    return (
        <div>
            <div className="flex items-center justify-between mb-1">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60">When</span>
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-4 w-4 p-0 text-muted-foreground hover:text-destructive"
                    onClick={onRemove}
                >
                    <Trash2 className="h-2.5 w-2.5" />
                </Button>
            </div>
            <ConditionEditor condition={condition} schema={schema} onChange={onChange} />
        </div>
    );
}

function SetStateSection({ setState, onChange }: {
    setState?: Record<string, unknown>;
    onChange: (setState?: Record<string, unknown>) => void;
}) {
    const hasState = setState && Object.keys(setState).length > 0;
    const entries = hasState ? Object.entries(setState!) : [];

    if (!hasState) {
        return (
            <Button
                variant="ghost"
                size="sm"
                className="h-6 text-xs text-muted-foreground"
                onClick={() => onChange({ "": "" })}
            >
                <Plus className="mr-1 h-3 w-3" />
                Add set_state
            </Button>
        );
    }

    return (
        <div>
            <div className="flex items-center justify-between mb-1">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60">Set State</span>
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-4 w-4 p-0 text-muted-foreground hover:text-destructive"
                    onClick={() => onChange(undefined)}
                >
                    <Trash2 className="h-2.5 w-2.5" />
                </Button>
            </div>
            <div className="space-y-1">
                {entries.map(([key, value], i) => (
                    <div key={i} className="flex items-center gap-1.5">
                        <Input
                            className="h-6 flex-1 text-xs font-mono"
                            placeholder="key"
                            value={key}
                            onChange={(e) => {
                                const next: Record<string, unknown> = {};
                                for (let j = 0; j < entries.length; j++) {
                                    const [k, v] = entries[j];
                                    next[j === i ? e.target.value : k] = v;
                                }
                                onChange(next);
                            }}
                        />
                        <span className="text-xs text-muted-foreground">=</span>
                        <Input
                            className="h-6 w-24 text-xs font-mono"
                            placeholder="value"
                            value={String(value ?? "")}
                            onChange={(e) => {
                                let v: unknown = e.target.value;
                                if (v === "true") v = true;
                                else if (v === "false") v = false;
                                else if (v !== "" && !isNaN(Number(v))) v = Number(v);
                                onChange({ ...setState!, [key]: v });
                            }}
                        />
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-5 w-5 p-0 text-muted-foreground hover:text-destructive"
                            onClick={() => {
                                const next = { ...setState! };
                                delete next[key];
                                onChange(Object.keys(next).length > 0 ? next : undefined);
                            }}
                        >
                            <Trash2 className="h-2.5 w-2.5" />
                        </Button>
                    </div>
                ))}
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-5 text-[10px] text-muted-foreground"
                    onClick={() => onChange({ ...setState!, "": "" })}
                >
                    <Plus className="mr-1 h-2.5 w-2.5" />
                    Add entry
                </Button>
            </div>
        </div>
    );
}
