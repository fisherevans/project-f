import { useState, useRef, useEffect, useMemo } from "react";
import { Input } from "@/components/ui/input";

interface SchemaSelectOption {
    value: string;
    label?: string;
    detail?: string;
    group?: string;
}

interface SchemaSelectProps {
    value: string;
    onChange: (value: string) => void;
    options: SchemaSelectOption[];
    placeholder?: string;
}

export function SchemaSelect({ value, onChange, options, placeholder }: SchemaSelectProps) {
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

    const filtered = useMemo(() => {
        const q = search.toLowerCase();
        return options.filter((o) =>
            !q ||
            o.value.toLowerCase().includes(q) ||
            o.label?.toLowerCase().includes(q) ||
            o.detail?.toLowerCase().includes(q)
        ).slice(0, 30);
    }, [options, search]);

    return (
        <div ref={containerRef} className="relative">
            <Input
                className="h-6 flex-1 text-xs font-mono"
                value={search}
                onChange={(e) => { setSearch(e.target.value); setOpen(true); }}
                onFocus={() => setOpen(true)}
                placeholder={placeholder}
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
            {open && filtered.length > 0 && (
                <div className="absolute z-50 mt-1 w-full rounded-md border border-border bg-popover shadow-lg max-h-48 overflow-y-auto">
                    {filtered.map((item) => (
                        <button
                            key={item.value}
                            className="flex w-full items-center gap-2 px-2 py-1 text-left text-xs hover:bg-accent"
                            onMouseDown={(e) => {
                                e.preventDefault();
                                setSearch(item.value);
                                onChange(item.value);
                                setOpen(false);
                            }}
                        >
                            <span className="font-mono flex-1 truncate">{item.label ?? item.value}</span>
                            {item.group && (
                                <span className="text-[10px] shrink-0 px-1 rounded bg-muted text-muted-foreground">
                                    {item.group}
                                </span>
                            )}
                            {item.detail && (
                                <span className="text-[10px] text-muted-foreground shrink-0 truncate max-w-[180px]">
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
