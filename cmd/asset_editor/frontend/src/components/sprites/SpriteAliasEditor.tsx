import { useState } from "react"
import { Plus, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { TilesheetCoordinates } from "@/types/sprites"

interface SpriteAliasEditorProps {
    sprites: Record<string, TilesheetCoordinates>
    onChange: (sprites: Record<string, TilesheetCoordinates>) => void
    pendingTile?: { col: number; row: number } | null
    onClearPendingTile?: () => void
}

export function SpriteAliasEditor({
    sprites,
    onChange,
    pendingTile,
    onClearPendingTile,
}: SpriteAliasEditorProps) {
    const [newName, setNewName] = useState("")

    const entries = Object.entries(sprites)

    const addSprite = () => {
        const name = newName.trim()
        if (!name || sprites[name]) return
        const coords = pendingTile
            ? { row: pendingTile.row, column: pendingTile.col }
            : { row: 1, column: 1 }
        onChange({ ...sprites, [name]: coords })
        setNewName("")
        onClearPendingTile?.()
    }

    const removeSprite = (name: string) => {
        const next = { ...sprites }
        delete next[name]
        onChange(next)
    }

    const updateSprite = (name: string, coords: TilesheetCoordinates) => {
        onChange({ ...sprites, [name]: coords })
    }

    const assignTile = (name: string) => {
        if (pendingTile) {
            updateSprite(name, { row: pendingTile.row, column: pendingTile.col })
            onClearPendingTile?.()
        }
    }

    return (
        <Card>
            <CardHeader className="py-3 px-4">
                <div className="flex items-center justify-between">
                    <CardTitle className="text-sm">Named Sprites</CardTitle>
                    <div className="flex items-center gap-2">
                        <Input
                            className="h-7 w-32 text-xs"
                            placeholder="sprite name"
                            value={newName}
                            onChange={(e) => setNewName(e.target.value)}
                            onKeyDown={(e) => e.key === "Enter" && addSprite()}
                        />
                        <Button variant="outline" size="sm" className="h-7" onClick={addSprite}>
                            <Plus className="h-3 w-3 mr-1" /> Add
                        </Button>
                    </div>
                </div>
                {pendingTile && (
                    <p className="text-xs text-accent-blue mt-1">
                        Click a sprite name to assign tile r{pendingTile.row} c{pendingTile.col}
                    </p>
                )}
            </CardHeader>
            <CardContent className="px-4 pb-3">
                <div className="space-y-1">
                    {entries.map(([name, coords]) => (
                        <div key={name} className="flex items-center gap-2 text-xs">
                            <button
                                className={`font-medium min-w-[80px] text-left ${pendingTile ? "text-accent-blue hover:underline cursor-pointer" : ""}`}
                                onClick={() => assignTile(name)}
                                disabled={!pendingTile}
                            >
                                {name}
                            </button>
                            <Label className="text-xs text-muted-foreground">R</Label>
                            <Input
                                type="number"
                                className="w-16 h-7"
                                min={1}
                                value={coords.row}
                                onChange={(e) =>
                                    updateSprite(name, { ...coords, row: parseInt(e.target.value) || 1 })
                                }
                            />
                            <Label className="text-xs text-muted-foreground">C</Label>
                            <Input
                                type="number"
                                className="w-16 h-7"
                                min={1}
                                value={coords.column}
                                onChange={(e) =>
                                    updateSprite(name, { ...coords, column: parseInt(e.target.value) || 1 })
                                }
                            />
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-7 w-7 p-0 text-destructive"
                                onClick={() => removeSprite(name)}
                            >
                                <Trash2 className="h-3 w-3" />
                            </Button>
                        </div>
                    ))}
                    {entries.length === 0 && (
                        <p className="text-xs text-muted-foreground py-2">
                            No named sprites. Click a tile then add a name.
                        </p>
                    )}
                </div>
            </CardContent>
        </Card>
    )
}
