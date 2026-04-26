import { useState, useRef, useEffect, useMemo } from "react";
import { Input } from "@/components/ui/input";
import { FileCode } from "lucide-react";
import { useScripts } from "@/api/scripts";
import type { ScriptSchema } from "@/types/scripts";

interface HandlerRefInputProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
    schema?: ScriptSchema;
}

export function HandlerRefInput({ value, onChange, placeholder, schema }: HandlerRefInputProps) {
    const { data: scripts } = useScripts();
    const [search, setSearch] = useState(value);
    const [open, setOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        setSearch(value);
    }, [value]);

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
        const items: { label: string; detail?: string; source: string }[] = [];

        if (schema?.actions) {
            for (const [name, def] of Object.entries(schema.actions)) {
                if (!q || name.toLowerCase().includes(q) || def.description.toLowerCase().includes(q)) {
                    items.push({ label: name, detail: def.description, source: "action" });
                }
            }
        }

        if (scripts) {
            const seen = new Set(items.map((i) => i.label));
            for (const script of scripts) {
                if (script.sequenceNames) {
                    for (const seqName of script.sequenceNames) {
                        if (seen.has(seqName)) continue;
                        if (!q || seqName.toLowerCase().includes(q) || script.path.toLowerCase().includes(q)) {
                            items.push({ label: seqName, detail: script.path, source: "sequence" });
                            seen.add(seqName);
                        }
                    }
                }
            }
        }

        items.sort((a, b) => a.label.localeCompare(b.label));
        return items.slice(0, 30);
    }, [schema, scripts, search]);

    return (
        <div ref={containerRef} className="relative">
            <div className="flex items-center gap-1">
                <FileCode className="h-3 w-3 shrink-0 text-muted-foreground" />
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    value={search}
                    onChange={(e) => {
                        setSearch(e.target.value);
                        setOpen(true);
                    }}
                    onFocus={() => setOpen(true)}
                    placeholder={placeholder ?? "Action or sequence name"}
                    onBlur={() => {
                        setTimeout(() => {
                            if (search !== value) onChange(search);
                        }, 150);
                    }}
                    onKeyDown={(e) => {
                        if (e.key === "Enter") {
                            onChange(search);
                            setOpen(false);
                        }
                        if (e.key === "Escape") {
                            setSearch(value);
                            setOpen(false);
                        }
                    }}
                />
            </div>
            {open && suggestions.length > 0 && (
                <div className="absolute z-50 mt-1 w-full rounded-md border border-border bg-popover shadow-lg max-h-48 overflow-y-auto">
                    {suggestions.map((item) => (
                        <button
                            key={`${item.source}:${item.label}`}
                            className="flex w-full items-center gap-2 px-2 py-1 text-left text-xs hover:bg-accent"
                            onMouseDown={(e) => {
                                e.preventDefault();
                                setSearch(item.label);
                                onChange(item.label);
                                setOpen(false);
                            }}
                        >
                            <span className="font-mono flex-1 truncate">{item.label}</span>
                            <span className={`text-[10px] shrink-0 px-1 rounded ${
                                item.source === "action"
                                    ? "bg-accent-violet-tint text-accent-violet"
                                    : "bg-accent-teal-tint text-accent-teal"
                            }`}>
                                {item.source}
                            </span>
                            {item.detail && (
                                <span className="text-[10px] text-muted-foreground shrink-0 truncate max-w-[140px]">
                                    {item.detail}
                                </span>
                            )}
                        </button>
                    ))}
                </div>
            )}
        </div>
    );
}
