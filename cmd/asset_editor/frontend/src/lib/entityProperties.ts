export type EditorKind = "handler-ref" | "enum" | "text" | "yaml-textarea" | "number" | "boolean";

export interface KnownEntityProperty {
    key: string;
    label: string;
    editorKind: EditorKind;
    enumValues?: string[];
    description: string;
}

export const ENTITY_CLASSES = ["ModeBasedEntity", "NPC", "DirectInteraction", "ShadowMob", "Zone"];

export const KNOWN_ENTITY_PROPERTIES: KnownEntityProperty[] = [
    {
        key: "script_ref",
        label: "Script handler",
        editorKind: "handler-ref",
        description: "Script handler bound to this entity",
    },
    {
        key: "class",
        label: "Entity class",
        editorKind: "enum",
        enumValues: ["", ...ENTITY_CLASSES],
        description: "Entity type that determines runtime behavior",
    },
    {
        key: "entity_id",
        label: "Entity ID",
        editorKind: "text",
        description: "Unique identifier for cross-references",
    },
    {
        key: "movement",
        label: "Movement",
        editorKind: "enum",
        enumValues: ["", "static", "horiz"],
        description: "NPC movement pattern (empty = wander)",
    },
    {
        key: "speed",
        label: "Speed",
        editorKind: "enum",
        enumValues: ["", "fast"],
        description: "NPC speed (empty = normal)",
    },
    {
        key: "idle_facing_direction",
        label: "Idle facing",
        editorKind: "enum",
        enumValues: ["", "up", "down", "left", "right"],
        description: "Direction the entity faces when idle",
    },
    {
        key: "active_player_zone",
        label: "Active zone",
        editorKind: "text",
        description: "Only active when player is in this zone",
    },
    {
        key: "zone_id",
        label: "Zone ID",
        editorKind: "text",
        description: "Zone identifier for zone_activity events",
    },
    {
        key: "target",
        label: "Target entity",
        editorKind: "text",
        description: "Entity ID that DirectInteraction redirects to",
    },
    {
        key: "presence_config",
        label: "Presence config",
        editorKind: "yaml-textarea",
        description: "YAML: block_ingress, is_interactable, impedance",
    },
    {
        key: "render_config",
        label: "Render config",
        editorKind: "yaml-textarea",
        description: "YAML: mode-based animations and lights",
    },
    {
        key: "shadow_mob",
        label: "Shadow mob",
        editorKind: "yaml-textarea",
        description: "YAML: combat encounter config",
    },
    {
        key: "metadata",
        label: "Metadata",
        editorKind: "yaml-textarea",
        description: "YAML: talker config and other metadata",
    },
];

const knownPropertyMap = new Map(KNOWN_ENTITY_PROPERTIES.map((p) => [p.key, p]));

export function getKnownProperty(key: string): KnownEntityProperty | undefined {
    return knownPropertyMap.get(key);
}
