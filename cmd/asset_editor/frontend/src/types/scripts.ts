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
