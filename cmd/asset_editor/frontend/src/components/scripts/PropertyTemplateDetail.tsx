import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Plus, Trash2 } from "lucide-react";

interface PropertyTemplateDetailProps {
    name: string;
    template: Record<string, unknown>;
    onChange: (template: Record<string, unknown>) => void;
}

export function PropertyTemplateDetail({ name, template, onChange }: PropertyTemplateDetailProps) {
    const entries = Object.entries(template);
    const [newKey, setNewKey] = useState("");

    const addEntry = () => {
        const key = newKey.trim() || `prop_${entries.length}`;
        onChange({ ...template, [key]: "" });
        setNewKey("");
    };

    const removeEntry = (key: string) => {
        const next = { ...template };
        delete next[key];
        onChange(next);
    };

    const updateKey = (oldKey: string, newKeyVal: string) => {
        if (newKeyVal === oldKey) return;
        const next: Record<string, unknown> = {};
        for (const [k, v] of Object.entries(template)) {
            next[k === oldKey ? newKeyVal : k] = v;
        }
        onChange(next);
    };

    const updateValue = (key: string, rawValue: string) => {
        let value: unknown = rawValue;
        if (rawValue === "true") value = true;
        else if (rawValue === "false") value = false;
        else if (rawValue !== "" && !isNaN(Number(rawValue))) value = Number(rawValue);
        onChange({ ...template, [key]: value });
    };

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
                    <div className="space-y-1">
                        {entries.map(([key, val]) => (
                            <div key={key} className="flex items-center gap-1 group/entry">
                                <Input
                                    className="h-7 w-40 text-xs font-mono"
                                    value={key}
                                    onChange={(e) => updateKey(key, e.target.value)}
                                />
                                <span className="text-muted-foreground text-[10px]">=</span>
                                <Input
                                    className="h-7 flex-1 text-xs font-mono"
                                    value={val !== undefined && val !== null ? String(val) : ""}
                                    onChange={(e) => updateValue(key, e.target.value)}
                                    placeholder="value"
                                />
                                <span className="text-[10px] text-muted-foreground/50 w-10 text-right shrink-0">
                                    {typeof val === "number" ? "num" : typeof val === "boolean" ? "bool" : "str"}
                                </span>
                                <button
                                    className="text-muted-foreground hover:text-destructive opacity-0 group-hover/entry:opacity-100 shrink-0"
                                    onClick={() => removeEntry(key)}
                                >
                                    <Trash2 className="h-3 w-3" />
                                </button>
                            </div>
                        ))}
                    </div>
                    <div className="flex items-center gap-1">
                        <Input
                            className="h-7 w-40 text-xs font-mono"
                            value={newKey}
                            onChange={(e) => setNewKey(e.target.value)}
                            placeholder="property name"
                            onKeyDown={(e) => {
                                if (e.key === "Enter") addEntry();
                            }}
                        />
                        <Button variant="outline" size="sm" className="h-7 text-xs" onClick={addEntry}>
                            <Plus className="mr-1 h-3 w-3" />
                            Add property
                        </Button>
                    </div>
                </div>
            </ScrollArea>
        </div>
    );
}
