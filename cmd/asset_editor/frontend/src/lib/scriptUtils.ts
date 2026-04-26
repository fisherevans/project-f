import YAML from "yaml";
import type {
    ParsedScript,
    HandlerDef,
    RuleDef,
    StepNode,
    ConditionNode,
    SequenceDef,
    StepKindDef,
    ScriptSchema,
} from "@/types/scripts";

const EVENT_HOOK_KEYS = new Set([
    "on_init",
    "on_interact_self",
    "on_interact",
    "on_zone_activity",
    "on_broadcast",
    "on_global_updated",
    "on_state_enter",
    "on_combat_complete",
    "on_timer_complete",
    "on_motion_complete_self",
    "on_motion_complete",
]);

export function isEventHookKey(key: string): boolean {
    return EVENT_HOOK_KEYS.has(key);
}

export function parseScript(yamlStr: string): ParsedScript {
    const raw = YAML.parse(yamlStr) as Record<string, unknown> | null;
    if (!raw) return { handlers: {} };

    const handlers: Record<string, HandlerDef> = {};
    const rawHandlers = (raw.handlers ?? {}) as Record<string, unknown>;
    for (const [name, handlerRaw] of Object.entries(rawHandlers)) {
        handlers[name] = parseHandler(handlerRaw as Record<string, unknown>);
    }

    let sequences: Record<string, SequenceDef> | undefined;
    if (raw.sequences) {
        sequences = {};
        const rawSeqs = raw.sequences as Record<string, unknown>;
        for (const [name, seqRaw] of Object.entries(rawSeqs)) {
            const s = seqRaw as Record<string, unknown>;
            sequences[name] = {
                params: s.params as string[] | undefined,
                steps: parseSteps((s.steps ?? s.effects ?? []) as unknown[]),
            };
        }
    }

    let data: Record<string, string[]> | undefined;
    if (raw.data && typeof raw.data === "object") {
        data = {};
        for (const [name, list] of Object.entries(raw.data as Record<string, unknown>)) {
            if (Array.isArray(list)) {
                data[name] = list.map(String);
            }
        }
    }

    let consts: Record<string, unknown> | undefined;
    if (raw.consts && typeof raw.consts === "object") {
        consts = raw.consts as Record<string, unknown>;
    }

    let custom_actions: Record<string, import("@/types/scripts").CustomActionDef> | undefined;
    if (raw.custom_actions && typeof raw.custom_actions === "object") {
        custom_actions = {};
        for (const [name, actionRaw] of Object.entries(raw.custom_actions as Record<string, unknown>)) {
            const a = actionRaw as Record<string, unknown>;
            custom_actions[name] = {
                description: a.description as string | undefined,
                params: a.params as import("@/types/scripts").CustomActionParam[] | undefined,
                steps: parseSteps((a.steps ?? []) as unknown[]),
            };
        }
    }

    let property_templates: Record<string, Record<string, unknown>> | undefined;
    if (raw.property_templates && typeof raw.property_templates === "object") {
        property_templates = raw.property_templates as Record<string, Record<string, unknown>>;
    }

    return { handlers, sequences, consts, custom_actions, data, property_templates };
}

function parseHandler(raw: Record<string, unknown>): HandlerDef {
    const handler: HandlerDef = {};
    if (raw.props && Array.isArray(raw.props)) {
        handler.props = raw.props as import("@/types/scripts").HandlerPropDef[];
    }
    if (raw.var && typeof raw.var === "object") {
        handler.var = raw.var as Record<string, unknown>;
    }
    for (const [key, value] of Object.entries(raw)) {
        if (isEventHookKey(key) && Array.isArray(value)) {
            handler[key] = value.map((r) => parseRule(r as Record<string, unknown>));
        }
    }
    return handler;
}

function parseRule(raw: Record<string, unknown>): RuleDef {
    const rule: RuleDef = {
        steps: parseSteps((raw.steps ?? []) as unknown[]),
    };
    if (raw.filter) {
        rule.filter = raw.filter as Record<string, unknown>;
    }
    if (raw.when) {
        rule.when = parseCondition(raw.when);
    }
    return rule;
}

export function parseSteps(raw: unknown[]): StepNode[] {
    return raw.map((item) => {
        const obj = item as Record<string, unknown>;
        const keys = Object.keys(obj);
        if (keys.length !== 1) {
            return { kind: keys[0] ?? "unknown", params: obj[keys[0]] };
        }
        return { kind: keys[0], params: obj[keys[0]] };
    });
}

export function parseCondition(raw: unknown): ConditionNode {
    if (typeof raw !== "object" || raw === null) {
        return { type: "unknown", params: raw };
    }
    const obj = raw as Record<string, unknown>;
    const keys = Object.keys(obj);
    if (keys.length !== 1) {
        return { type: "unknown", params: raw };
    }
    const type = keys[0];
    const params = obj[type];

    if (type === "all" || type === "any") {
        const children = Array.isArray(params) ? params.map(parseCondition) : [];
        return { type, params: children };
    }
    if (type === "not") {
        return { type, params: parseCondition(params) };
    }
    if (type === "check") {
        return { type, params: params as Record<string, unknown> };
    }
    return { type, params };
}

export function stringifyScript(script: ParsedScript): string {
    const obj: Record<string, unknown> = {};

    if (script.data && Object.keys(script.data).length > 0) {
        obj.data = script.data;
    }

    const handlers: Record<string, unknown> = {};
    for (const [name, handler] of Object.entries(script.handlers)) {
        handlers[name] = serializeHandler(handler);
    }
    if (Object.keys(handlers).length > 0) {
        obj.handlers = handlers;
    }

    if (script.consts && Object.keys(script.consts).length > 0) {
        obj.consts = script.consts;
    }

    if (script.sequences && Object.keys(script.sequences).length > 0) {
        const seqs: Record<string, unknown> = {};
        for (const [name, seq] of Object.entries(script.sequences)) {
            const s: Record<string, unknown> = {};
            if (seq.params && seq.params.length > 0) {
                s.params = seq.params;
            }
            s.steps = serializeSteps(seq.steps);
            seqs[name] = s;
        }
        obj.sequences = seqs;
    }

    if (script.custom_actions && Object.keys(script.custom_actions).length > 0) {
        const actions: Record<string, unknown> = {};
        for (const [name, action] of Object.entries(script.custom_actions)) {
            const a: Record<string, unknown> = {};
            if (action.description) a.description = action.description;
            if (action.params && action.params.length > 0) a.params = action.params;
            a.steps = serializeSteps(action.steps);
            actions[name] = a;
        }
        obj.custom_actions = actions;
    }

    if (script.property_templates && Object.keys(script.property_templates).length > 0) {
        obj.property_templates = script.property_templates;
    }

    return YAML.stringify(obj, { indent: 2, lineWidth: 0 });
}

function serializeHandler(handler: HandlerDef): Record<string, unknown> {
    const obj: Record<string, unknown> = {};
    if (handler.props && handler.props.length > 0) {
        obj.props = handler.props;
    }
    if (handler.var && Object.keys(handler.var).length > 0) {
        obj.var = handler.var;
    }
    for (const [key, rules] of Object.entries(handler)) {
        if (isEventHookKey(key)) {
            obj[key] = (rules as RuleDef[]).map(serializeRule);
        }
    }
    return obj;
}

function serializeRule(rule: RuleDef): Record<string, unknown> {
    const obj: Record<string, unknown> = {};
    if (rule.filter && Object.keys(rule.filter).length > 0) {
        obj.filter = rule.filter;
    }
    if (rule.when) {
        obj.when = serializeCondition(rule.when);
    }
    obj.steps = serializeSteps(rule.steps);
    return obj;
}

export function serializeSteps(steps: StepNode[]): unknown[] {
    return steps.map((step) => ({ [step.kind]: step.params }));
}

export function serializeCondition(cond: ConditionNode): unknown {
    if (cond.type === "all" || cond.type === "any") {
        const children = cond.params as ConditionNode[];
        return { [cond.type]: children.map(serializeCondition) };
    }
    if (cond.type === "not") {
        return { [cond.type]: serializeCondition(cond.params as ConditionNode) };
    }
    return { [cond.type]: cond.params };
}

export function createEmptyRule(): RuleDef {
    return { steps: [] };
}

export function createEmptyStep(kind: string, schema?: StepKindDef): StepNode {
    if (!schema) return { kind, params: null };

    switch (schema.paramStyle) {
        case "string":
            return { kind, params: "" };
        case "number":
            return { kind, params: 0 };
        case "map": {
            const params: Record<string, unknown> = {};
            for (const p of schema.params ?? []) {
                if (p.default !== undefined) {
                    params[p.name] = p.default;
                } else if (p.required) {
                    switch (p.type) {
                        case "string":
                            params[p.name] = "";
                            break;
                        case "number":
                            params[p.name] = 0;
                            break;
                        case "bool":
                            params[p.name] = false;
                            break;
                        case "list":
                        case "steps":
                            params[p.name] = [];
                            break;
                        default:
                            params[p.name] = "";
                    }
                }
            }
            return { kind, params };
        }
        case "list":
            return { kind, params: [] };
        case "string_or_map":
            return { kind, params: "" };
        default:
            return { kind, params: null };
    }
}

export function getHandlerHookKeys(handler: HandlerDef): string[] {
    return Object.keys(handler).filter(isEventHookKey);
}

export function getAvailableHooks(handler: HandlerDef, schema: ScriptSchema): string[] {
    const used = new Set(getHandlerHookKeys(handler));
    return Object.values(schema.eventHooks)
        .map((h) => h.yamlKey)
        .filter((k) => !used.has(k))
        .sort();
}

export function moveItem<T>(arr: T[], from: number, to: number): T[] {
    const result = [...arr];
    const [item] = result.splice(from, 1);
    result.splice(to, 0, item);
    return result;
}

export function getStepSubSteps(step: StepNode): { key: string; steps: StepNode[] }[] {
    if (typeof step.params !== "object" || step.params === null || Array.isArray(step.params)) {
        return [];
    }
    const obj = step.params as Record<string, unknown>;
    const result: { key: string; steps: StepNode[] }[] = [];
    for (const key of ["effects", "pre_effects", "post_effects", "interstitial", "steps", "then", "else", "default"]) {
        if (Array.isArray(obj[key])) {
            result.push({ key, steps: parseSteps(obj[key] as unknown[]) });
        }
    }
    // Handle switch cases - each case has its own steps
    if (step.kind === "switch" && Array.isArray(obj.cases)) {
        for (let i = 0; i < (obj.cases as unknown[]).length; i++) {
            const c = (obj.cases as Record<string, unknown>[])[i];
            if (c && Array.isArray(c.steps)) {
                result.push({ key: `cases[${i}].steps`, steps: parseSteps(c.steps as unknown[]) });
            }
        }
    }
    return result;
}

export function setStepSubSteps(step: StepNode, key: string, steps: StepNode[]): StepNode {
    const params = { ...(step.params as Record<string, unknown>) };
    const casesMatch = key.match(/^cases\[(\d+)\]\.steps$/);
    if (casesMatch) {
        const idx = parseInt(casesMatch[1]);
        const cases = [...(params.cases as Record<string, unknown>[])];
        cases[idx] = { ...cases[idx], steps: serializeSteps(steps) };
        params.cases = cases;
    } else {
        params[key] = serializeSteps(steps);
    }
    return { ...step, params };
}
