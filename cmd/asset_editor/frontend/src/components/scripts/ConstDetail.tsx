import { useState } from "react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

interface ConstDetailProps {
    name: string;
    value: unknown;
    onChange: (value: unknown) => void;
}

type ConstType = "string" | "number" | "boolean" | "list" | "map" | "json";

function detectType(value: unknown): ConstType {
    if (typeof value === "string") return "string";
    if (typeof value === "number") return "number";
    if (typeof value === "boolean") return "boolean";
    if (Array.isArray(value)) return "list";
    if (typeof value === "object" && value !== null) return "map";
    return "json";
}

function ListEditor({ items, onChange }: { items: unknown[]; onChange: (items: unknown[]) => void }) {
    const addItem = () => onChange([...items, ""]);
    const removeItem = (idx: number) => onChange(items.filter((_, i) => i !== idx));
    const updateItem = (idx: number, raw: string) => {
        const next = [...items];
        next[idx] = coerceInput(raw);
        onChange(next);
    };

    return (
        <div className="space-y-1">
            <div className="flex items-center gap-1.5 mb-1">
                <span className="text-xs text-muted-foreground">{items.length} item{items.length !== 1 ? "s" : ""}</span>
                <div className="flex-1" />
                <Button variant="ghost" size="sm" className="h-5 text-[10px] text-muted-foreground" onClick={addItem}>
                    + Add
                </Button>
            </div>
            {items.map((item, i) => (
                <div key={i} className="flex items-center gap-1 group/item">
                    <span className="text-[10px] text-muted-foreground/40 w-4 text-right shrink-0">{i}</span>
                    <Input
                        className="h-6 flex-1 text-xs font-mono"
                        value={displayValue(item)}
                        onChange={(e) => updateItem(i, e.target.value)}
                    />
                    <button
                        className="text-muted-foreground hover:text-destructive opacity-0 group-hover/item:opacity-100 shrink-0"
                        onClick={() => removeItem(i)}
                    >
                        <span className="text-xs">x</span>
                    </button>
                </div>
            ))}
        </div>
    );
}

function MapEditor({ map, onChange }: { map: Record<string, unknown>; onChange: (map: Record<string, unknown>) => void }) {
    const entries = Object.entries(map);
    const addEntry = () => {
        const key = `key_${entries.length}`;
        onChange({ ...map, [key]: "" });
    };
    const removeEntry = (key: string) => {
        const next = { ...map };
        delete next[key];
        onChange(next);
    };
    const updateKey = (oldKey: string, newKey: string) => {
        if (newKey === oldKey || !newKey) return;
        const next: Record<string, unknown> = {};
        for (const [k, v] of entries) {
            next[k === oldKey ? newKey : k] = v;
        }
        onChange(next);
    };
    const updateValue = (key: string, raw: string) => {
        onChange({ ...map, [key]: coerceInput(raw) });
    };

    return (
        <div className="space-y-1">
            <div className="flex items-center gap-1.5 mb-1">
                <span className="text-xs text-muted-foreground">{entries.length} entr{entries.length !== 1 ? "ies" : "y"}</span>
                <div className="flex-1" />
                <Button variant="ghost" size="sm" className="h-5 text-[10px] text-muted-foreground" onClick={addEntry}>
                    + Add
                </Button>
            </div>
            {entries.map(([k, v]) => (
                <div key={k} className="flex items-center gap-1 group/entry">
                    <Input
                        className="h-6 w-32 text-xs font-mono"
                        value={k}
                        onChange={(e) => updateKey(k, e.target.value)}
                    />
                    <Input
                        className="h-6 flex-1 text-xs font-mono"
                        value={displayValue(v)}
                        onChange={(e) => updateValue(k, e.target.value)}
                    />
                    <button
                        className="text-muted-foreground hover:text-destructive opacity-0 group-hover/entry:opacity-100 shrink-0"
                        onClick={() => removeEntry(k)}
                    >
                        <span className="text-xs">x</span>
                    </button>
                </div>
            ))}
        </div>
    );
}

function coerceInput(raw: string): unknown {
    if (raw === "true") return true;
    if (raw === "false") return false;
    if (raw !== "" && !isNaN(Number(raw))) return Number(raw);
    return raw;
}

function displayValue(val: unknown): string {
    if (val === null || val === undefined) return "";
    if (typeof val === "object") return JSON.stringify(val);
    return String(val);
}

export function ConstDetail({ name, value, onChange }: ConstDetailProps) {
    const detectedType = detectType(value);
    const [jsonMode, setJsonMode] = useState(false);
    const [jsonText, setJsonText] = useState(() => JSON.stringify(value, null, 2));
    const [jsonError, setJsonError] = useState<string | null>(null);

    const applyJson = () => {
        try {
            const parsed = JSON.parse(jsonText);
            onChange(parsed);
            setJsonError(null);
            setJsonMode(false);
        } catch (e) {
            setJsonError(String(e));
        }
    };

    return (
        <div className="flex h-full flex-col overflow-hidden">
            <div className="flex items-center gap-2 border-b border-border px-3 py-1.5">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-accent-amber px-1.5 py-0.5 rounded bg-accent-amber-tint">const</span>
                <code className="text-sm font-mono font-semibold">{name}</code>
                <span className="text-[10px] text-muted-foreground/50">{detectedType}</span>
                <div className="flex-1" />
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-5 text-[10px] text-muted-foreground"
                    onClick={() => {
                        if (!jsonMode) {
                            setJsonText(JSON.stringify(value, null, 2));
                            setJsonError(null);
                        }
                        setJsonMode(!jsonMode);
                    }}
                >
                    {jsonMode ? "Visual" : "JSON"}
                </Button>
            </div>
            <ScrollArea className="flex-1 overflow-hidden">
                <div className="p-3">
                    <div className="text-[10px] text-muted-foreground/60 mb-2">
                        Constants are read-only at runtime. Access as <code className="font-mono">const.{name}</code> in expressions.
                    </div>
                    {jsonMode ? (
                        <div className="space-y-2">
                            <textarea
                                className="min-h-[200px] w-full rounded-md border border-border bg-background px-3 py-2 font-mono text-xs outline-none focus:ring-1 focus:ring-ring"
                                value={jsonText}
                                onChange={(e) => {
                                    setJsonText(e.target.value);
                                    setJsonError(null);
                                }}
                                spellCheck={false}
                            />
                            {jsonError && (
                                <div className="text-xs text-destructive">{jsonError}</div>
                            )}
                            <Button size="sm" className="h-6 text-xs" onClick={applyJson}>
                                Apply
                            </Button>
                        </div>
                    ) : (
                        <div>
                            {detectedType === "string" && (
                                <Input
                                    className="h-7 text-xs font-mono"
                                    value={value as string}
                                    onChange={(e) => onChange(e.target.value)}
                                />
                            )}
                            {detectedType === "number" && (
                                <Input
                                    className="h-7 text-xs font-mono w-40"
                                    type="number"
                                    value={value as number}
                                    onChange={(e) => onChange(Number(e.target.value))}
                                />
                            )}
                            {detectedType === "boolean" && (
                                <label className="flex items-center gap-2 text-xs">
                                    <input
                                        type="checkbox"
                                        checked={value as boolean}
                                        onChange={(e) => onChange(e.target.checked)}
                                    />
                                    {String(value)}
                                </label>
                            )}
                            {detectedType === "list" && (
                                <ListEditor
                                    items={value as unknown[]}
                                    onChange={onChange}
                                />
                            )}
                            {detectedType === "map" && (
                                <MapEditor
                                    map={value as Record<string, unknown>}
                                    onChange={onChange}
                                />
                            )}
                            {detectedType === "json" && (
                                <div className="text-xs text-muted-foreground">
                                    Unsupported type. Switch to JSON mode to edit.
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </ScrollArea>
        </div>
    );
}
