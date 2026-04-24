import { useState } from "react"
import { Plus, Trash2, Pencil, Check, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { AnimationConfigForm } from "./AnimationConfigForm"
import { AnimationPreview } from "./AnimationPreview"
import type { SpriteTilesheetAnimation } from "@/types/sprites"

interface AnimationEditorProps {
    animations: Record<string, SpriteTilesheetAnimation>
    onChange: (animations: Record<string, SpriteTilesheetAnimation>) => void
    imageSrc: string
    tileWidth: number
    tileHeight: number
    totalCols: number
    totalRows: number
    selectedAnimation?: string
    onSelectAnimation: (name: string | undefined) => void
    bgClass?: string
}

export function AnimationEditor({
    animations,
    onChange,
    imageSrc,
    tileWidth,
    tileHeight,
    totalCols,
    totalRows,
    selectedAnimation,
    onSelectAnimation,
    bgClass,
}: AnimationEditorProps) {
    const [newName, setNewName] = useState("")
    const [renamingKey, setRenamingKey] = useState<string | null>(null)
    const [renameValue, setRenameValue] = useState("")

    const names = Object.keys(animations)

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
        onSelectAnimation(name)
    }

    const deleteAnimation = (name: string) => {
        const next = { ...animations }
        delete next[name]
        onChange(next)
        if (selectedAnimation === name) onSelectAnimation(undefined)
    }

    const startRename = (name: string) => {
        setRenamingKey(name)
        setRenameValue(name)
    }

    const confirmRename = () => {
        if (!renamingKey || !renameValue.trim()) return
        const trimmed = renameValue.trim()
        if (trimmed === renamingKey) {
            setRenamingKey(null)
            return
        }
        if (animations[trimmed]) return
        const next: Record<string, SpriteTilesheetAnimation> = {}
        for (const [k, v] of Object.entries(animations)) {
            next[k === renamingKey ? trimmed : k] = v
        }
        onChange(next)
        if (selectedAnimation === renamingKey) onSelectAnimation(trimmed)
        setRenamingKey(null)
    }

    const updateAnimation = (name: string, anim: SpriteTilesheetAnimation) => {
        onChange({ ...animations, [name]: anim })
    }

    const selected = selectedAnimation ? animations[selectedAnimation] : undefined

    return (
        <Card>
            <CardHeader className="py-3 px-4">
                <div className="flex items-center justify-between">
                    <CardTitle className="text-sm">Animations</CardTitle>
                    <div className="flex items-center gap-2">
                        <Input
                            className="h-7 w-32 text-xs"
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
            </CardHeader>
            <CardContent className="px-4 pb-4 space-y-3">
                <div className="flex flex-wrap gap-1">
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
                                        onClick={() => onSelectAnimation(selectedAnimation === name ? undefined : name)}
                                    >
                                        {name}
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        className="h-7 w-7 p-0"
                                        onClick={() => startRename(name)}
                                    >
                                        <Pencil className="h-3 w-3" />
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        className="h-7 w-7 p-0 text-destructive"
                                        onClick={() => deleteAnimation(name)}
                                    >
                                        <Trash2 className="h-3 w-3" />
                                    </Button>
                                </>
                            )}
                        </div>
                    ))}
                </div>

                {selectedAnimation && selected && (
                    <>
                        <Separator />
                        <AnimationConfigForm
                            animation={selected}
                            onChange={(anim) => updateAnimation(selectedAnimation, anim)}
                            totalCols={totalCols}
                            totalRows={totalRows}
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
                    </>
                )}
            </CardContent>
        </Card>
    )
}
