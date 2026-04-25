import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Trash2 } from "lucide-react";
import type { ConditionNode, ScriptSchema } from "@/types/scripts";
import { parseCondition, serializeCondition } from "@/lib/scriptUtils";

interface ConditionEditorProps {
    condition: ConditionNode;
    schema: ScriptSchema;
    onChange: (condition: ConditionNode) => void;
    onRemove?: () => void;
    depth?: number;
}

const BUILTIN_CONDITION_TYPES = [
    "global_eq", "global_ne", "global_gt", "global_gte", "global_lt", "global_lte",
    "global_exists", "global_not_exists",
    "handler_state_eq", "handler_state_ne",
    "all", "any", "not",
    "check",
];

const COMPOSITE_TYPES = new Set(["all", "any", "not"]);
const KEY_VALUE_TYPES = new Set([
    "global_eq", "global_ne", "global_gt", "global_gte", "global_lt", "global_lte",
    "handler_state_eq", "handler_state_ne",
]);
const KEY_ONLY_TYPES = new Set(["global_exists", "global_not_exists"]);

function getConditionLabel(type: string): string {
    switch (type) {
        case "all": return "ALL of";
        case "any": return "ANY of";
        case "not": return "NOT";
        case "check": return "Check";
        default: return type;
    }
}

function getConditionColor(type: string): string {
    if (COMPOSITE_TYPES.has(type)) return "text-blue-400";
    if (KEY_VALUE_TYPES.has(type) || KEY_ONLY_TYPES.has(type)) return "text-amber-400";
    return "text-emerald-400";
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

export function ConditionEditor({ condition, schema, onChange, onRemove, depth = 0 }: ConditionEditorProps) {
    const maxDepth = 6;
    if (depth > maxDepth) {
        return <div className="text-xs text-muted-foreground italic">Max nesting depth reached</div>;
    }

    const handleTypeChange = (newType: string) => {
        if (COMPOSITE_TYPES.has(newType)) {
            if (newType === "not") {
                onChange({ type: newType, params: { type: "global_eq", params: { key: "", value: "" } } });
            } else {
                onChange({ type: newType, params: [] });
            }
        } else if (KEY_VALUE_TYPES.has(newType)) {
            onChange({ type: newType, params: { key: "", value: "" } });
        } else if (KEY_ONLY_TYPES.has(newType)) {
            onChange({ type: newType, params: { key: "" } });
        } else if (newType === "check") {
            const firstCondName = Object.keys(schema.conditions)[0] ?? "";
            onChange({ type: newType, params: { name: firstCondName } });
        } else {
            onChange({ type: newType, params: {} });
        }
    };

    return (
        <div className={`rounded border border-border/60 bg-muted/20 ${depth > 0 ? "ml-3" : ""}`}>
            <div className="flex items-center gap-1.5 px-2 py-1">
                <Select value={condition.type} onValueChange={handleTypeChange}>
                    <SelectTrigger className="h-6 w-auto min-w-[140px] text-xs">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="_label_composite" disabled>Composite</SelectItem>
                        {["all", "any", "not"].map((t) => (
                            <SelectItem key={t} value={t}>{t}</SelectItem>
                        ))}
                        <SelectItem value="_label_global" disabled>Global State</SelectItem>
                        {["global_eq", "global_ne", "global_gt", "global_gte", "global_lt", "global_lte", "global_exists", "global_not_exists"].map((t) => (
                            <SelectItem key={t} value={t}>{t}</SelectItem>
                        ))}
                        <SelectItem value="_label_handler" disabled>Handler State</SelectItem>
                        {["handler_state_eq", "handler_state_ne"].map((t) => (
                            <SelectItem key={t} value={t}>{t}</SelectItem>
                        ))}
                        <SelectItem value="_label_named" disabled>Named Conditions</SelectItem>
                        {Object.keys(schema.conditions).sort().map((name) => (
                            <SelectItem key={name} value={name}>{name}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                {onRemove && (
                    <Button variant="ghost" size="sm" className="h-5 w-5 p-0 text-muted-foreground hover:text-destructive" onClick={onRemove}>
                        <Trash2 className="h-3 w-3" />
                    </Button>
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
                        const newChild: ConditionNode = { type: "global_eq", params: { key: "", value: "" } };
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

    if (KEY_VALUE_TYPES.has(condition.type)) {
        const { key, value } = extractKeyValue(condition.params);
        return (
            <div className="flex items-center gap-1.5">
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    placeholder="key"
                    value={key}
                    onChange={(e) => onChange({ ...condition, params: { key: e.target.value, value } })}
                />
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
                            {Object.keys(schema.conditions).sort().map((n) => (
                                <SelectItem key={n} value={n}>{n}</SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                )}
                {activeDef?.params?.map((p) => (
                    <div key={p.name} className="flex items-center gap-1.5">
                        <span className="text-xs text-muted-foreground w-20 shrink-0">{p.name}</span>
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
                    </div>
                ))}
            </div>
        );
    }

    // Fallback: raw JSON editor for unknown condition types
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
    return { type: "global_eq", params: { key: "", value: "" } };
}
