import type { StepKindDef, CallableDef, ConditionDef, EventHookDef, ParamDef } from "@/types/scripts";

function placeholderValue(p: ParamDef): string {
    if (p.default !== undefined && p.default !== null) return String(p.default);
    if (p.enum && p.enum.length > 0) return p.enum[0];
    switch (p.type) {
        case "string": return `"${p.name}_value"`;
        case "number": return "1";
        case "bool": return "true";
        case "steps": return "[]";
        case "condition": return '{ expr: "var.x > 0" }';
        case "map": return "{}";
        case "list": return "[]";
        default: return `"${p.name}_value"`;
    }
}

export function generateStepExample(def: StepKindDef): string {
    const params = def.params?.filter(p => p.type !== "steps") ?? [];
    const subStepParams = def.params?.filter(p => p.type === "steps") ?? [];

    if (def.paramStyle === "string") {
        const val = params[0] ? placeholderValue(params[0]) : '"value"';
        return `- ${def.name}: ${val}`;
    }

    if (def.paramStyle === "number") {
        const val = params[0] ? placeholderValue(params[0]) : "1";
        return `- ${def.name}: ${val}`;
    }

    if (def.paramStyle === "value") {
        if (params.length === 0) return `- ${def.name}:`;
        return `- ${def.name}: ${placeholderValue(params[0])}`;
    }

    if (def.paramStyle === "list") {
        return `- ${def.name}:\n    - { step_kind: params }`;
    }

    if (def.paramStyle === "string_or_map") {
        const lines: string[] = [];
        const firstParam = params[0];
        if (firstParam) {
            lines.push(`# Short form:`);
            lines.push(`- ${def.name}: ${placeholderValue(firstParam)}`);
        }
        if (params.length > 1) {
            lines.push(`# Expanded form:`);
            lines.push(`- ${def.name}:`);
            for (const p of params) {
                lines.push(`    ${p.name}: ${placeholderValue(p)}`);
            }
        }
        return lines.join("\n");
    }

    // map style
    const required = params.filter(p => p.required);
    const optional = params.filter(p => !p.required);
    const lines: string[] = [`- ${def.name}:`];
    for (const p of required) {
        lines.push(`    ${p.name}: ${placeholderValue(p)}`);
    }
    for (const p of optional.slice(0, 2)) {
        lines.push(`    ${p.name}: ${placeholderValue(p)}`);
    }
    for (const p of subStepParams.slice(0, 1)) {
        lines.push(`    ${p.name}:`);
        lines.push(`      - { step_kind: params }`);
    }
    return lines.join("\n");
}

export function generateActionExample(def: CallableDef): string {
    const params = def.params ?? [];
    if (params.length === 0) {
        return `- action: ${def.name}`;
    }
    const lines: string[] = [`- action:`, `    name: ${def.name}`];
    for (const p of params) {
        lines.push(`    ${p.name}: ${placeholderValue(p)}`);
    }
    return lines.join("\n");
}

export function generateConditionExample(def: CallableDef | ConditionDef): string {
    const isComposite = "isComposite" in def && def.isComposite;
    if (isComposite) {
        if (def.name === "not") {
            return `when:\n  not: { expr: "var.x > 0" }`;
        }
        return `when:\n  ${def.name}:\n    - { expr: "var.x > 0" }\n    - { expr: "var.y > 0" }`;
    }
    if (def.name === "expr") {
        return `when:\n  expr: "global.quest_stage == 'done'"`;
    }
    const params = def.params ?? [];
    if (params.length === 0) {
        return `when:\n  ${def.name}:`;
    }
    const lines: string[] = [`when:`, `  ${def.name}:`];
    for (const p of params) {
        lines.push(`    ${p.name}: ${placeholderValue(p)}`);
    }
    return lines.join("\n");
}

export function generateHookExample(def: EventHookDef): string {
    const lines: string[] = [`${def.yamlKey}:`, `  rules:`];
    if (def.filterFields && def.filterFields.length > 0) {
        lines.push(`    - filter:`);
        for (const f of def.filterFields.slice(0, 1)) {
            lines.push(`        ${f.name}: ${placeholderValue(f)}`);
        }
        lines.push(`      steps:`);
    } else {
        lines.push(`    - steps:`);
    }
    lines.push(`        - dialogue: "Hello!"`);
    return lines.join("\n");
}
