export interface ScriptFileEntry {
    path: string;
    directory: string;
    name: string;
    handlerCount: number;
    sequenceCount: number;
    handlerNames: string[];
    sequenceNames?: string[];
    customActionNames?: string[];
    constNames?: string[];
    dataListNames?: string[];
    propertyTemplateNames?: string[];
}

export interface ScriptFileDetail extends ScriptFileEntry {
    rawYaml: string;
}

export interface ScriptSchema {
    stepKinds: Record<string, StepKindDef>;
    actions: Record<string, CallableDef>;
    conditions: Record<string, CallableDef>;
    builtinConditions: Record<string, ConditionDef>;
    eventHooks: Record<string, EventHookDef>;
    templateVars: TemplateVarDef[];
}

export interface ParamDef {
    name: string;
    type: string;
    required?: boolean;
    default?: unknown;
    description: string;
    enum?: string[];
}

export interface StepKindDef {
    name: string;
    description: string;
    category: string;
    paramStyle: string;
    params?: ParamDef[];
    acceptsSubSteps?: boolean;
}

export interface CallableDef {
    name: string;
    description: string;
    params?: ParamDef[];
}

export interface ConditionDef {
    name: string;
    description: string;
    params?: ParamDef[];
    isComposite?: boolean;
}

export interface EventHookDef {
    name: string;
    yamlKey: string;
    description: string;
    eventType: string;
    filterFields?: ParamDef[];
}

export interface TemplateVarDef {
    pattern: string;
    description: string;
}

// Tiled cross-reference types

export interface TiledEntityRef {
    mapFile: string;
    objectId: number;
    entityId?: string;
    objectType?: string;
    x: number;
    y: number;
    properties: Record<string, string>;
}

export interface TiledHandlerUsage {
    handlerName: string;
    entities: TiledEntityRef[];
}

// Parsed script tree types - mirrors the YAML structure

export interface HandlerPropDef {
    name: string;
    type: string;
    required?: boolean;
    default?: unknown;
    description?: string;
}

export interface ParsedScript {
    handlers: Record<string, HandlerDef>;
    sequences?: Record<string, SequenceDef>;
    consts?: Record<string, unknown>;
    custom_actions?: Record<string, CustomActionDef>;
    data?: Record<string, string[]>;
    property_templates?: Record<string, Record<string, unknown>>;
}

export type ScriptItemSelection =
    | { type: "handler"; name: string }
    | { type: "custom_action"; name: string }
    | { type: "sequence"; name: string }
    | { type: "const"; name: string }
    | { type: "data"; name: string }
    | { type: "property_template"; name: string };

export interface CustomActionDef {
    description?: string;
    params?: CustomActionParam[];
    steps: StepNode[];
}

export interface CustomActionParam {
    name: string;
    description?: string;
    default?: unknown;
}

export type HandlerDef = {
    [hookKey: string]: RuleDef[];
} & {
    props?: HandlerPropDef[];
    var?: Record<string, unknown>;
};

export interface RuleDef {
    filter?: Record<string, unknown>;
    when?: ConditionNode;
    steps: StepNode[];
    set_state?: Record<string, unknown>;
}

export interface StepNode {
    kind: string;
    params: unknown;
}

export interface ConditionNode {
    type: string;
    params: unknown;
}

export interface SequenceDef {
    params?: string[];
    steps: StepNode[];
}
