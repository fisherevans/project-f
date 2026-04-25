import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import type { ParamDef, StepKindDef } from "@/types/scripts";

interface StepParamFormProps {
    stepKind: StepKindDef;
    params: unknown;
    onChange: (params: unknown) => void;
    excludeKeys?: Set<string>;
}

export function StepParamForm({ stepKind, params, onChange, excludeKeys }: StepParamFormProps) {
    const style = stepKind.paramStyle;

    if (style === "string") {
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
            return (
                <div className="space-y-1">
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
                            Switch to map form
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
                    <MapParamFields defs={defs} params={mapParams} onChange={(updated) => onChange(updated)} />
                    <button
                        className="text-[10px] text-muted-foreground hover:text-foreground"
                        onClick={() => {
                            const firstParam = stepKind.params?.[0];
                            onChange(firstParam ? mapParams[firstParam.name] ?? "" : "");
                        }}
                    >
                        Switch to simple form
                    </button>
                </div>
            );
        }

        return <MapParamFields defs={defs} params={mapParams} onChange={(updated) => onChange(updated)} />;
    }

    return null;
}

function MapParamFields({ defs, params, onChange }: {
    defs: ParamDef[];
    params: Record<string, unknown>;
    onChange: (params: Record<string, unknown>) => void;
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
                <ParamField key={p.name} def={p} value={params[p.name]} onChange={(v) => updateField(p.name, v)} />
            ))}
        </div>
    );
}

function ParamField({ def, value, onChange }: {
    def: ParamDef;
    value: unknown;
    onChange: (value: unknown) => void;
}) {
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
