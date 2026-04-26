import { useState, useEffect, useMemo, useRef, useCallback } from "react";
import { Input } from "@/components/ui/input";
import { useSprites, useSprite } from "@/api/sprites";
import { apiImageUrl } from "@/api/client";
import { AnimationPlayer, resolveFrames } from "@/lib/animationEngine";
import type { SpriteTilesheetAnimation } from "@/types/sprites";

interface AnimationPickerProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
}

export function AnimationPicker({ value, onChange, placeholder }: AnimationPickerProps) {
    const { data: sprites } = useSprites();
    const [search, setSearch] = useState(value);
    const [open, setOpen] = useState(false);
    const [expandedSheet, setExpandedSheet] = useState<string | null>(null);
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

    const tilesheets = useMemo(() => {
        if (!sprites) return [];
        return sprites.filter((s) => s.type === "tilesheet").map((s) => s.path);
    }, [sprites]);

    const filtered = useMemo(() => {
        const q = search.toLowerCase().split(":")[0];
        if (!q) return tilesheets.slice(0, 30);
        return tilesheets.filter((p) => p.toLowerCase().includes(q)).slice(0, 30);
    }, [tilesheets, search]);

    const commit = useCallback((val: string) => {
        setSearch(val);
        onChange(val);
        setOpen(false);
        setExpandedSheet(null);
    }, [onChange]);

    const parsedRef = parseAnimRef(value);

    return (
        <div ref={containerRef} className="relative space-y-1">
            <Input
                className="h-6 flex-1 text-xs font-mono"
                value={search}
                onChange={(e) => { setSearch(e.target.value); setOpen(true); setExpandedSheet(null); }}
                onFocus={() => setOpen(true)}
                onBlur={() => { setTimeout(() => { if (search !== value) onChange(search); }, 200); }}
                onKeyDown={(e) => {
                    if (e.key === "Enter") { onChange(search); setOpen(false); }
                    if (e.key === "Escape") { setSearch(value); setOpen(false); }
                }}
                placeholder={placeholder ?? "sprites/path:animation"}
            />
            {parsedRef && <AnimationMiniPreview sheetPath={parsedRef.sheet} animName={parsedRef.anim} />}
            {open && filtered.length > 0 && (
                <div className="absolute z-50 mt-1 w-full rounded-md border border-border bg-popover shadow-lg max-h-56 overflow-y-auto">
                    {filtered.map((path) => (
                        <SheetDropdownItem
                            key={path}
                            path={path}
                            expanded={expandedSheet === path}
                            onExpand={() => setExpandedSheet(expandedSheet === path ? null : path)}
                            onSelect={(ref) => commit(ref)}
                        />
                    ))}
                </div>
            )}
        </div>
    );
}

function SheetDropdownItem({
    path,
    expanded,
    onExpand,
    onSelect,
}: {
    path: string;
    expanded: boolean;
    onExpand: () => void;
    onSelect: (ref: string) => void;
}) {
    const { data: detail } = useSprite(expanded ? path : "");
    const animNames = useMemo(() => {
        if (!detail?.metadata?.animations) return [];
        return Object.keys(detail.metadata.animations);
    }, [detail]);

    return (
        <div className="border-b border-border last:border-b-0">
            <button
                className="flex w-full items-center gap-2 px-2 py-1 text-left text-xs hover:bg-accent"
                onMouseDown={(e) => { e.preventDefault(); onExpand(); }}
            >
                <span className="font-mono flex-1 truncate">{path}</span>
                <span className="text-[10px] text-muted-foreground shrink-0">
                    {expanded ? "▾" : "▸"}
                </span>
            </button>
            {expanded && animNames.length > 0 && (
                <div className="bg-muted/30">
                    {animNames.map((anim) => {
                        const ref = anim === "default" ? path : `${path}:${anim}`;
                        return (
                            <button
                                key={anim}
                                className="flex w-full items-center gap-2 pl-5 pr-2 py-0.5 text-left text-[11px] hover:bg-accent"
                                onMouseDown={(e) => { e.preventDefault(); onSelect(ref); }}
                            >
                                <span className="font-mono text-accent-violet">{anim}</span>
                                <span className="font-mono text-muted-foreground text-[10px] truncate">{ref}</span>
                            </button>
                        );
                    })}
                </div>
            )}
        </div>
    );
}

function parseAnimRef(ref: string): { sheet: string; anim: string } | null {
    if (!ref) return null;
    const colonIdx = ref.indexOf(":");
    if (colonIdx === -1) return { sheet: ref, anim: "default" };
    return { sheet: ref.substring(0, colonIdx), anim: ref.substring(colonIdx + 1) };
}

export function AnimationMiniPreview({ sheetPath, animName }: { sheetPath: string; animName: string }) {
    const { data: detail } = useSprite(sheetPath);
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const imageRef = useRef<HTMLImageElement | null>(null);
    const playerRef = useRef<AnimationPlayer | null>(null);
    const rafRef = useRef<number>(0);
    const lastTimeRef = useRef<number>(0);

    const animConfig = detail?.metadata?.animations?.[animName];
    const tileWidth = detail?.metadata?.tilesheet?.tileWidth ?? 16;
    const tileHeight = detail?.metadata?.tilesheet?.tileHeight ?? 16;
    const totalCols = detail?.computed?.columns ?? 1;
    const totalRows = detail?.computed?.rows ?? 1;

    useEffect(() => {
        if (!detail || !animConfig) return;
        const img = new Image();
        img.src = apiImageUrl(`sprites/${sheetPath}.png`);
        img.onload = () => {
            imageRef.current = img;
            const frames = resolveFrames(animConfig, totalCols, totalRows);
            if (frames.length === 0) return;
            playerRef.current = new AnimationPlayer(animConfig, frames);
            lastTimeRef.current = performance.now();
            const animate = (now: number) => {
                const dt = (now - lastTimeRef.current) / 1000;
                lastTimeRef.current = now;
                if (playerRef.current && imageRef.current && canvasRef.current) {
                    playerRef.current.update(dt);
                    drawFrame();
                }
                rafRef.current = requestAnimationFrame(animate);
            };
            rafRef.current = requestAnimationFrame(animate);
        };
        return () => {
            cancelAnimationFrame(rafRef.current);
            imageRef.current = null;
            playerRef.current = null;
        };
    }, [detail, animConfig, sheetPath, totalCols, totalRows]);

    const drawFrame = useCallback(() => {
        const canvas = canvasRef.current;
        const img = imageRef.current;
        const player = playerRef.current;
        if (!canvas || !img || !player) return;
        const ctx = canvas.getContext("2d");
        if (!ctx) return;
        const frame = player.currentTile;
        if (!frame) return;
        const sx = (frame.col - 1) * tileWidth;
        const sy = (frame.row - 1) * tileHeight;
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.imageSmoothingEnabled = false;
        ctx.drawImage(img, sx, sy, tileWidth, tileHeight, 0, 0, canvas.width, canvas.height);
    }, [tileWidth, tileHeight]);

    if (!animConfig) return null;

    const scale = Math.max(1, Math.floor(48 / Math.max(tileWidth, tileHeight)));

    return (
        <canvas
            ref={canvasRef}
            width={tileWidth * scale}
            height={tileHeight * scale}
            className="rounded border border-border bg-canvas"
        />
    );
}
