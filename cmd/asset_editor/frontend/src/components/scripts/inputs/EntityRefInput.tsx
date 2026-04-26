import { useState, useRef, useEffect, useMemo } from "react";
import { Input } from "@/components/ui/input";
import { User } from "lucide-react";
import { useEntities } from "@/api/debug";

interface EntityRefInputProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
}

const TEMPLATE_VARS = [
    { pattern: "{{self}}", description: "This entity (the handler owner)" },
    { pattern: "{{source}}", description: "Entity that triggered the event" },
    { pattern: "{{player}}", description: "The player entity" },
];

export function EntityRefInput({ value, onChange, placeholder }: EntityRefInputProps) {
    const { data: entityData, isError: debugUnavailable } = useEntities();
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
        const items: { label: string; detail?: string; isTemplate: boolean }[] = [];

        for (const tv of TEMPLATE_VARS) {
            if (!q || tv.pattern.toLowerCase().includes(q) || tv.description.toLowerCase().includes(q)) {
                items.push({ label: tv.pattern, detail: tv.description, isTemplate: true });
            }
        }

        if (entityData?.entities) {
            for (const e of entityData.entities) {
                if (!q || e.id.toLowerCase().includes(q) || e.debug_type?.toLowerCase().includes(q) || e.handler_ref?.toLowerCase().includes(q)) {
                    const detail = [e.debug_type, e.handler_ref].filter(Boolean).join(" - ");
                    items.push({ label: e.id, detail: detail || undefined, isTemplate: false });
                }
            }
        }

        return items.slice(0, 30);
    }, [entityData, search]);

    return (
        <div ref={containerRef} className="relative">
            <div className="flex items-center gap-1">
                <User className="h-3 w-3 shrink-0 text-muted-foreground" />
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    value={search}
                    onChange={(e) => {
                        setSearch(e.target.value);
                        setOpen(true);
                    }}
                    onFocus={() => setOpen(true)}
                    placeholder={placeholder ?? "Entity reference"}
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
                            key={item.label}
                            className="flex w-full items-center gap-2 px-2 py-1 text-left text-xs hover:bg-accent"
                            onMouseDown={(e) => {
                                e.preventDefault();
                                setSearch(item.label);
                                onChange(item.label);
                                setOpen(false);
                            }}
                        >
                            <span className={`font-mono flex-1 truncate ${item.isTemplate ? "text-accent-violet" : ""}`}>
                                {item.label}
                            </span>
                            {item.detail && (
                                <span className="text-[10px] text-muted-foreground shrink-0 truncate max-w-[140px]">
                                    {item.detail}
                                </span>
                            )}
                        </button>
                    ))}
                    {debugUnavailable && (
                        <div className="px-2 py-1 text-[10px] text-muted-foreground/50 border-t border-border/30">
                            Game not running - showing templates only
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
