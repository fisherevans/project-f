import { useState, useMemo, useEffect } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { Input } from "@/components/ui/input"
import { useSprites } from "@/api/sprites"
import { apiImageUrl } from "@/api/client"
import { cn } from "@/lib/utils"
import { ScaffoldDialog } from "@/components/sprites/ScaffoldDialog"
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

const typeBadgeClass: Record<string, string> = {
    tilesheet: "bg-blue-500/10 text-blue-500",
    frame: "bg-green-500/10 text-green-500",
    plain: "bg-zinc-500/10 text-zinc-400",
    nonAtlas: "bg-orange-500/10 text-orange-500",
}

function DirectoryTreeItem({
    node,
    depth,
    selectedDir,
    onSelect,
}: {
    node: DirectoryNode
    depth: number
    selectedDir: string
    onSelect: (path: string) => void
}) {
    const isSelected = selectedDir === node.path
    const children = Array.from(node.children.values()).sort((a, b) => a.name.localeCompare(b.name))

    return (
        <>
            {node.name && (
                <button
                    className={cn(
                        "w-full text-left px-2 py-1 text-sm rounded transition-colors",
                        isSelected
                            ? "bg-accent text-accent-foreground"
                            : "hover:bg-accent/50 text-foreground",
                    )}
                    style={{ paddingLeft: `${depth * 12 + 8}px` }}
                    onClick={() => onSelect(node.path)}
                >
                    {node.name}
                    {node.count > 0 && (
                        <span className="ml-1 text-xs text-muted-foreground">({node.count})</span>
                    )}
                </button>
            )}
            {children.map((child) => (
                <DirectoryTreeItem
                    key={child.path}
                    node={child}
                    depth={node.name ? depth + 1 : depth}
                    selectedDir={selectedDir}
                    onSelect={onSelect}
                />
            ))}
        </>
    )
}

export function SpriteBrowser() {
    const { data: sprites, isLoading } = useSprites()
    const [search, setSearch] = useState("")
    const [selectedDir, setSelectedDir] = useState("")
    const navigate = useNavigate()
    const [searchParams] = useSearchParams()

    useEffect(() => {
        const dir = searchParams.get("dir")
        if (dir) setSelectedDir(dir)
    }, [searchParams])

    const tree = useMemo(() => {
        if (!sprites) return null
        return buildTree(sprites)
    }, [sprites])

    const directories = useMemo(() => {
        if (!sprites) return []
        const dirs = new Set<string>()
        for (const s of sprites) {
            if (s.directory) dirs.add(s.directory)
        }
        return Array.from(dirs).sort()
    }, [sprites])

    const filtered = useMemo(() => {
        if (!sprites) return []
        let items = sprites
        if (selectedDir) {
            items = items.filter(
                (s) => s.directory === selectedDir || s.directory.startsWith(selectedDir + "/"),
            )
        }
        if (search) {
            const q = search.toLowerCase()
            items = items.filter((s) => s.path.toLowerCase().includes(q))
        }
        return items
    }, [sprites, selectedDir, search])

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-full text-muted-foreground">
                Loading sprites...
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
                                selectedDir === ""
                                    ? "bg-accent text-accent-foreground"
                                    : "hover:bg-accent/50 text-foreground",
                            )}
                            onClick={() => setSelectedDir("")}
                        >
                            All ({sprites?.length ?? 0})
                        </button>
                        {tree && (
                            <DirectoryTreeItem
                                node={tree}
                                depth={0}
                                selectedDir={selectedDir}
                                onSelect={setSelectedDir}
                            />
                        )}
                    </div>
                </div>
            </div>

            <div className="flex-1 flex flex-col overflow-hidden">
                <div className="p-3 border-b border-border flex items-center gap-3">
                    <Input
                        placeholder="Filter sprites..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        className="max-w-sm"
                    />
                    <ScaffoldDialog directories={directories} />
                </div>
                <div className="flex-1 overflow-y-auto">
                    <div className="p-3 grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-3">
                        {filtered.map((sprite) => (
                            <button
                                key={sprite.path}
                                onClick={() => navigate(`/sprites/${sprite.path}`)}
                                className="group border border-border rounded-lg p-2 hover:bg-accent/50 transition-colors text-left"
                            >
                                <div className="aspect-square bg-zinc-900 rounded flex items-center justify-center overflow-hidden mb-2">
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
                        {filtered.length === 0 && (
                            <div className="col-span-full text-center text-muted-foreground py-8">
                                No sprites found.
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}
