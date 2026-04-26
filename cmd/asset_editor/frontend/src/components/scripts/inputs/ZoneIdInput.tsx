import { useState, useRef, useEffect, useMemo } from "react";
import { Input } from "@/components/ui/input";
import { Layers } from "lucide-react";
import { useZones } from "@/api/debug";

interface ZoneIdInputProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
}

export function ZoneIdInput({ value, onChange, placeholder }: ZoneIdInputProps) {
    const { data: zones, isError: debugUnavailable } = useZones();
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
        if (!zones) return [];
        const seen = new Map<string, { count: number; x: number; y: number; w: number; h: number }>();
        for (const z of zones) {
            const existing = seen.get(z.id);
            if (existing) {
                existing.count++;
            } else {
                seen.set(z.id, { count: 1, x: z.x, y: z.y, w: z.w, h: z.h });
            }
        }
        const q = search.toLowerCase();
        return [...seen.entries()]
            .filter(([id]) => !q || id.toLowerCase().includes(q))
            .sort(([a], [b]) => a.localeCompare(b))
            .slice(0, 30);
    }, [zones, search]);

    return (
        <div ref={containerRef} className="relative">
            <div className="flex items-center gap-1">
                <Layers className="h-3 w-3 shrink-0 text-muted-foreground" />
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    value={search}
                    onChange={(e) => {
                        setSearch(e.target.value);
                        setOpen(true);
                    }}
                    onFocus={() => setOpen(true)}
                    placeholder={placeholder ?? "Zone ID"}
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
                    {suggestions.map(([id, info]) => (
                        <button
                            key={id}
                            className="flex w-full items-center gap-2 px-2 py-1 text-left text-xs hover:bg-accent"
                            onMouseDown={(e) => {
                                e.preventDefault();
                                setSearch(id);
                                onChange(id);
                                setOpen(false);
                            }}
                        >
                            <span className="font-mono flex-1 truncate">{id}</span>
                            <span className="text-[10px] text-muted-foreground shrink-0">
                                ({info.x},{info.y}) {info.w}x{info.h}
                                {info.count > 1 && ` +${info.count - 1}`}
                            </span>
                        </button>
                    ))}
                </div>
            )}
            {!open && debugUnavailable && value && (
                <div className="text-[10px] text-muted-foreground/50 mt-0.5">Game not running - no zone suggestions</div>
            )}
        </div>
    );
}
