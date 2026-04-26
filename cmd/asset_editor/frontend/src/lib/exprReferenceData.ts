export const ENV_VARS = [
    { name: "var", type: "map", desc: "Handler-local state. Initialized from the handler's var block, mutable via set_var." },
    { name: "global", type: "map", desc: "World and run state. Read-only in expressions - write via set_world_state / set_run_state steps." },
    { name: "const", type: "map", desc: "Script-defined constants from consts: blocks. Shared across all handlers." },
    { name: "save", type: "map", desc: "Current game save data. Read-only. Includes save.animech.level, save.character_name, etc." },
    { name: "prop", type: "map", desc: "Entity properties from the Tiled map editor (custom properties on the entity object)." },
    { name: "param", type: "map", desc: "Parameters passed to this custom action via the 'params' map." },
    { name: "self", type: "string", desc: "Entity ID of the handler's owner." },
    { name: "player", type: "string", desc: "Entity ID of the player." },
    { name: "source", type: "string", desc: "Entity ID that triggered the current event." },
];

export const FUNCTIONS = [
    { name: "len(x)", desc: "Length of string, array, or map.", example: "len(var.messages)" },
    { name: "min(a, b)", desc: "Minimum of two numbers.", example: "min(var.count, 5)" },
    { name: "max(a, b)", desc: "Maximum of two numbers.", example: "max(0, var.hp - 10)" },
    { name: "clamp(v, lo, hi)", desc: "Clamp value to range [lo, hi].", example: "clamp(var.x, 0, 100)" },
    { name: "str(x)", desc: "Convert to string.", example: "str(var.count)" },
    { name: "int(x)", desc: "Convert to integer.", example: "int(var.ratio * 100)" },
    { name: "float(x)", desc: "Convert to float.", example: "float(var.level)" },
    { name: "rand(n)", desc: "Random integer in [0, n).", example: "rand(len(const.messages))" },
    { name: "randf()", desc: "Random float in [0, 1).", example: "randf() < 0.5" },
    { name: "keys(map)", desc: "Array of map keys.", example: "keys(var.inventory)" },
    { name: "values(map)", desc: "Array of map values.", example: "values(var.scores)" },
    { name: "hasKey(map, key)", desc: "True if map has the key.", example: "hasKey(global, 'quest_done')" },
];

export const OPERATORS = [
    { op: "==  !=  <  <=  >  >=", desc: "Comparison" },
    { op: "&&  ||  !", desc: "Logical AND, OR, NOT" },
    { op: "+  -  *  /  %  **", desc: "Arithmetic (** is exponent)" },
    { op: "in", desc: "Membership test: 'x' in ['a','b','x']" },
    { op: "contains", desc: "String contains: 'hello' contains 'ell'" },
    { op: "startsWith  endsWith", desc: "String prefix/suffix check" },
    { op: "matches", desc: "Regex match: 'hello' matches '^h.*o$'" },
    { op: "?:", desc: "Ternary: condition ? yes : no" },
    { op: "??", desc: "Nil coalesce: value ?? fallback" },
    { op: "[]", desc: "Index: var.list[0], var.map['key']" },
    { op: ".", desc: "Property access: save.animech.level" },
];

export const EXAMPLES = [
    { expr: 'var.count + 1', desc: "Increment a counter" },
    { expr: 'var.messages[min(var.attempts, len(var.messages) - 1)]', desc: "Index-clamped message lookup" },
    { expr: 'global.quest_stage == "complete"', desc: "Check world state" },
    { expr: 'var.hp > 0 && var.shield > 0', desc: "Compound boolean" },
    { expr: 'rand(len(const.greetings))', desc: "Random index into a const list" },
    { expr: 'save.animech.level >= 5 ? "strong" : "weak"', desc: "Ternary expression" },
    { expr: 'hasKey(var.visited, prop.zone_id)', desc: "Check if zone was visited" },
];
