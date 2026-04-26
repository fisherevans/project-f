import { useState, useMemo, useEffect, useCallback } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { Input } from "@/components/ui/input"
import { useSprites } from "@/api/sprites"
import { apiImageUrl } from "@/api/client"
import { cn } from "@/lib/utils"
import { ScaffoldDialog } from "@/components/sprites/ScaffoldDialog"
import { Folder, ChevronRight, ChevronDown, ArrowLeft, PanelLeftClose, PanelLeft } from "lucide-react"
import { usePageTitle } from "@/hooks/usePageTitle"
import type { SpriteEntry } from "@/types/sprites"

interface DirectoryNode {
    name: string
    path: string
    children: Map<string, DirectoryNode>
    count: number
}

function buildTree(sprites: SpriteEntry[]): DirectoryNode {
    const root: DirectoryNode = { name: "", path: "", children: new Map(), count: 0 }
    for (const sprite of sprites) {
        const parts = (sprite.directory || "").split("/").filter(Boolean)
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

const typeBadgeClass: Record<string, string> = {
    tilesheet: "bg-accent-blue-tint text-accent-blue",
    frame: "bg-accent-green-tint text-accent-green",
    plain: "bg-muted text-muted-foreground",
    nonAtlas: "bg-accent-orange-tint text-accent-orange",
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

function getSubdirectories(sprites: SpriteEntry[], currentDir: string): string[] {
    const subdirs = new Set<string>()
    for (const sprite of sprites) {
        const dir = sprite.directory || ""
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

function getDirectFiles(sprites: SpriteEntry[], currentDir: string): SpriteEntry[] {
    return sprites.filter((s) => s.directory === currentDir)
}

const SPRITE_TYPES = ["tilesheet", "frame", "plain", "nonAtlas"] as const

export function SpriteBrowser() {
    const { data: sprites, isLoading } = useSprites()
    const [search, setSearch] = useState("")
    const [currentDir, setCurrentDir] = useState("")
    const [treeExpanded, setTreeExpanded] = useState<Set<string>>(new Set())
    const navigate = useNavigate()
    const [searchParams, setSearchParams] = useSearchParams()
    const [showSidebar, setShowSidebar] = useState(() => searchParams.get("explorer") !== "1")
    const [activeTypes, setActiveTypes] = useState<Set<string>>(() => {
        const typesParam = searchParams.get("types")
        return typesParam ? new Set(typesParam.split(",")) : new Set(SPRITE_TYPES)
    })

    usePageTitle(currentDir ? `${currentDir} - Sprites` : "Sprites")

    useEffect(() => {
        const dir = searchParams.get("dir")
        if (dir != null) setCurrentDir(dir)
    }, [searchParams])

    const toggleType = useCallback((type: string) => {
        setActiveTypes(prev => {
            const next = new Set(prev)
            if (next.has(type)) {
                if (next.size > 1) next.delete(type)
            } else {
                next.add(type)
            }
            const params = new URLSearchParams(searchParams)
            if (next.size === SPRITE_TYPES.length) {
                params.delete("types")
            } else {
                params.set("types", Array.from(next).join(","))
            }
            setSearchParams(params, { replace: true })
            return next
        })
    }, [searchParams, setSearchParams])

    const toggleSidebar = useCallback(() => {
        setShowSidebar(prev => {
            const next = !prev
            const params = new URLSearchParams(searchParams)
            if (next) {
                params.delete("explorer")
            } else {
                params.set("explorer", "1")
            }
            setSearchParams(params, { replace: true })
            return next
        })
    }, [searchParams, setSearchParams])

    const typeFilteredSprites = useMemo(() => {
        if (!sprites) return undefined
        if (activeTypes.size === SPRITE_TYPES.length) return sprites
        return sprites.filter(s => activeTypes.has(s.type))
    }, [sprites, activeTypes])

    const tree = useMemo(() => {
        if (!typeFilteredSprites) return null
        return buildTree(typeFilteredSprites)
    }, [typeFilteredSprites])

    const directories = useMemo(() => {
        if (!sprites) return []
        const dirs = new Set<string>()
        for (const s of sprites) {
            if (s.directory) dirs.add(s.directory)
        }
        return Array.from(dirs).sort()
    }, [sprites])

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
        if (!typeFilteredSprites) return []
        return getSubdirectories(typeFilteredSprites, currentDir)
    }, [typeFilteredSprites, currentDir])

    const directFiles = useMemo(() => {
        if (!typeFilteredSprites) return []
        let items = getDirectFiles(typeFilteredSprites, currentDir)
        if (search) {
            const q = search.toLowerCase()
            items = items.filter((s) => s.name.toLowerCase().includes(q) || s.path.toLowerCase().includes(q))
        }
        return items
    }, [typeFilteredSprites, currentDir, search])

    const isSearching = search.length > 0
    const searchResults = useMemo(() => {
        if (!typeFilteredSprites || !isSearching) return []
        const q = search.toLowerCase()
        return typeFilteredSprites.filter((s) => {
            if (currentDir && !s.directory.startsWith(currentDir)) return false
            return s.name.toLowerCase().includes(q) || s.path.toLowerCase().includes(q)
        })
    }, [typeFilteredSprites, search, currentDir, isSearching])

    const displayFiles = isSearching ? searchResults : directFiles

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-full text-muted-foreground">
                Loading sprites...
            </div>
        )
    }

    return (
        <div className="flex h-full">
            {showSidebar && (
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
                                All ({typeFilteredSprites?.length ?? 0})
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
            )}

            <div className="flex-1 flex flex-col overflow-hidden">
                <div className="p-3 border-b border-border flex items-center gap-3">
                    <button
                        onClick={toggleSidebar}
                        className="flex items-center justify-center h-7 w-7 rounded hover:bg-accent transition-colors text-muted-foreground hover:text-foreground"
                        title={showSidebar ? "Hide sidebar (explorer mode)" : "Show sidebar"}
                    >
                        {showSidebar ? <PanelLeftClose className="h-4 w-4" /> : <PanelLeft className="h-4 w-4" />}
                    </button>
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
                    <div className="flex items-center gap-1">
                        {SPRITE_TYPES.map(type => (
                            <button
                                key={type}
                                onClick={() => toggleType(type)}
                                className={cn(
                                    "text-[10px] px-1.5 py-0.5 rounded-full font-medium border transition-colors",
                                    activeTypes.has(type)
                                        ? typeBadgeClass[type] + " border-current/20"
                                        : "bg-muted/30 text-muted-foreground/40 border-transparent",
                                )}
                                title={`${activeTypes.has(type) ? "Hide" : "Show"} ${type} sprites`}
                            >
                                {type}
                            </button>
                        ))}
                    </div>
                    <Input
                        placeholder="Filter..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        className="max-w-xs"
                    />
                    <ScaffoldDialog directories={directories} />
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
                        <div className="p-3 grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-3">
                            {displayFiles.map((sprite) => (
                                <button
                                    key={sprite.path}
                                    onClick={() => navigate(`/sprites/${sprite.path}`)}
                                    className="group border border-border rounded-lg p-2 hover:bg-accent/50 transition-colors text-left"
                                >
                                    <div className="aspect-square bg-canvas rounded flex items-center justify-center overflow-hidden mb-2">
                                        <img
                                            src={apiImageUrl(sprite.path)}
                                            alt={sprite.name}
                                            loading="lazy"
                                            className="max-w-full max-h-full object-contain"
                                            style={{ imageRendering: "pixelated" }}
                                        />
                                    </div>
                                    <p className="text-xs font-medium truncate" title={sprite.name}>
                                        {sprite.name}
                                    </p>
                                    <div className="flex items-center gap-1 mt-1">
                                        <span
                                            className={cn(
                                                "text-[10px] px-1.5 py-0.5 rounded-full font-medium",
                                                typeBadgeClass[sprite.type] ?? "",
                                            )}
                                        >
                                            {sprite.type}
                                        </span>
                                        <span className="text-[10px] text-muted-foreground">
                                            {sprite.imageWidth}x{sprite.imageHeight}
                                        </span>
                                    </div>
                                </button>
                            ))}
                        </div>
                    )}
                    {displayFiles.length === 0 && subdirs.length === 0 && (
                        <div className="text-center text-muted-foreground py-8">
                            No sprites found.
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
