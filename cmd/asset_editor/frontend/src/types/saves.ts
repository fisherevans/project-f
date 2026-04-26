export interface SaveSummary {
    save_id: string
    character_name: string
}

export interface SaveDetail extends SaveSummary {
    raw_yaml: string
}
