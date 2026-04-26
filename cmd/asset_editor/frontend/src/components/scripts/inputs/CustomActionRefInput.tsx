import { useState, useRef, useEffect, useMemo } from "react";
import { Input } from "@/components/ui/input";
import { Zap } from "lucide-react";
import { useScripts } from "@/api/scripts";
import { useExprContext } from "../ExprContext";

interface CustomActionRefInputProps {
    value: string;
    onChange: (value: string) => void;
}

export function CustomActionRefInput({ value, onChange }: CustomActionRefInputProps) {
    const { customActions } = useExprContext();
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
        const seen = new Set<string>();

        for (const [name, def] of Object.entries(customActions)) {
            if (!q || name.toLowerCase().includes(q) || (def.description ?? "").toLowerCase().includes(q)) {
                items.push({ label: name, detail: def.description, source: "this file" });
                seen.add(name);
            }
        }

        if (scripts) {
            for (const script of scripts) {
                if (script.customActionNames) {
                    for (const name of script.customActionNames) {
                        if (seen.has(name)) continue;
                        if (!q || name.toLowerCase().includes(q) || script.path.toLowerCase().includes(q)) {
                            items.push({ label: name, detail: script.path, source: "other" });
                            seen.add(name);
                        }
                    }
                }
            }
        }

        items.sort((a, b) => {
            if (a.source !== b.source) return a.source === "this file" ? -1 : 1;
            return a.label.localeCompare(b.label);
        });
        return items.slice(0, 30);
    }, [customActions, scripts, search]);

    return (
        <div ref={containerRef} className="relative">
            <div className="flex items-center gap-1">
                <Zap className="h-3 w-3 shrink-0 text-muted-foreground" />
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    value={search}
                    onChange={(e) => {
                        setSearch(e.target.value);
                        setOpen(true);
                    }}
                    onFocus={() => setOpen(true)}
                    placeholder="Custom action name"
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
                            {item.source === "this file" ? (
                                <span className="text-[10px] shrink-0 px-1 rounded bg-accent-violet-tint text-accent-violet">local</span>
                            ) : (
                                <span className="text-[10px] shrink-0 px-1 rounded bg-muted text-muted-foreground">other</span>
                            )}
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
