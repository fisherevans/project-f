import { useState } from "react"
import { Plus, Trash2, Pencil, Check, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { AnimationConfigForm } from "./AnimationConfigForm"
import { AnimationPreview } from "./AnimationPreview"
import { TilesheetViewer } from "./TilesheetViewer"
import type { SpriteTilesheetAnimation, SpriteTilesheet, TilesheetCoordinates } from "@/types/sprites"

interface AnimationModalProps {
    animations: Record<string, SpriteTilesheetAnimation>
    onChange: (animations: Record<string, SpriteTilesheetAnimation>) => void
    imageSrc: string
    tileWidth: number
    tileHeight: number
    imageWidth: number
    imageHeight: number
    totalCols: number
    totalRows: number
    bgClass?: string
    sprites?: Record<string, TilesheetCoordinates>
    tilesheet: SpriteTilesheet
    onTilesheetChange: (ts: SpriteTilesheet) => void
}

export function AnimationModal({
    animations,
    onChange,
    imageSrc,
    tileWidth,
    tileHeight,
    imageWidth,
    imageHeight,
    totalCols,
    totalRows,
    bgClass,
    sprites,
    tilesheet,
    onTilesheetChange,
}: AnimationModalProps) {
    const [open, setOpen] = useState(false)
    const [selectedAnimation, setSelectedAnimationRaw] = useState<string | undefined>()
    const [selectedTileIndex, setSelectedTileIndex] = useState<number | undefined>()

    const setSelectedAnimation = (name: string | undefined) => {
        setSelectedAnimationRaw(name)
        setSelectedTileIndex(undefined)
    }
    const [newName, setNewName] = useState("")
    const [renamingKey, setRenamingKey] = useState<string | null>(null)
    const [renameValue, setRenameValue] = useState("")

    const names = Object.keys(animations)
    const selected = selectedAnimation ? animations[selectedAnimation] : undefined

    const addAnimation = () => {
        const name = newName.trim() || `anim_${names.length + 1}`
        if (animations[name]) return
        onChange({
            ...animations,
            [name]: {
                framesPerSecond: 8,
                h_sequence: { row: 1, fromColumn: 1, toColumn: totalCols },
            },
        })
        setNewName("")
        setSelectedAnimation(name)
    }

    const deleteAnimation = (name: string) => {
        const next = { ...animations }
        delete next[name]
        onChange(next)
        if (selectedAnimation === name) setSelectedAnimation(undefined)
    }

    const startRename = (name: string) => {
        setRenamingKey(name)
        setRenameValue(name)
    }

    const confirmRename = () => {
        if (!renamingKey || !renameValue.trim()) return
        const trimmed = renameValue.trim()
        if (trimmed === renamingKey) { setRenamingKey(null); return }
        if (animations[trimmed]) return
        const next: Record<string, SpriteTilesheetAnimation> = {}
        for (const [k, v] of Object.entries(animations)) {
            next[k === renamingKey ? trimmed : k] = v
        }
        onChange(next)
        if (selectedAnimation === renamingKey) setSelectedAnimation(trimmed)
        setRenamingKey(null)
    }

    const updateAnimation = (name: string, anim: SpriteTilesheetAnimation) => {
        onChange({ ...animations, [name]: anim })
    }

    const handleTileClick = (col: number, row: number) => {
        if (!selectedAnimation || !selected) return
        if (!selected.tiles) return
        const tiles = [...(selected.tiles ?? [])]
        if (selectedTileIndex !== undefined && selectedTileIndex < tiles.length) {
            tiles[selectedTileIndex] = { ...tiles[selectedTileIndex], row, column: col }
        } else {
            tiles.push({ row, column: col })
            setSelectedTileIndex(tiles.length - 1)
        }
        updateAnimation(selectedAnimation, { ...selected, tiles })
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={
                <Button variant="outline" size="sm">
                    Animations ({names.length})
                </Button>
            } />
            <DialogContent
                className="!max-w-[90vw] !w-[90vw] !max-h-[85vh] flex flex-col"
                showCloseButton
            >
                <DialogHeader>
                    <DialogTitle>Animations</DialogTitle>
                </DialogHeader>

                <div className="flex items-center gap-2 flex-wrap">
                    {names.map((name) => (
                        <div key={name} className="flex items-center gap-0.5">
                            {renamingKey === name ? (
                                <div className="flex items-center gap-1">
                                    <Input
                                        className="h-7 w-28 text-xs"
                                        value={renameValue}
                                        onChange={(e) => setRenameValue(e.target.value)}
                                        onKeyDown={(e) => {
                                            if (e.key === "Enter") confirmRename()
                                            if (e.key === "Escape") setRenamingKey(null)
                                        }}
                                        autoFocus
                                    />
                                    <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={confirmRename}>
                                        <Check className="h-3 w-3" />
                                    </Button>
                                    <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={() => setRenamingKey(null)}>
                                        <X className="h-3 w-3" />
                                    </Button>
                                </div>
                            ) : (
                                <>
                                    <Button
                                        variant={selectedAnimation === name ? "default" : "outline"}
                                        size="sm"
                                        onClick={() => setSelectedAnimation(selectedAnimation === name ? undefined : name)}
                                    >
                                        {name}
                                    </Button>
                                    <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={() => startRename(name)}>
                                        <Pencil className="h-3 w-3" />
                                    </Button>
                                    <Button variant="ghost" size="sm" className="h-7 w-7 p-0 text-destructive" onClick={() => deleteAnimation(name)}>
                                        <Trash2 className="h-3 w-3" />
                                    </Button>
                                </>
                            )}
                        </div>
                    ))}
                    <div className="flex items-center gap-1 ml-auto">
                        <Input
                            className="h-7 w-28 text-xs"
                            placeholder="name"
                            value={newName}
                            onChange={(e) => setNewName(e.target.value)}
                            onKeyDown={(e) => e.key === "Enter" && addAnimation()}
                        />
                        <Button variant="outline" size="sm" className="h-7" onClick={addAnimation}>
                            <Plus className="h-3 w-3 mr-1" /> Add
                        </Button>
                    </div>
                </div>

                {selectedAnimation && selected ? (
                    <>
                        <Separator />
                        <div className="flex gap-4 flex-1 min-h-0 overflow-hidden">
                            <div className="w-72 shrink-0 overflow-y-auto space-y-4 pr-2">
                                <AnimationConfigForm
                                    animation={selected}
                                    onChange={(anim) => updateAnimation(selectedAnimation, anim)}
                                    totalCols={totalCols}
                                    totalRows={totalRows}
                                    selectedTileIndex={selectedTileIndex}
                                    onSelectTileIndex={setSelectedTileIndex}
                                />
                                <Separator />
                                <AnimationPreview
                                    imageSrc={imageSrc}
                                    tileWidth={tileWidth}
                                    tileHeight={tileHeight}
                                    totalCols={totalCols}
                                    totalRows={totalRows}
                                    animation={selected}
                                    bgClass={bgClass}
                                />
                            </div>

                            <Separator orientation="vertical" className="h-auto" />

                            <div className="flex-1 overflow-auto min-w-0">
                                <TilesheetViewer
                                    imageSrc={imageSrc}
                                    imageWidth={imageWidth}
                                    imageHeight={imageHeight}
                                    tileWidth={tileWidth}
                                    tileHeight={tileHeight}
                                    sprites={sprites}
                                    animations={animations}
                                    selectedAnimation={selectedAnimation}
                                    onTileClick={handleTileClick}
                                    bgClass={bgClass}
                                    tilesheet={tilesheet}
                                    onTilesheetChange={onTilesheetChange}
                                />
                            </div>
                        </div>
                    </>
                ) : (
                    <div className="flex-1 flex items-center justify-center text-muted-foreground text-sm">
                        {names.length === 0
                            ? "No animations. Add one to get started."
                            : "Select an animation to edit."}
                    </div>
                )}
            </DialogContent>
        </Dialog>
    )
}
