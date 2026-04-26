import { useState, useMemo, useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { X } from "lucide-react";
import type { ScriptSchema, StepKindDef } from "@/types/scripts";

interface StepKindPickerProps {
    schema: ScriptSchema;
    onSelect: (kind: string) => void;
    onCancel: () => void;
}

const CATEGORY_ORDER = ["text", "flow", "state", "entity", "camera", "audio", "transition", "rpg", "combat"];

const CATEGORY_COLORS: Record<string, string> = {
    text: "bg-accent-blue-tint text-accent-blue",
    flow: "bg-accent-violet-tint text-accent-violet",
    state: "bg-accent-amber-tint text-accent-amber",
    entity: "bg-accent-teal-tint text-accent-teal",
    camera: "bg-accent-blue-tint text-accent-blue",
    audio: "bg-accent-violet-tint text-accent-violet",
    transition: "bg-accent-orange-tint text-accent-orange",
    rpg: "bg-accent-red-tint text-accent-red",
    combat: "bg-accent-red-tint text-accent-red",
};

export function getCategoryColor(category: string): string {
    return CATEGORY_COLORS[category] ?? "bg-muted text-muted-foreground";
}

function paramSummary(def: StepKindDef): string {
    if (!def.params || def.params.length === 0) return "";
    return def.params
        .filter((p) => p.type !== "steps")
        .map((p) => p.required ? p.name : `${p.name}?`)
        .join(", ");
}

export function StepKindPicker({ schema, onSelect, onCancel }: StepKindPickerProps) {
    const [search, setSearch] = useState("");
    const [activeCategory, setActiveCategory] = useState<string | null>(null);
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        inputRef.current?.focus();
    }, []);

    const grouped = useMemo(() => {
        const map = new Map<string, StepKindDef[]>();
        for (const def of Object.values(schema.stepKinds)) {
            const cat = def.category;
            if (!map.has(cat)) map.set(cat, []);
            map.get(cat)!.push(def);
        }
        for (const defs of map.values()) {
            defs.sort((a, b) => a.name.localeCompare(b.name));
        }
        return map;
    }, [schema]);

    const filtered = useMemo(() => {
        let source = grouped;
        if (search) {
            const q = search.toLowerCase();
            const result = new Map<string, StepKindDef[]>();
            for (const [cat, defs] of source) {
                const matching = defs.filter(
                    (d) => d.name.toLowerCase().includes(q) || d.description.toLowerCase().includes(q)
                );
                if (matching.length > 0) result.set(cat, matching);
            }
            source = result;
        }
        if (activeCategory) {
            const result = new Map<string, StepKindDef[]>();
            if (source.has(activeCategory)) {
                result.set(activeCategory, source.get(activeCategory)!);
            }
            return result;
        }
        return source;
    }, [grouped, search, activeCategory]);

    const categories = [
        ...CATEGORY_ORDER.filter((c) => grouped.has(c)),
        ...[...grouped.keys()].filter((c) => !CATEGORY_ORDER.includes(c)).sort(),
    ];
    const visibleCategories = [
        ...CATEGORY_ORDER.filter((c) => filtered.has(c)),
        ...[...filtered.keys()].filter((c) => !CATEGORY_ORDER.includes(c)).sort(),
    ];

    return (
        <div className="rounded-lg border border-border bg-popover shadow-xl">
            <div className="flex items-center gap-2 border-b border-border p-2">
                <Input
                    ref={inputRef}
                    className="h-7 text-xs flex-1"
                    placeholder="Search steps..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    onKeyDown={(e) => {
                        if (e.key === "Escape") onCancel();
                    }}
                />
                <button className="text-muted-foreground hover:text-foreground shrink-0" onClick={onCancel}>
                    <X className="h-4 w-4" />
                </button>
            </div>
            <div className="flex flex-wrap gap-1 px-2 py-1.5 border-b border-border/50">
                {categories.map((cat) => (
                    <button
                        key={cat}
                        className={`text-[10px] px-1.5 py-0.5 rounded transition-colors ${
                            activeCategory === cat
                                ? CATEGORY_COLORS[cat] ?? "bg-muted text-foreground"
                                : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
                        }`}
                        onClick={() => setActiveCategory(activeCategory === cat ? null : cat)}
                    >
                        {cat}
                    </button>
                ))}
            </div>
            <div className="max-h-[60vh] overflow-y-auto p-1">
                {visibleCategories.map((cat) => (
                    <div key={cat} className="mb-1">
                        <div className="px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60 sticky top-0 bg-popover">
                            {cat}
                        </div>
                        {filtered.get(cat)!.map((def) => {
                            const params = paramSummary(def);
                            return (
                                <button
                                    key={def.name}
                                    className="flex w-full items-start gap-2 rounded px-2 py-1.5 text-left hover:bg-accent group/item"
                                    onClick={() => onSelect(def.name)}
                                >
                                    <div className="flex-1 min-w-0">
                                        <div className="flex items-center gap-2">
                                            <code className="text-xs font-mono font-semibold">{def.name}</code>
                                            {params && (
                                                <span className="text-[10px] text-muted-foreground/60 font-mono truncate">({params})</span>
                                            )}
                                        </div>
                                        <div className="text-[11px] text-muted-foreground leading-tight mt-0.5">{def.description}</div>
                                    </div>
                                </button>
                            );
                        })}
                    </div>
                ))}
                {visibleCategories.length === 0 && (
                    <div className="px-2 py-4 text-center text-xs text-muted-foreground">No matches</div>
                )}
            </div>
        </div>
    );
}
