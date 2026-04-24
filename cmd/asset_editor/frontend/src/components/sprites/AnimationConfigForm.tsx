import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { X, Plus, GripVertical } from "lucide-react"
import type { SpriteTilesheetAnimation, SpriteTilesheetAnimationTile } from "@/types/sprites"

type AnimType = "h_sequence" | "v_sequence" | "tiles"

interface AnimationConfigFormProps {
    animation: SpriteTilesheetAnimation
    onChange: (anim: SpriteTilesheetAnimation) => void
    totalCols: number
    totalRows: number
}

function getAnimType(anim: SpriteTilesheetAnimation): AnimType {
    if (anim.h_sequence) return "h_sequence"
    if (anim.v_sequence) return "v_sequence"
    return "tiles"
}

export function AnimationConfigForm({ animation, onChange, totalCols, totalRows }: AnimationConfigFormProps) {
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
}: {
    tiles: SpriteTilesheetAnimationTile[]
    onChange: (tiles: SpriteTilesheetAnimationTile[]) => void
}) {
    const addTile = () => {
        onChange([...tiles, { row: 1, column: 1 }])
    }

    const removeTile = (index: number) => {
        onChange(tiles.filter((_, i) => i !== index))
    }

    const updateTile = (index: number, patch: Partial<SpriteTilesheetAnimationTile>) => {
        onChange(tiles.map((t, i) => (i === index ? { ...t, ...patch } : t)))
    }

    const moveTile = (from: number, to: number) => {
        if (to < 0 || to >= tiles.length) return
        const next = [...tiles]
        const [item] = next.splice(from, 1)
        next.splice(to, 0, item)
        onChange(next)
    }

    return (
        <div className="space-y-2">
            <div className="flex items-center justify-between">
                <Label className="text-xs">Tiles ({tiles.length})</Label>
                <Button variant="outline" size="sm" onClick={addTile}>
                    <Plus className="h-3 w-3 mr-1" /> Add
                </Button>
            </div>
            <div className="space-y-1 max-h-48 overflow-y-auto">
                {tiles.map((tile, i) => (
                    <div key={i} className="flex items-center gap-2 text-xs">
                        <button
                            className="cursor-grab text-muted-foreground hover:text-foreground"
                            onMouseDown={(e) => {
                                e.preventDefault()
                                const onMouseUp = () => {
                                    document.removeEventListener("mouseup", onMouseUp)
                                }
                                document.addEventListener("mouseup", onMouseUp)
                            }}
                            onClick={() => {
                                if (i > 0) moveTile(i, i - 1)
                            }}
                        >
                            <GripVertical className="h-3 w-3" />
                        </button>
                        <span className="text-muted-foreground w-4">{i + 1}.</span>
                        <Label className="text-xs">R</Label>
                        <Input
                            type="number"
                            className="w-16 h-7"
                            min={1}
                            value={tile.row}
                            onChange={(e) => updateTile(i, { row: parseInt(e.target.value) || 1 })}
                        />
                        <Label className="text-xs">C</Label>
                        <Input
                            type="number"
                            className="w-16 h-7"
                            min={1}
                            value={tile.column}
                            onChange={(e) => updateTile(i, { column: parseInt(e.target.value) || 1 })}
                        />
                        <Label className="text-xs">W</Label>
                        <Input
                            type="number"
                            className="w-16 h-7"
                            min={0}
                            step={0.1}
                            value={tile.weight ?? 1}
                            onChange={(e) => updateTile(i, { weight: parseFloat(e.target.value) || 1 })}
                        />
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-7 w-7 p-0"
                            onClick={() => removeTile(i)}
                        >
                            <X className="h-3 w-3" />
                        </Button>
                    </div>
                ))}
            </div>
        </div>
    )
}
