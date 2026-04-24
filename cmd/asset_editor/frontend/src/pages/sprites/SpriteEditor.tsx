import { useCallback, useEffect, useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { ArrowLeft, Save, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useSprite, useSaveSprite, useDeleteSpriteSidecar } from "@/api/sprites"
import { apiImageUrl } from "@/api/client"
import { TilesheetViewer } from "@/components/sprites/TilesheetViewer"
import { TilesheetConfigForm } from "@/components/sprites/TilesheetConfigForm"
import { SpriteAliasEditor } from "@/components/sprites/SpriteAliasEditor"
import { AnimationEditor } from "@/components/sprites/AnimationEditor"
import { FrameEditor } from "@/components/sprites/FrameEditor"
import type { SpriteMetadata } from "@/types/sprites"

type SpriteType = "plain" | "tilesheet" | "frame" | "nonAtlas"

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

export function SpriteEditor() {
    const params = useParams()
    const path = params["*"] ?? ""
    const navigate = useNavigate()
    const { data: sprite, isLoading, error } = useSprite(path)
    const saveMutation = useSaveSprite()
    const deleteMutation = useDeleteSpriteSidecar()
    const [selectedAnimation, setSelectedAnimation] = useState<string | undefined>()
    const [pendingTile, setPendingTile] = useState<{ col: number; row: number } | null>(null)
    const [editedMeta, setEditedMeta] = useState<SpriteMetadata | null>(null)
    const [dirty, setDirty] = useState(false)

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

    const handleTileClick = (col: number, row: number) => {
        setPendingTile({ col, row })
    }

    if (isLoading) {
        return <div className="flex items-center justify-center h-full text-muted-foreground">Loading...</div>
    }

    if (error || !sprite) {
        return (
            <div className="flex flex-col items-center justify-center h-full gap-4 text-muted-foreground">
                <p>Sprite not found: {path}</p>
                <Button variant="outline" onClick={() => navigate("/sprites")}>
                    <ArrowLeft className="mr-2 h-4 w-4" />
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

    return (
        <div className="flex h-full flex-col">
            <div className="flex items-center gap-3 border-b border-border px-4 py-3">
                <Button variant="ghost" size="sm" onClick={() => navigate("/sprites")}>
                    <ArrowLeft className="h-4 w-4" />
                </Button>
                <div className="flex-1 min-w-0">
                    <h1 className="text-sm font-medium truncate">{sprite.path}</h1>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <span>{sprite.imageWidth}x{sprite.imageHeight}px</span>
                        {hasTilesheet && (
                            <>
                                <Separator orientation="vertical" className="h-3" />
                                <span>{tileWidth}x{tileHeight} tiles</span>
                                <Separator orientation="vertical" className="h-3" />
                                <span>{totalCols}x{totalRows} grid</span>
                            </>
                        )}
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <Select value={currentType} onValueChange={(v) => handleTypeChange(v as SpriteType)}>
                        <SelectTrigger className="w-32 h-8">
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
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={handleDelete}
                            disabled={deleteMutation.isPending}
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                    )}
                    <Button
                        size="sm"
                        onClick={handleSave}
                        disabled={!dirty || saveMutation.isPending}
                    >
                        <Save className="h-4 w-4 mr-1" />
                        {saveMutation.isPending ? "Saving..." : "Save"}
                        {dirty && <Badge variant="secondary" className="ml-1 text-[10px] px-1 py-0">modified</Badge>}
                    </Button>
                </div>
            </div>

            <ScrollArea className="flex-1">
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
                                        selectedAnimation={selectedAnimation}
                                        onTileClick={handleTileClick}
                                    />
                                    <TilesheetConfigForm
                                        tilesheet={meta.tilesheet!}
                                        onChange={(tilesheet) => updateMeta((prev) => ({ ...prev, tilesheet }))}
                                        imageWidth={sprite.imageWidth}
                                        imageHeight={sprite.imageHeight}
                                    />
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
                                    <AnimationEditor
                                        animations={meta.animations ?? {}}
                                        onChange={(animations) =>
                                            updateMeta((prev) => ({
                                                ...prev,
                                                animations: Object.keys(animations).length > 0 ? animations : undefined,
                                            }))
                                        }
                                        imageSrc={apiImageUrl(sprite.path)}
                                        imageWidth={sprite.imageWidth}
                                        imageHeight={sprite.imageHeight}
                                        tileWidth={tileWidth}
                                        tileHeight={tileHeight}
                                        totalCols={totalCols}
                                        totalRows={totalRows}
                                        selectedAnimation={selectedAnimation}
                                        onSelectAnimation={setSelectedAnimation}
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
                                <div className="overflow-auto rounded border border-border bg-[repeating-conic-gradient(#808080_0%_25%,transparent_0%_50%)] bg-[length:16px_16px] inline-block">
                                    <img
                                        src={apiImageUrl(sprite.path)}
                                        alt={sprite.path}
                                        style={{ imageRendering: "pixelated" }}
                                    />
                                </div>
                            )}
                        </TabsContent>

                        <TabsContent value="preview" className="space-y-4">
                            <div className="overflow-auto rounded border border-border bg-[repeating-conic-gradient(#808080_0%_25%,transparent_0%_50%)] bg-[length:16px_16px] inline-block">
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
            </ScrollArea>
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
