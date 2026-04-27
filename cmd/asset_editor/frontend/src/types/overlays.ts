export interface OverlayRect {
    x: number
    y: number
    w: number
    h: number
}

export interface OverlayTargetMessage {
    text: string
    placement: string
    wrap?: number
}

export interface OverlayTargetBadge {
    placement?: string
    label?: string
}

export interface OverlayTarget {
    rect?: string
    region?: OverlayRect
    message?: OverlayTargetMessage
    badge?: OverlayTargetBadge
    noPadding?: boolean
}

export interface OverlayFlow {
    description?: string
    targets: OverlayTarget[]
}

export interface OverlayFlowEntry {
    name: string
    description?: string
    targetCount: number
}

export interface OverlayFlowDetail {
    name: string
    description?: string
    targetCount: number
    flow: OverlayFlow
    rawYaml: string
}

export interface NamedRectsDetail {
    rects: Record<string, OverlayRect>
    rawYaml: string
}
