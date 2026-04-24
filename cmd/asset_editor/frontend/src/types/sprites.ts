export interface SpriteEntry {
    path: string
    directory: string
    name: string
    type: "plain" | "tilesheet" | "frame" | "nonAtlas"
    imageWidth: number
    imageHeight: number
    hasYaml: boolean
}

export interface SpriteTilesheet {
    tileWidth: number
    tileHeight: number
}

export interface TilesheetCoordinates {
    row: number
    column: number
}

export interface SpriteTilesheetAnimationHSequence {
    row: number
    fromColumn?: number
    toColumn?: number
    frameWeights?: number[]
}

export interface SpriteTilesheetAnimationVSequence {
    column: number
    fromRow?: number
    toRow?: number
    frameWeights?: number[]
}

export interface SpriteTilesheetAnimationTile {
    row: number
    column: number
    weight?: number
}

export interface SpriteTilesheetAnimation {
    h_sequence?: SpriteTilesheetAnimationHSequence
    v_sequence?: SpriteTilesheetAnimationVSequence
    tiles?: SpriteTilesheetAnimationTile[]
    framesPerSecond: number
    jitterPercent?: number
    repeat?: boolean
    pingPong?: boolean
    reverse?: boolean
    randomize?: boolean
}

export type FrameSide = "top" | "top-left" | "left" | "bottom-left" | "bottom" | "bottom-right" | "right" | "top-right" | "middle"
export type FrameMode = "stretch" | "repeat"

export interface SpriteFrameDefaults {
    cutMargin: number
    padding: number
    frameMode: FrameMode
}

export interface SpriteFrame {
    cutMargin?: Partial<Record<FrameSide, number>>
    padding?: Partial<Record<FrameSide, number>>
    frameModes?: Partial<Record<FrameSide, FrameMode>>
    defaults: SpriteFrameDefaults
}

export interface SpriteMetadata {
    frame?: SpriteFrame
    tilesheet?: SpriteTilesheet
    animations?: Record<string, SpriteTilesheetAnimation>
    sprites?: Record<string, TilesheetCoordinates>
    nonAtlasSprite?: boolean
}

export interface ComputedFields {
    columns: number
    rows: number
}

export interface SpriteDetail extends SpriteEntry {
    metadata?: SpriteMetadata
    computed?: ComputedFields
    rawYaml?: string
}

export interface ScaffoldRequest {
    name: string
    tileWidth: number
    tileHeight: number
    cols: number
    rows: number
    sprites?: string[]
    force?: boolean
}
