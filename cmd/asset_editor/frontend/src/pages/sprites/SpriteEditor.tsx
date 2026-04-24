import { useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { ArrowLeft } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import { useSprite } from "@/api/sprites"
import { apiImageUrl } from "@/api/client"
import { TilesheetViewer } from "@/components/sprites/TilesheetViewer"

export function SpriteEditor() {
    const params = useParams()
    const path = params["*"] ?? ""
    const navigate = useNavigate()
    const { data: sprite, isLoading, error } = useSprite(path)
    const [selectedAnimation, setSelectedAnimation] = useState<string | undefined>()

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

    const hasTilesheet = !!sprite.metadata?.tilesheet
    const animationNames = sprite.metadata?.animations ? Object.keys(sprite.metadata.animations) : []
    const namedSprites = sprite.metadata?.sprites ? Object.entries(sprite.metadata.sprites) : []

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
                        {hasTilesheet && sprite.computed && (
                            <>
                                <Separator orientation="vertical" className="h-3" />
                                <span>
                                    {sprite.metadata!.tilesheet!.tileWidth}x{sprite.metadata!.tilesheet!.tileHeight} tiles
                                </span>
                                <Separator orientation="vertical" className="h-3" />
                                <span>{sprite.computed.columns}x{sprite.computed.rows} grid</span>
                            </>
                        )}
                        <Badge variant="outline">{sprite.type}</Badge>
                    </div>
                </div>
            </div>

            <ScrollArea className="flex-1">
                <div className="p-4 space-y-4">
                    <Tabs defaultValue="preview">
                        <TabsList>
                            <TabsTrigger value="preview">Preview</TabsTrigger>
                            {sprite.rawYaml && <TabsTrigger value="yaml">Raw YAML</TabsTrigger>}
                        </TabsList>

                        <TabsContent value="preview" className="space-y-4">
                            {hasTilesheet ? (
                                <TilesheetViewer
                                    imageSrc={apiImageUrl(sprite.path)}
                                    imageWidth={sprite.imageWidth}
                                    imageHeight={sprite.imageHeight}
                                    tileWidth={sprite.metadata!.tilesheet!.tileWidth}
                                    tileHeight={sprite.metadata!.tilesheet!.tileHeight}
                                    sprites={sprite.metadata?.sprites}
                                    animations={sprite.metadata?.animations}
                                    selectedAnimation={selectedAnimation}
                                />
                            ) : (
                                <div className="overflow-auto rounded border border-border bg-[repeating-conic-gradient(#808080_0%_25%,transparent_0%_50%)] bg-[length:16px_16px] inline-block">
                                    <img
                                        src={apiImageUrl(sprite.path)}
                                        alt={sprite.path}
                                        style={{ imageRendering: "pixelated" }}
                                    />
                                </div>
                            )}

                            {animationNames.length > 0 && (
                                <Card>
                                    <CardHeader className="py-3 px-4">
                                        <CardTitle className="text-sm">Animations</CardTitle>
                                    </CardHeader>
                                    <CardContent className="px-4 pb-3">
                                        <div className="flex flex-wrap gap-1">
                                            {animationNames.map((name) => (
                                                <Button
                                                    key={name}
                                                    variant={selectedAnimation === name ? "default" : "outline"}
                                                    size="sm"
                                                    onClick={() =>
                                                        setSelectedAnimation(
                                                            selectedAnimation === name ? undefined : name,
                                                        )
                                                    }
                                                >
                                                    {name}
                                                </Button>
                                            ))}
                                        </div>
                                        {selectedAnimation && sprite.metadata?.animations?.[selectedAnimation] && (
                                            <div className="mt-3 text-xs text-muted-foreground space-y-1">
                                                <p>FPS: {sprite.metadata.animations[selectedAnimation].framesPerSecond}</p>
                                                {sprite.metadata.animations[selectedAnimation].repeat !== undefined && (
                                                    <p>Repeat: {String(sprite.metadata.animations[selectedAnimation].repeat)}</p>
                                                )}
                                                {sprite.metadata.animations[selectedAnimation].pingPong && (
                                                    <p>Ping-pong: true</p>
                                                )}
                                            </div>
                                        )}
                                    </CardContent>
                                </Card>
                            )}

                            {namedSprites.length > 0 && (
                                <Card>
                                    <CardHeader className="py-3 px-4">
                                        <CardTitle className="text-sm">Named Sprites</CardTitle>
                                    </CardHeader>
                                    <CardContent className="px-4 pb-3">
                                        <div className="grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-2">
                                            {namedSprites.map(([name, coords]) => (
                                                <div
                                                    key={name}
                                                    className="flex items-center gap-2 rounded border border-border px-2 py-1 text-xs"
                                                >
                                                    <Badge variant="secondary">{name}</Badge>
                                                    <span className="text-muted-foreground">
                                                        r{coords.row} c{coords.column}
                                                    </span>
                                                </div>
                                            ))}
                                        </div>
                                    </CardContent>
                                </Card>
                            )}
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
