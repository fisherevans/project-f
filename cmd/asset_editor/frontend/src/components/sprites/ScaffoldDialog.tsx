import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Plus, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { useScaffoldSprite } from "@/api/sprites"

interface ScaffoldDialogProps {
    directories: string[]
}

export function ScaffoldDialog({ directories }: ScaffoldDialogProps) {
    const [open, setOpen] = useState(false)
    const navigate = useNavigate()
    const scaffold = useScaffoldSprite()

    const [path, setPath] = useState("")
    const [tileWidth, setTileWidth] = useState(16)
    const [tileHeight, setTileHeight] = useState(16)
    const [cols, setCols] = useState(4)
    const [rows, setRows] = useState(4)
    const [spriteNames, setSpriteNames] = useState("")
    const [force, setForce] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const filteredDirs = path
        ? directories.filter((d) => d.startsWith(path) && d !== path).slice(0, 5)
        : []

    const handleSubmit = async () => {
        setError(null)
        const trimmed = path.trim()
        if (!trimmed) {
            setError("Path is required")
            return
        }
        const sprites = spriteNames
            .split(",")
            .map((s) => s.trim())
            .filter(Boolean)

        try {
            await scaffold.mutateAsync({
                name: `assets/sprites/${trimmed}`,
                tileWidth,
                tileHeight,
                cols,
                rows,
                sprites: sprites.length > 0 ? sprites : undefined,
                force,
            })
            setOpen(false)
            setPath("")
            setSpriteNames("")
            navigate(`/sprites/${trimmed}`)
        } catch (e) {
            setError(e instanceof Error ? e.message : "Scaffold failed")
        }
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button variant="outline" size="sm" />}>
                <Plus className="h-3.5 w-3.5 mr-1" /> New Tilesheet
            </DialogTrigger>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>New Tilesheet</DialogTitle>
                </DialogHeader>
                <div className="space-y-4 pt-2">
                    <div>
                        <Label>Path (relative to assets/sprites/)</Label>
                        <Input
                            className="mt-1"
                            placeholder="e.g. ui/icons"
                            value={path}
                            onChange={(e) => setPath(e.target.value)}
                        />
                        {filteredDirs.length > 0 && (
                            <div className="mt-1 border border-border rounded text-xs">
                                {filteredDirs.map((d) => (
                                    <button
                                        key={d}
                                        className="block w-full text-left px-2 py-1 hover:bg-muted"
                                        onClick={() => setPath(d)}
                                    >
                                        {d}
                                    </button>
                                ))}
                            </div>
                        )}
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                        <div>
                            <Label>Tile Width</Label>
                            <Input
                                type="number"
                                className="mt-1"
                                min={1}
                                value={tileWidth}
                                onChange={(e) => setTileWidth(parseInt(e.target.value) || 16)}
                            />
                        </div>
                        <div>
                            <Label>Tile Height</Label>
                            <Input
                                type="number"
                                className="mt-1"
                                min={1}
                                value={tileHeight}
                                onChange={(e) => setTileHeight(parseInt(e.target.value) || 16)}
                            />
                        </div>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                        <div>
                            <Label>Columns</Label>
                            <Input
                                type="number"
                                className="mt-1"
                                min={1}
                                value={cols}
                                onChange={(e) => setCols(parseInt(e.target.value) || 1)}
                            />
                        </div>
                        <div>
                            <Label>Rows</Label>
                            <Input
                                type="number"
                                className="mt-1"
                                min={1}
                                value={rows}
                                onChange={(e) => setRows(parseInt(e.target.value) || 1)}
                            />
                        </div>
                    </div>
                    <div>
                        <Label>Sprite Names (comma-separated, use - to skip)</Label>
                        <Input
                            className="mt-1"
                            placeholder="e.g. icon_a,icon_b,-,icon_d"
                            value={spriteNames}
                            onChange={(e) => setSpriteNames(e.target.value)}
                        />
                        <p className="text-xs text-muted-foreground mt-1">
                            Positional, row-major. {cols * rows} cells total.
                        </p>
                    </div>
                    <div className="flex items-center gap-2">
                        <Switch checked={force} onCheckedChange={setForce} />
                        <Label>Force overwrite existing files</Label>
                    </div>
                    {error && (
                        <div className="flex items-center gap-2 text-sm text-destructive">
                            <X className="h-4 w-4" />
                            {error}
                        </div>
                    )}
                    <Button onClick={handleSubmit} disabled={scaffold.isPending} className="w-full">
                        {scaffold.isPending ? "Creating..." : "Create Tilesheet"}
                    </Button>
                </div>
            </DialogContent>
        </Dialog>
    )
}
