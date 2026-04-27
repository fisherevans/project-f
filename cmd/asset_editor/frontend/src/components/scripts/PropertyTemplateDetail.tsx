import { useState, useRef, useEffect, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Plus, Trash2 } from "lucide-react";
import { useScripts } from "@/api/scripts";
import { KNOWN_ENTITY_PROPERTIES, getKnownProperty } from "@/lib/entityProperties";
import type { KnownEntityProperty } from "@/lib/entityProperties";

interface PropertyTemplateDetailProps {
    name: string;
    template: Record<string, unknown>;
    onChange: (template: Record<string, unknown>) => void;
}

export function PropertyTemplateDetail({ name, template, onChange }: PropertyTemplateDetailProps) {
    const entries = Object.entries(template);
    const [addingKey, setAddingKey] = useState<string | null>(null);

    const addEntry = (key: string) => {
        const k = key.trim();
        if (!k || template[k] !== undefined) return;
        const known = getKnownProperty(k);
        const defaultValue = known?.editorKind === "boolean" ? false : "";
        onChange({ ...template, [k]: defaultValue });
        setAddingKey(null);
    };

    const removeEntry = (key: string) => {
        const next = { ...template };
        delete next[key];
        onChange(next);
    };

    const updateValue = (key: string, value: unknown) => {
        onChange({ ...template, [key]: value });
    };

    const usedKeys = new Set(Object.keys(template));

    return (
        <div className="flex h-full flex-col overflow-hidden">
            <div className="flex items-center gap-2 border-b border-border px-3 py-1.5">
                <code className="text-sm font-mono font-semibold text-accent-green">{name}</code>
                <span className="text-xs text-muted-foreground">property template</span>
            </div>
            <ScrollArea className="flex-1 overflow-hidden">
                <div className="p-3 space-y-3">
                    <div className="text-[10px] text-muted-foreground/60">
                        Default property values applied to Tiled entities that reference this template.
                        Entities inherit these as base properties, overridable per-entity in Tiled.
                    </div>
                    <div className="space-y-2">
                        {entries.map(([key, val]) => (
                            <PropertyRow
                                key={key}
                                propKey={key}
                                value={val}
                                onChange={(v) => updateValue(key, v)}
                                onRemove={() => removeEntry(key)}
                            />
                        ))}
                    </div>
                    {addingKey !== null ? (
                        <PropertyKeyPicker
                            usedKeys={usedKeys}
                            onSelect={addEntry}
                            onCancel={() => setAddingKey(null)}
                        />
                    ) : (
                        <Button variant="outline" size="sm" className="h-7 text-xs gap-1" onClick={() => setAddingKey("")}>
                            <Plus className="h-3 w-3" />
                            Add property
                        </Button>
                    )}
                </div>
            </ScrollArea>
        </div>
    );
}

function PropertyRow({
    propKey,
    value,
    onChange,
    onRemove,
}: {
    propKey: string;
    value: unknown;
    onChange: (value: unknown) => void;
    onRemove: () => void;
}) {
    const known = getKnownProperty(propKey);

    return (
        <div className="group/entry space-y-0.5">
            <div className="flex items-center gap-1.5">
                <span className="text-[11px] font-mono font-medium shrink-0">{propKey}</span>
                {known && (
                    <span className="text-[10px] text-muted-foreground/50">{known.label}</span>
                )}
                <span className="flex-1" />
                <button
                    className="text-muted-foreground hover:text-destructive opacity-0 group-hover/entry:opacity-100 shrink-0"
                    onClick={onRemove}
                >
                    <Trash2 className="h-3 w-3" />
                </button>
            </div>
            {known?.description && (
                <p className="text-[10px] text-muted-foreground/40">{known.description}</p>
            )}
            <PropertyValueEditor propKey={propKey} known={known} value={value} onChange={onChange} />
        </div>
    );
}

function PropertyValueEditor({
    propKey,
    known,
    value,
    onChange,
}: {
    propKey: string;
    known: KnownEntityProperty | undefined;
    value: unknown;
    onChange: (value: unknown) => void;
}) {
    if (known?.editorKind === "handler-ref") {
        return <HandlerSearchInput value={String(value ?? "")} onChange={onChange} />;
    }

    if (known?.editorKind === "enum" && known.enumValues) {
        return (
            <select
                className="h-7 w-full text-xs font-mono rounded border border-border bg-background px-2"
                value={String(value ?? "")}
                onChange={(e) => onChange(e.target.value)}
            >
                {known.enumValues.map((v) => (
                    <option key={v} value={v}>{v || `(default)`}</option>
                ))}
            </select>
        );
    }

    if (known?.editorKind === "yaml-textarea") {
        return (
            <textarea
                className="w-full min-h-[60px] text-xs font-mono rounded border border-border bg-background px-2 py-1 resize-y"
                value={String(value ?? "")}
                onChange={(e) => onChange(e.target.value)}
                placeholder="YAML value"
                rows={3}
            />
        );
    }

    if (known?.editorKind === "boolean") {
        return (
            <label className="flex items-center gap-2 text-xs">
                <input
                    type="checkbox"
                    className="h-3.5 w-3.5"
                    checked={value === true || value === "true"}
                    onChange={(e) => onChange(e.target.checked)}
                />
                <span className="text-muted-foreground">{value === true || value === "true" ? "true" : "false"}</span>
            </label>
        );
    }

    if (known?.editorKind === "number") {
        return (
            <Input
                className="h-7 w-full text-xs font-mono"
                type="number"
                value={value !== undefined && value !== null ? String(value) : ""}
                onChange={(e) => {
                    const v = e.target.value;
                    onChange(v === "" ? "" : Number(v));
                }}
                placeholder="0"
            />
        );
    }

    // Default: text input with auto-typing
    return (
        <RawValueInput value={value} onChange={onChange} />
    );
}

function RawValueInput({ value, onChange }: { value: unknown; onChange: (v: unknown) => void }) {
    const strValue = value !== undefined && value !== null ? String(value) : "";

    const handleChange = (rawValue: string) => {
        let typed: unknown = rawValue;
        if (rawValue === "true") typed = true;
        else if (rawValue === "false") typed = false;
        else if (rawValue !== "" && !isNaN(Number(rawValue))) typed = Number(rawValue);
        onChange(typed);
    };

    const typeLabel = typeof value === "number" ? "num" : typeof value === "boolean" ? "bool" : "str";

    return (
        <div className="flex items-center gap-1">
            <Input
                className="h-7 flex-1 text-xs font-mono"
                value={strValue}
                onChange={(e) => handleChange(e.target.value)}
                placeholder="value"
            />
            <span className="text-[10px] text-muted-foreground/50 w-8 text-right shrink-0">
                {typeLabel}
            </span>
        </div>
    );
}

function HandlerSearchInput({ value, onChange }: { value: string; onChange: (v: unknown) => void }) {
    const { data: scripts } = useScripts();
    const [search, setSearch] = useState(value);
    const [open, setOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);

    useEffect(() => { setSearch(value); }, [value]);

    useEffect(() => {
        function handleClickOutside(e: MouseEvent) {
            if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
                setOpen(false);
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);

    const suggestions = useMemo(() => {
        const q = search.toLowerCase();
        const items: { name: string; file: string }[] = [];
        if (!scripts) return items;
        for (const script of scripts) {
            for (const handlerName of script.handlerNames) {
                if (!q || handlerName.toLowerCase().includes(q) || script.path.toLowerCase().includes(q)) {
                    items.push({ name: handlerName, file: script.path });
                }
            }
        }
        items.sort((a, b) => a.name.localeCompare(b.name));
        return items.slice(0, 30);
    }, [scripts, search]);

    return (
        <div ref={containerRef} className="relative">
            <Input
                className="h-7 text-xs font-mono"
                value={search}
                onChange={(e) => { setSearch(e.target.value); setOpen(true); }}
                onFocus={() => setOpen(true)}
                placeholder="handler.name"
                onBlur={() => {
                    setTimeout(() => {
                        if (search !== value) onChange(search);
                    }, 150);
                }}
                onKeyDown={(e) => {
                    if (e.key === "Enter") { onChange(search); setOpen(false); }
                    if (e.key === "Escape") { setSearch(value); setOpen(false); }
                }}
            />
            {open && suggestions.length > 0 && (
                <div className="absolute z-50 mt-1 w-full rounded-md border border-border bg-popover shadow-lg max-h-48 overflow-y-auto">
                    {suggestions.map((item) => (
                        <button
                            key={item.name}
                            className="flex w-full items-center gap-2 px-2 py-1 text-left text-xs hover:bg-accent border-b border-border last:border-b-0"
                            onMouseDown={(e) => {
                                e.preventDefault();
                                setSearch(item.name);
                                onChange(item.name);
                                setOpen(false);
                            }}
                        >
                            <span className="font-mono flex-1 truncate">{item.name}</span>
                            <span className="text-[10px] text-muted-foreground truncate max-w-[140px]">{item.file}</span>
                        </button>
                    ))}
                </div>
            )}
        </div>
    );
}

function PropertyKeyPicker({
    usedKeys,
    onSelect,
    onCancel,
}: {
    usedKeys: Set<string>;
    onSelect: (key: string) => void;
    onCancel: () => void;
}) {
    const [search, setSearch] = useState("");
    const [showSuggestions, setShowSuggestions] = useState(true);
    const containerRef = useRef<HTMLDivElement>(null);

    const suggestions = useMemo(() => {
        const q = search.toLowerCase();
        return KNOWN_ENTITY_PROPERTIES.filter((p) => {
            if (usedKeys.has(p.key)) return false;
            if (p.key === "template") return false;
            if (!q) return true;
            return p.key.toLowerCase().includes(q) || p.label.toLowerCase().includes(q);
        });
    }, [search, usedKeys]);

    return (
        <div ref={containerRef} className="space-y-1">
            <div className="flex items-center gap-1">
                <Input
                    className="h-7 flex-1 text-xs font-mono"
                    value={search}
                    onChange={(e) => { setSearch(e.target.value); setShowSuggestions(true); }}
                    placeholder="property name"
                    autoFocus
                    onFocus={() => setShowSuggestions(true)}
                    onKeyDown={(e) => {
                        if (e.key === "Enter" && search.trim()) { onSelect(search.trim()); }
                        if (e.key === "Escape") onCancel();
                    }}
                />
                <Button variant="ghost" size="sm" className="h-7 text-xs px-2" onClick={onCancel}>
                    Cancel
                </Button>
            </div>
            {showSuggestions && suggestions.length > 0 && (
                <div className="border border-border rounded-md max-h-48 overflow-y-auto">
                    {suggestions.map((prop) => (
                        <button
                            key={prop.key}
                            className="flex w-full items-center gap-2 px-2 py-1.5 text-left text-xs hover:bg-accent border-b border-border last:border-b-0"
                            onMouseDown={(e) => {
                                e.preventDefault();
                                onSelect(prop.key);
                            }}
                        >
                            <span className="font-mono font-medium shrink-0">{prop.key}</span>
                            <span className="text-[10px] text-muted-foreground truncate">{prop.description}</span>
                        </button>
                    ))}
                </div>
            )}
            {search.trim() && !suggestions.some((s) => s.key === search.trim()) && (
                <p className="text-[10px] text-muted-foreground">
                    Press Enter to add custom property <code className="font-mono">{search.trim()}</code>
                </p>
            )}
        </div>
    );
}
