import { useEffect, useRef } from "react";
import { EditorView, keymap, placeholder as cmPlaceholder, lineNumbers } from "@codemirror/view";
import type { ViewUpdate } from "@codemirror/view";
import { EditorState } from "@codemirror/state";
import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands";
import { yaml } from "@codemirror/lang-yaml";
import { linter } from "@codemirror/lint";
import type { Diagnostic } from "@codemirror/lint";
import { search, searchKeymap, openSearchPanel } from "@codemirror/search";
import { indentOnInput, bracketMatching, foldGutter, foldKeymap } from "@codemirror/language";
import { parse as yamlParse } from "yaml";

const yamlEditorTheme = EditorView.theme({
    "&": {
        fontSize: "12px",
        fontFamily: "var(--font-mono, ui-monospace, monospace)",
        backgroundColor: "transparent",
        height: "100%",
    },
    ".cm-content": {
        padding: "8px 0",
        caretColor: "oklch(0.15 0 0)",
    },
    "&.cm-focused": {
        outline: "none",
    },
    ".cm-scroller": {
        overflow: "auto",
        lineHeight: "1.5",
    },
    ".cm-gutters": {
        backgroundColor: "oklch(0.97 0 0)",
        borderRight: "1px solid oklch(0.90 0 0)",
        color: "oklch(0.55 0 0)",
        fontSize: "10px",
        minWidth: "32px",
    },
    ".cm-activeLineGutter": {
        backgroundColor: "oklch(0.94 0.01 260)",
        color: "oklch(0.35 0 0)",
    },
    ".cm-activeLine": {
        backgroundColor: "oklch(0.97 0.005 260)",
    },
    ".cm-cursor": {
        borderLeftColor: "oklch(0.15 0 0)",
    },
    ".cm-selectionBackground": {
        backgroundColor: "oklch(0.90 0.03 260) !important",
    },
    "&.cm-focused .cm-selectionBackground": {
        backgroundColor: "oklch(0.87 0.05 260) !important",
    },
    ".cm-foldGutter .cm-gutterElement": {
        padding: "0 2px",
        cursor: "pointer",
    },
    ".cm-matchingBracket": {
        backgroundColor: "oklch(0.90 0.08 170)",
        outline: "1px solid oklch(0.75 0.10 170)",
    },
    // Search panel
    ".cm-panels": {
        backgroundColor: "oklch(0.97 0 0)",
        borderBottom: "1px solid oklch(0.88 0 0)",
        color: "oklch(0.15 0 0)",
    },
    ".cm-panel.cm-search": {
        padding: "4px 8px",
        fontSize: "12px",
    },
    ".cm-panel.cm-search input": {
        fontSize: "12px",
        border: "1px solid oklch(0.85 0 0)",
        borderRadius: "4px",
        padding: "2px 6px",
        outline: "none",
    },
    ".cm-panel.cm-search input:focus": {
        borderColor: "oklch(0.65 0.10 260)",
    },
    ".cm-panel.cm-search button": {
        fontSize: "11px",
        borderRadius: "4px",
        padding: "2px 8px",
        cursor: "pointer",
        backgroundColor: "oklch(0.94 0 0)",
        border: "1px solid oklch(0.85 0 0)",
    },
    ".cm-panel.cm-search button:hover": {
        backgroundColor: "oklch(0.90 0 0)",
    },
    ".cm-panel.cm-search label": {
        fontSize: "11px",
    },
    ".cm-searchMatch": {
        backgroundColor: "oklch(0.90 0.12 85)",
        outline: "1px solid oklch(0.80 0.12 85)",
    },
    ".cm-searchMatch-selected": {
        backgroundColor: "oklch(0.82 0.15 85)",
    },
    // Lint
    ".cm-diagnostic-error": {
        borderLeft: "2px solid oklch(0.55 0.2 25)",
        backgroundColor: "oklch(0.95 0.03 25)",
        padding: "2px 6px",
        fontSize: "10px",
        borderRadius: "0 4px 4px 0",
    },
    ".cm-diagnostic-warning": {
        borderLeft: "2px solid oklch(0.60 0.18 85)",
        backgroundColor: "oklch(0.96 0.03 85)",
        padding: "2px 6px",
        fontSize: "10px",
        borderRadius: "0 4px 4px 0",
    },
    ".cm-lintRange-error": {
        backgroundImage: "none",
        textDecoration: "underline wavy oklch(0.55 0.2 25)",
        textUnderlineOffset: "2px",
    },
    ".cm-lintRange-warning": {
        backgroundImage: "none",
        textDecoration: "underline wavy oklch(0.60 0.18 85)",
        textUnderlineOffset: "2px",
    },
    ".cm-placeholder": {
        color: "oklch(0.55 0 0)",
        fontStyle: "normal",
    },
});

function yamlLinter() {
    return linter((view): Diagnostic[] => {
        const text = view.state.doc.toString();
        if (!text.trim()) return [];
        try {
            yamlParse(text, { strict: true });
            return [];
        } catch (e: unknown) {
            const err = e as { message?: string; linePos?: [{ line: number; col: number }] };
            const msg = err.message ?? "Invalid YAML";
            let from = 0;
            let to = Math.min(text.length, 80);
            if (err.linePos?.[0]) {
                const line = Math.min(err.linePos[0].line, view.state.doc.lines);
                const lineInfo = view.state.doc.line(line);
                from = lineInfo.from;
                to = lineInfo.to;
            }
            return [{ from, to, severity: "error", message: msg }];
        }
    }, { delay: 300 });
}

interface YamlEditorProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
    readOnly?: boolean;
}

export function YamlEditor({
    value,
    onChange,
    placeholder = "",
    readOnly = false,
}: YamlEditorProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const viewRef = useRef<EditorView | null>(null);
    const onChangeRef = useRef(onChange);
    onChangeRef.current = onChange;

    useEffect(() => {
        if (!containerRef.current) return;

        const state = EditorState.create({
            doc: value,
            extensions: [
                yamlEditorTheme,
                lineNumbers(),
                foldGutter(),
                history(),
                indentOnInput(),
                bracketMatching(),
                yaml(),
                search(),
                yamlLinter(),
                keymap.of([
                    ...defaultKeymap,
                    ...historyKeymap,
                    ...searchKeymap,
                    ...foldKeymap,
                    indentWithTab,
                    { key: "Mod-f", run: openSearchPanel },
                ]),
                EditorView.updateListener.of((update: ViewUpdate) => {
                    if (update.docChanged) {
                        onChangeRef.current(update.state.doc.toString());
                    }
                }),
                ...(placeholder ? [cmPlaceholder(placeholder)] : []),
                ...(readOnly ? [EditorState.readOnly.of(true)] : []),
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

    return (
        <div
            ref={containerRef}
            className="h-full w-full overflow-hidden [&_.cm-editor]:h-full"
        />
    );
}
