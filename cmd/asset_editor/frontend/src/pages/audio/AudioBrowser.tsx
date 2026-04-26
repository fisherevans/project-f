import { useState, useMemo, useEffect, useRef, useCallback } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { Input } from "@/components/ui/input"
import { useAudioFiles } from "@/api/audio"
import { apiAudioUrl } from "@/api/client"
import { cn } from "@/lib/utils"
import { Play, Square, Volume2, Copy, Check, Folder, ChevronRight, ChevronDown, ArrowLeft } from "lucide-react"
import { usePageTitle } from "@/hooks/usePageTitle"
import type { AudioEntry } from "@/types/audio"

interface DirectoryNode {
    name: string
    path: string
    children: Map<string, DirectoryNode>
    count: number
}

function buildTree(files: AudioEntry[]): DirectoryNode {
    const root: DirectoryNode = { name: "", path: "", children: new Map(), count: 0 }
    for (const file of files) {
        const parts = (file.directory || "").split("/").filter(Boolean)
        let node = root
        let currentPath = ""
        for (const part of parts) {
            currentPath = currentPath ? `${currentPath}/${part}` : part
            if (!node.children.has(part)) {
                node.children.set(part, { name: part, path: currentPath, children: new Map(), count: 0 })
            }
            node = node.children.get(part)!
        }
        node.count++
    }
    return root
}

function countDescendants(node: DirectoryNode): number {
    let total = node.count
    for (const child of node.children.values()) {
        total += countDescendants(child)
    }
    return total
}

const formatBadgeClass: Record<string, string> = {
    wav: "bg-accent-blue-tint text-accent-blue",
    mp3: "bg-accent-green-tint text-accent-green",
    ogg: "bg-accent-violet-tint text-accent-violet",
    flac: "bg-accent-orange-tint text-accent-orange",
}

function formatFileSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function DirectoryTreeItem({
    node,
    depth,
    selectedDir,
    onSelect,
    expanded,
    onToggle,
}: {
    node: DirectoryNode
    depth: number
    selectedDir: string
    onSelect: (path: string) => void
    expanded: Set<string>
    onToggle: (path: string) => void
}) {
    const isSelected = selectedDir === node.path
    const isExpanded = expanded.has(node.path)
    const children = Array.from(node.children.values()).sort((a, b) => a.name.localeCompare(b.name))
    const hasChildren = children.length > 0

    return (
        <>
            {node.name && (
                <button
                    className={cn(
                        "w-full text-left px-2 py-1 text-sm rounded transition-colors flex items-center gap-1",
                        isSelected
                            ? "bg-accent text-accent-foreground"
                            : "hover:bg-accent/50 text-foreground",
                    )}
                    style={{ paddingLeft: `${depth * 12 + 4}px` }}
                    onClick={() => {
                        onSelect(node.path)
                        if (hasChildren) onToggle(node.path)
                    }}
                >
                    {hasChildren ? (
                        isExpanded ? <ChevronDown className="h-3 w-3 shrink-0" /> : <ChevronRight className="h-3 w-3 shrink-0" />
                    ) : (
                        <span className="w-3 shrink-0" />
                    )}
                    <span className="truncate">{node.name}</span>
                    <span className="ml-auto text-xs text-muted-foreground">{countDescendants(node)}</span>
                </button>
            )}
            {(isExpanded || !node.name) && children.map((child) => (
                <DirectoryTreeItem
                    key={child.path}
                    node={child}
                    depth={node.name ? depth + 1 : depth}
                    selectedDir={selectedDir}
                    onSelect={onSelect}
                    expanded={expanded}
                    onToggle={onToggle}
                />
            ))}
        </>
    )
}

function InlinePlayer({ path, gain }: { path: string; gain?: number }) {
    const audioRef = useRef<HTMLAudioElement | null>(null)
    const [playing, setPlaying] = useState(false)

    const toggle = useCallback((e: React.MouseEvent) => {
        e.stopPropagation()
        e.preventDefault()
        if (!audioRef.current) {
            const audio = new Audio(apiAudioUrl(path))
            audio.volume = gain ?? 1.0
            audio.onended = () => setPlaying(false)
            audioRef.current = audio
        }
        if (playing) {
            audioRef.current.pause()
            audioRef.current.currentTime = 0
            setPlaying(false)
        } else {
            audioRef.current.volume = gain ?? 1.0
            audioRef.current.play()
            setPlaying(true)
        }
    }, [path, gain, playing])

    useEffect(() => {
        return () => {
            if (audioRef.current) {
                audioRef.current.pause()
                audioRef.current = null
            }
        }
    }, [])

    return (
        <button
            onClick={toggle}
            className={cn(
                "flex items-center justify-center w-7 h-7 rounded-full transition-colors",
                playing
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted hover:bg-primary/20 text-foreground",
            )}
            title={playing ? "Stop" : "Play"}
        >
            {playing ? <Square className="h-3 w-3" /> : <Play className="h-3 w-3 ml-0.5" />}
        </button>
    )
}

function CopyButton({ text }: { text: string }) {
    const [copied, setCopied] = useState(false)

    const copy = useCallback((e: React.MouseEvent) => {
        e.stopPropagation()
        e.preventDefault()
        navigator.clipboard.writeText(text)
        setCopied(true)
        setTimeout(() => setCopied(false), 1500)
    }, [text])

    return (
        <button
            onClick={copy}
            className="p-1 rounded hover:bg-accent transition-colors text-muted-foreground hover:text-foreground"
            title={`Copy: ${text}`}
        >
            {copied ? <Check className="h-3.5 w-3.5 text-accent-green" /> : <Copy className="h-3.5 w-3.5" />}
        </button>
    )
}

function getSubdirectories(files: AudioEntry[], currentDir: string): string[] {
    const subdirs = new Set<string>()
    for (const file of files) {
        const dir = file.directory || ""
        if (currentDir === "") {
            const top = dir.split("/")[0]
            if (top) subdirs.add(top)
        } else if (dir.startsWith(currentDir + "/")) {
            const rest = dir.slice(currentDir.length + 1)
            const next = rest.split("/")[0]
            if (next) subdirs.add(next)
        }
    }
    return Array.from(subdirs).sort()
}

function getDirectFiles(files: AudioEntry[], currentDir: string): AudioEntry[] {
    return files.filter((f) => f.directory === currentDir)
}

export function AudioBrowser() {
    const { data: files, isLoading } = useAudioFiles()
    const [search, setSearch] = useState("")
    const [currentDir, setCurrentDir] = useState("")
    const [treeExpanded, setTreeExpanded] = useState<Set<string>>(new Set())
    const navigate = useNavigate()
    const [searchParams, setSearchParams] = useSearchParams()

    usePageTitle(currentDir ? `${currentDir} - Audio` : "Audio")

    useEffect(() => {
        const dir = searchParams.get("dir")
        if (dir != null) setCurrentDir(dir)
    }, [searchParams])

    const tree = useMemo(() => {
        if (!files) return null
        return buildTree(files)
    }, [files])

    const navigateToDir = useCallback((dir: string) => {
        setCurrentDir(dir)
        setSearch("")
        setSearchParams(dir ? { dir } : {})
    }, [setSearchParams])

    const toggleTreeNode = useCallback((path: string) => {
        setTreeExpanded((prev) => {
            const next = new Set(prev)
            if (next.has(path)) {
                next.delete(path)
            } else {
                next.add(path)
            }
            return next
        })
    }, [])

    const parentDir = useMemo(() => {
        if (!currentDir) return null
        const idx = currentDir.lastIndexOf("/")
        return idx === -1 ? "" : currentDir.slice(0, idx)
    }, [currentDir])

    const subdirs = useMemo(() => {
        if (!files) return []
        return getSubdirectories(files, currentDir)
    }, [files, currentDir])

    const directFiles = useMemo(() => {
        if (!files) return []
        let items = getDirectFiles(files, currentDir)
        if (search) {
            const q = search.toLowerCase()
            items = items.filter((s) => s.name.toLowerCase().includes(q) || s.resourceName.toLowerCase().includes(q))
        }
        return items
    }, [files, currentDir, search])

    const isSearching = search.length > 0
    const searchResults = useMemo(() => {
        if (!files || !isSearching) return []
        const q = search.toLowerCase()
        return files.filter((s) => {
            if (currentDir && !s.directory.startsWith(currentDir)) return false
            return s.name.toLowerCase().includes(q) || s.resourceName.toLowerCase().includes(q) || s.path.toLowerCase().includes(q)
        })
    }, [files, search, currentDir, isSearching])

    const displayFiles = isSearching ? searchResults : directFiles

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-full text-muted-foreground">
                Loading audio files...
            </div>
        )
    }

    return (
        <div className="flex h-full">
            <div className="w-56 border-r border-border flex flex-col">
                <div className="p-2 border-b border-border">
                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                        Directories
                    </span>
                </div>
                <div className="flex-1 overflow-y-auto">
                    <div className="p-1">
                        <button
                            className={cn(
                                "w-full text-left px-2 py-1 text-sm rounded transition-colors",
                                currentDir === ""
                                    ? "bg-accent text-accent-foreground"
                                    : "hover:bg-accent/50 text-foreground",
                            )}
                            onClick={() => navigateToDir("")}
                        >
                            All ({files?.length ?? 0})
                        </button>
                        {tree && (
                            <DirectoryTreeItem
                                node={tree}
                                depth={0}
                                selectedDir={currentDir}
                                onSelect={navigateToDir}
                                expanded={treeExpanded}
                                onToggle={toggleTreeNode}
                            />
                        )}
                    </div>
                </div>
            </div>

            <div className="flex-1 flex flex-col overflow-hidden">
                <div className="p-3 border-b border-border flex items-center gap-3">
                    {parentDir != null && (
                        <button
                            onClick={() => navigateToDir(parentDir)}
                            className="flex items-center gap-1 px-2 py-1 text-sm rounded hover:bg-accent transition-colors text-muted-foreground hover:text-foreground"
                        >
                            <ArrowLeft className="h-3.5 w-3.5" />
                        </button>
                    )}
                    {currentDir && (
                        <div className="flex items-center gap-1 text-sm">
                            {currentDir.split("/").map((part, i, arr) => {
                                const path = arr.slice(0, i + 1).join("/")
                                return (
                                    <span key={path} className="flex items-center gap-1">
                                        {i > 0 && <span className="text-muted-foreground">/</span>}
                                        <button
                                            onClick={() => navigateToDir(path)}
                                            className="hover:underline text-foreground"
                                        >
                                            {part}
                                        </button>
                                    </span>
                                )
                            })}
                        </div>
                    )}
                    <div className="flex-1" />
                    <Input
                        placeholder="Filter..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        className="max-w-xs"
                    />
                </div>
                <div className="flex-1 overflow-y-auto">
                    {!isSearching && subdirs.length > 0 && (
                        <div className="p-3 grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-2">
                            {subdirs.map((dir) => {
                                const fullPath = currentDir ? `${currentDir}/${dir}` : dir
                                return (
                                    <button
                                        key={dir}
                                        onClick={() => navigateToDir(fullPath)}
                                        className="flex items-center gap-2 p-3 rounded-lg border border-border hover:bg-accent/50 transition-colors text-left"
                                    >
                                        <Folder className="h-5 w-5 text-muted-foreground shrink-0" />
                                        <span className="text-sm font-medium truncate">{dir}</span>
                                    </button>
                                )
                            })}
                        </div>
                    )}
                    {displayFiles.length > 0 && (
                        <table className="w-full text-sm">
                            <thead className="sticky top-0 bg-background border-b border-border">
                                <tr className="text-left text-muted-foreground">
                                    <th className="px-3 py-2 w-10"></th>
                                    <th className="px-3 py-2">Name</th>
                                    <th className="px-3 py-2">Resource Name</th>
                                    <th className="px-3 py-2 w-16">Format</th>
                                    <th className="px-3 py-2 w-20">Size</th>
                                    <th className="px-3 py-2 w-16">Gain</th>
                                </tr>
                            </thead>
                            <tbody>
                                {displayFiles.map((file) => (
                                    <tr
                                        key={file.path}
                                        className="border-b border-border/50 hover:bg-accent/30 transition-colors group"
                                    >
                                        <td className="px-3 py-1.5">
                                            <InlinePlayer path={file.path} gain={file.gain} />
                                        </td>
                                        <td className="px-3 py-1.5">
                                            <button
                                                onClick={() => navigate(`/audio/${file.path}`)}
                                                className="text-left hover:underline font-medium"
                                            >
                                                {file.name}
                                            </button>
                                        </td>
                                        <td className="px-3 py-1.5">
                                            <span className="inline-flex items-center gap-1">
                                                <code className="text-xs text-muted-foreground font-mono">{file.resourceName}</code>
                                                <span className="opacity-0 group-hover:opacity-100 transition-opacity">
                                                    <CopyButton text={file.resourceName} />
                                                </span>
                                            </span>
                                        </td>
                                        <td className="px-3 py-1.5">
                                            <span
                                                className={cn(
                                                    "text-[10px] px-1.5 py-0.5 rounded-full font-medium",
                                                    formatBadgeClass[file.format] ?? "",
                                                )}
                                            >
                                                {file.format}
                                            </span>
                                        </td>
                                        <td className="px-3 py-1.5 text-muted-foreground">
                                            {formatFileSize(file.fileSize)}
                                        </td>
                                        <td className="px-3 py-1.5">
                                            {file.gain != null ? (
                                                <span className="flex items-center gap-1 text-muted-foreground">
                                                    <Volume2 className="h-3 w-3" />
                                                    {Math.round(file.gain * 100)}%
                                                </span>
                                            ) : (
                                                <span className="text-muted-foreground/50">-</span>
                                            )}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    )}
                    {displayFiles.length === 0 && subdirs.length === 0 && (
                        <div className="text-center text-muted-foreground py-8">
                            No audio files found.
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
