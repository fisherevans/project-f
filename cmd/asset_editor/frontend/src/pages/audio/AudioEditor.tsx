import { useState, useRef, useEffect, useCallback } from "react"
import { useLocation, useNavigate } from "react-router-dom"
import { usePageTitle } from "@/hooks/usePageTitle"
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges"
import { Button } from "@/components/ui/button"
import { Slider } from "@/components/ui/slider"
import { useAudioFile, useSaveAudio } from "@/api/audio"
import { apiAudioUrl } from "@/api/client"
import { ArrowLeft, Play, Square, Pause, RotateCcw, Save, Volume2 } from "lucide-react"

function formatDuration(seconds: number): string {
    const m = Math.floor(seconds / 60)
    const s = Math.floor(seconds % 60)
    return `${m}:${s.toString().padStart(2, "0")}`
}

function formatFileSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function AudioEditor() {
    const location = useLocation()
    const navigate = useNavigate()
    const audioPath = location.pathname.replace(/^\/audio\//, "")
    const { data: detail, isLoading } = useAudioFile(audioPath)
    const saveMutation = useSaveAudio()

    const audioRef = useRef<HTMLAudioElement | null>(null)
    const [playing, setPlaying] = useState(false)
    const [paused, setPaused] = useState(false)
    const [currentTime, setCurrentTime] = useState(0)
    const [duration, setDuration] = useState(0)
    const [gain, setGain] = useState<number>(1.0)
    const [hasChanges, setHasChanges] = useState(false)

    const audioName = audioPath.split("/").pop() ?? audioPath
    usePageTitle(`${audioName} - Audio`)
    useUnsavedChanges(hasChanges)

    useEffect(() => {
        if (detail) {
            setGain(detail.gain ?? 1.0)
            setHasChanges(false)
        }
    }, [detail])

    useEffect(() => {
        const audio = new Audio(apiAudioUrl(audioPath))
        audio.addEventListener("loadedmetadata", () => setDuration(audio.duration))
        audio.addEventListener("timeupdate", () => setCurrentTime(audio.currentTime))
        audio.addEventListener("ended", () => {
            setPlaying(false)
            setPaused(false)
        })
        audioRef.current = audio
        return () => {
            audio.pause()
            audio.src = ""
        }
    }, [audioPath])

    useEffect(() => {
        if (audioRef.current) {
            audioRef.current.volume = Math.max(0, Math.min(2, gain))
        }
    }, [gain])

    const play = useCallback(() => {
        if (!audioRef.current) return
        audioRef.current.play()
        setPlaying(true)
        setPaused(false)
    }, [])

    const pause = useCallback(() => {
        if (!audioRef.current) return
        audioRef.current.pause()
        setPaused(true)
    }, [])

    const stop = useCallback(() => {
        if (!audioRef.current) return
        audioRef.current.pause()
        audioRef.current.currentTime = 0
        setPlaying(false)
        setPaused(false)
    }, [])

    const seek = useCallback((value: number | readonly number[]) => {
        if (!audioRef.current) return
        const t = Array.isArray(value) ? value[0] : value
        audioRef.current.currentTime = t
        setCurrentTime(t)
    }, [])

    const handleGainChange = useCallback((value: number | readonly number[]) => {
        const v = Array.isArray(value) ? value[0] : value
        setGain(v)
        setHasChanges(true)
    }, [])

    const handleSave = useCallback(() => {
        const metadata = gain === 1.0 ? {} : { gain }
        saveMutation.mutate({ path: audioPath, metadata }, {
            onSuccess: () => setHasChanges(false),
        })
    }, [audioPath, gain, saveMutation])

    const handleReset = useCallback(() => {
        setGain(1.0)
        setHasChanges(true)
    }, [])

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-full text-muted-foreground">
                Loading...
            </div>
        )
    }

    if (!detail) {
        return (
            <div className="flex items-center justify-center h-full text-muted-foreground">
                Audio file not found.
            </div>
        )
    }

    const originalGain = detail.gain ?? 1.0

    return (
        <div className="flex flex-col h-full">
            <div className="p-3 border-b border-border flex items-center gap-3">
                <Button variant="ghost" size="sm" onClick={() => navigate(`/audio?dir=${detail.directory}`)}>
                    <ArrowLeft className="h-4 w-4 mr-1" />
                    Back
                </Button>
                <div className="flex-1">
                    <h1 className="text-lg font-semibold">{detail.name}</h1>
                    <p className="text-sm text-muted-foreground">{detail.path}.{detail.format}</p>
                </div>
                <Button
                    size="sm"
                    onClick={handleSave}
                    disabled={!hasChanges || saveMutation.isPending}
                >
                    <Save className="h-4 w-4 mr-1" />
                    {saveMutation.isPending ? "Saving..." : "Save"}
                </Button>
            </div>

            <div className="flex-1 overflow-y-auto p-6">
                <div className="max-w-2xl mx-auto space-y-8">
                    {/* File info */}
                    <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                            <span className="text-muted-foreground">Resource name</span>
                            <p className="font-mono mt-0.5">{detail.resourceName}</p>
                        </div>
                        <div>
                            <span className="text-muted-foreground">Asset path</span>
                            <p className="font-mono mt-0.5">audio/{detail.path}.{detail.format}</p>
                        </div>
                        <div>
                            <span className="text-muted-foreground">Format</span>
                            <p className="mt-0.5">{detail.format.toUpperCase()}</p>
                        </div>
                        <div>
                            <span className="text-muted-foreground">File size</span>
                            <p className="mt-0.5">{formatFileSize(detail.fileSize)}</p>
                        </div>
                        <div>
                            <span className="text-muted-foreground">Category</span>
                            <p className="mt-0.5">{detail.category}</p>
                        </div>
                        <div>
                            <span className="text-muted-foreground">Duration</span>
                            <p className="mt-0.5">{duration > 0 ? formatDuration(duration) : "-"}</p>
                        </div>
                    </div>

                    {/* Playback */}
                    <div className="space-y-3">
                        <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                            Playback
                        </h2>
                        <div className="flex items-center gap-3">
                            {!playing || paused ? (
                                <Button variant="outline" size="sm" onClick={play}>
                                    <Play className="h-4 w-4" />
                                </Button>
                            ) : (
                                <Button variant="outline" size="sm" onClick={pause}>
                                    <Pause className="h-4 w-4" />
                                </Button>
                            )}
                            <Button variant="outline" size="sm" onClick={stop} disabled={!playing}>
                                <Square className="h-4 w-4" />
                            </Button>
                            <span className="text-sm text-muted-foreground tabular-nums min-w-[70px]">
                                {formatDuration(currentTime)} / {formatDuration(duration)}
                            </span>
                        </div>
                        {duration > 0 && (
                            <div className="w-full">
                                <Slider
                                    min={0}
                                    max={duration}
                                    step={0.01}
                                    value={[currentTime]}
                                    onValueChange={seek}
                                />
                            </div>
                        )}
                    </div>

                    {/* Gain control */}
                    <div className="space-y-3">
                        <div className="flex items-center justify-between">
                            <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                                Gain
                            </h2>
                            <Button variant="ghost" size="sm" onClick={handleReset}>
                                <RotateCcw className="h-3.5 w-3.5 mr-1" />
                                Reset to 100%
                            </Button>
                        </div>
                        <div className="flex items-center gap-4">
                            <Volume2 className="h-4 w-4 text-muted-foreground shrink-0" />
                            <div className="flex-1">
                                <Slider
                                    min={0}
                                    max={2}
                                    step={0.01}
                                    value={[gain]}
                                    onValueChange={handleGainChange}
                                />
                            </div>
                            <span className="text-sm tabular-nums min-w-[50px] text-right">
                                {Math.round(gain * 100)}%
                            </span>
                        </div>
                        {gain !== originalGain && (
                            <p className="text-xs text-muted-foreground">
                                Changed from {Math.round(originalGain * 100)}%
                            </p>
                        )}
                        <p className="text-xs text-muted-foreground">
                            Sets the default volume multiplier for this audio file. 100% is unity gain.
                            The game applies this value whenever the sound plays.
                        </p>
                    </div>

                    {/* Raw YAML */}
                    {detail.rawYaml && (
                        <div className="space-y-2">
                            <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                                YAML Sidecar
                            </h2>
                            <pre className="text-xs font-mono bg-muted p-3 rounded-md overflow-x-auto whitespace-pre">
                                {detail.rawYaml}
                            </pre>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
