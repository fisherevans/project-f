import { useState, useRef, useEffect, useMemo } from "react";
import { Input } from "@/components/ui/input";
import { useGlobals } from "@/api/debug";

interface GlobalKeyInputProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
}

export function GlobalKeyInput({ value, onChange, placeholder }: GlobalKeyInputProps) {
    const { data: liveGlobals, isError: debugUnavailable } = useGlobals();
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
        const keys = new Map<string, { value?: string; exists: boolean }>();
        if (liveGlobals) {
            for (const g of liveGlobals) {
                keys.set(g.key, { value: g.exists ? String(g.value) : undefined, exists: g.exists });
            }
        }
        const q = search.toLowerCase();
        const results = [...keys.entries()]
            .filter(([key]) => !q || key.toLowerCase().includes(q))
            .sort(([a], [b]) => a.localeCompare(b))
            .slice(0, 30);
        return results;
    }, [liveGlobals, search]);

    const currentLiveValue = useMemo(() => {
        if (!liveGlobals || !value) return null;
        const entry = liveGlobals.find((g) => g.key === value);
        if (!entry || !entry.exists) return null;
        return String(entry.value);
    }, [liveGlobals, value]);

    return (
        <div ref={containerRef} className="relative">
            <div className="flex items-center gap-1">
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    value={search}
                    onChange={(e) => {
                        setSearch(e.target.value);
                        setOpen(true);
                    }}
                    onFocus={() => setOpen(true)}
                    placeholder={placeholder ?? "Global key"}
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
                {currentLiveValue !== null && (
                    <span className="shrink-0 text-[10px] text-accent-green font-mono bg-accent-green-tint px-1 rounded" title="Live value from running game">
                        = {currentLiveValue.length > 20 ? currentLiveValue.slice(0, 20) + "..." : currentLiveValue}
                    </span>
                )}
                {value && !debugUnavailable && currentLiveValue === null && liveGlobals && (
                    <span className="shrink-0 text-[10px] text-muted-foreground/50 font-mono" title="Key not set in running game">
                        unset
                    </span>
                )}
            </div>
            {open && suggestions.length > 0 && (
                <div className="absolute z-50 mt-1 w-full rounded-md border border-border bg-popover shadow-lg max-h-48 overflow-y-auto">
                    {suggestions.map(([key, info]) => (
                        <button
                            key={key}
                            className="flex w-full items-center gap-2 px-2 py-1 text-left text-xs hover:bg-accent"
                            onMouseDown={(e) => {
                                e.preventDefault();
                                setSearch(key);
                                onChange(key);
                                setOpen(false);
                            }}
                        >
                            <span className="font-mono flex-1 truncate">{key}</span>
                            {info.exists && info.value !== undefined && (
                                <span className="text-[10px] text-accent-green/70 font-mono shrink-0 truncate max-w-[100px]">
                                    = {info.value}
                                </span>
                            )}
                        </button>
                    ))}
                </div>
            )}
        </div>
    );
}
