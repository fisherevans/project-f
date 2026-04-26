import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Plus, X } from "lucide-react";
import { useMemo } from "react";
import { SoundPicker } from "./inputs/SoundPicker";
import { GlobalKeyInput } from "./inputs/GlobalKeyInput";
import { EntityRefInput } from "./inputs/EntityRefInput";
import { HandlerRefInput } from "./inputs/HandlerRefInput";
import { ZoneIdInput } from "./inputs/ZoneIdInput";
import { ExpressionInput } from "./inputs/ExpressionInput";
import { SchemaSelect } from "./inputs/SchemaSelect";
import { ExprHint } from "./ConditionEditor";
import { useExprContext } from "./ExprContext";
import type { ParamDef, StepKindDef, ScriptSchema } from "@/types/scripts";

interface StepParamFormProps {
    stepKind: StepKindDef;
    params: unknown;
    onChange: (params: unknown) => void;
    excludeKeys?: Set<string>;
    schema?: ScriptSchema;
}

const ENTITY_PARAM_NAMES = new Set(["entity", "target", "to_entity", "entity_id", "follow_entity", "facing_entity", "created_by"]);
const ZONE_PARAM_NAMES = new Set(["zone", "zone_id", "active_player_zone"]);
const SOUND_PARAM_NAMES = new Set(["sound"]);
const GLOBAL_KEY_STEP_KINDS = new Set(["set_world_state", "set_run_state"]);
const ACTION_NAME_STEP_KINDS = new Set(["action", "ref", "custom_action"]);
const EXPR_PARAM_HINTS: Record<string, Set<string>> = {
    "set_var": new Set(["value"]),
    "if": new Set(["when"]),
    "while": new Set(["when"]),
    "switch": new Set(["on"]),
};

export function StepParamForm({ stepKind, params, onChange, excludeKeys, schema }: StepParamFormProps) {
    const style = stepKind.paramStyle;

    if (style === "string") {
        if (stepKind.name === "play_sound") {
            return <SoundPicker value={String(params ?? "")} onChange={onChange} />;
        }
        return (
            <Input
                className="h-6 text-xs font-mono"
                value={String(params ?? "")}
                onChange={(e) => onChange(e.target.value)}
                placeholder={stepKind.params?.[0]?.description}
            />
        );
    }

    if (style === "number") {
        return (
            <Input
                className="h-6 w-24 text-xs font-mono"
                type="number"
                step="any"
                value={String(params ?? 0)}
                onChange={(e) => onChange(Number(e.target.value))}
                placeholder={stepKind.params?.[0]?.description}
            />
        );
    }

    if (style === "string_or_map") {
        if (typeof params === "string" || typeof params === "number" || typeof params === "boolean") {
            const hasMapParams = (stepKind.params?.length ?? 0) > 1;
            const isActionLike = stepKind.name === "action" || stepKind.name === "ref" || stepKind.name === "custom_action";
            const inlineInput = stepKind.name === "play_sound" ? (
                <SoundPicker value={String(params)} onChange={onChange} />
            ) : isActionLike ? (
                <HandlerRefInput value={String(params)} onChange={(v) => onChange(v || undefined)} schema={schema} />
            ) : (
                <Input
                    className="h-6 text-xs font-mono"
                    value={String(params)}
                    onChange={(e) => {
                        let v: unknown = e.target.value;
                        if (!isNaN(Number(v)) && v !== "") v = Number(v);
                        onChange(v);
                    }}
                    placeholder={stepKind.params?.[0]?.description}
                />
            );
            return (
                <div className="space-y-1">
                    {inlineInput}
                    {hasMapParams && (
                        <button
                            className="text-[10px] text-muted-foreground hover:text-foreground"
                            onClick={() => {
                                const mapParams: Record<string, unknown> = {};
                                const firstParam = stepKind.params?.[0];
                                if (firstParam) mapParams[firstParam.name] = params;
                                onChange(mapParams);
                            }}
                        >
                            Show all options
                        </button>
                    )}
                </div>
            );
        }
        // Fall through to map rendering
    }

    if (style === "map" || style === "string_or_map") {
        const mapParams = (typeof params === "object" && params !== null && !Array.isArray(params))
            ? params as Record<string, unknown>
            : {};
        const defs = (stepKind.params ?? []).filter((p) => p.type !== "steps" && !(excludeKeys?.has(p.name)));

        if (style === "string_or_map" && defs.length > 0) {
            return (
                <div className="space-y-1">
                    <MapParamFields defs={defs} params={mapParams} onChange={(updated) => onChange(updated)} stepKindName={stepKind.name} schema={schema} />
                    <button
                        className="text-[10px] text-muted-foreground hover:text-foreground"
                        onClick={() => {
                            const firstParam = stepKind.params?.[0];
                            onChange(firstParam ? mapParams[firstParam.name] ?? "" : "");
                        }}
                    >
                        Collapse
                    </button>
                </div>
            );
        }

        return <MapParamFields defs={defs} params={mapParams} onChange={(updated) => onChange(updated)} stepKindName={stepKind.name} schema={schema} />;
    }

    return null;
}

function MapParamFields({ defs, params, onChange, stepKindName, schema }: {
    defs: ParamDef[];
    params: Record<string, unknown>;
    onChange: (params: Record<string, unknown>) => void;
    stepKindName?: string;
    schema?: ScriptSchema;
}) {
    const updateField = (name: string, value: unknown) => {
        const next = { ...params, [name]: value };
        if (value === "" || value === undefined || value === null) {
            delete next[name];
        }
        onChange(next);
    };

    return (
        <div className="space-y-1">
            {defs.map((p) => (
                <ParamField key={p.name} def={p} value={params[p.name]} onChange={(v) => updateField(p.name, v)} stepKindName={stepKindName} schema={schema} parentParams={params} />
            ))}
        </div>
    );
}

function DynamicMapEditor({ value, onChange, valueLabel }: {
    value: unknown;
    onChange: (value: unknown) => void;
    valueLabel?: string;
}) {
    const map = (typeof value === "object" && value !== null && !Array.isArray(value))
        ? value as Record<string, unknown>
        : {};
    const entries = Object.entries(map);

    const addEntry = () => {
        const key = `param_${entries.length}`;
        onChange({ ...map, [key]: "" });
    };

    const removeEntry = (key: string) => {
        const next = { ...map };
        delete next[key];
        onChange(next);
    };

    const updateKey = (oldKey: string, newKey: string) => {
        if (newKey === oldKey) return;
        const next: Record<string, unknown> = {};
        for (const [k, v] of Object.entries(map)) {
            next[k === oldKey ? newKey : k] = v;
        }
        onChange(next);
    };

    const updateValue = (key: string, val: unknown) => {
        onChange({ ...map, [key]: val });
    };

    return (
        <div className="space-y-1">
            {entries.map(([key, val], i) => (
                <div key={i} className="flex items-center gap-1">
                    <Input
                        className="h-6 w-28 text-xs font-mono"
                        value={key}
                        onChange={(e) => updateKey(key, e.target.value)}
                        placeholder="key"
                    />
                    <span className="text-muted-foreground text-[10px]">=</span>
                    <Input
                        className="h-6 flex-1 text-xs font-mono"
                        value={String(val ?? "")}
                        onChange={(e) => updateValue(key, e.target.value)}
                        placeholder={valueLabel ?? "value (expression)"}
                    />
                    <button className="text-muted-foreground hover:text-destructive shrink-0" onClick={() => removeEntry(key)}>
                        <X className="h-3 w-3" />
                    </button>
                </div>
            ))}
            <Button variant="ghost" size="sm" className="h-5 text-[10px] text-muted-foreground" onClick={addEntry}>
                <Plus className="mr-1 h-2.5 w-2.5" />
                Add entry
            </Button>
        </div>
    );
}

function CustomActionParamsEditor({ value, onChange, actionName }: {
    value: unknown;
    onChange: (value: unknown) => void;
    actionName: string;
}) {
    const exprCtx = useExprContext();
    const actionDef = exprCtx.customActions[actionName];

    const map = (typeof value === "object" && value !== null && !Array.isArray(value))
        ? value as Record<string, unknown>
        : {};

    const updateField = (key: string, val: unknown) => {
        const next = { ...map, [key]: val };
        if (val === "" || val === undefined || val === null) {
            delete next[key];
        }
        onChange(Object.keys(next).length > 0 ? next : undefined);
    };

    const declaredParams = actionDef?.params ?? [];
    const declaredNames = new Set(declaredParams.map((p) => p.name));
    const extraKeys = Object.keys(map).filter((k) => !declaredNames.has(k));

    if (declaredParams.length === 0 && extraKeys.length === 0) {
        return (
            <div className="space-y-1">
                <span className="text-xs text-muted-foreground">params</span>
                <DynamicMapEditor value={value} onChange={onChange} valueLabel="value (expression)" />
            </div>
        );
    }

    return (
        <div className="space-y-1">
            <span className="text-xs text-muted-foreground">params</span>
            {declaredParams.map((p) => (
                <div key={p.name} className="flex items-start gap-1.5">
                    <span className="text-xs text-muted-foreground w-24 shrink-0 pt-1">
                        {p.name}{p.default === undefined ? "*" : ""}
                    </span>
                    <div className="flex-1 space-y-0.5">
                        <ExpressionInput
                            value={String(map[p.name] ?? "")}
                            onChange={(v) => updateField(p.name, v || undefined)}
                            placeholder={p.default !== undefined ? `default: ${p.default}` : undefined}
                            handlerVarKeys={exprCtx.handlerVarKeys}
                            constKeys={exprCtx.constKeys}
                        />
                        {p.description && (
                            <div className="text-[10px] text-muted-foreground/60 leading-tight">{p.description}</div>
                        )}
                    </div>
                </div>
            ))}
            {extraKeys.map((key) => (
                <div key={key} className="flex items-center gap-1.5">
                    <Input
                        className="h-6 w-24 text-xs font-mono text-accent-amber"
                        value={key}
                        readOnly
                        title="Unrecognized param"
                    />
                    <span className="text-muted-foreground text-[10px]">=</span>
                    <div className="flex-1">
                        <ExpressionInput
                            value={String(map[key] ?? "")}
                            onChange={(v) => updateField(key, v || undefined)}
                            handlerVarKeys={exprCtx.handlerVarKeys}
                            constKeys={exprCtx.constKeys}
                        />
                    </div>
                    <button className="text-muted-foreground hover:text-destructive shrink-0" onClick={() => {
                        const next = { ...map };
                        delete next[key];
                        onChange(Object.keys(next).length > 0 ? next : undefined);
                    }}>
                        <X className="h-3 w-3" />
                    </button>
                </div>
            ))}
            <Button variant="ghost" size="sm" className="h-5 text-[10px] text-muted-foreground" onClick={() => {
                const key = `param_${Object.keys(map).length}`;
                onChange({ ...map, [key]: "" });
            }}>
                <Plus className="mr-1 h-2.5 w-2.5" />
                Add extra param
            </Button>
        </div>
    );
}

function ParamField({ def, value, onChange, stepKindName, schema, parentParams }: {
    def: ParamDef;
    value: unknown;
    onChange: (value: unknown) => void;
    stepKindName?: string;
    schema?: ScriptSchema;
    parentParams?: Record<string, unknown>;
}) {
    const exprCtx = useExprContext();
    const isExprField = stepKindName ? EXPR_PARAM_HINTS[stepKindName]?.has(def.name) : false;

    if (isExprField) {
        return (
            <div className="space-y-0.5">
                <div className="flex items-center gap-1.5">
                    <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
                    <div className="flex-1">
                        <ExpressionInput
                            value={String(value ?? "")}
                            onChange={(v) => onChange(v || undefined)}
                            placeholder={def.default !== undefined ? String(def.default) : def.description}
                            handlerVarKeys={exprCtx.handlerVarKeys}
                            constKeys={exprCtx.constKeys}
                        />
                    </div>
                </div>
                <div className="pl-[6.5rem]"><ExprHint /></div>
            </div>
        );
    }

    if (ENTITY_PARAM_NAMES.has(def.name)) {
        return (
            <div className="flex items-center gap-1.5">
                <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
                <div className="flex-1"><EntityRefInput value={String(value ?? "")} onChange={(v) => onChange(v || undefined)} /></div>
            </div>
        );
    }

    if (ZONE_PARAM_NAMES.has(def.name)) {
        return (
            <div className="flex items-center gap-1.5">
                <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
                <div className="flex-1"><ZoneIdInput value={String(value ?? "")} onChange={(v) => onChange(v || undefined)} /></div>
            </div>
        );
    }

    if (def.name === "key" && stepKindName && GLOBAL_KEY_STEP_KINDS.has(stepKindName)) {
        return (
            <div className="flex items-center gap-1.5">
                <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
                <div className="flex-1"><GlobalKeyInput value={String(value ?? "")} onChange={(v) => onChange(v || undefined)} /></div>
            </div>
        );
    }

    if (SOUND_PARAM_NAMES.has(def.name)) {
        return (
            <div className="flex items-center gap-1.5">
                <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
                <div className="flex-1"><SoundPicker value={String(value ?? "")} onChange={(v) => onChange(v || undefined)} /></div>
            </div>
        );
    }

    if (def.name === "name" && stepKindName && ACTION_NAME_STEP_KINDS.has(stepKindName)) {
        return (
            <div className="flex items-center gap-1.5">
                <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
                <div className="flex-1"><HandlerRefInput value={String(value ?? "")} onChange={(v) => onChange(v || undefined)} schema={schema} /></div>
            </div>
        );
    }

    if (def.name === "condition" && def.type === "string" && schema) {
        return <ConditionNameField def={def} value={value} onChange={onChange} schema={schema} />;
    }

    if (def.name === "sequence" && def.type === "string" && schema) {
        return <SequenceNameField def={def} value={value} onChange={onChange} schema={schema} />;
    }

    if (def.type === "map") {
        if (def.name === "params" && stepKindName === "custom_action" && parentParams) {
            return <CustomActionParamsEditor value={value} onChange={onChange} actionName={String(parentParams.name ?? "")} />;
        }
        return (
            <div className="space-y-1">
                <span className="text-xs text-muted-foreground">{def.name}{def.required ? "*" : ""}</span>
                <DynamicMapEditor value={value} onChange={onChange} valueLabel={def.description} />
                {stepKindName && EXPR_PARAM_HINTS[stepKindName]?.has(def.name) && <ExprHint />}
            </div>
        );
    }

    if (def.type === "bool") {
        return (
            <div className="flex items-center gap-2">
                <Switch
                    checked={Boolean(value)}
                    onCheckedChange={(checked) => onChange(checked)}
                    className="h-4 w-7"
                />
                <Label className="text-xs text-muted-foreground">{def.name}</Label>
            </div>
        );
    }

    if (def.enum && def.enum.length > 0) {
        return (
            <div className="flex items-center gap-1.5">
                <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
                <Select value={String(value ?? "")} onValueChange={onChange}>
                    <SelectTrigger className="h-6 flex-1 text-xs">
                        <SelectValue placeholder={`Select ${def.name}`} />
                    </SelectTrigger>
                    <SelectContent>
                        {def.enum.map((v) => (
                            <SelectItem key={v} value={v}>{v}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>
        );
    }

    if (def.type === "number") {
        return (
            <div className="flex items-center gap-1.5">
                <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    type="number"
                    step="any"
                    value={value !== undefined && value !== null ? String(value) : ""}
                    placeholder={def.default !== undefined ? String(def.default) : def.description}
                    onChange={(e) => {
                        const v = e.target.value;
                        onChange(v === "" ? undefined : Number(v));
                    }}
                />
            </div>
        );
    }

    if (def.type === "any") {
        return (
            <div className="flex items-center gap-1.5">
                <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    value={value !== undefined && value !== null ? String(value) : ""}
                    placeholder={def.description}
                    onChange={(e) => {
                        let v: unknown = e.target.value;
                        if (v === "true") v = true;
                        else if (v === "false") v = false;
                        else if (v !== "" && !isNaN(Number(v))) v = Number(v);
                        onChange(v === "" ? undefined : v);
                    }}
                />
            </div>
        );
    }

    // Default: string
    return (
        <div className="flex items-center gap-1.5">
            <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
            <Input
                className="h-6 flex-1 text-xs font-mono"
                value={String(value ?? "")}
                placeholder={def.default !== undefined ? String(def.default) : def.description}
                onChange={(e) => onChange(e.target.value || undefined)}
            />
        </div>
    );
}

function ConditionNameField({ def, value, onChange, schema }: {
    def: ParamDef;
    value: unknown;
    onChange: (value: unknown) => void;
    schema: ScriptSchema;
}) {
    const options = useMemo(() => {
        const items: { value: string; label?: string; detail?: string; group?: string }[] = [];
        for (const [name, cond] of Object.entries(schema.builtinConditions ?? {})) {
            items.push({ value: name, detail: cond.description, group: cond.isComposite ? "composite" : "builtin" });
        }
        for (const [name, cond] of Object.entries(schema.conditions ?? {})) {
            items.push({ value: name, detail: cond.description, group: "named" });
        }
        items.sort((a, b) => a.value.localeCompare(b.value));
        return items;
    }, [schema]);

    return (
        <div className="flex items-center gap-1.5">
            <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
            <div className="flex-1">
                <SchemaSelect
                    value={String(value ?? "")}
                    onChange={(v) => onChange(v || undefined)}
                    options={options}
                    placeholder="Condition name"
                />
            </div>
        </div>
    );
}

function SequenceNameField({ def, value, onChange, schema }: {
    def: ParamDef;
    value: unknown;
    onChange: (value: unknown) => void;
    schema: ScriptSchema;
}) {
    const options = useMemo(() => {
        const items: { value: string; detail?: string; group?: string }[] = [];
        if (schema.actions) {
            for (const [name, action] of Object.entries(schema.actions)) {
                items.push({ value: name, detail: action.description, group: "action" });
            }
        }
        items.sort((a, b) => a.value.localeCompare(b.value));
        return items;
    }, [schema]);

    return (
        <div className="flex items-center gap-1.5">
            <span className="text-xs text-muted-foreground w-24 shrink-0">{def.name}{def.required ? "*" : ""}</span>
            <div className="flex-1">
                <SchemaSelect
                    value={String(value ?? "")}
                    onChange={(v) => onChange(v || undefined)}
                    options={options}
                    placeholder="Sequence name"
                />
            </div>
        </div>
    );
}
