import { useState, useMemo, useRef, useEffect, useLayoutEffect, useCallback } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useEntities, useEntity, useTeleports, useTeleport, useZones } from "@/api/debug"
import { useScripts } from "@/api/scripts"
import { MapPin, X, ArrowUpDown, Map, List, ZoomIn, ZoomOut, Crosshair, Lock, Layers } from "lucide-react"
import { cn } from "@/lib/utils"
import { Link, useSearchParams } from "react-router-dom"
import { usePageTitle } from "@/hooks/usePageTitle"
import type { EntitySnapshot, TeleportEntry as TeleportEntryType, ZoneRect } from "@/types/debug"

type SortKey = "id" | "x" | "y" | "direction" | "proximity"
type SortDir = "asc" | "desc"

function distance(ax: number, ay: number, bx: number, by: number): number {
    return Math.sqrt((ax - bx) ** 2 + (ay - by) ** 2)
}

export function DebugEntities() {
    const { data: entityResponse, isLoading, error } = useEntities()
    const { data: teleports } = useTeleports()
    const { data: zones } = useZones()
    const [searchParams, setSearchParams] = useSearchParams()
    const [selectedId, setSelectedId] = useState(() => searchParams.get("entity") ?? "")
    const [view, setView] = useState<"list" | "map">(() => (searchParams.get("view") as "list" | "map") || "list")

    usePageTitle("Entities - Debug")

    useEffect(() => {
        const params = new URLSearchParams(searchParams)
        let changed = false
        if (selectedId) {
            if (params.get("entity") !== selectedId) { params.set("entity", selectedId); changed = true }
        } else {
            if (params.has("entity")) { params.delete("entity"); changed = true }
        }
        if (view !== "list") {
            if (params.get("view") !== view) { params.set("view", view); changed = true }
        } else {
            if (params.has("view")) { params.delete("view"); changed = true }
        }
        if (changed) setSearchParams(params, { replace: true })
    }, [selectedId, view])
    const [filter, setFilter] = useState("")
    const [sortKey, setSortKey] = useState<SortKey>("id")
    const [sortDir, setSortDir] = useState<SortDir>("asc")
    const [filterMoving, setFilterMoving] = useState<boolean | null>(null)
    const [filterDirections, setFilterDirections] = useState<Set<string>>(new Set())
    const [filterTypes, setFilterTypes] = useState<Set<string>>(new Set())
    const [filterHandler, setFilterHandler] = useState<string>("")
    const teleport = useTeleport()

    const entities = entityResponse?.entities ?? []
    const mapWidth = entityResponse?.map_width ?? 0
    const mapHeight = entityResponse?.map_height ?? 0

    const player = useMemo(() => entities.find(e => e.is_player), [entities])

    const toggleSort = useCallback((key: SortKey) => {
        if (sortKey === key) {
            setSortDir(d => d === "asc" ? "desc" : "asc")
        } else {
            setSortKey(key)
            setSortDir("asc")
        }
    }, [sortKey])

    const filtered = useMemo(() => {
        let items = entities
        if (filter) {
            const q = filter.toLowerCase()
            items = items.filter(e => e.id.toLowerCase().includes(q))
        }
        if (filterMoving !== null) {
            items = items.filter(e => e.is_moving === filterMoving)
        }
        if (filterDirections.size > 0) {
            items = items.filter(e => filterDirections.has(e.direction))
        }
        if (filterTypes.size > 0) {
            items = items.filter(e => filterTypes.has(e.debug_type || ""))
        }
        if (filterHandler) {
            const hq = filterHandler.toLowerCase()
            items = items.filter(e => e.handler_ref?.toLowerCase().includes(hq))
        }
        return items
    }, [entities, filter, filterMoving, filterDirections, filterTypes, filterHandler])

    const sorted = useMemo(() => {
        const arr = [...filtered]
        const dir = sortDir === "asc" ? 1 : -1
        arr.sort((a, b) => {
            switch (sortKey) {
                case "id": return dir * a.id.localeCompare(b.id)
                case "x": return dir * (a.x - b.x)
                case "y": return dir * (a.y - b.y)
                case "direction": return dir * a.direction.localeCompare(b.direction)
                case "proximity": {
                    if (!player) return 0
                    const da = distance(a.x, a.y, player.x, player.y)
                    const db = distance(b.x, b.y, player.x, player.y)
                    return dir * (da - db)
                }
                default: return 0
            }
        })
        return arr
    }, [filtered, sortKey, sortDir, player])

    const notAdventure = error?.message?.includes("409")
    const directions = useMemo(() => {
        const ds = new Set<string>()
        for (const e of entities) ds.add(e.direction)
        return Array.from(ds).sort()
    }, [entities])
    const debugTypes = useMemo(() => {
        const ts = new Set<string>()
        for (const e of entities) if (e.debug_type) ts.add(e.debug_type)
        return Array.from(ts).sort()
    }, [entities])
    const handlerRefs = useMemo(() => {
        const hs = new Set<string>()
        for (const e of entities) if (e.handler_ref) hs.add(e.handler_ref)
        return Array.from(hs).sort()
    }, [entities])

    return (
        <div className="flex h-full overflow-hidden">
            <div className="flex-1 flex flex-col overflow-hidden p-4 gap-3">
                <div className="flex items-center gap-2">
                    <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                        Entities
                        {entities.length > 0 && <span className="ml-1 text-xs">({filtered.length}/{entities.length})</span>}
                    </h2>
                    <div className="flex-1" />
                    <div className="flex border rounded overflow-hidden">
                        <button
                            className={cn("px-2 py-1 text-xs", view === "list" ? "bg-accent" : "hover:bg-accent/50")}
                            onClick={() => setView("list")}
                        >
                            <List className="h-3.5 w-3.5" />
                        </button>
                        <button
                            className={cn("px-2 py-1 text-xs border-l", view === "map" ? "bg-accent" : "hover:bg-accent/50")}
                            onClick={() => setView("map")}
                        >
                            <Map className="h-3.5 w-3.5" />
                        </button>
                    </div>
                </div>

                {isLoading && <div className="text-sm text-muted-foreground">Loading...</div>}
                {notAdventure && (
                    <div className="text-sm text-muted-foreground">
                        Not in adventure state. Entity data is only available during exploration.
                    </div>
                )}
                {error && !notAdventure && (
                    <div className="text-sm text-destructive">Cannot reach debug API. Is the game running?</div>
                )}

                {entities.length > 0 && (
                    <div className="flex flex-col gap-1.5">
                        <div className="flex items-center gap-2">
                            <Input
                                placeholder="Filter by ID..."
                                value={filter}
                                onChange={e => setFilter(e.target.value)}
                                className="h-7 text-xs max-w-[200px]"
                            />
                            <select
                                className="h-7 text-xs bg-background border rounded px-1.5"
                                value={filterMoving === null ? "" : filterMoving ? "yes" : "no"}
                                onChange={e => {
                                    const v = e.target.value
                                    setFilterMoving(v === "" ? null : v === "yes")
                                }}
                            >
                                <option value="">All</option>
                                <option value="yes">Moving</option>
                                <option value="no">Idle</option>
                            </select>
                            {handlerRefs.length > 0 && (
                                <Input
                                    placeholder="Handler..."
                                    value={filterHandler}
                                    onChange={e => setFilterHandler(e.target.value)}
                                    className="h-7 text-xs max-w-[150px]"
                                />
                            )}
                        </div>
                        <div className="flex items-center gap-1 flex-wrap">
                            <span className="text-[10px] text-muted-foreground mr-0.5">Type:</span>
                            {debugTypes.map(t => {
                                const active = filterTypes.has(t)
                                const style = ENTITY_TYPE_STYLES[t]
                                return (
                                    <button
                                        key={t}
                                        className={cn(
                                            "h-5 px-1.5 text-[10px] rounded border transition-colors",
                                            active
                                                ? "border-current font-medium"
                                                : "border-border text-muted-foreground hover:text-foreground hover:border-foreground/30"
                                        )}
                                        style={active && style ? { color: style.color, borderColor: style.color } : undefined}
                                        onClick={() => setFilterTypes(prev => {
                                            const next = new Set(prev)
                                            if (next.has(t)) next.delete(t); else next.add(t)
                                            return next
                                        })}
                                    >
                                        {t}
                                    </button>
                                )
                            })}
                            {filterTypes.size > 0 && (
                                <button className="h-5 px-1 text-[10px] text-muted-foreground hover:text-foreground" onClick={() => setFilterTypes(new Set())}>clear</button>
                            )}
                        </div>
                        {directions.length > 1 && (
                            <div className="flex items-center gap-1 flex-wrap">
                                <span className="text-[10px] text-muted-foreground mr-0.5">Dir:</span>
                                {directions.map(d => (
                                    <button
                                        key={d}
                                        className={cn(
                                            "h-5 px-1.5 text-[10px] rounded border transition-colors",
                                            filterDirections.has(d) ? "border-foreground/50 text-foreground" : "border-border text-muted-foreground hover:text-foreground hover:border-foreground/30"
                                        )}
                                        onClick={() => setFilterDirections(prev => {
                                            const next = new Set(prev)
                                            if (next.has(d)) next.delete(d); else next.add(d)
                                            return next
                                        })}
                                    >
                                        {d}
                                    </button>
                                ))}
                                <button className="h-5 px-1 text-[10px] text-muted-foreground hover:text-foreground" onClick={() => setFilterDirections(new Set())}>clear</button>
                            </div>
                        )}
                    </div>
                )}

                {entities.length > 0 && view === "list" && (
                    <EntityTable
                        entities={sorted}
                        player={player}
                        selectedId={selectedId}
                        sortKey={sortKey}
                        sortDir={sortDir}
                        onSort={toggleSort}
                        onSelect={setSelectedId}
                        onTeleport={id => teleport.mutate({ target: id })}
                    />
                )}

                {entities.length > 0 && view === "map" && (
                    <EntityMap
                        entities={filtered}
                        teleports={teleports ?? []}
                        zones={zones ?? []}
                        player={player}
                        mapWidth={mapWidth}
                        mapHeight={mapHeight}
                        selectedId={selectedId}
                        onSelect={setSelectedId}
                        onTeleport={(x, y) => teleport.mutate({ x, y })}
                    />
                )}

                {teleports && teleports.length > 0 && view === "list" && (
                    <>
                        <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mt-1">Teleports</h2>
                        <div className="overflow-auto rounded-md border max-h-48">
                            <table className="w-full text-sm">
                                <thead className="sticky top-0 bg-muted">
                                    <tr>
                                        <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Reference</th>
                                        <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Pos</th>
                                        <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Dest</th>
                                        <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Exit</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {teleports.map(t => (
                                        <tr key={t.reference} className="border-t hover:bg-muted/50">
                                            <td className="px-3 py-1.5 font-mono text-xs font-medium">{t.reference}</td>
                                            <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">({t.x}, {t.y})</td>
                                            <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">{t.destination}</td>
                                            <td className="px-3 py-1.5 text-xs text-muted-foreground">{t.exit_direction}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </>
                )}
            </div>

            {selectedId && (
                <div className="w-72 border-l pl-4 pr-4 pt-4 overflow-auto shrink-0">
                    <EntityDetailPanel id={selectedId} snapshot={entities.find(e => e.id === selectedId)} onClose={() => setSelectedId("")} />
                </div>
            )}
        </div>
    )
}

function SortHeader({ label, sortKey: key, currentKey, currentDir, onSort, className }: {
    label: string
    sortKey: SortKey
    currentKey: SortKey
    currentDir: SortDir
    onSort: (k: SortKey) => void
    className?: string
}) {
    const active = currentKey === key
    return (
        <th
            className={cn("px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase cursor-pointer hover:text-foreground select-none", className)}
            onClick={() => onSort(key)}
        >
            <span className="inline-flex items-center gap-0.5">
                {label}
                {active && (
                    <ArrowUpDown className={cn("h-3 w-3", currentDir === "desc" && "rotate-180")} />
                )}
            </span>
        </th>
    )
}

function EntityTable({ entities, player, selectedId, sortKey, sortDir, onSort, onSelect, onTeleport }: {
    entities: EntitySnapshot[]
    player?: EntitySnapshot
    selectedId: string
    sortKey: SortKey
    sortDir: SortDir
    onSort: (k: SortKey) => void
    onSelect: (id: string) => void
    onTeleport: (id: string) => void
}) {
    return (
        <div className="flex-1 overflow-auto rounded-md border">
            <table className="w-full text-sm">
                <thead className="sticky top-0 bg-muted">
                    <tr>
                        <SortHeader label="ID" sortKey="id" currentKey={sortKey} currentDir={sortDir} onSort={onSort} />
                        <SortHeader label="X" sortKey="x" currentKey={sortKey} currentDir={sortDir} onSort={onSort} className="w-16" />
                        <SortHeader label="Y" sortKey="y" currentKey={sortKey} currentDir={sortDir} onSort={onSort} className="w-16" />
                        <SortHeader label="Dir" sortKey="direction" currentKey={sortKey} currentDir={sortDir} onSort={onSort} className="w-20" />
                        <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase w-16">Move</th>
                        {player && (
                            <SortHeader label="Dist" sortKey="proximity" currentKey={sortKey} currentDir={sortDir} onSort={onSort} className="w-16" />
                        )}
                        <th className="px-3 py-2 w-12"></th>
                    </tr>
                </thead>
                <tbody>
                    {entities.map(e => {
                        const dist = player ? distance(e.x, e.y, player.x, player.y) : null
                        return (
                            <tr
                                key={e.id}
                                className={cn(
                                    "border-t cursor-pointer transition-colors",
                                    e.id === selectedId ? "bg-accent" : "hover:bg-muted/50",
                                    e.is_player && "font-semibold",
                                )}
                                onClick={() => onSelect(e.id)}
                            >
                                <td className="px-3 py-1.5 font-mono text-xs">
                                    {e.id}
                                    {e.is_player && <Badge variant="secondary" className="ml-1 text-[9px]">player</Badge>}
                                </td>
                                <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">{e.x}</td>
                                <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">{e.y}</td>
                                <td className="px-3 py-1.5 text-xs text-muted-foreground">{e.direction}</td>
                                <td className="px-3 py-1.5">
                                    {e.is_moving && <Badge variant="secondary" className="text-[10px]">moving</Badge>}
                                </td>
                                {player && (
                                    <td className="px-3 py-1.5 font-mono text-xs text-muted-foreground">
                                        {e.is_player ? "-" : dist!.toFixed(1)}
                                    </td>
                                )}
                                <td className="px-3 py-1.5">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        className="h-6 px-2 text-xs"
                                        onClick={(ev) => {
                                            ev.stopPropagation()
                                            onTeleport(e.id)
                                        }}
                                    >
                                        <MapPin className="h-3 w-3" />
                                    </Button>
                                </td>
                            </tr>
                        )
                    })}
                </tbody>
            </table>
        </div>
    )
}

const MAP_PADDING = 24
const ZOOM_LEVELS = [4, 6, 8, 12, 16, 24, 32]
const DEFAULT_ZOOM_IDX = 2

const ENTITY_TYPE_STYLES: Record<string, { color: string; selectedColor: string; shape: "circle" | "square"; showDirection: boolean }> = {
    player:      { color: "#3b82f6", selectedColor: "#60a5fa", shape: "circle", showDirection: true },
    npc:         { color: "#4ade80", selectedColor: "#22d3ee", shape: "circle", showDirection: true },
    collision:   { color: "#52525b", selectedColor: "#a1a1aa", shape: "square", showDirection: false },
    light:       { color: "#fbbf24", selectedColor: "#fde68a", shape: "circle", showDirection: false },
    pickup:      { color: "#f472b6", selectedColor: "#f9a8d4", shape: "circle", showDirection: false },
    interactable:{ color: "#a78bfa", selectedColor: "#c4b5fd", shape: "circle", showDirection: true },
    trigger:     { color: "#fb923c", selectedColor: "#fdba74", shape: "square", showDirection: false },
    system:      { color: "#6b7280", selectedColor: "#9ca3af", shape: "square", showDirection: false },
}
const DEFAULT_STYLE = { color: "#94a3b8", selectedColor: "#cbd5e1", shape: "circle" as const, showDirection: true }

function drawArrow(ctx: CanvasRenderingContext2D, cx: number, cy: number, dir: string, len: number, headSize: number, color: string, width: number) {
    let dx = 0, dy = 0
    switch (dir) {
        case "Up": dy = -len; break
        case "Down": dy = len; break
        case "Left": dx = -len; break
        case "Right": dx = len; break
        default: return
    }
    const ex = cx + dx
    const ey = cy + dy
    ctx.beginPath()
    ctx.moveTo(cx, cy)
    ctx.lineTo(ex, ey)
    ctx.strokeStyle = color
    ctx.lineWidth = width
    ctx.stroke()

    // arrowhead
    const angle = Math.atan2(dy, dx)
    ctx.beginPath()
    ctx.moveTo(ex, ey)
    ctx.lineTo(ex - headSize * Math.cos(angle - 0.5), ey - headSize * Math.sin(angle - 0.5))
    ctx.lineTo(ex - headSize * Math.cos(angle + 0.5), ey - headSize * Math.sin(angle + 0.5))
    ctx.closePath()
    ctx.fillStyle = color
    ctx.fill()
}

function EntityMap({ entities, teleports, zones, player, mapWidth, mapHeight, selectedId, onSelect, onTeleport }: {
    entities: EntitySnapshot[]
    teleports: TeleportEntryType[]
    zones: ZoneRect[]
    player?: EntitySnapshot
    mapWidth: number
    mapHeight: number
    selectedId: string
    onSelect: (id: string) => void
    onTeleport: (x: number, y: number) => void
}) {
    const canvasRef = useRef<HTMLCanvasElement>(null)
    const containerRef = useRef<HTMLDivElement>(null)
    const [hoveredId, setHoveredId] = useState<string | null>(null)
    const [tooltip, setTooltip] = useState<{ x: number; y: number; text: string } | null>(null)
    const [ctxMenu, setCtxMenu] = useState<{ x: number; y: number; entities: EntitySnapshot[]; gx: number; gy: number } | null>(null)
    const [zoomIdx, setZoomIdx] = useState(DEFAULT_ZOOM_IDX)
    const [followPlayer, setFollowPlayer] = useState(true)
    const [showZones, setShowZones] = useState(true)
    const [viewportSize, setViewportSize] = useState({ w: 0, h: 0 })

    useEffect(() => {
        const container = containerRef.current
        if (!container) return
        const ro = new ResizeObserver(([entry]) => {
            setViewportSize({ w: entry.contentRect.width, h: entry.contentRect.height })
        })
        ro.observe(container)
        return () => ro.disconnect()
    }, [])

    const cellSize = ZOOM_LEVELS[zoomIdx]
    const overscrollX = Math.floor(viewportSize.w / 2)
    const overscrollY = Math.floor(viewportSize.h / 2)
    const canvasWidth = mapWidth * cellSize + MAP_PADDING * 2 + overscrollX
    const canvasHeight = mapHeight * cellSize + MAP_PADDING * 2 + overscrollY

    const zoomIn = useCallback(() => setZoomIdx(i => Math.min(i + 1, ZOOM_LEVELS.length - 1)), [])
    const zoomOut = useCallback(() => setZoomIdx(i => Math.max(i - 1, 0)), [])

    const zoomIdxRef = useRef(zoomIdx)
    zoomIdxRef.current = zoomIdx
    const pinchAccum = useRef(0)
    const pendingScroll = useRef<{ left: number; top: number } | null>(null)

    useEffect(() => {
        const container = containerRef.current
        if (!container) return
        const onWheel = (ev: WheelEvent) => {
            if (!(ev.ctrlKey || ev.metaKey)) return
            ev.preventDefault()

            const THRESHOLD = 30
            pinchAccum.current += ev.deltaY
            if (Math.abs(pinchAccum.current) < THRESHOLD) return

            const direction = pinchAccum.current < 0 ? 1 : -1
            pinchAccum.current = 0

            const oldIdx = zoomIdxRef.current
            const newIdx = Math.max(0, Math.min(ZOOM_LEVELS.length - 1, oldIdx + direction))
            if (newIdx === oldIdx) return

            const oldCell = ZOOM_LEVELS[oldIdx]
            const newCell = ZOOM_LEVELS[newIdx]
            const rect = container.getBoundingClientRect()
            const viewX = ev.clientX - rect.left
            const viewY = ev.clientY - rect.top
            const mapFracX = (container.scrollLeft + viewX - MAP_PADDING) / oldCell
            const mapFracY = (container.scrollTop + viewY - MAP_PADDING) / oldCell

            pendingScroll.current = {
                left: MAP_PADDING + mapFracX * newCell - viewX,
                top: MAP_PADDING + mapFracY * newCell - viewY,
            }
            setZoomIdx(newIdx)
        }
        container.addEventListener("wheel", onWheel, { passive: false })
        return () => container.removeEventListener("wheel", onWheel)
    }, [])

    useLayoutEffect(() => {
        if (!pendingScroll.current || !containerRef.current) return
        containerRef.current.scrollLeft = pendingScroll.current.left
        containerRef.current.scrollTop = pendingScroll.current.top
        pendingScroll.current = null
    }, [zoomIdx])

    const scrollToPlayer = useCallback((smooth: boolean) => {
        if (!player || !containerRef.current) return
        const px = MAP_PADDING + player.x * cellSize + cellSize / 2
        const py = MAP_PADDING + (mapHeight - 1 - player.y) * cellSize + cellSize / 2
        const container = containerRef.current
        container.scrollTo({
            left: px - container.clientWidth / 2,
            top: py - container.clientHeight / 2,
            behavior: smooth ? "smooth" : "instant",
        })
    }, [player, cellSize, mapHeight])

    const isUserScrolling = useRef(false)
    useEffect(() => {
        const container = containerRef.current
        if (!container) return
        let scrollTimeout: ReturnType<typeof setTimeout>
        const onScroll = () => {
            if (isUserScrolling.current && followPlayer) {
                setFollowPlayer(false)
            }
            isUserScrolling.current = true
            clearTimeout(scrollTimeout)
            scrollTimeout = setTimeout(() => { isUserScrolling.current = false }, 100)
        }
        container.addEventListener("scroll", onScroll, { passive: true })
        return () => { container.removeEventListener("scroll", onScroll); clearTimeout(scrollTimeout) }
    }, [followPlayer])

    useEffect(() => {
        if (followPlayer && player) {
            isUserScrolling.current = false
            scrollToPlayer(false)
        }
    }, [followPlayer, player?.x, player?.y, scrollToPlayer])

    const toCanvas = useCallback((ev: React.MouseEvent<HTMLCanvasElement>): { x: number; y: number } | null => {
        const rect = canvasRef.current?.getBoundingClientRect()
        if (!rect) return null
        return { x: ev.clientX - rect.left, y: ev.clientY - rect.top }
    }, [])

    const canvasToGrid = useCallback((cx: number, cy: number): { gx: number; gy: number } => {
        const gx = Math.floor((cx - MAP_PADDING) / cellSize)
        const gy = mapHeight - 1 - Math.floor((cy - MAP_PADDING) / cellSize)
        return { gx, gy }
    }, [cellSize, mapHeight])

    const entitiesAt = useCallback((cx: number, cy: number): EntitySnapshot[] => {
        const hitR = Math.max(5, cellSize * 0.4)
        const hits: EntitySnapshot[] = []
        for (const e of entities) {
            const ex = MAP_PADDING + e.x * cellSize + cellSize / 2
            const ey = MAP_PADDING + (mapHeight - 1 - e.y) * cellSize + cellSize / 2
            if (Math.abs(cx - ex) <= hitR && Math.abs(cy - ey) <= hitR) {
                hits.push(e)
            }
        }
        return hits
    }, [entities, mapHeight, cellSize])

    const handleCanvasClick = useCallback((ev: React.MouseEvent<HTMLCanvasElement>) => {
        setCtxMenu(null)
        const pos = toCanvas(ev)
        if (!pos) return
        const hits = entitiesAt(pos.x, pos.y)
        if (hits.length === 1) {
            onSelect(hits[0].id)
        } else if (hits.length > 1) {
            const { gx, gy } = canvasToGrid(pos.x, pos.y)
            const container = containerRef.current
            const menuX = ev.clientX - (container?.getBoundingClientRect().left ?? 0) + (container?.scrollLeft ?? 0)
            const menuY = ev.clientY - (container?.getBoundingClientRect().top ?? 0) + (container?.scrollTop ?? 0)
            setCtxMenu({ x: menuX, y: menuY, entities: hits, gx, gy })
        }
    }, [entitiesAt, toCanvas, canvasToGrid, onSelect])

    const handleContextMenu = useCallback((ev: React.MouseEvent<HTMLCanvasElement>) => {
        ev.preventDefault()
        const pos = toCanvas(ev)
        if (!pos) return
        const hits = entitiesAt(pos.x, pos.y)
        const { gx, gy } = canvasToGrid(pos.x, pos.y)
        const container = containerRef.current
        const menuX = ev.clientX - (container?.getBoundingClientRect().left ?? 0) + (container?.scrollLeft ?? 0)
        const menuY = ev.clientY - (container?.getBoundingClientRect().top ?? 0) + (container?.scrollTop ?? 0)
        setCtxMenu({ x: menuX, y: menuY, entities: hits, gx, gy })
    }, [entitiesAt, toCanvas, canvasToGrid])

    const handleCanvasMove = useCallback((ev: React.MouseEvent<HTMLCanvasElement>) => {
        const pos = toCanvas(ev)
        if (!pos) return
        const hits = entitiesAt(pos.x, pos.y)
        if (hits.length > 0) {
            setHoveredId(hits[0].id)
            const loc = `(${hits[0].x}, ${hits[0].y})`
            if (hits.length === 1) {
                const typeLabel = hits[0].debug_type ? ` [${hits[0].debug_type}]` : ""
                setTooltip({ x: pos.x, y: pos.y, text: `${hits[0].id}${typeLabel} ${loc}` })
            } else {
                const names = hits.map(e => e.id).join(", ")
                setTooltip({ x: pos.x, y: pos.y, text: `${hits.length} entities ${loc}: ${names}` })
            }
        } else {
            setHoveredId(null)
            const { gx, gy } = canvasToGrid(pos.x, pos.y)
            if (gx >= 0 && gx < mapWidth && gy >= 0 && gy < mapHeight) {
                setTooltip({ x: pos.x, y: pos.y, text: `(${gx}, ${gy})` })
            } else {
                setTooltip(null)
            }
        }
    }, [entitiesAt, toCanvas, canvasToGrid, mapWidth, mapHeight])

    useEffect(() => {
        const canvas = canvasRef.current
        if (!canvas) return
        const ctx = canvas.getContext("2d")
        if (!ctx) return

        ctx.clearRect(0, 0, canvasWidth, canvasHeight)

        // Map background
        ctx.fillStyle = "#1a1a2e"
        ctx.fillRect(MAP_PADDING, MAP_PADDING, mapWidth * cellSize, mapHeight * cellSize)

        // Grid lines (only when zoomed in enough)
        if (cellSize >= 8) {
            ctx.strokeStyle = "#ffffff0a"
            ctx.lineWidth = 1
            for (let x = 0; x <= mapWidth; x++) {
                const px = MAP_PADDING + x * cellSize + 0.5
                ctx.beginPath()
                ctx.moveTo(px, MAP_PADDING)
                ctx.lineTo(px, MAP_PADDING + mapHeight * cellSize)
                ctx.stroke()
            }
            for (let y = 0; y <= mapHeight; y++) {
                const py = MAP_PADDING + y * cellSize + 0.5
                ctx.beginPath()
                ctx.moveTo(MAP_PADDING, py)
                ctx.lineTo(MAP_PADDING + mapWidth * cellSize, py)
                ctx.stroke()
            }
        }

        // Zone overlays
        if (showZones && zones.length > 0) {
            const zoneColors = ["#06b6d4", "#8b5cf6", "#f59e0b", "#10b981", "#ef4444", "#ec4899", "#6366f1", "#14b8a6"]
            const zoneIds = [...new Set(zones.map(z => z.id))].sort()
            for (const z of zones) {
                const colorIdx = zoneIds.indexOf(z.id) % zoneColors.length
                const base = zoneColors[colorIdx]
                const rx = MAP_PADDING + z.x * cellSize
                const ry = MAP_PADDING + (mapHeight - 1 - z.y - z.h) * cellSize
                const rw = (z.w + 1) * cellSize
                const rh = (z.h + 1) * cellSize
                ctx.fillStyle = base + "18"
                ctx.fillRect(rx, ry, rw, rh)
                ctx.strokeStyle = base + "80"
                ctx.lineWidth = 1.5
                ctx.strokeRect(rx + 0.5, ry + 0.5, rw - 1, rh - 1)
                if (cellSize >= 6) {
                    const fontSize = Math.max(8, Math.min(11, cellSize * 0.8))
                    ctx.font = `${fontSize}px monospace`
                    ctx.fillStyle = base + "cc"
                    ctx.textAlign = "left"
                    ctx.fillText(z.id, rx + 3, ry + fontSize + 2)
                }
            }
        }

        // Teleport markers
        for (const tp of teleports) {
            const tpX = MAP_PADDING + tp.x * cellSize
            const tpY = MAP_PADDING + (mapHeight - 1 - tp.y) * cellSize
            ctx.fillStyle = "#8b5cf640"
            ctx.fillRect(tpX + 1, tpY + 1, cellSize - 2, cellSize - 2)
            ctx.strokeStyle = "#8b5cf6"
            ctx.lineWidth = 1
            ctx.strokeRect(tpX + 1.5, tpY + 1.5, cellSize - 3, cellSize - 3)
        }

        // Entity drawing helper
        const drawEntity = (e: EntitySnapshot) => {
            const px = MAP_PADDING + e.x * cellSize + cellSize / 2
            const py = MAP_PADDING + (mapHeight - 1 - e.y) * cellSize + cellSize / 2
            const isSelected = e.id === selectedId
            const isHovered = e.id === hoveredId
            const isPlayerEntity = !!e.is_player
            const style = ENTITY_TYPE_STYLES[e.debug_type || ""] || DEFAULT_STYLE
            const r = isPlayerEntity ? Math.max(4, cellSize * 0.35) : Math.max(3, cellSize * 0.25)
            const drawR = isSelected || isHovered ? r + 1 : r
            const fillColor = isSelected ? style.selectedColor : style.color

            if (isPlayerEntity) {
                ctx.beginPath()
                ctx.arc(px, py, r + 4, 0, Math.PI * 2)
                ctx.fillStyle = style.color + "20"
                ctx.fill()
            }

            if (style.shape === "square") {
                ctx.fillStyle = fillColor
                ctx.fillRect(px - drawR, py - drawR, drawR * 2, drawR * 2)
                if (isSelected) {
                    ctx.strokeStyle = "#ffffff"
                    ctx.lineWidth = 1
                    ctx.strokeRect(px - drawR, py - drawR, drawR * 2, drawR * 2)
                }
            } else {
                ctx.beginPath()
                ctx.arc(px, py, drawR, 0, Math.PI * 2)
                ctx.fillStyle = fillColor
                ctx.fill()
                if (isSelected || isPlayerEntity) {
                    ctx.strokeStyle = "#ffffff"
                    ctx.lineWidth = isPlayerEntity ? 1.5 : 1
                    ctx.stroke()
                }
            }

            if (style.showDirection) {
                const arrLen = Math.max(6, cellSize * 0.6)
                const headSize = Math.max(3, cellSize * 0.25)
                const arrColor = isPlayerEntity ? "#93c5fd" : isSelected ? style.selectedColor : style.color + "cc"
                drawArrow(ctx, px, py, e.direction, arrLen, headSize, arrColor, isPlayerEntity ? 2 : 1.5)
            }
        }

        // Non-player entities (collision first/behind, then others)
        for (const e of entities) {
            if (e.is_player || e.debug_type !== "collision") continue
            drawEntity(e)
        }
        for (const e of entities) {
            if (e.is_player || e.debug_type === "collision") continue
            drawEntity(e)
        }

        // Player on top
        if (player) drawEntity(player)

        // Axis labels
        ctx.fillStyle = "#a1a1aa"
        const fontSize = Math.max(9, Math.min(11, cellSize))
        ctx.font = `${fontSize}px monospace`
        ctx.textAlign = "center"
        const xStep = Math.max(1, Math.ceil(40 / cellSize))
        for (let x = 0; x < mapWidth; x += xStep) {
            ctx.fillText(String(x), MAP_PADDING + x * cellSize + cellSize / 2, MAP_PADDING - 6)
        }
        ctx.textAlign = "right"
        const yStep = Math.max(1, Math.ceil(40 / cellSize))
        for (let y = 0; y < mapHeight; y += yStep) {
            ctx.fillText(String(y), MAP_PADDING - 6, MAP_PADDING + (mapHeight - 1 - y) * cellSize + cellSize / 2 + fontSize * 0.35)
        }

    }, [entities, teleports, zones, showZones, player, mapWidth, mapHeight, selectedId, hoveredId, canvasWidth, canvasHeight, cellSize])

    if (mapWidth === 0 || mapHeight === 0) {
        return <div className="text-sm text-muted-foreground p-4">No map data available.</div>
    }

    return (
        <div className="flex-1 rounded-md border bg-canvas relative">
            <div className="absolute inset-0 overflow-auto" ref={containerRef}>
                <div className="relative" style={{ width: canvasWidth, height: canvasHeight }}>
                    <canvas
                        ref={canvasRef}
                        width={canvasWidth}
                        height={canvasHeight}
                        className="cursor-crosshair"
                        onClick={handleCanvasClick}
                        onContextMenu={handleContextMenu}
                        onMouseMove={handleCanvasMove}
                        onMouseLeave={() => { setHoveredId(null); setTooltip(null) }}
                    />
                    {tooltip && !ctxMenu && (
                        <div
                            className="absolute pointer-events-none bg-popover text-popover-foreground border border-border rounded px-2 py-1 text-xs font-mono shadow-md"
                            style={{ left: tooltip.x + 12, top: tooltip.y - 8 }}
                        >
                            {tooltip.text}
                        </div>
                    )}
                    {ctxMenu && (<>
                        <div className="fixed inset-0 z-20" onClick={() => setCtxMenu(null)} onContextMenu={e => { e.preventDefault(); setCtxMenu(null) }} />
                        <div
                            className="absolute z-30 bg-popover border border-border rounded shadow-lg py-1 text-xs min-w-[160px]"
                            style={{ left: ctxMenu.x, top: ctxMenu.y }}
                        >
                            <div className="px-3 py-1 text-muted-foreground font-mono text-[10px]">({ctxMenu.gx}, {ctxMenu.gy})</div>
                            {ctxMenu.entities.length > 0 && (
                                <>
                                    <div className="border-t border-border my-0.5" />
                                    {ctxMenu.entities.map(e => (
                                        <button
                                            key={e.id}
                                            className="w-full text-left px-3 py-1 text-foreground hover:bg-accent font-mono truncate"
                                            onClick={() => { onSelect(e.id); setCtxMenu(null) }}
                                        >
                                            {e.id}
                                            {e.debug_type && <span className="text-muted-foreground ml-1">[{e.debug_type}]</span>}
                                        </button>
                                    ))}
                                </>
                            )}
                            {ctxMenu.gx >= 0 && ctxMenu.gx < mapWidth && ctxMenu.gy >= 0 && ctxMenu.gy < mapHeight && (
                                <>
                                    <div className="border-t border-border my-0.5" />
                                    <button
                                        className="w-full text-left px-3 py-1 text-foreground hover:bg-accent"
                                        onClick={() => { onTeleport(ctxMenu.gx, ctxMenu.gy); setCtxMenu(null) }}
                                    >
                                        Teleport here
                                    </button>
                                </>
                            )}
                        </div>
                    </>)}
                </div>
            </div>
            <div className="absolute top-2 right-2 flex items-center gap-1 z-20 pointer-events-auto">
                <button
                    className="w-7 h-7 flex items-center justify-center rounded bg-popover border border-border text-foreground hover:bg-accent transition-colors"
                    onClick={zoomOut}
                    disabled={zoomIdx === 0}
                    title="Zoom out"
                >
                    <ZoomOut className="h-3.5 w-3.5" />
                </button>
                <span className="text-[10px] font-mono text-muted-foreground w-8 text-center">{cellSize}px</span>
                <button
                    className="w-7 h-7 flex items-center justify-center rounded bg-popover border border-border text-foreground hover:bg-accent transition-colors"
                    onClick={zoomIn}
                    disabled={zoomIdx === ZOOM_LEVELS.length - 1}
                    title="Zoom in"
                >
                    <ZoomIn className="h-3.5 w-3.5" />
                </button>
                {player && (
                    <button
                        className={cn(
                            "w-7 h-7 flex items-center justify-center rounded border transition-colors",
                            followPlayer
                                ? "bg-accent-blue-tint border-accent-blue-edge text-accent-blue"
                                : "bg-popover border-border text-foreground hover:bg-accent"
                        )}
                        onClick={() => { setFollowPlayer(f => !f); if (!followPlayer) scrollToPlayer(true) }}
                        title={followPlayer ? "Following player (click to unlock)" : "Follow player"}
                    >
                        {followPlayer ? <Lock className="h-3.5 w-3.5" /> : <Crosshair className="h-3.5 w-3.5" />}
                    </button>
                )}
                <button
                    className={cn(
                        "w-7 h-7 flex items-center justify-center rounded border transition-colors",
                        showZones
                            ? "bg-accent-teal-tint border-accent-teal-edge text-accent-teal"
                            : "bg-popover border-border text-foreground hover:bg-accent"
                    )}
                    onClick={() => setShowZones(z => !z)}
                    title={showZones ? "Hide zones" : "Show zones"}
                >
                    <Layers className="h-3.5 w-3.5" />
                </button>
            </div>
            <div className="absolute bottom-2 right-2 flex flex-col gap-1 text-[10px] rounded p-2 border border-border bg-popover/95 text-foreground z-20 pointer-events-auto">
                {Object.entries(ENTITY_TYPE_STYLES).filter(([k]) => k !== "system").map(([type, s]) => (
                    <span key={type} className="flex items-center gap-1.5">
                        <span
                            className={cn("w-2.5 h-2.5 inline-block", s.shape === "square" ? "rounded-sm" : "rounded-full", type === "player" && "border border-white")}
                            style={{ backgroundColor: s.color }}
                        />
                        {type}
                    </span>
                ))}
                <span className="flex items-center gap-1.5">
                    <span className="w-2.5 h-2.5 rounded border border-accent-violet-edge bg-accent-violet-tint inline-block" /> teleport
                </span>
                <span className="text-muted-foreground mt-1">Click to select</span>
                <span className="text-muted-foreground">Right-click for menu</span>
                <span className="text-muted-foreground">Ctrl+scroll to zoom</span>
            </div>
        </div>
    )
}

function EntityDetailPanel({ id, snapshot, onClose }: { id: string; snapshot?: EntitySnapshot; onClose: () => void }) {
    const { data, isLoading, error } = useEntity(id)

    if (isLoading) return <div className="text-sm text-muted-foreground">Loading...</div>
    if (error) return <div className="text-sm text-destructive">{error.message}</div>
    if (!data) return null

    return (
        <div>
            <div className="flex items-center justify-between mb-3">
                <h3 className="font-mono text-sm font-bold">{data.id}</h3>
                <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={onClose}>
                    <X className="h-3 w-3" />
                </Button>
            </div>
            <div className="grid grid-cols-[90px_1fr] gap-x-3 gap-y-1.5 text-xs">
                {snapshot?.debug_type && (<>
                    <Label>Type</Label>
                    <Badge variant="secondary" className="w-fit text-[10px]">{snapshot.debug_type}</Badge>
                </>)}
                <Label>Grid</Label><span className="font-mono">({data.x}, {data.y})</span>
                <Label>Precise</Label><span className="font-mono">({data.precise_x.toFixed(2)}, {data.precise_y.toFixed(2)})</span>
                <Label>Direction</Label><span>{data.direction}</span>
                <Label>Moving</Label><span>{data.is_moving ? "yes" : "no"}</span>
                <Label>Behavior</Label>
                <span>
                    {data.has_behavior ? "yes" : "no"}
                    {!data.behavior_enabled && <Badge variant="destructive" className="ml-1 text-[9px]">disabled</Badge>}
                </span>
                <Label>Presence</Label><span>{data.has_presence ? "yes" : "no"}</span>
                <Label>Renderer</Label><span>{data.has_renderer ? "yes" : "no"}</span>
                <Label>Sound</Label><span>{data.sound_enabled ? "on" : "off"}</span>
                {snapshot?.handler_ref && (<>
                    <Label>Handler</Label>
                    <HandlerLink handlerRef={snapshot.handler_ref} />
                </>)}
            </div>
            {data.metadata != null && (
                <div className="mt-3">
                    <Label>Metadata</Label>
                    <pre className="mt-1 rounded-md bg-muted p-2 text-[10px] font-mono overflow-auto max-h-48">
                        {JSON.stringify(data.metadata, null, 2)}
                    </pre>
                </div>
            )}
        </div>
    )
}

function HandlerLink({ handlerRef }: { handlerRef: string }) {
    const { data: scripts } = useScripts()
    const scriptPath = useMemo(() => {
        if (!scripts) return null
        for (const s of scripts) {
            if (s.handlerNames?.includes(handlerRef)) {
                return s.path.replace(/\.yaml$/, "")
            }
        }
        return null
    }, [scripts, handlerRef])
    return (
        <span className="font-mono text-[10px]">
            {scriptPath ? (
                <Link to={`/scripts/${scriptPath}`} className="text-accent-blue hover:text-accent-blue/80 underline">{handlerRef}</Link>
            ) : handlerRef}
        </span>
    )
}

function Label({ children }: { children: React.ReactNode }) {
    return <span className="text-muted-foreground">{children}</span>
}
