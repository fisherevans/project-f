import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ChevronDown, ChevronRight, Plus, Trash2, BookOpen } from "lucide-react";
import { ReferencePopover } from "./ReferencePopover";
import { RuleEditor } from "./RuleEditor";
import { createEmptyRule, moveItem } from "@/lib/scriptUtils";
import type { HookDef, ScriptSchema, EventHookDef } from "@/types/scripts";

interface HookSectionProps {
    hookKey: string;
    hookDef?: EventHookDef;
    hook: HookDef;
    schema: ScriptSchema;
    onChange: (hook: HookDef) => void;
    onRemove: () => void;
}

export function HookSection({ hookKey, hookDef, hook, schema, onChange, onRemove }: HookSectionProps) {
    const [expanded, setExpanded] = useState(true);
    const rules = hook.rules;
    const isAllMode = hook.mode === "all";

    const updateRules = (newRules: typeof rules) => {
        onChange({ ...hook, rules: newRules });
    };

    return (
        <div className="rounded-md border border-border">
            <div className="flex items-center gap-2 px-2 py-1.5 bg-muted/30">
                <button className="shrink-0 text-muted-foreground" onClick={() => setExpanded(!expanded)}>
                    {expanded ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </button>
                <ReferencePopover target={{ type: "hook", yamlKey: hookKey }}>
                    <span className="inline-flex items-center gap-1 hover:text-accent-violet cursor-help">
                        <code className="text-xs font-mono font-semibold">{hookKey}</code>
                        <BookOpen className="h-2.5 w-2.5 text-muted-foreground/40" />
                    </span>
                </ReferencePopover>
                <Badge variant="outline" className="text-[10px] px-1 py-0">{rules.length} rule{rules.length !== 1 ? "s" : ""}</Badge>
                <button
                    className={`text-[10px] px-1.5 py-0.5 rounded transition-colors ${
                        isAllMode
                            ? "bg-accent-teal-tint text-accent-teal"
                            : "bg-muted text-muted-foreground hover:text-foreground"
                    }`}
                    onClick={() => onChange({ ...hook, mode: isAllMode ? undefined : "all" })}
                    title={isAllMode ? "All matching rules run" : "First matching rule wins"}
                >
                    {isAllMode ? "all" : "first match"}
                </button>
                {hookDef && (
                    <span className="text-[10px] text-muted-foreground truncate">{hookDef.description}</span>
                )}
                <div className="flex-1" />
                <Button variant="ghost" size="sm" className="h-5 w-5 p-0 text-muted-foreground hover:text-destructive" onClick={onRemove} title="Remove hook">
                    <Trash2 className="h-3 w-3" />
                </Button>
            </div>
            {expanded && (
                <div className="space-y-2 p-2">
                    {rules.map((rule, i) => (
                        <RuleEditor
                            key={i}
                            rule={rule}
                            ruleIndex={i}
                            hookDef={hookDef}
                            schema={schema}
                            onChange={(updated) => {
                                const next = [...rules];
                                next[i] = updated;
                                updateRules(next);
                            }}
                            onRemove={() => updateRules(rules.filter((_, j) => j !== i))}
                            onMoveUp={i > 0 ? () => updateRules(moveItem(rules, i, i - 1)) : undefined}
                            onMoveDown={i < rules.length - 1 ? () => updateRules(moveItem(rules, i, i + 1)) : undefined}
                        />
                    ))}
                    <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 text-xs text-muted-foreground"
                        onClick={() => updateRules([...rules, createEmptyRule()])}
                    >
                        <Plus className="mr-1 h-3 w-3" />
                        Add rule
                    </Button>
                </div>
            )}
        </div>
    );
}
