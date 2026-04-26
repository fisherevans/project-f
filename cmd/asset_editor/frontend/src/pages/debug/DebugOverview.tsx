import { useState } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { useGameState, useLoadMap, useTeleport, useSetGlobal } from "@/api/debug"
import { usePageTitle } from "@/hooks/usePageTitle"

export function DebugOverview() {
    const { data: state, isLoading, error } = useGameState()
    const loadMap = useLoadMap()
    const teleport = useTeleport()
    const setGlobal = useSetGlobal()

    usePageTitle("Overview - Debug")

    const [mapName, setMapName] = useState("")
    const [waypoint, setWaypoint] = useState("")
    const [tpTarget, setTpTarget] = useState("")
    const [elythium, setElythium] = useState("")

    if (isLoading) {
        return <div className="p-6 text-muted-foreground">Connecting to game...</div>
    }
    if (error) {
        return (
            <div className="p-6">
                <div className="rounded-md border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
                    Cannot reach debug API. Make sure the game is running in development mode.
                </div>
            </div>
        )
    }
    if (!state) return null

    return (
        <div className="flex h-full flex-col gap-6 overflow-auto p-6">
            <div>
                <h2 className="mb-3 text-sm font-medium text-muted-foreground uppercase tracking-wider">Game State</h2>
                <div className="rounded-md border bg-card p-4 grid grid-cols-[120px_1fr] gap-x-4 gap-y-2 text-sm">
                    <span className="text-muted-foreground">State</span>
                    <span className="font-mono text-xs">{state.type}</span>
                    <span className="text-muted-foreground">Capabilities</span>
                    <div className="flex flex-wrap gap-1">
                        {state.capabilities.map(c => (
                            <Badge key={c} variant="secondary" className="text-xs">{c}</Badge>
                        ))}
                    </div>
                    {state.map_name && (
                        <>
                            <span className="text-muted-foreground">Map</span>
                            <span>{state.map_name}</span>
                        </>
                    )}
                    {state.player_position && (
                        <>
                            <span className="text-muted-foreground">Player Pos</span>
                            <span className="font-mono text-xs">
                                ({state.player_position.x.toFixed(1)}, {state.player_position.y.toFixed(1)})
                            </span>
                        </>
                    )}
                </div>
            </div>

            <div>
                <h2 className="mb-3 text-sm font-medium text-muted-foreground uppercase tracking-wider">Quick Actions</h2>
                <div className="flex flex-col gap-3">
                    <div className="flex items-center gap-2">
                        <span className="w-24 text-sm text-muted-foreground shrink-0">Load Map</span>
                        <Input
                            placeholder="map name"
                            value={mapName}
                            onChange={e => setMapName(e.target.value)}
                            className="w-40"
                        />
                        <Input
                            placeholder="waypoint"
                            value={waypoint}
                            onChange={e => setWaypoint(e.target.value)}
                            className="w-32"
                        />
                        <Button
                            size="sm"
                            onClick={() => { if (mapName) loadMap.mutate({ name: mapName, waypoint: waypoint || undefined }) }}
                            disabled={loadMap.isPending || !mapName}
                        >
                            Go
                        </Button>
                    </div>

                    <div className="flex items-center gap-2">
                        <span className="w-24 text-sm text-muted-foreground shrink-0">Teleport</span>
                        <Input
                            placeholder="entity id"
                            value={tpTarget}
                            onChange={e => setTpTarget(e.target.value)}
                            className="w-52"
                        />
                        <Button
                            size="sm"
                            onClick={() => { if (tpTarget) teleport.mutate({ target: tpTarget }) }}
                            disabled={teleport.isPending || !tpTarget}
                        >
                            Teleport
                        </Button>
                    </div>

                    <div className="flex items-center gap-2">
                        <span className="w-24 text-sm text-muted-foreground shrink-0">Elythium</span>
                        <Input
                            type="number"
                            placeholder="amount"
                            value={elythium}
                            onChange={e => setElythium(e.target.value)}
                            className="w-28"
                        />
                        <Button
                            size="sm"
                            onClick={() => {
                                const n = parseInt(elythium)
                                if (!isNaN(n)) setGlobal.mutate({ key: "elythium", type: "int", value: n })
                            }}
                            disabled={setGlobal.isPending}
                        >
                            Set
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    )
}
