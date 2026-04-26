import { useEffect, useRef } from "react";
import { EditorView, ViewPlugin, Decoration, keymap, placeholder as cmPlaceholder } from "@codemirror/view";
import type { DecorationSet, ViewUpdate } from "@codemirror/view";
import { EditorState, Compartment, RangeSetBuilder } from "@codemirror/state";
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
import { autocompletion } from "@codemirror/autocomplete";
import type { CompletionContext, CompletionResult } from "@codemirror/autocomplete";
import { linter } from "@codemirror/lint";
import type { Diagnostic } from "@codemirror/lint";
import { validateExpr } from "@/api/scripts";

// --- Theme ---

const exprTheme = EditorView.theme({
    "&": {
        fontSize: "12px",
        fontFamily: "var(--font-mono, ui-monospace, monospace)",
        backgroundColor: "transparent",
    },
    ".cm-content": {
        padding: "3px 0",
        caretColor: "oklch(0.985 0 0)",
        minHeight: "20px",
    },
    "&.cm-focused": {
        outline: "none",
    },
    ".cm-line": {
        padding: "0",
    },
    ".cm-scroller": {
        overflow: "hidden",
        lineHeight: "1.4",
    },
    ".cm-tooltip.cm-tooltip-autocomplete": {
        backgroundColor: "oklch(0.21 0.006 285.75)",
        border: "1px solid oklch(0.35 0.006 285.75)",
        borderRadius: "6px",
        fontSize: "11px",
        fontFamily: "var(--font-mono, ui-monospace, monospace)",
        maxHeight: "180px",
    },
    ".cm-tooltip-autocomplete ul li": {
        padding: "2px 8px",
    },
    ".cm-tooltip-autocomplete ul li[aria-selected]": {
        backgroundColor: "oklch(0.30 0.010 285.75)",
    },
    ".cm-completionLabel": {
        color: "oklch(0.90 0 0)",
    },
    ".cm-completionDetail": {
        color: "oklch(0.55 0 0)",
        marginLeft: "8px",
        fontStyle: "normal",
    },
    ".cm-completionMatchedText": {
        textDecoration: "none",
        color: "oklch(0.75 0.15 300)",
        fontWeight: "600",
    },
    ".cm-diagnostic-error": {
        borderLeft: "2px solid oklch(0.65 0.2 25)",
        backgroundColor: "oklch(0.25 0.04 25 / 0.3)",
        padding: "2px 6px",
        fontSize: "10px",
        borderRadius: "0 4px 4px 0",
    },
    ".cm-lintRange-error": {
        backgroundImage: "none",
        textDecoration: "underline wavy oklch(0.65 0.2 25)",
        textUnderlineOffset: "2px",
    },
    ".cm-placeholder": {
        color: "oklch(0.45 0 0)",
        fontStyle: "normal",
    },
});

// --- Token-based highlighting ---

const KEYWORDS = new Set(["true", "false", "nil", "null", "in", "not", "and", "or", "matches", "contains", "startsWith", "endsWith"]);
const FUNCTIONS = new Set(["len", "min", "max", "clamp", "str", "int", "float", "rand", "randf", "keys", "values", "hasKey"]);
const ENV_ROOTS = new Set(["var", "global", "const", "save", "prop", "param", "self", "player", "source"]);

type TokenType = "keyword" | "function" | "number" | "string" | "operator" | "envRoot" | "paren" | "bracket" | "property";

interface Token {
    from: number;
    to: number;
    type: TokenType;
}

function tokenize(text: string): Token[] {
    const tokens: Token[] = [];
    let i = 0;
    while (i < text.length) {
        if (text[i] === " " || text[i] === "\t" || text[i] === "\n") { i++; continue; }

        // Strings
        if (text[i] === '"' || text[i] === "'") {
            const quote = text[i];
            const start = i;
            i++;
            while (i < text.length && text[i] !== quote) {
                if (text[i] === "\\" && i + 1 < text.length) i++;
                i++;
            }
            if (i < text.length) i++;
            tokens.push({ from: start, to: i, type: "string" });
            continue;
        }

        // Numbers
        if (/[0-9]/.test(text[i]) || (text[i] === "." && i + 1 < text.length && /[0-9]/.test(text[i + 1]))) {
            const start = i;
            while (i < text.length && /[0-9.]/.test(text[i])) i++;
            tokens.push({ from: start, to: i, type: "number" });
            continue;
        }

        // Identifiers
        if (/[a-zA-Z_]/.test(text[i])) {
            const start = i;
            while (i < text.length && /[a-zA-Z0-9_]/.test(text[i])) i++;
            const word = text.substring(start, i);

            // "not in" - emit "not" and "in" as separate keyword tokens, skip whitespace
            if (word === "not") {
                const rest = text.substring(i);
                const m = rest.match(/^(\s+)(in)\b/);
                if (m) {
                    tokens.push({ from: start, to: i, type: "keyword" });
                    i += m[1].length;
                    tokens.push({ from: i, to: i + m[2].length, type: "keyword" });
                    i += m[2].length;
                    continue;
                }
            }

            if (KEYWORDS.has(word)) {
                tokens.push({ from: start, to: i, type: "keyword" });
            } else if (FUNCTIONS.has(word) && i < text.length && text[i] === "(") {
                tokens.push({ from: start, to: i, type: "function" });
            } else if (ENV_ROOTS.has(word)) {
                tokens.push({ from: start, to: i, type: "envRoot" });
            } else {
                tokens.push({ from: start, to: i, type: "property" });
            }
            continue;
        }

        // Operators
        if ("+-*/%!<>=&|?:.".includes(text[i])) {
            const start = i;
            if (i + 1 < text.length) {
                const two = text.substring(i, i + 2);
                if (["==", "!=", "<=", ">=", "&&", "||", "??", "**"].includes(two)) {
                    i += 2;
                    tokens.push({ from: start, to: i, type: "operator" });
                    continue;
                }
            }
            i++;
            tokens.push({ from: start, to: i, type: "operator" });
            continue;
        }

        // Parens & brackets
        if ("()".includes(text[i])) {
            tokens.push({ from: i, to: i + 1, type: "paren" });
            i++;
            continue;
        }
        if ("[]".includes(text[i])) {
            tokens.push({ from: i, to: i + 1, type: "bracket" });
            i++;
            continue;
        }

        i++;
    }
    return tokens;
}

const TOKEN_MARKS: Record<TokenType, string> = {
    keyword: "oklch(0.75 0.15 300)",
    function: "oklch(0.80 0.14 190)",
    number: "oklch(0.75 0.15 170)",
    string: "oklch(0.75 0.12 140)",
    operator: "oklch(0.70 0.15 50)",
    envRoot: "oklch(0.80 0.14 260)",
    paren: "oklch(0.55 0 0)",
    bracket: "oklch(0.55 0 0)",
    property: "oklch(0.85 0 0)",
};

const tokenDecorations = Object.fromEntries(
    Object.entries(TOKEN_MARKS).map(([type, color]) => [
        type,
        Decoration.mark({ attributes: { style: `color: ${color}` } }),
    ])
);

const highlightPlugin = ViewPlugin.fromClass(
    class {
        decorations: DecorationSet;
        constructor(view: EditorView) {
            this.decorations = this.build(view);
        }
        update(update: ViewUpdate) {
            if (update.docChanged || update.viewportChanged) {
                this.decorations = this.build(update.view);
            }
        }
        build(view: EditorView): DecorationSet {
            const builder = new RangeSetBuilder<Decoration>();
            const text = view.state.doc.toString();
            const tokens = tokenize(text);
            for (const tok of tokens) {
                const deco = tokenDecorations[tok.type];
                if (deco) builder.add(tok.from, tok.to, deco);
            }
            return builder.finish();
        }
    },
    { decorations: (v) => v.decorations }
);

// --- Autocomplete ---

interface CompletionItem {
    label: string;
    detail?: string;
    type?: string;
    boost?: number;
}

function buildCompletions(handlerVarKeys?: string[], constKeys?: string[]): CompletionItem[] {
    const items: CompletionItem[] = [];

    items.push({ label: "var", detail: "handler-local state", type: "variable", boost: 10 });
    items.push({ label: "global", detail: "world/run state", type: "variable", boost: 9 });
    items.push({ label: "const", detail: "script constants", type: "variable", boost: 8 });
    items.push({ label: "save", detail: "game save data", type: "variable", boost: 7 });
    items.push({ label: "prop", detail: "entity properties", type: "variable", boost: 6 });
    items.push({ label: "param", detail: "action parameters", type: "variable", boost: 5 });
    items.push({ label: "self", detail: "entity ID", type: "variable", boost: 4 });
    items.push({ label: "player", detail: "player entity ID", type: "variable", boost: 4 });
    items.push({ label: "source", detail: "event source ID", type: "variable", boost: 4 });

    items.push({ label: "len", detail: "(x) - length", type: "function", boost: 3 });
    items.push({ label: "min", detail: "(a, b) - minimum", type: "function", boost: 3 });
    items.push({ label: "max", detail: "(a, b) - maximum", type: "function", boost: 3 });
    items.push({ label: "clamp", detail: "(v, lo, hi)", type: "function", boost: 3 });
    items.push({ label: "str", detail: "(x) - to string", type: "function", boost: 2 });
    items.push({ label: "int", detail: "(x) - to integer", type: "function", boost: 2 });
    items.push({ label: "float", detail: "(x) - to float", type: "function", boost: 2 });
    items.push({ label: "rand", detail: "(n) - random [0,n)", type: "function", boost: 2 });
    items.push({ label: "randf", detail: "() - random [0,1)", type: "function", boost: 2 });
    items.push({ label: "keys", detail: "(map) - key list", type: "function", boost: 2 });
    items.push({ label: "values", detail: "(map) - value list", type: "function", boost: 2 });
    items.push({ label: "hasKey", detail: "(map, key) - has key?", type: "function", boost: 2 });

    items.push({ label: "true", type: "keyword", boost: 1 });
    items.push({ label: "false", type: "keyword", boost: 1 });
    items.push({ label: "nil", type: "keyword", boost: 1 });

    if (handlerVarKeys) {
        for (const k of handlerVarKeys) {
            items.push({ label: `var.${k}`, detail: "handler variable", type: "property", boost: 8 });
        }
    }

    if (constKeys) {
        for (const k of constKeys) {
            items.push({ label: `const.${k}`, detail: "constant", type: "property", boost: 6 });
        }
    }

    const savePaths = [
        "save.animech.level", "save.animech.experience",
        "save.character_name", "save.save_id",
    ];
    for (const p of savePaths) {
        items.push({ label: p, detail: "save data", type: "property", boost: 5 });
    }

    return items;
}

function exprCompletions(completionItems: CompletionItem[]) {
    return (context: CompletionContext): CompletionResult | null => {
        const word = context.matchBefore(/[a-zA-Z_][\w.]*/);
        if (!word && !context.explicit) return null;
        const from = word?.from ?? context.pos;
        const text = word?.text ?? "";

        const filtered = completionItems
            .filter((item) => item.label.toLowerCase().startsWith(text.toLowerCase()))
            .map((item) => ({
                label: item.label,
                detail: item.detail,
                type: item.type,
                boost: item.boost ?? 0,
            }));

        if (filtered.length === 0) return null;
        return { from, options: filtered };
    };
}

// --- Linting ---

function exprLinter() {
    return linter(async (view): Promise<Diagnostic[]> => {
        const text = view.state.doc.toString().trim();
        if (!text) return [];

        // Debounce by waiting, then checking if the doc changed while we slept
        const docAtStart = view.state.doc.toString();
        await new Promise((r) => setTimeout(r, 500));
        if (view.state.doc.toString() !== docAtStart) return [];

        try {
            const result = await validateExpr(text);
            // Verify doc hasn't changed during the fetch
            if (view.state.doc.toString().trim() !== text) return [];
            if (!result.valid && result.error) {
                return [{
                    from: 0,
                    to: view.state.doc.length,
                    severity: "error",
                    message: result.error,
                }];
            }
            return [];
        } catch {
            return [];
        }
    }, { delay: 0 });
}

// --- Main Component ---

interface ExpressionInputProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
    handlerVarKeys?: string[];
    constKeys?: string[];
    singleLine?: boolean;
}

export function ExpressionInput({
    value,
    onChange,
    placeholder = "expression",
    handlerVarKeys,
    constKeys,
    singleLine = true,
}: ExpressionInputProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const viewRef = useRef<EditorView | null>(null);
    const onChangeRef = useRef(onChange);
    onChangeRef.current = onChange;

    const completionItems = useRef<CompletionItem[]>([]);
    const completionCompartment = useRef(new Compartment());

    useEffect(() => {
        completionItems.current = buildCompletions(handlerVarKeys, constKeys);
    }, [handlerVarKeys, constKeys]);

    useEffect(() => {
        if (!containerRef.current) return;

        const state = EditorState.create({
            doc: value,
            extensions: [
                exprTheme,
                highlightPlugin,
                history(),
                keymap.of([
                    ...defaultKeymap,
                    ...historyKeymap,
                    ...(singleLine ? [{ key: "Enter", run: () => true }] : []),
                ]),
                completionCompartment.current.of(
                    autocompletion({
                        override: [exprCompletions(completionItems.current)],
                        activateOnTyping: true,
                        icons: false,
                    })
                ),
                exprLinter(),
                cmPlaceholder(placeholder),
                EditorView.updateListener.of((update: ViewUpdate) => {
                    if (update.docChanged) {
                        onChangeRef.current(update.state.doc.toString());
                    }
                }),
                ...(singleLine ? [EditorState.transactionFilter.of((tr) => {
                    if (tr.newDoc.lines > 1) return [];
                    return tr;
                })] : []),
            ],
        });

        const view = new EditorView({
            state,
            parent: containerRef.current,
        });
        viewRef.current = view;

        return () => {
            view.destroy();
            viewRef.current = null;
        };
    }, []);

    // Sync value from outside
    useEffect(() => {
        const view = viewRef.current;
        if (!view) return;
        const current = view.state.doc.toString();
        if (current !== value) {
            view.dispatch({
                changes: { from: 0, to: current.length, insert: value },
            });
        }
    }, [value]);

    // Update completions when context changes
    useEffect(() => {
        const view = viewRef.current;
        if (!view) return;
        view.dispatch({
            effects: completionCompartment.current.reconfigure(
                autocompletion({
                    override: [exprCompletions(completionItems.current)],
                    activateOnTyping: true,
                    icons: false,
                })
            ),
        });
    }, [handlerVarKeys, constKeys]);

    return (
        <div
            ref={containerRef}
            className="rounded border border-border bg-background px-2 min-h-[24px] flex items-center [&_.cm-editor]:flex-1 [&_.cm-editor]:min-w-0"
        />
    );
}
