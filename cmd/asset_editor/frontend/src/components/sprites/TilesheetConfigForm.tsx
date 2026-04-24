import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { SpriteTilesheet } from "@/types/sprites"

interface TilesheetConfigFormProps {
    tilesheet: SpriteTilesheet
    onChange: (tilesheet: SpriteTilesheet) => void
    imageWidth: number
    imageHeight: number
}

export function TilesheetConfigForm({ tilesheet, onChange, imageWidth, imageHeight }: TilesheetConfigFormProps) {
    const cols = tilesheet.tileWidth > 0 ? Math.floor(imageWidth / tilesheet.tileWidth) : 0
    const rows = tilesheet.tileHeight > 0 ? Math.floor(imageHeight / tilesheet.tileHeight) : 0

    return (
        <Card>
            <CardHeader className="py-3 px-4">
                <CardTitle className="text-sm">Tilesheet</CardTitle>
            </CardHeader>
            <CardContent className="px-4 pb-3">
                <div className="grid grid-cols-2 gap-3">
                    <div>
                        <Label className="text-xs">Tile Width</Label>
                        <Input
                            type="number"
                            className="mt-1"
                            min={1}
                            max={imageWidth}
                            value={tilesheet.tileWidth}
                            onChange={(e) =>
                                onChange({ ...tilesheet, tileWidth: parseInt(e.target.value) || 1 })
                            }
                        />
                    </div>
                    <div>
                        <Label className="text-xs">Tile Height</Label>
                        <Input
                            type="number"
                            className="mt-1"
                            min={1}
                            max={imageHeight}
                            value={tilesheet.tileHeight}
                            onChange={(e) =>
                                onChange({ ...tilesheet, tileHeight: parseInt(e.target.value) || 1 })
                            }
                        />
                    </div>
                </div>
                <p className="text-xs text-muted-foreground mt-2">
                    {cols} columns x {rows} rows = {cols * rows} tiles
                </p>
            </CardContent>
        </Card>
    )
}
