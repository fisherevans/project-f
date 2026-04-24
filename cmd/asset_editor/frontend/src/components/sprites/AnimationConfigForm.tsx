import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { X, Plus, ChevronUp, ChevronDown } from "lucide-react"
import type { SpriteTilesheetAnimation, SpriteTilesheetAnimationTile } from "@/types/sprites"

type AnimType = "h_sequence" | "v_sequence" | "tiles"

interface AnimationConfigFormProps {
    animation: SpriteTilesheetAnimation
    onChange: (anim: SpriteTilesheetAnimation) => void
    totalCols: number
    totalRows: number
    selectedTileIndex?: number
    onSelectTileIndex?: (index: number | undefined) => void
}

function getAnimType(anim: SpriteTilesheetAnimation): AnimType {
    if (anim.h_sequence) return "h_sequence"
    if (anim.v_sequence) return "v_sequence"
    return "tiles"
}

export function AnimationConfigForm({ animation, onChange, totalCols, totalRows, selectedTileIndex, onSelectTileIndex }: AnimationConfigFormProps) {
    const animType = getAnimType(animation)

    const setType = (type: AnimType) => {
        const base: SpriteTilesheetAnimation = {
            framesPerSecond: animation.framesPerSecond,
            jitterPercent: animation.jitterPercent,
            repeat: animation.repeat,
            pingPong: animation.pingPong,
            reverse: animation.reverse,
            randomize: animation.randomize,
        }
        if (type === "h_sequence") {
            onChange({ ...base, h_sequence: { row: 1, fromColumn: 1, toColumn: totalCols } })
        } else if (type === "v_sequence") {
            onChange({ ...base, v_sequence: { column: 1, fromRow: 1, toRow: totalRows } })
        } else {
            onChange({ ...base, tiles: [{ row: 1, column: 1 }] })
        }
    }

    const update = (patch: Partial<SpriteTilesheetAnimation>) => {
        onChange({ ...animation, ...patch })
    }

    const updateBool = (field: "repeat" | "pingPong" | "reverse" | "randomize", value: boolean) => {
        update({ [field]: value || undefined })
    }

    return (
        <div className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
                <div>
                    <Label className="text-xs">Type</Label>
                    <Select value={animType} onValueChange={(v) => setType(v as AnimType)}>
                        <SelectTrigger className="mt-1">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="h_sequence">Horizontal Sequence</SelectItem>
                            <SelectItem value="v_sequence">Vertical Sequence</SelectItem>
                            <SelectItem value="tiles">Explicit Tiles</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
                <div>
                    <Label className="text-xs">Frames Per Second</Label>
                    <Input
                        type="number"
                        className="mt-1"
                        min={0.1}
                        step={0.5}
                        value={animation.framesPerSecond}
                        onChange={(e) => update({ framesPerSecond: parseFloat(e.target.value) || 1 })}
                    />
                </div>
            </div>

            {animType === "h_sequence" && animation.h_sequence && (
                <HSequenceFields
                    seq={animation.h_sequence}
                    totalCols={totalCols}
                    onChange={(h_sequence) => onChange({ ...animation, h_sequence })}
                />
            )}

            {animType === "v_sequence" && animation.v_sequence && (
                <VSequenceFields
                    seq={animation.v_sequence}
                    totalRows={totalRows}
                    onChange={(v_sequence) => onChange({ ...animation, v_sequence })}
                />
            )}

            {animType === "tiles" && (
                <TilesFields
                    tiles={animation.tiles ?? []}
                    onChange={(tiles) => onChange({ ...animation, tiles })}
                    selectedIndex={selectedTileIndex}
                    onSelectIndex={onSelectTileIndex}
                />
            )}

            <div className="grid grid-cols-2 gap-3">
                <div>
                    <Label className="text-xs">Jitter %</Label>
                    <Input
                        type="number"
                        className="mt-1"
                        min={0}
                        max={1}
                        step={0.05}
                        value={animation.jitterPercent ?? 0}
                        onChange={(e) => update({ jitterPercent: parseFloat(e.target.value) || undefined })}
                    />
                </div>
            </div>

            <div className="flex flex-wrap gap-x-6 gap-y-2">
                <div className="flex items-center gap-2">
                    <Switch
                        checked={animation.repeat ?? true}
                        onCheckedChange={(v) => updateBool("repeat", v)}
                    />
                    <Label className="text-xs">Repeat</Label>
                </div>
                <div className="flex items-center gap-2">
                    <Switch
                        checked={!!animation.pingPong}
                        onCheckedChange={(v) => updateBool("pingPong", v)}
                    />
                    <Label className="text-xs">Ping-Pong</Label>
                </div>
                <div className="flex items-center gap-2">
                    <Switch
                        checked={!!animation.reverse}
                        onCheckedChange={(v) => updateBool("reverse", v)}
                    />
                    <Label className="text-xs">Reverse</Label>
                </div>
                <div className="flex items-center gap-2">
                    <Switch
                        checked={!!animation.randomize}
                        onCheckedChange={(v) => updateBool("randomize", v)}
                    />
                    <Label className="text-xs">Randomize</Label>
                </div>
            </div>
        </div>
    )
}

function HSequenceFields({
    seq,
    totalCols,
    onChange,
}: {
    seq: NonNullable<SpriteTilesheetAnimation["h_sequence"]>
    totalCols: number
    onChange: (seq: NonNullable<SpriteTilesheetAnimation["h_sequence"]>) => void
}) {
    return (
        <div className="grid grid-cols-3 gap-3">
            <div>
                <Label className="text-xs">Row</Label>
                <Input
                    type="number"
                    className="mt-1"
                    min={1}
                    value={seq.row}
                    onChange={(e) => onChange({ ...seq, row: parseInt(e.target.value) || 1 })}
                />
            </div>
            <div>
                <Label className="text-xs">From Column</Label>
                <Input
                    type="number"
                    className="mt-1"
                    min={1}
                    max={totalCols}
                    value={seq.fromColumn ?? 1}
                    onChange={(e) => onChange({ ...seq, fromColumn: parseInt(e.target.value) || 1 })}
                />
            </div>
            <div>
                <Label className="text-xs">To Column</Label>
                <Input
                    type="number"
                    className="mt-1"
                    min={1}
                    max={totalCols}
                    value={seq.toColumn ?? totalCols}
                    onChange={(e) => onChange({ ...seq, toColumn: parseInt(e.target.value) || totalCols })}
                />
            </div>
        </div>
    )
}

function VSequenceFields({
    seq,
    totalRows,
    onChange,
}: {
    seq: NonNullable<SpriteTilesheetAnimation["v_sequence"]>
    totalRows: number
    onChange: (seq: NonNullable<SpriteTilesheetAnimation["v_sequence"]>) => void
}) {
    return (
        <div className="grid grid-cols-3 gap-3">
            <div>
                <Label className="text-xs">Column</Label>
                <Input
                    type="number"
                    className="mt-1"
                    min={1}
                    value={seq.column}
                    onChange={(e) => onChange({ ...seq, column: parseInt(e.target.value) || 1 })}
                />
            </div>
            <div>
                <Label className="text-xs">From Row</Label>
                <Input
                    type="number"
                    className="mt-1"
                    min={1}
                    max={totalRows}
                    value={seq.fromRow ?? 1}
                    onChange={(e) => onChange({ ...seq, fromRow: parseInt(e.target.value) || 1 })}
                />
            </div>
            <div>
                <Label className="text-xs">To Row</Label>
                <Input
                    type="number"
                    className="mt-1"
                    min={1}
                    max={totalRows}
                    value={seq.toRow ?? totalRows}
                    onChange={(e) => onChange({ ...seq, toRow: parseInt(e.target.value) || totalRows })}
                />
            </div>
        </div>
    )
}

function TilesFields({
    tiles,
    onChange,
    selectedIndex,
    onSelectIndex,
}: {
    tiles: SpriteTilesheetAnimationTile[]
    onChange: (tiles: SpriteTilesheetAnimationTile[]) => void
    selectedIndex?: number
    onSelectIndex?: (index: number | undefined) => void
}) {
    const addTile = () => {
        onChange([...tiles, { row: 1, column: 1 }])
        onSelectIndex?.(tiles.length)
    }

    const removeTile = (index: number) => {
        onChange(tiles.filter((_, i) => i !== index))
        if (selectedIndex === index) onSelectIndex?.(undefined)
        else if (selectedIndex !== undefined && selectedIndex > index) onSelectIndex?.(selectedIndex - 1)
    }

    const moveTile = (from: number, to: number) => {
        if (to < 0 || to >= tiles.length) return
        const next = [...tiles]
        const [item] = next.splice(from, 1)
        next.splice(to, 0, item)
        onChange(next)
        if (selectedIndex === from) onSelectIndex?.(to)
    }

    return (
        <div className="space-y-2">
            <div className="flex items-center justify-between">
                <Label className="text-xs">Tiles ({tiles.length})</Label>
                <Button variant="outline" size="sm" onClick={addTile}>
                    <Plus className="h-3 w-3 mr-1" /> Add
                </Button>
            </div>
            {selectedIndex !== undefined && (
                <p className="text-xs text-blue-400">
                    Tile {selectedIndex + 1} selected - click grid to reassign
                </p>
            )}
            <div className="space-y-0.5 max-h-60 overflow-y-auto">
                {tiles.map((tile, i) => {
                    const isSelected = selectedIndex === i
                    return (
                        <div
                            key={i}
                            className={`flex items-center gap-1 text-xs rounded px-1 py-0.5 cursor-pointer ${isSelected ? "bg-blue-500/20 ring-1 ring-blue-500/40" : "hover:bg-muted"}`}
                            onClick={() => onSelectIndex?.(isSelected ? undefined : i)}
                        >
                            <span className="text-muted-foreground w-5 text-right shrink-0">{i + 1}.</span>
                            <span className="text-muted-foreground">r</span>
                            <span>{tile.row}</span>
                            <span className="text-muted-foreground ml-1">c</span>
                            <span>{tile.column}</span>
                            <div className="flex items-center gap-0 ml-auto">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-5 w-5 p-0"
                                    disabled={i === 0}
                                    onClick={(e) => { e.stopPropagation(); moveTile(i, i - 1) }}
                                >
                                    <ChevronUp className="h-3 w-3" />
                                </Button>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-5 w-5 p-0"
                                    disabled={i === tiles.length - 1}
                                    onClick={(e) => { e.stopPropagation(); moveTile(i, i + 1) }}
                                >
                                    <ChevronDown className="h-3 w-3" />
                                </Button>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-5 w-5 p-0 text-destructive"
                                    onClick={(e) => { e.stopPropagation(); removeTile(i) }}
                                >
                                    <X className="h-3 w-3" />
                                </Button>
                            </div>
                        </div>
                    )
                })}
            </div>
        </div>
    )
}
