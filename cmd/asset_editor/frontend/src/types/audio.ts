export interface AudioEntry {
    path: string
    directory: string
    name: string
    format: string
    fileSize: number
    category: string
    hasYaml: boolean
    gain?: number
    resourceName: string
}

export interface AudioMetadata {
    gain?: number
}

export interface AudioDetail extends AudioEntry {
    resourceName: string
    metadata?: AudioMetadata
    rawYaml?: string
}
