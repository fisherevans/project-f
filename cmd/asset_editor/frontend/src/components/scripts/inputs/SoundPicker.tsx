import { useState, useRef, useEffect, useMemo, useCallback } from "react";
import { Input } from "@/components/ui/input";
import { Play, Square, Volume2 } from "lucide-react";
import { useAudioFiles } from "@/api/audio";
import { apiAudioUrl } from "@/api/client";

interface SoundPickerProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
}

export function SoundPicker({ value, onChange, placeholder }: SoundPickerProps) {
    const { data: audioFiles } = useAudioFiles();
    const [search, setSearch] = useState(value);
    const [open, setOpen] = useState(false);
    const [playing, setPlaying] = useState(false);
    const audioRef = useRef<HTMLAudioElement | null>(null);
    const containerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        setSearch(value);
    }, [value]);

    useEffect(() => {
        return () => {
            if (audioRef.current) {
                audioRef.current.pause();
                audioRef.current.src = "";
            }
        };
    }, []);

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
        if (!audioFiles) return [];
        const q = search.toLowerCase();
        return audioFiles
            .filter((f) => f.category === "sounds")
            .filter((f) => !q || f.resourceName.toLowerCase().includes(q) || f.path.toLowerCase().includes(q))
            .slice(0, 30);
    }, [audioFiles, search]);

    const playSound = useCallback((resourceName: string) => {
        if (audioRef.current) {
            audioRef.current.pause();
            audioRef.current.src = "";
        }
        const file = audioFiles?.find((f) => f.resourceName === resourceName);
        if (!file) return;
        const audio = new Audio(apiAudioUrl(file.path));
        audio.addEventListener("ended", () => setPlaying(false));
        audio.play();
        audioRef.current = audio;
        setPlaying(true);
    }, [audioFiles]);

    const stopSound = useCallback(() => {
        if (audioRef.current) {
            audioRef.current.pause();
            audioRef.current.currentTime = 0;
            setPlaying(false);
        }
    }, []);

    return (
        <div ref={containerRef} className="relative">
            <div className="flex items-center gap-1">
                <Volume2 className="h-3 w-3 shrink-0 text-muted-foreground" />
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    value={search}
                    onChange={(e) => {
                        setSearch(e.target.value);
                        setOpen(true);
                    }}
                    onFocus={() => setOpen(true)}
                    placeholder={placeholder ?? "Sound resource name"}
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
                {value && (
                    <button
                        className="shrink-0 text-muted-foreground hover:text-foreground"
                        onClick={() => playing ? stopSound() : playSound(value)}
                    >
                        {playing ? <Square className="h-3 w-3" /> : <Play className="h-3 w-3" />}
                    </button>
                )}
            </div>
            {open && suggestions.length > 0 && (
                <div className="absolute z-50 mt-1 w-full rounded-md border border-border bg-popover shadow-lg max-h-48 overflow-y-auto">
                    {suggestions.map((file) => (
                        <button
                            key={file.path}
                            className="flex w-full items-center gap-2 px-2 py-1 text-left text-xs hover:bg-accent"
                            onMouseDown={(e) => {
                                e.preventDefault();
                                setSearch(file.resourceName);
                                onChange(file.resourceName);
                                setOpen(false);
                            }}
                        >
                            <span className="font-mono flex-1 truncate">{file.resourceName}</span>
                            <span className="text-[10px] text-muted-foreground shrink-0">{file.format}</span>
                            <button
                                className="shrink-0 text-muted-foreground hover:text-foreground"
                                onMouseDown={(e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    playSound(file.resourceName);
                                }}
                            >
                                <Play className="h-2.5 w-2.5" />
                            </button>
                        </button>
                    ))}
                </div>
            )}
        </div>
    );
}
