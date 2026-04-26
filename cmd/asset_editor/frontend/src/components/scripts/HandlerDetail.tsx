import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { ChevronDown, ChevronRight, Plus, Trash2, X, MapPin } from "lucide-react";
import { HookSection } from "./HookSection";
import { getHandlerHookKeys, getAvailableHooks } from "@/lib/scriptUtils";
import { useTiledUsages } from "@/api/scripts";
import type { HandlerDef, HandlerPropDef, RuleDef, ScriptSchema, TiledEntityRef } from "@/types/scripts";

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

const PROP_TYPES = ["string", "number", "bool", "entity_ref"];

function PropsSection({ props, onChange }: { props: HandlerPropDef[]; onChange: (props: HandlerPropDef[] | undefined) => void }) {
    const [expanded, setExpanded] = useState(props.length > 0);

    const addProp = () => {
        onChange([...props, { name: "new_prop", type: "string", description: "" }]);
        if (!expanded) setExpanded(true);
    };

    const removeProp = (idx: number) => {
        const next = props.filter((_, i) => i !== idx);
        onChange(next.length > 0 ? next : undefined);
    };

    const updateProp = (idx: number, patch: Partial<HandlerPropDef>) => {
        const next = props.map((p, i) => i === idx ? { ...p, ...patch } : p);
        onChange(next);
    };

    return (
        <div className="rounded border border-accent-teal-edge/40">
            <div className="flex items-center gap-1.5 px-2 py-1">
                <button className="shrink-0 text-muted-foreground hover:text-foreground" onClick={() => setExpanded(!expanded)}>
                    {expanded ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </button>
                <span className="text-xs font-semibold text-accent-teal">Properties</span>
                <span className="text-[10px] text-muted-foreground/50">
                    {props.length > 0 ? `(${props.length})` : ""}
                </span>
                <div className="flex-1" />
                <Button variant="ghost" size="sm" className="h-5 text-[10px] text-muted-foreground" onClick={addProp}>
                    <Plus className="mr-1 h-2.5 w-2.5" />
                    Add
                </Button>
            </div>
            {expanded && props.length > 0 && (
                <div className="border-t border-border/30 px-2 py-1.5 space-y-1.5">
                    <div className="text-[10px] text-muted-foreground/50 mb-1">
                        Expected entity properties from Tiled. Access as prop.* in expressions.
                    </div>
                    {props.map((p, i) => (
                        <div key={i} className="flex items-start gap-1 group/prop border border-border/30 rounded px-1.5 py-1">
                            <div className="flex flex-col gap-1 flex-1 min-w-0">
                                <div className="flex items-center gap-1">
                                    <Input
                                        className="h-6 w-28 text-xs font-mono"
                                        value={p.name}
                                        onChange={(e) => updateProp(i, { name: e.target.value })}
                                        placeholder="name"
                                    />
                                    <select
                                        className="h-6 rounded border border-input bg-background px-1 text-xs"
                                        value={p.type}
                                        onChange={(e) => updateProp(i, { type: e.target.value })}
                                    >
                                        {PROP_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
                                    </select>
                                    <label className="flex items-center gap-1 text-[10px] text-muted-foreground">
                                        <input
                                            type="checkbox"
                                            checked={p.required ?? false}
                                            onChange={(e) => updateProp(i, { required: e.target.checked || undefined })}
                                            className="h-3 w-3"
                                        />
                                        req
                                    </label>
                                    {!p.required && (
                                        <Input
                                            className="h-6 w-20 text-xs font-mono"
                                            value={p.default !== undefined && p.default !== null ? String(p.default) : ""}
                                            onChange={(e) => {
                                                const v = e.target.value;
                                                updateProp(i, { default: v === "" ? undefined : v });
                                            }}
                                            placeholder="default"
                                        />
                                    )}
                                </div>
                                <Input
                                    className="h-6 text-xs text-muted-foreground"
                                    value={p.description ?? ""}
                                    onChange={(e) => updateProp(i, { description: e.target.value })}
                                    placeholder="description"
                                />
                            </div>
                            <button
                                className="text-muted-foreground hover:text-destructive opacity-0 group-hover/prop:opacity-100 shrink-0 mt-1"
                                onClick={() => removeProp(i)}
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

function TiledRefsSection({ handlerName }: { handlerName: string }) {
    const { data: usages } = useTiledUsages();
    const [expanded, setExpanded] = useState(false);

    const entities: TiledEntityRef[] = usages
        ?.find((u) => u.handlerName === handlerName)
        ?.entities ?? [];

    if (entities.length === 0) return null;

    return (
        <div className="rounded border border-accent-orange-edge/40">
            <div className="flex items-center gap-1.5 px-2 py-1">
                <button className="shrink-0 text-muted-foreground hover:text-foreground" onClick={() => setExpanded(!expanded)}>
                    {expanded ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </button>
                <MapPin className="h-3 w-3 text-accent-orange" />
                <span className="text-xs font-semibold text-accent-orange">Tiled Entities</span>
                <span className="text-[10px] text-muted-foreground/50">({entities.length})</span>
            </div>
            {expanded && (
                <div className="border-t border-border/30 px-2 py-1.5 space-y-1">
                    <div className="text-[10px] text-muted-foreground/50 mb-1">
                        Map entities that reference this handler via script_ref.
                    </div>
                    {entities.map((ent, i) => (
                        <div key={i} className="border border-border/30 rounded px-2 py-1 text-xs">
                            <div className="flex items-center gap-2">
                                <span className="font-mono text-accent-orange">{ent.mapFile}</span>
                                <span className="text-muted-foreground/50">obj#{ent.objectId}</span>
                                {ent.entityId && (
                                    <code className="bg-accent-teal-tint text-accent-teal px-1 rounded text-[10px]">{ent.entityId}</code>
                                )}
                                {ent.objectType && (
                                    <span className="bg-accent-violet-tint text-accent-violet px-1 rounded text-[10px]">{ent.objectType}</span>
                                )}
                            </div>
                            {Object.keys(ent.properties).length > 0 && (
                                <div className="mt-1 flex flex-wrap gap-1">
                                    {Object.entries(ent.properties).map(([k, v]) => (
                                        <span key={k} className="text-[10px] font-mono text-muted-foreground">
                                            <span className="text-accent-amber">{k}</span>=<span className="text-foreground">{v || '""'}</span>
                                        </span>
                                    ))}
                                </div>
                            )}
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
                    <PropsSection
                        props={handler.props ?? []}
                        onChange={(props) => {
                            const next = { ...handler };
                            if (props && props.length > 0) {
                                next.props = props;
                            } else {
                                delete next.props;
                            }
                            onChange(next);
                        }}
                    />
                    <TiledRefsSection handlerName={handlerName} />
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
