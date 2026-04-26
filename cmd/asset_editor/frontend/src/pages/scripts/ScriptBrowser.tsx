import { useMemo } from "react"
import { useNavigate } from "react-router-dom"
import { useScripts } from "@/api/scripts"
import { FileText, ChevronRight } from "lucide-react"
import { usePageTitle } from "@/hooks/usePageTitle"
import type { ScriptFileEntry } from "@/types/scripts"

export function ScriptBrowser() {
    const navigate = useNavigate()
    const { data: scripts, isLoading, error } = useScripts()
    usePageTitle("Scripts")

    const grouped = useMemo(() => {
        if (!scripts) return new Map<string, ScriptFileEntry[]>()
        const map = new Map<string, ScriptFileEntry[]>()
        for (const s of scripts) {
            const dir = s.directory || "(root)"
            if (!map.has(dir)) map.set(dir, [])
            map.get(dir)!.push(s)
        }
        return map
    }, [scripts])

    if (isLoading) return <div className="p-4 text-muted-foreground">Loading scripts...</div>
    if (error) return <div className="p-4 text-destructive">Error: {error.message}</div>

    return (
        <div className="h-full overflow-auto p-4">
            <h1 className="mb-4 text-lg font-semibold">Scripts</h1>
            {scripts && scripts.length === 0 && (
                <p className="text-muted-foreground">No script files found in assets/scripts/</p>
            )}
            <div className="space-y-6">
                {Array.from(grouped.entries()).map(([dir, files]) => (
                    <div key={dir}>
                        <h2 className="mb-2 text-sm font-medium text-muted-foreground">{dir}</h2>
                        <div className="space-y-1">
                            {files.map((file) => (
                                <button
                                    key={file.path}
                                    onClick={() => navigate(`/scripts/${file.path}`)}
                                    className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm transition-colors hover:bg-accent"
                                >
                                    <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                                    <div className="min-w-0 flex-1">
                                        <div className="font-medium">{file.name}</div>
                                        <div className="text-xs text-muted-foreground">
                                            {file.handlerCount} handler{file.handlerCount !== 1 ? "s" : ""}
                                            {file.sequenceCount > 0 && (
                                                <>, {file.sequenceCount} sequence{file.sequenceCount !== 1 ? "s" : ""}</>
                                            )}
                                        </div>
                                        {file.handlerNames && file.handlerNames.length > 0 && (
                                            <div className="mt-1 flex flex-wrap gap-1">
                                                {file.handlerNames.map((name) => (
                                                    <span
                                                        key={name}
                                                        className="inline-block rounded bg-primary/10 px-1.5 py-0.5 text-xs text-primary"
                                                    >
                                                        {name}
                                                    </span>
                                                ))}
                                            </div>
                                        )}
                                    </div>
                                    <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
                                </button>
                            ))}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    )
}
