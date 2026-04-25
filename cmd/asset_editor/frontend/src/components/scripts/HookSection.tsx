import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ChevronDown, ChevronRight, Plus, Trash2 } from "lucide-react";
import { RuleEditor } from "./RuleEditor";
import { createEmptyRule } from "@/lib/scriptUtils";
import type { RuleDef, ScriptSchema, EventHookDef } from "@/types/scripts";

interface HookSectionProps {
    hookKey: string;
    hookDef?: EventHookDef;
    rules: RuleDef[];
    schema: ScriptSchema;
    onChange: (rules: RuleDef[]) => void;
    onRemove: () => void;
}

export function HookSection({ hookKey, hookDef, rules, schema, onChange, onRemove }: HookSectionProps) {
    const [expanded, setExpanded] = useState(true);

    return (
        <div className="rounded-md border border-border">
            <div className="flex items-center gap-2 px-2 py-1.5 bg-muted/30">
                <button className="shrink-0 text-muted-foreground" onClick={() => setExpanded(!expanded)}>
                    {expanded ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </button>
                <code className="text-xs font-mono font-semibold">{hookKey}</code>
                <Badge variant="outline" className="text-[10px] px-1 py-0">{rules.length} rule{rules.length !== 1 ? "s" : ""}</Badge>
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
                                onChange(next);
                            }}
                            onRemove={() => onChange(rules.filter((_, j) => j !== i))}
                        />
                    ))}
                    <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 text-xs text-muted-foreground"
                        onClick={() => onChange([...rules, createEmptyRule()])}
                    >
                        <Plus className="mr-1 h-3 w-3" />
                        Add rule
                    </Button>
                </div>
            )}
        </div>
    );
}
