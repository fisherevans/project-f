import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Trash2, Plus, ArrowUp, ArrowDown, Maximize2, Copy, ClipboardPaste } from "lucide-react";
import { StepList } from "./StepList";
import { ConditionEditor, createEmptyCondition } from "./ConditionEditor";
import { ZoneIdInput } from "./inputs/ZoneIdInput";
import { useScriptClipboard, pasteItem } from "./ScriptClipboard";
import type { RuleDef, ScriptSchema, EventHookDef } from "@/types/scripts";

interface RuleEditorProps {
    rule: RuleDef;
    ruleIndex: number;
    hookDef?: EventHookDef;
    schema: ScriptSchema;
    onChange: (rule: RuleDef) => void;
    onRemove: () => void;
    onMoveUp?: () => void;
    onMoveDown?: () => void;
    onClone?: () => void;
    onFullscreen?: () => void;
}

export function RuleEditor({ rule, ruleIndex, hookDef, schema, onChange, onRemove, onMoveUp, onMoveDown, onClone, onFullscreen }: RuleEditorProps) {
    const hasFilterFields = hookDef?.filterFields && hookDef.filterFields.length > 0;
    const clipboard = useScriptClipboard();
    const canPasteCondition = clipboard.item?.type === "condition";

    return (
        <div className="rounded border border-border/50 bg-muted/10">
            <div className="flex items-center justify-between border-b border-border/30 px-2 py-1">
                <span className="text-xs font-medium text-muted-foreground">Rule {ruleIndex + 1}</span>
                <div className="flex items-center">
                    {onFullscreen && (
                        <Button variant="ghost" size="sm" className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground" onClick={onFullscreen} title="Edit fullscreen">
                            <Maximize2 className="h-3 w-3" />
                        </Button>
                    )}
                    {onClone && (
                        <Button variant="ghost" size="sm" className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground" onClick={onClone} title="Clone rule">
                            <Copy className="h-3 w-3" />
                        </Button>
                    )}
                    <Button variant="ghost" size="sm" className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground" onClick={() => clipboard.copyRule(rule)} title="Copy rule to clipboard">
                        <ClipboardPaste className="h-3 w-3" />
                    </Button>
                    {onMoveUp && (
                        <Button variant="ghost" size="sm" className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground" onClick={onMoveUp}>
                            <ArrowUp className="h-3 w-3" />
                        </Button>
                    )}
                    {onMoveDown && (
                        <Button variant="ghost" size="sm" className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground" onClick={onMoveDown}>
                            <ArrowDown className="h-3 w-3" />
                        </Button>
                    )}
                    <Button variant="ghost" size="sm" className="h-5 w-5 p-0 text-muted-foreground hover:text-destructive" onClick={onRemove}>
                        <Trash2 className="h-3 w-3" />
                    </Button>
                </div>
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
                    canPaste={canPasteCondition}
                    onPaste={() => {
                        if (clipboard.item?.type === "condition") {
                            onChange({ ...rule, when: pasteItem(clipboard.item.data) });
                        }
                    }}
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
                        <span className="text-xs text-muted-foreground shrink-0">{field.name}</span>
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

function ConditionSection({ condition, schema, onChange, onRemove, canPaste, onPaste }: {
    condition?: import("@/types/scripts").ConditionNode;
    schema: ScriptSchema;
    onChange: (condition: import("@/types/scripts").ConditionNode) => void;
    onRemove: () => void;
    canPaste?: boolean;
    onPaste?: () => void;
}) {
    if (!condition) {
        return (
            <div className="flex items-center gap-1">
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 text-xs text-muted-foreground"
                    onClick={() => onChange(createEmptyCondition())}
                >
                    <Plus className="mr-1 h-3 w-3" />
                    Add condition
                </Button>
                {canPaste && onPaste && (
                    <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 text-xs text-accent-violet"
                        onClick={onPaste}
                    >
                        <ClipboardPaste className="mr-1 h-3 w-3" />
                        Paste condition
                    </Button>
                )}
            </div>
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
