import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { ChevronDown, ChevronRight, Plus, Trash2, X } from "lucide-react";
import { HookSection } from "./HookSection";
import { getHandlerHookKeys, getAvailableHooks } from "@/lib/scriptUtils";
import type { HandlerDef, RuleDef, ScriptSchema } from "@/types/scripts";

interface HandlerDetailProps {
    handlerName: string;
    handler: HandlerDef;
    schema: ScriptSchema;
    onChange: (handler: HandlerDef) => void;
}

const HOOK_GROUPS: Record<string, string[]> = {
    "Interaction": ["on_interact_self", "on_interact"],
    "Lifecycle": ["on_init", "on_state_enter"],
    "State": ["on_global_updated", "on_zone_activity", "on_broadcast", "on_combat_complete", "on_timer_complete"],
    "Motion": ["on_motion_complete_self", "on_motion_complete"],
};

function VarSection({ vars, onChange }: { vars: Record<string, unknown>; onChange: (vars: Record<string, unknown> | undefined) => void }) {
    const [expanded, setExpanded] = useState(Object.keys(vars).length > 0);
    const entries = Object.entries(vars);

    const addVar = () => {
        const key = `new_var_${entries.length}`;
        onChange({ ...vars, [key]: "" });
        if (!expanded) setExpanded(true);
    };

    const removeVar = (key: string) => {
        const next = { ...vars };
        delete next[key];
        if (Object.keys(next).length === 0) {
            onChange(undefined);
        } else {
            onChange(next);
        }
    };

    const updateKey = (oldKey: string, newKey: string) => {
        if (newKey === oldKey) return;
        const next: Record<string, unknown> = {};
        for (const [k, v] of Object.entries(vars)) {
            next[k === oldKey ? newKey : k] = v;
        }
        onChange(next);
    };

    const updateValue = (key: string, rawValue: string) => {
        let value: unknown = rawValue;
        if (rawValue === "true") value = true;
        else if (rawValue === "false") value = false;
        else if (rawValue !== "" && !isNaN(Number(rawValue))) value = Number(rawValue);
        onChange({ ...vars, [key]: value });
    };

    return (
        <div className="rounded border border-border/60">
            <div className="flex items-center gap-1.5 px-2 py-1">
                <button className="shrink-0 text-muted-foreground hover:text-foreground" onClick={() => setExpanded(!expanded)}>
                    {expanded ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </button>
                <span className="text-xs font-semibold text-muted-foreground">Variables</span>
                <span className="text-[10px] text-muted-foreground/50">
                    {entries.length > 0 ? `(${entries.length})` : ""}
                </span>
                <div className="flex-1" />
                <Button variant="ghost" size="sm" className="h-5 text-[10px] text-muted-foreground" onClick={addVar}>
                    <Plus className="mr-1 h-2.5 w-2.5" />
                    Add
                </Button>
            </div>
            {expanded && entries.length > 0 && (
                <div className="border-t border-border/30 px-2 py-1.5 space-y-1">
                    <div className="text-[10px] text-muted-foreground/50 mb-1">
                        Initial values for handler-local state. Access as var.* in expressions.
                    </div>
                    {entries.map(([key, val], i) => (
                        <div key={i} className="flex items-center gap-1 group/var">
                            <Input
                                className="h-6 w-32 text-xs font-mono"
                                value={key}
                                onChange={(e) => updateKey(key, e.target.value)}
                                placeholder="key"
                            />
                            <span className="text-muted-foreground text-[10px]">=</span>
                            <Input
                                className="h-6 flex-1 text-xs font-mono"
                                value={val !== undefined && val !== null ? String(val) : ""}
                                onChange={(e) => updateValue(key, e.target.value)}
                                placeholder="initial value"
                            />
                            <span className="text-[10px] text-muted-foreground/50 w-10 text-right shrink-0">
                                {typeof val === "number" ? "num" : typeof val === "boolean" ? "bool" : Array.isArray(val) ? "list" : typeof val === "object" && val !== null ? "map" : "str"}
                            </span>
                            <button
                                className="text-muted-foreground hover:text-destructive opacity-0 group-hover/var:opacity-100 shrink-0"
                                onClick={() => removeVar(key)}
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

export function HandlerDetail({ handlerName, handler, schema, onChange }: HandlerDetailProps) {
    const hookKeys = getHandlerHookKeys(handler);
    const availableHooks = getAvailableHooks(handler, schema);
    const [showHookPicker, setShowHookPicker] = useState(false);

    const handleAddHook = (hookKey: string) => {
        onChange({ ...handler, [hookKey]: [] });
        setShowHookPicker(false);
    };

    const handleRemoveHook = (hookKey: string) => {
        const next = { ...handler };
        delete next[hookKey];
        onChange(next);
    };

    const handleUpdateHook = (hookKey: string, rules: RuleDef[]) => {
        onChange({ ...handler, [hookKey]: rules });
    };

    return (
        <div className="flex h-full flex-col overflow-hidden">
            <div className="flex items-center gap-2 border-b border-border px-3 py-1.5">
                <code className="text-sm font-mono font-semibold">{handlerName}</code>
                <div className="flex-1" />
                {availableHooks.length > 0 && (
                    <Button variant="outline" size="sm" className="h-7 text-xs gap-1" onClick={() => setShowHookPicker(!showHookPicker)}>
                        <Plus className="h-3 w-3" />
                        Add hook
                    </Button>
                )}
            </div>
            {showHookPicker && (
                <div className="border-b border-border bg-muted/30 p-2">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-xs font-medium text-muted-foreground">Select event hook</span>
                        <button className="text-muted-foreground hover:text-foreground" onClick={() => setShowHookPicker(false)}>
                            <X className="h-3.5 w-3.5" />
                        </button>
                    </div>
                    <div className="space-y-2">
                        {Object.entries(HOOK_GROUPS).map(([group, keys]) => {
                            const available = keys.filter((k) => availableHooks.includes(k));
                            if (available.length === 0) return null;
                            return (
                                <div key={group}>
                                    <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60 mb-1">{group}</div>
                                    <div className="space-y-0.5">
                                        {available.map((hook) => {
                                            const hookDef = Object.values(schema.eventHooks).find((h) => h.yamlKey === hook);
                                            return (
                                                <button
                                                    key={hook}
                                                    className="flex w-full items-start gap-2 rounded px-2 py-1.5 text-left hover:bg-accent"
                                                    onClick={() => handleAddHook(hook)}
                                                >
                                                    <code className="text-xs font-mono font-semibold shrink-0">{hook}</code>
                                                    {hookDef && (
                                                        <span className="text-[11px] text-muted-foreground leading-tight">{hookDef.description}</span>
                                                    )}
                                                </button>
                                            );
                                        })}
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                </div>
            )}
            <ScrollArea className="flex-1 overflow-hidden">
                <div className="space-y-3 p-3">
                    <VarSection
                        vars={handler.var ?? {}}
                        onChange={(vars) => {
                            const next = { ...handler };
                            if (vars && Object.keys(vars).length > 0) {
                                next.var = vars;
                            } else {
                                delete next.var;
                            }
                            onChange(next);
                        }}
                    />
                    {hookKeys.length === 0 && (
                        <div className="text-center text-sm text-muted-foreground py-8">
                            No event hooks defined. Add one to get started.
                        </div>
                    )}
                    {hookKeys.map((hookKey) => {
                        const hookDef = Object.values(schema.eventHooks).find((h) => h.yamlKey === hookKey);
                        return (
                            <HookSection
                                key={hookKey}
                                hookKey={hookKey}
                                hookDef={hookDef}
                                rules={handler[hookKey]}
                                schema={schema}
                                onChange={(rules) => handleUpdateHook(hookKey, rules)}
                                onRemove={() => handleRemoveHook(hookKey)}
                            />
                        );
                    })}
                </div>
            </ScrollArea>
        </div>
    );
}
