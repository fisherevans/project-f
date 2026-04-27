import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ChevronDown, ChevronRight, Plus, Trash2, BookOpen, ClipboardPaste, X } from "lucide-react";
import { ReferencePopover } from "./ReferencePopover";
import { RuleEditor } from "./RuleEditor";
import { useScriptClipboard, pasteItem } from "./ScriptClipboard";
import { createEmptyRule, moveItem } from "@/lib/scriptUtils";
import type { HookDef, RuleDef, ScriptSchema, EventHookDef } from "@/types/scripts";

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
    const [fullscreenRule, setFullscreenRule] = useState<number | null>(null);
    const rules = hook.rules;
    const isAllMode = hook.mode === "all";
    const clipboard = useScriptClipboard();
    const canPasteRule = clipboard.item?.type === "rule";

    const updateRules = (newRules: typeof rules) => {
        onChange({ ...hook, rules: newRules });
    };

    const cloneRule = (index: number) => {
        const cloned: RuleDef = JSON.parse(JSON.stringify(rules[index]));
        const next = [...rules];
        next.splice(index + 1, 0, cloned);
        updateRules(next);
    };

    const pasteRuleAt = (index: number) => {
        if (clipboard.item?.type !== "rule") return;
        const pasted = pasteItem(clipboard.item.data);
        const next = [...rules];
        next.splice(index, 0, pasted);
        updateRules(next);
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
                            onClone={() => cloneRule(i)}
                            onFullscreen={() => setFullscreenRule(i)}
                        />
                    ))}
                    <div className="flex items-center gap-1">
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 text-xs text-muted-foreground"
                            onClick={() => updateRules([...rules, createEmptyRule()])}
                        >
                            <Plus className="mr-1 h-3 w-3" />
                            Add rule
                        </Button>
                        {canPasteRule && (
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 text-xs text-accent-violet"
                                onClick={() => pasteRuleAt(rules.length)}
                            >
                                <ClipboardPaste className="mr-1 h-3 w-3" />
                                Paste rule
                            </Button>
                        )}
                    </div>
                </div>
            )}
            {fullscreenRule !== null && rules[fullscreenRule] && (
                <FullscreenRuleModal
                    rule={rules[fullscreenRule]}
                    ruleIndex={fullscreenRule}
                    hookKey={hookKey}
                    hookDef={hookDef}
                    schema={schema}
                    onChange={(updated) => {
                        const next = [...rules];
                        next[fullscreenRule] = updated;
                        updateRules(next);
                    }}
                    onClose={() => setFullscreenRule(null)}
                />
            )}
        </div>
    );
}

function FullscreenRuleModal({
    rule,
    ruleIndex,
    hookKey,
    hookDef,
    schema,
    onChange,
    onClose,
}: {
    rule: RuleDef;
    ruleIndex: number;
    hookKey: string;
    hookDef?: EventHookDef;
    schema: ScriptSchema;
    onChange: (rule: RuleDef) => void;
    onClose: () => void;
}) {
    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
            <div
                className="bg-popover border border-border rounded-lg shadow-xl flex flex-col"
                style={{ width: "min(90vw, 900px)", height: "min(85vh, 800px)" }}
                onClick={(e) => e.stopPropagation()}
            >
                <div className="flex items-center justify-between border-b border-border px-4 py-2.5 shrink-0">
                    <div className="flex items-center gap-2">
                        <code className="text-sm font-mono font-semibold">{hookKey}</code>
                        <span className="text-xs text-muted-foreground">Rule {ruleIndex + 1}</span>
                    </div>
                    <button className="text-muted-foreground hover:text-foreground" onClick={onClose}>
                        <X className="h-4 w-4" />
                    </button>
                </div>
                <div className="flex-1 overflow-y-auto p-4">
                    <RuleEditor
                        rule={rule}
                        ruleIndex={ruleIndex}
                        hookDef={hookDef}
                        schema={schema}
                        onChange={onChange}
                        onRemove={() => {}}
                    />
                </div>
            </div>
        </div>
    );
}
