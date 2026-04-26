import { useState, useEffect, useCallback } from "react";
import { Input } from "@/components/ui/input";

interface ColorInputProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
}

export function ColorInput({ value, onChange, placeholder }: ColorInputProps) {
    const [text, setText] = useState(value);

    useEffect(() => {
        setText(value);
    }, [value]);

    const commit = useCallback((val: string) => {
        if (val !== value) onChange(val);
    }, [value, onChange]);

    const nativeColor = toNativeColor(value);

    return (
        <div className="flex items-center gap-1">
            <input
                type="color"
                className="h-6 w-6 shrink-0 cursor-pointer rounded border border-border p-0"
                value={nativeColor}
                onChange={(e) => {
                    const hex = e.target.value;
                    setText(hex);
                    onChange(hex);
                }}
            />
            <Input
                className="h-6 flex-1 text-xs font-mono"
                value={text}
                onChange={(e) => setText(e.target.value)}
                onBlur={() => commit(text)}
                onKeyDown={(e) => { if (e.key === "Enter") commit(text); }}
                placeholder={placeholder ?? "#rrggbb"}
            />
        </div>
    );
}

function toNativeColor(val: string): string {
    if (!val) return "#ffffff";
    const v = val.trim();
    if (v.match(/^#[0-9a-fA-F]{6}$/)) return v;
    if (v.match(/^#[0-9a-fA-F]{3}$/)) {
        return `#${v[1]}${v[1]}${v[2]}${v[2]}${v[3]}${v[3]}`;
    }
    return "#ffffff";
}
