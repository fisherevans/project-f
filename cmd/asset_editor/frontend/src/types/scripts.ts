export interface ScriptFileEntry {
    path: string;
    directory: string;
    name: string;
    handlerCount: number;
    sequenceCount: number;
    handlerNames: string[];
    sequenceNames?: string[];
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

// Parsed script tree types - mirrors the YAML structure

export interface ParsedScript {
    handlers: Record<string, HandlerDef>;
    sequences?: Record<string, SequenceDef>;
    consts?: Record<string, unknown>;
    custom_actions?: Record<string, CustomActionDef>;
}

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
