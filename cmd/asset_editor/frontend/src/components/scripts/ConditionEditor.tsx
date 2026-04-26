import { useState, useRef, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ChevronDown, Plus, Trash2 } from "lucide-react";
import { GlobalKeyInput } from "./inputs/GlobalKeyInput";
import { ZoneIdInput } from "./inputs/ZoneIdInput";
import { ExpressionInput } from "./inputs/ExpressionInput";
import { ExpressionHelpLink } from "./ExpressionHelpModal";
import { useExprContext } from "./ExprContext";
import type { ConditionNode, ScriptSchema } from "@/types/scripts";

interface ConditionEditorProps {
    condition: ConditionNode;
    schema: ScriptSchema;
    onChange: (condition: ConditionNode) => void;
    onRemove?: () => void;
    depth?: number;
}

const COMPOSITE_TYPES = new Set(["all", "any", "not"]);
const KEY_VALUE_TYPES = new Set([
    "global_eq", "global_ne", "global_gt", "global_gte", "global_lt", "global_lte",
    "handler_state_eq", "handler_state_ne",
]);
const KEY_ONLY_TYPES = new Set(["global_exists", "global_not_exists"]);

const EXPRESSION_TYPES = new Set(["expr"]);

const DEPRECATED_TYPES = new Set([
    "global_eq", "global_ne", "global_gt", "global_gte", "global_lt", "global_lte",
    "global_exists", "global_not_exists", "handler_state_eq", "handler_state_ne",
]);

const CONDITION_GROUPS: { label: string; keys: string[]; deprecated?: boolean }[] = [
    { label: "Composite", keys: ["all", "any", "not"] },
    { label: "Expression", keys: ["expr"] },
    { label: "State (deprecated - use expr)", keys: ["global_eq", "global_ne", "global_gt", "global_gte", "global_lt", "global_lte", "global_exists", "global_not_exists"], deprecated: true },
    { label: "Handler (deprecated - use expr)", keys: ["handler_state_eq", "handler_state_ne"], deprecated: true },
];

function getConditionColor(type: string): string {
    if (COMPOSITE_TYPES.has(type)) return "text-accent-blue";
    if (EXPRESSION_TYPES.has(type)) return "text-accent-violet";
    if (KEY_VALUE_TYPES.has(type) || KEY_ONLY_TYPES.has(type)) return "text-accent-amber";
    return "text-accent-teal";
}

function getConditionDescription(type: string, schema: ScriptSchema): string | undefined {
    return schema.builtinConditions?.[type]?.description ?? schema.conditions?.[type]?.description;
}

function extractKeyValue(params: unknown): { key: string; value: unknown } {
    if (typeof params !== "object" || params === null) return { key: "", value: "" };
    const obj = params as Record<string, unknown>;
    if ("key" in obj && "value" in obj) return { key: String(obj.key), value: obj.value };
    if ("key" in obj) return { key: String(obj.key), value: "" };
    const entries = Object.entries(obj);
    if (entries.length === 1) return { key: entries[0][0], value: entries[0][1] };
    return { key: "", value: "" };
}

function extractKey(params: unknown): string {
    if (typeof params === "string") return params;
    if (typeof params !== "object" || params === null) return "";
    const obj = params as Record<string, unknown>;
    if ("key" in obj) return String(obj.key);
    const entries = Object.entries(obj);
    if (entries.length === 1) return entries[0][0];
    return "";
}

function ConditionTypePicker({ schema, currentType, onSelect, onClose }: {
    schema: ScriptSchema;
    currentType: string;
    onSelect: (type: string) => void;
    onClose: () => void;
}) {
    const ref = useRef<HTMLDivElement>(null);

    useEffect(() => {
        function handleClickOutside(e: MouseEvent) {
            if (ref.current && !ref.current.contains(e.target as Node)) {
                onClose();
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, [onClose]);

    const namedConditions = Object.entries(schema.conditions).sort(([a], [b]) => a.localeCompare(b));

    return (
        <div ref={ref} className="absolute z-50 mt-1 left-0 w-80 rounded-md border border-border bg-popover shadow-lg max-h-72 overflow-y-auto">
            {CONDITION_GROUPS.map(({ label, keys, deprecated }) => (
                <div key={label}>
                    <div className={`px-2 py-1 text-[10px] font-semibold uppercase tracking-wider sticky top-0 bg-popover ${deprecated ? "text-muted-foreground/30" : "text-muted-foreground/60"}`}>
                        {label}
                    </div>
                    {keys.map((key) => {
                        const desc = getConditionDescription(key, schema);
                        return (
                            <button
                                key={key}
                                className={`flex w-full items-start gap-2 rounded px-2 py-1.5 text-left hover:bg-accent ${key === currentType ? "bg-accent/50" : ""} ${deprecated ? "opacity-50" : ""}`}
                                onMouseDown={(e) => {
                                    e.preventDefault();
                                    onSelect(key);
                                }}
                            >
                                <code className={`text-xs font-mono font-semibold shrink-0 ${getConditionColor(key)}`}>{key}</code>
                                {desc && <span className="text-[11px] text-muted-foreground leading-tight">{desc}</span>}
                            </button>
                        );
                    })}
                </div>
            ))}
            {namedConditions.length > 0 && (
                <div>
                    <div className="px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60 sticky top-0 bg-popover">
                        Named Conditions
                    </div>
                    {namedConditions.map(([name, def]) => (
                        <button
                            key={name}
                            className={`flex w-full items-start gap-2 rounded px-2 py-1.5 text-left hover:bg-accent ${name === currentType ? "bg-accent/50" : ""}`}
                            onMouseDown={(e) => {
                                e.preventDefault();
                                onSelect(name);
                            }}
                        >
                            <code className={`text-xs font-mono font-semibold shrink-0 ${getConditionColor(name)}`}>{name}</code>
                            {def.description && <span className="text-[11px] text-muted-foreground leading-tight">{def.description}</span>}
                        </button>
                    ))}
                </div>
            )}
        </div>
    );
}

export function ExprHint() {
    const { openHelp } = useExprContext();
    return (
        <div className="flex items-center gap-2">
            <span className="text-[10px] text-muted-foreground/50 font-mono">
                var.* global.* const.* save.* self player source prop.* param.*
            </span>
            <ExpressionHelpLink onClick={openHelp} />
        </div>
    );
}

export function ConditionEditor({ condition, schema, onChange, onRemove, depth = 0 }: ConditionEditorProps) {
    const [pickerOpen, setPickerOpen] = useState(false);
    const maxDepth = 6;
    if (depth > maxDepth) {
        return <div className="text-xs text-muted-foreground italic">Max nesting depth reached</div>;
    }

    const handleTypeChange = (newType: string) => {
        setPickerOpen(false);
        if (newType === condition.type) return;
        if (COMPOSITE_TYPES.has(newType)) {
            if (newType === "not") {
                onChange({ type: newType, params: { type: "expr", params: "" } });
            } else {
                onChange({ type: newType, params: [] });
            }
        } else if (EXPRESSION_TYPES.has(newType)) {
            onChange({ type: newType, params: "" });
        } else if (KEY_VALUE_TYPES.has(newType)) {
            onChange({ type: newType, params: { key: "", value: "" } });
        } else if (KEY_ONLY_TYPES.has(newType)) {
            onChange({ type: newType, params: { key: "" } });
        } else if (schema.conditions[newType]) {
            onChange({ type: newType, params: {} });
        } else {
            onChange({ type: newType, params: {} });
        }
    };

    const desc = getConditionDescription(condition.type, schema);

    return (
        <div className={`rounded border border-border/60 bg-muted/20 ${depth > 0 ? "ml-3" : ""}`}>
            <div className="flex items-center gap-1.5 px-2 py-1 relative">
                <button
                    className="flex items-center gap-1 rounded border border-border/60 bg-background px-1.5 py-0.5 text-xs hover:bg-accent"
                    onClick={() => setPickerOpen(!pickerOpen)}
                >
                    <code className={`font-mono font-semibold ${getConditionColor(condition.type)}`}>{condition.type}</code>
                    <ChevronDown className="h-3 w-3 text-muted-foreground" />
                </button>
                {DEPRECATED_TYPES.has(condition.type) && (
                    <span className="text-[9px] text-accent-amber/60">deprecated</span>
                )}
                {desc && <span className="text-[10px] text-muted-foreground truncate">{desc}</span>}
                <div className="flex-1" />
                {onRemove && (
                    <Button variant="ghost" size="sm" className="h-5 w-5 p-0 text-muted-foreground hover:text-destructive" onClick={onRemove}>
                        <Trash2 className="h-3 w-3" />
                    </Button>
                )}
                {pickerOpen && (
                    <ConditionTypePicker
                        schema={schema}
                        currentType={condition.type}
                        onSelect={handleTypeChange}
                        onClose={() => setPickerOpen(false)}
                    />
                )}
            </div>
            <div className="px-2 pb-2">
                <ConditionParams condition={condition} schema={schema} onChange={onChange} depth={depth} />
            </div>
        </div>
    );
}

function ConditionParams({ condition, schema, onChange, depth }: {
    condition: ConditionNode;
    schema: ScriptSchema;
    onChange: (c: ConditionNode) => void;
    depth: number;
}) {
    const exprCtx = useExprContext();

    if (condition.type === "all" || condition.type === "any") {
        const children = (condition.params as ConditionNode[]) ?? [];
        return (
            <div className="space-y-1.5">
                {children.map((child, i) => (
                    <ConditionEditor
                        key={i}
                        condition={child}
                        schema={schema}
                        depth={depth + 1}
                        onChange={(updated) => {
                            const next = [...children];
                            next[i] = updated;
                            onChange({ ...condition, params: next });
                        }}
                        onRemove={() => {
                            onChange({ ...condition, params: children.filter((_, j) => j !== i) });
                        }}
                    />
                ))}
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 text-xs"
                    onClick={() => {
                        const newChild: ConditionNode = { type: "expr", params: "" };
                        onChange({ ...condition, params: [...children, newChild] });
                    }}
                >
                    <Plus className="mr-1 h-3 w-3" />
                    Add condition
                </Button>
            </div>
        );
    }

    if (condition.type === "not") {
        const child = condition.params as ConditionNode;
        return (
            <ConditionEditor
                condition={child}
                schema={schema}
                depth={depth + 1}
                onChange={(updated) => onChange({ ...condition, params: updated })}
            />
        );
    }

    if (EXPRESSION_TYPES.has(condition.type)) {
        return (
            <div className="space-y-1">
                <ExpressionInput
                    value={String(condition.params ?? "")}
                    onChange={(v) => onChange({ ...condition, params: v })}
                    placeholder='e.g. var.count > 3 && global.quest_stage == "complete"'
                    handlerVarKeys={exprCtx.handlerVarKeys}
                    constKeys={exprCtx.constKeys}
                />
                <ExprHint />
            </div>
        );
    }

    if (KEY_VALUE_TYPES.has(condition.type)) {
        const { key, value } = extractKeyValue(condition.params);
        const isGlobal = condition.type.startsWith("global_");
        return (
            <div className="flex items-center gap-1.5">
                <div className="flex-1">
                    {isGlobal ? (
                        <GlobalKeyInput value={key} onChange={(k) => onChange({ ...condition, params: { key: k, value } })} placeholder="key" />
                    ) : (
                        <Input className="h-6 text-xs font-mono" placeholder="key" value={key} onChange={(e) => onChange({ ...condition, params: { key: e.target.value, value } })} />
                    )}
                </div>
                <span className="text-xs text-muted-foreground">=</span>
                <Input
                    className="h-6 w-28 text-xs font-mono"
                    placeholder="value"
                    value={String(value ?? "")}
                    onChange={(e) => {
                        let v: unknown = e.target.value;
                        if (v === "true") v = true;
                        else if (v === "false") v = false;
                        else if (!isNaN(Number(v)) && v !== "") v = Number(v);
                        onChange({ ...condition, params: { key, value: v } });
                    }}
                />
            </div>
        );
    }

    if (KEY_ONLY_TYPES.has(condition.type)) {
        const key = extractKey(condition.params);
        const isGlobal = condition.type.startsWith("global_");
        if (isGlobal) {
            return <GlobalKeyInput value={key} onChange={(k) => onChange({ ...condition, params: { key: k } })} placeholder="key" />;
        }
        return (
            <Input
                className="h-6 text-xs font-mono"
                placeholder="key"
                value={key}
                onChange={(e) => onChange({ ...condition, params: { key: e.target.value } })}
            />
        );
    }

    if (condition.type === "check" || schema.conditions[condition.type]) {
        const condDef = condition.type === "check"
            ? null
            : schema.conditions[condition.type];
        const params = (condition.params ?? {}) as Record<string, unknown>;
        const condName = condition.type === "check" ? String(params.name ?? "") : condition.type;
        const activeDef = condDef ?? schema.conditions[condName];

        return (
            <div className="space-y-1">
                {condition.type === "check" && (
                    <Select value={condName} onValueChange={(name) => onChange({ ...condition, params: { name } })}>
                        <SelectTrigger className="h-6 text-xs">
                            <SelectValue placeholder="Select condition" />
                        </SelectTrigger>
                        <SelectContent>
                            {Object.entries(schema.conditions).sort(([a], [b]) => a.localeCompare(b)).map(([n, def]) => (
                                <SelectItem key={n} value={n}>
                                    <span className="font-mono">{n}</span>
                                    {def.description && <span className="ml-2 text-muted-foreground text-[11px]">{def.description}</span>}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                )}
                {activeDef?.params?.map((p) => (
                    <div key={p.name} className="flex items-center gap-1.5">
                        <span className="text-xs text-muted-foreground w-20 shrink-0">{p.name}</span>
                        {p.name === "zone" || p.name === "zone_id" ? (
                            <div className="flex-1">
                                <ZoneIdInput
                                    value={String(params[p.name] ?? "")}
                                    placeholder={p.description}
                                    onChange={(v) => onChange({ ...condition, params: { ...params, [p.name]: v || undefined } })}
                                />
                            </div>
                        ) : (
                            <Input
                                className="h-6 flex-1 text-xs font-mono"
                                placeholder={p.description}
                                value={String(params[p.name] ?? "")}
                                onChange={(e) => {
                                    let v: unknown = e.target.value;
                                    if (p.type === "number" && !isNaN(Number(v)) && v !== "") v = Number(v);
                                    if (p.type === "bool") v = v === "true";
                                    onChange({ ...condition, params: { ...params, [p.name]: v } });
                                }}
                            />
                        )}
                    </div>
                ))}
            </div>
        );
    }

    return <RawConditionEditor condition={condition} onChange={onChange} />;
}

function RawConditionEditor({ condition, onChange }: { condition: ConditionNode; onChange: (c: ConditionNode) => void }) {
    const [rawText, setRawText] = useState(JSON.stringify(condition.params, null, 2));
    return (
        <div className="space-y-1">
            <span className="text-xs text-muted-foreground">Unknown condition type: {condition.type}</span>
            <textarea
                className="w-full rounded border border-border bg-background p-1 font-mono text-xs"
                rows={3}
                value={rawText}
                onChange={(e) => {
                    setRawText(e.target.value);
                    try {
                        const parsed = JSON.parse(e.target.value);
                        onChange({ ...condition, params: parsed });
                    } catch {
                        // invalid JSON, keep local state
                    }
                }}
            />
        </div>
    );
}

export function createEmptyCondition(): ConditionNode {
    return { type: "expr", params: "" };
}
