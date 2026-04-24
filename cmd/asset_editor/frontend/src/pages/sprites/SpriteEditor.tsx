import { useCallback, useEffect, useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { Save, Trash2, Copy, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useSprite, useSaveSprite, useDeleteSpriteSidecar } from "@/api/sprites"
import { apiImageUrl } from "@/api/client"
import { TilesheetViewer } from "@/components/sprites/TilesheetViewer"
import { SpriteAliasEditor } from "@/components/sprites/SpriteAliasEditor"

import { AnimationModal } from "@/components/sprites/AnimationModal"
import { FrameEditor } from "@/components/sprites/FrameEditor"
import type { SpriteMetadata } from "@/types/sprites"

type SpriteType = "plain" | "tilesheet" | "frame" | "nonAtlas"

const BG_OPTIONS = [
    { value: "checker", label: "Checker", style: "bg-[repeating-conic-gradient(#808080_0%_25%,transparent_0%_50%)] bg-[length:16px_16px]" },
    { value: "black", label: "Black", style: "bg-black" },
    { value: "white", label: "White", style: "bg-white" },
    { value: "magenta", label: "Magenta", style: "bg-fuchsia-600" },
] as const

function detectType(meta: SpriteMetadata | undefined): SpriteType {
    if (!meta) return "plain"
    if (meta.nonAtlasSprite) return "nonAtlas"
    if (meta.frame) return "frame"
    if (meta.tilesheet) return "tilesheet"
    return "plain"
}

function cloneMetadata(meta: SpriteMetadata): SpriteMetadata {
    return JSON.parse(JSON.stringify(meta))
}

function Breadcrumb({ path, navigate }: { path: string; navigate: (path: string) => void }) {
    const parts = path.split("/")
    const name = parts.pop()!

    return (
        <div className="flex items-center gap-0.5 text-sm min-w-0">
            {parts.map((part, i) => {
                const dirPath = parts.slice(0, i + 1).join("/")
                return (
                    <span key={i} className="flex items-center gap-0.5">
                        <button
                            className="text-muted-foreground hover:text-foreground transition-colors"
                            onClick={() => navigate(`/sprites?dir=${encodeURIComponent(dirPath)}`)}
                        >
                            {part}
                        </button>
                        <span className="text-muted-foreground">/</span>
                    </span>
                )
            })}
            <span className="font-medium truncate">{name}</span>
        </div>
    )
}

function CopyButton({ text }: { text: string }) {
    const [copied, setCopied] = useState(false)
    const handleCopy = () => {
        navigator.clipboard.writeText(text)
        setCopied(true)
        setTimeout(() => setCopied(false), 1500)
    }
    return (
        <Button variant="ghost" size="sm" className="h-6 px-1.5" onClick={handleCopy} title="Copy sprite path">
            {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
        </Button>
    )
}

function BgPicker({ value, onChange }: { value: string; onChange: (v: string) => void }) {
    return (
        <div className="flex items-center gap-1">
            {BG_OPTIONS.map((opt) => (
                <button
                    key={opt.value}
                    className={`w-5 h-5 rounded border ${opt.style} ${value === opt.value ? "ring-2 ring-primary ring-offset-1 ring-offset-background" : "border-border"}`}
                    onClick={() => onChange(opt.value)}
                    title={opt.label}
                />
            ))}
        </div>
    )
}

export function SpriteEditor() {
    const params = useParams()
    const path = params["*"] ?? ""
    const navigate = useNavigate()
    const { data: sprite, isLoading, error } = useSprite(path)
    const saveMutation = useSaveSprite()
    const deleteMutation = useDeleteSpriteSidecar()
    const [pendingTile, setPendingTile] = useState<{ col: number; row: number } | null>(null)
    const [editedMeta, setEditedMeta] = useState<SpriteMetadata | null>(null)
    const [dirty, setDirty] = useState(false)
    const [bg, setBg] = useState("checker")

    useEffect(() => {
        if (sprite?.metadata) {
            setEditedMeta(cloneMetadata(sprite.metadata))
            setDirty(false)
        } else if (sprite) {
            setEditedMeta({})
            setDirty(false)
        }
    }, [sprite])

    const updateMeta = useCallback((updater: (prev: SpriteMetadata) => SpriteMetadata) => {
        setEditedMeta((prev) => {
            const next = updater(prev ?? {})
            setDirty(true)
            return next
        })
    }, [])

    const handleSave = async () => {
        if (!editedMeta || !path) return
        const cleaned = cleanMetadata(editedMeta)
        await saveMutation.mutateAsync({ path, metadata: cleaned })
        setDirty(false)
    }

    const handleDelete = async () => {
        if (!path) return
        await deleteMutation.mutateAsync(path)
        setDirty(false)
    }

    const handleTypeChange = (type: SpriteType) => {
        updateMeta((prev) => {
            const next: SpriteMetadata = {}
            if (type === "tilesheet") {
                next.tilesheet = prev.tilesheet ?? { tileWidth: sprite?.imageWidth ?? 16, tileHeight: sprite?.imageHeight ?? 16 }
                next.animations = prev.animations
                next.sprites = prev.sprites
            } else if (type === "frame") {
                next.frame = prev.frame ?? {
                    defaults: { cutMargin: 2, padding: 2, frameMode: "stretch" },
                }
            } else if (type === "nonAtlas") {
                next.nonAtlasSprite = true
            }
            return next
        })
    }

    if (isLoading) {
        return <div className="flex items-center justify-center h-full text-muted-foreground">Loading...</div>
    }

    if (error || !sprite) {
        return (
            <div className="flex flex-col items-center justify-center h-full gap-4 text-muted-foreground">
                <p>Sprite not found: {path}</p>
                <Button variant="outline" onClick={() => navigate("/sprites")}>
                    Back to browser
                </Button>
            </div>
        )
    }

    const meta = editedMeta ?? {}
    const currentType = detectType(meta)
    const hasTilesheet = !!meta.tilesheet
    const hasFrame = !!meta.frame
    const tileWidth = meta.tilesheet?.tileWidth ?? sprite.imageWidth
    const tileHeight = meta.tilesheet?.tileHeight ?? sprite.imageHeight
    const totalCols = tileWidth > 0 ? Math.floor(sprite.imageWidth / tileWidth) : 1
    const totalRows = tileHeight > 0 ? Math.floor(sprite.imageHeight / tileHeight) : 1
    const bgClass = BG_OPTIONS.find((o) => o.value === bg)?.style ?? BG_OPTIONS[0].style

    return (
        <div className="flex h-full flex-col">
            <div className="flex items-center gap-3 border-b border-border px-4 py-2 shrink-0">
                <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-1">
                        <Breadcrumb path={sprite.path} navigate={navigate} />
                        <CopyButton text={sprite.path} />
                    </div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <span>{sprite.imageWidth}x{sprite.imageHeight}px</span>
                        {hasTilesheet && (
                            <>
                                <Separator orientation="vertical" className="h-3" />
                                <span>{tileWidth}x{tileHeight} tiles, {totalCols}x{totalRows} grid</span>
                            </>
                        )}
                    </div>
                </div>
                <BgPicker value={bg} onChange={setBg} />
                <Select value={currentType} onValueChange={(v) => handleTypeChange(v as SpriteType)}>
                    <SelectTrigger className="w-28 h-7 text-xs">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="plain">Plain</SelectItem>
                        <SelectItem value="tilesheet">Tilesheet</SelectItem>
                        <SelectItem value="frame">Frame</SelectItem>
                        <SelectItem value="nonAtlas">Non-Atlas</SelectItem>
                    </SelectContent>
                </Select>
                {sprite.hasYaml && (
                    <Button variant="outline" size="sm" className="h-7" onClick={handleDelete} disabled={deleteMutation.isPending}>
                        <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                )}
                <Button size="sm" className="h-7" onClick={handleSave} disabled={!dirty || saveMutation.isPending}>
                    <Save className="h-3.5 w-3.5 mr-1" />
                    {saveMutation.isPending ? "Saving..." : "Save"}
                    {dirty && <Badge variant="secondary" className="ml-1 text-[10px] px-1 py-0">*</Badge>}
                </Button>
            </div>

            <div className="flex-1 overflow-y-auto">
                <div className="p-4 space-y-4">
                    <Tabs defaultValue="edit">
                        <TabsList>
                            <TabsTrigger value="edit">Edit</TabsTrigger>
                            <TabsTrigger value="preview">Preview</TabsTrigger>
                            {sprite.rawYaml && <TabsTrigger value="yaml">Raw YAML</TabsTrigger>}
                        </TabsList>

                        <TabsContent value="edit" className="space-y-4">
                            {hasTilesheet && (
                                <>
                                    <TilesheetViewer
                                        imageSrc={apiImageUrl(sprite.path)}
                                        imageWidth={sprite.imageWidth}
                                        imageHeight={sprite.imageHeight}
                                        tileWidth={tileWidth}
                                        tileHeight={tileHeight}
                                        sprites={meta.sprites}
                                        animations={meta.animations}
                                        onTileClick={(col, row) => setPendingTile({ col, row })}
                                        bgClass={bgClass}
                                        tilesheet={meta.tilesheet!}
                                        onTilesheetChange={(ts) => updateMeta((prev) => ({ ...prev, tilesheet: ts }))}
                                    />
                                    <div className="flex items-center gap-2">
                                        <AnimationModal
                                            animations={meta.animations ?? {}}
                                            onChange={(animations) =>
                                                updateMeta((prev) => ({
                                                    ...prev,
                                                    animations: Object.keys(animations).length > 0 ? animations : undefined,
                                                }))
                                            }
                                            imageSrc={apiImageUrl(sprite.path)}
                                            tileWidth={tileWidth}
                                            tileHeight={tileHeight}
                                            imageWidth={sprite.imageWidth}
                                            imageHeight={sprite.imageHeight}
                                            totalCols={totalCols}
                                            totalRows={totalRows}
                                            bgClass={bgClass}
                                            sprites={meta.sprites}
                                            tilesheet={meta.tilesheet!}
                                            onTilesheetChange={(ts) => updateMeta((prev) => ({ ...prev, tilesheet: ts }))}
                                        />
                                    </div>
                                    <SpriteAliasEditor
                                        sprites={meta.sprites ?? {}}
                                        onChange={(sprites) =>
                                            updateMeta((prev) => ({
                                                ...prev,
                                                sprites: Object.keys(sprites).length > 0 ? sprites : undefined,
                                            }))
                                        }
                                        pendingTile={pendingTile}
                                        onClearPendingTile={() => setPendingTile(null)}
                                    />
                                </>
                            )}

                            {hasFrame && (
                                <FrameEditor
                                    frame={meta.frame!}
                                    onChange={(frame) => updateMeta((prev) => ({ ...prev, frame }))}
                                    imageSrc={apiImageUrl(sprite.path)}
                                    imageWidth={sprite.imageWidth}
                                    imageHeight={sprite.imageHeight}
                                />
                            )}

                            {currentType === "nonAtlas" && (
                                <Card>
                                    <CardContent className="p-4">
                                        <div className="flex items-center gap-2">
                                            <Switch
                                                checked={!!meta.nonAtlasSprite}
                                                onCheckedChange={(v) =>
                                                    updateMeta((prev) => ({ ...prev, nonAtlasSprite: v || undefined }))
                                                }
                                            />
                                            <Label>Non-Atlas Sprite</Label>
                                        </div>
                                        <p className="text-xs text-muted-foreground mt-2">
                                            Excluded from the texture atlas. Loaded as a standalone texture.
                                        </p>
                                    </CardContent>
                                </Card>
                            )}

                            {currentType === "plain" && !hasTilesheet && !hasFrame && (
                                <div className={`overflow-auto rounded border border-border ${bgClass} inline-block`}>
                                    <img
                                        src={apiImageUrl(sprite.path)}
                                        alt={sprite.path}
                                        style={{ imageRendering: "pixelated" }}
                                    />
                                </div>
                            )}
                        </TabsContent>

                        <TabsContent value="preview" className="space-y-4">
                            <div className={`overflow-auto rounded border border-border ${bgClass} inline-block`}>
                                <img
                                    src={apiImageUrl(sprite.path)}
                                    alt={sprite.path}
                                    style={{ imageRendering: "pixelated" }}
                                />
                            </div>
                        </TabsContent>

                        {sprite.rawYaml && (
                            <TabsContent value="yaml">
                                <Card>
                                    <CardContent className="p-4">
                                        <pre className="text-xs font-mono whitespace-pre-wrap overflow-auto max-h-[600px]">
                                            {sprite.rawYaml}
                                        </pre>
                                    </CardContent>
                                </Card>
                            </TabsContent>
                        )}
                    </Tabs>
                </div>
            </div>
        </div>
    )
}

function cleanMetadata(meta: SpriteMetadata): SpriteMetadata {
    const cleaned: SpriteMetadata = {}
    if (meta.tilesheet) cleaned.tilesheet = meta.tilesheet
    if (meta.frame) cleaned.frame = meta.frame
    if (meta.animations && Object.keys(meta.animations).length > 0) cleaned.animations = meta.animations
    if (meta.sprites && Object.keys(meta.sprites).length > 0) cleaned.sprites = meta.sprites
    if (meta.nonAtlasSprite) cleaned.nonAtlasSprite = true
    return cleaned
}
