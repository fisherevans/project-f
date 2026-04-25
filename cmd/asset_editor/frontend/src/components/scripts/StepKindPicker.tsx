import { useState, useMemo } from "react";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { ScriptSchema, StepKindDef } from "@/types/scripts";

interface StepKindPickerProps {
    schema: ScriptSchema;
    onSelect: (kind: string) => void;
    onCancel: () => void;
}

const CATEGORY_ORDER = ["text", "flow", "state", "entity", "camera", "audio", "transition", "rpg", "combat"];

const CATEGORY_COLORS: Record<string, string> = {
    text: "bg-blue-500/20 text-blue-300",
    flow: "bg-purple-500/20 text-purple-300",
    state: "bg-amber-500/20 text-amber-300",
    entity: "bg-emerald-500/20 text-emerald-300",
    camera: "bg-cyan-500/20 text-cyan-300",
    audio: "bg-pink-500/20 text-pink-300",
    transition: "bg-orange-500/20 text-orange-300",
    rpg: "bg-red-500/20 text-red-300",
    combat: "bg-red-500/20 text-red-300",
};

export function getCategoryColor(category: string): string {
    return CATEGORY_COLORS[category] ?? "bg-muted text-muted-foreground";
}

export function StepKindPicker({ schema, onSelect, onCancel }: StepKindPickerProps) {
    const [search, setSearch] = useState("");

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
        if (!search) return grouped;
        const q = search.toLowerCase();
        const result = new Map<string, StepKindDef[]>();
        for (const [cat, defs] of grouped) {
            const matching = defs.filter(
                (d) => d.name.toLowerCase().includes(q) || d.description.toLowerCase().includes(q)
            );
            if (matching.length > 0) result.set(cat, matching);
        }
        return result;
    }, [grouped, search]);

    const categories = CATEGORY_ORDER.filter((c) => filtered.has(c));

    return (
        <div className="rounded-md border border-border bg-popover shadow-lg">
            <div className="border-b border-border p-2">
                <Input
                    className="h-7 text-xs"
                    placeholder="Search step kinds..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    autoFocus
                    onKeyDown={(e) => {
                        if (e.key === "Escape") onCancel();
                    }}
                />
            </div>
            <ScrollArea className="max-h-64">
                <div className="p-1">
                    {categories.map((cat) => (
                        <div key={cat} className="mb-1">
                            <div className="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60">
                                {cat}
                            </div>
                            {filtered.get(cat)!.map((def) => (
                                <button
                                    key={def.name}
                                    className="flex w-full items-start gap-2 rounded px-2 py-1 text-left text-xs hover:bg-accent"
                                    onClick={() => onSelect(def.name)}
                                >
                                    <code className="shrink-0 font-mono font-semibold">{def.name}</code>
                                    <span className="text-muted-foreground line-clamp-1">{def.description}</span>
                                </button>
                            ))}
                        </div>
                    ))}
                    {categories.length === 0 && (
                        <div className="px-2 py-3 text-center text-xs text-muted-foreground">No matches</div>
                    )}
                </div>
            </ScrollArea>
        </div>
    );
}
