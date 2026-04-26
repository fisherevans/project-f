import { useMemo } from "react"
import { useNavigate } from "react-router-dom"
import { useScripts } from "@/api/scripts"
import { FileText, ChevronRight } from "lucide-react"
import { usePageTitle } from "@/hooks/usePageTitle"
import type { ScriptFileEntry } from "@/types/scripts"

function ScriptCounts({ file }: { file: ScriptFileEntry }) {
    const parts: string[] = [];
    if (file.handlerCount > 0) parts.push(`${file.handlerCount} handler${file.handlerCount !== 1 ? "s" : ""}`);
    if (file.sequenceCount > 0) parts.push(`${file.sequenceCount} sequence${file.sequenceCount !== 1 ? "s" : ""}`);
    if (file.customActionNames && file.customActionNames.length > 0) parts.push(`${file.customActionNames.length} action${file.customActionNames.length !== 1 ? "s" : ""}`);
    if (file.dataListNames && file.dataListNames.length > 0) parts.push(`${file.dataListNames.length} data list${file.dataListNames.length !== 1 ? "s" : ""}`);
    if (file.constNames && file.constNames.length > 0) parts.push(`${file.constNames.length} const${file.constNames.length !== 1 ? "s" : ""}`);
    return <>{parts.join(", ") || "empty"}</>;
}

const tagStyles: Record<string, string> = {
    handler: "bg-accent-blue-tint text-accent-blue",
    sequence: "bg-accent-teal-tint text-accent-teal",
    action: "bg-accent-violet-tint text-accent-violet",
    data: "bg-accent-orange-tint text-accent-orange",
    const: "bg-accent-amber-tint text-accent-amber",
    template: "bg-accent-green-tint text-accent-green",
};

function ScriptNameTags({ file }: { file: ScriptFileEntry }) {
    const groups: { label: string; names: string[] }[] = [];
    if (file.handlerNames?.length) groups.push({ label: "handler", names: file.handlerNames });
    if (file.customActionNames?.length) groups.push({ label: "action", names: file.customActionNames });
    if (file.sequenceNames?.length) groups.push({ label: "sequence", names: file.sequenceNames });
    if (file.dataListNames?.length) groups.push({ label: "data", names: file.dataListNames });
    if (file.constNames?.length) groups.push({ label: "const", names: file.constNames });
    if (file.propertyTemplateNames?.length) groups.push({ label: "template", names: file.propertyTemplateNames });

    if (groups.length === 0) return null;

    return (
        <div className="mt-1 flex flex-wrap gap-1">
            {groups.map(({ label, names }) =>
                names.map((name) => (
                    <span
                        key={`${label}-${name}`}
                        className={`inline-block rounded px-1.5 py-0.5 text-[10px] font-mono ${tagStyles[label]}`}
                    >
                        {name}
                    </span>
                )),
            )}
        </div>
    );
}

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
                                            <ScriptCounts file={file} />
                                        </div>
                                        <ScriptNameTags file={file} />
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
