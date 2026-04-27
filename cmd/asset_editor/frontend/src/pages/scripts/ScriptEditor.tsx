import { useState, useEffect, useCallback, useMemo } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { useScript, useScripts, useSaveScript, useScriptSchema } from "@/api/scripts";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { YamlEditor } from "@/components/ui/yaml-editor";
import { ArrowLeft, Save, GitCompare, X } from "lucide-react";
import { HandlerList } from "@/components/scripts/HandlerList";
import { HandlerDetail } from "@/components/scripts/HandlerDetail";
import { CustomActionDetail } from "@/components/scripts/CustomActionDetail";
import { SequenceDetail } from "@/components/scripts/SequenceDetail";
import { ConstDetail } from "@/components/scripts/ConstDetail";
import { PropertyTemplateDetail } from "@/components/scripts/PropertyTemplateDetail";
import { ExprContextProvider } from "@/components/scripts/ExprContext";
import { ScriptClipboardProvider } from "@/components/scripts/ScriptClipboard";
import { parseScript, stringifyScript, normalizeYaml } from "@/lib/scriptUtils";
import { usePageTitle } from "@/hooks/usePageTitle";
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";
import type { ParsedScript, ScriptItemSelection, ScriptSchema } from "@/types/scripts";
import type { CrossFileEntry } from "@/components/scripts/ExprContext";

function renderDetail(
    parsed: ParsedScript,
    selection: ScriptItemSelection | null,
    schema: ScriptSchema,
    onChange: (script: ParsedScript) => void,
) {
    if (!selection) {
        return (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
                Select an item to edit.
            </div>
        );
    }
    switch (selection.type) {
        case "handler": {
            const handler = parsed.handlers[selection.name];
            if (!handler) return null;
            return (
                <HandlerDetail
                    handlerName={selection.name}
                    handler={handler}
                    schema={schema}
                    onChange={(h) => onChange({ ...parsed, handlers: { ...parsed.handlers, [selection.name]: h } })}
                />
            );
        }
        case "custom_action": {
            const action = parsed.custom_actions?.[selection.name];
            if (!action) return null;
            return (
                <CustomActionDetail
                    name={selection.name}
                    action={action}
                    schema={schema}
                    onChange={(a) => onChange({ ...parsed, custom_actions: { ...parsed.custom_actions, [selection.name]: a } })}
                />
            );
        }
        case "sequence": {
            const seq = parsed.sequences?.[selection.name];
            if (!seq) return null;
            return (
                <SequenceDetail
                    name={selection.name}
                    sequence={seq}
                    schema={schema}
                    onChange={(s) => onChange({ ...parsed, sequences: { ...parsed.sequences, [selection.name]: s } })}
                />
            );
        }
        case "const": {
            const value = parsed.consts?.[selection.name];
            if (value === undefined) return null;
            return (
                <ConstDetail
                    name={selection.name}
                    value={value}
                    onChange={(v) => onChange({ ...parsed, consts: { ...parsed.consts, [selection.name]: v } })}
                />
            );
        }
        case "property_template": {
            const template = parsed.property_templates?.[selection.name];
            if (!template) return null;
            return (
                <PropertyTemplateDetail
                    name={selection.name}
                    template={template}
                    onChange={(t) => onChange({ ...parsed, property_templates: { ...parsed.property_templates, [selection.name]: t } })}
                />
            );
        }
    }
}

interface DiffLine {
    type: "context" | "add" | "remove";
    text: string;
    oldNum?: number;
    newNum?: number;
}

function computeUnifiedDiff(oldText: string, newText: string): DiffLine[] {
    const oldLines = oldText.split("\n");
    const newLines = newText.split("\n");

    const n = oldLines.length;
    const m = newLines.length;
    const max = n + m;
    const v = new Int32Array(2 * max + 1);
    const trace: Int32Array[] = [];

    for (let d = 0; d <= max; d++) {
        trace.push(v.slice());
        for (let k = -d; k <= d; k += 2) {
            let x: number;
            if (k === -d || (k !== d && v[k - 1 + max] < v[k + 1 + max])) {
                x = v[k + 1 + max];
            } else {
                x = v[k - 1 + max] + 1;
            }
            let y = x - k;
            while (x < n && y < m && oldLines[x] === newLines[y]) {
                x++;
                y++;
            }
            v[k + max] = x;
            if (x >= n && y >= m) {
                const edits: DiffLine[] = [];
                let cx = n, cy = m;
                for (let dd = d; dd > 0; dd--) {
                    const vv = trace[dd];
                    const kk = cx - cy;
                    const isDown = kk === -dd || (kk !== dd && vv[kk - 1 + max] < vv[kk + 1 + max]);
                    const prevK = isDown ? kk + 1 : kk - 1;
                    const endX = vv[prevK + max];
                    const endY = endX - prevK;
                    while (cx > endX && cy > endY) {
                        cx--;
                        cy--;
                        edits.push({ type: "context", text: oldLines[cx], oldNum: cx + 1, newNum: cy + 1 });
                    }
                    if (isDown) {
                        cy--;
                        edits.push({ type: "add", text: newLines[cy], newNum: cy + 1 });
                    } else {
                        cx--;
                        edits.push({ type: "remove", text: oldLines[cx], oldNum: cx + 1 });
                    }
                }
                while (cx > 0 && cy > 0) {
                    cx--;
                    cy--;
                    edits.push({ type: "context", text: oldLines[cx], oldNum: cx + 1, newNum: cy + 1 });
                }
                edits.reverse();

                const contextLines = 3;
                const result: DiffLine[] = [];
                let lastShown = -1;
                const changeIndices = edits.map((e, i) => e.type !== "context" ? i : -1).filter(i => i >= 0);
                if (changeIndices.length === 0) return [];

                for (const ci of changeIndices) {
                    const start = Math.max(0, ci - contextLines);
                    const end = Math.min(edits.length - 1, ci + contextLines);
                    if (start > lastShown + 1) {
                        result.push({ type: "context", text: "···" });
                    }
                    for (let j = Math.max(start, lastShown + 1); j <= end; j++) {
                        result.push(edits[j]);
                    }
                    lastShown = end;
                }
                return result;
            }
        }
    }
    return [];
}

export function ScriptEditor() {
    const location = useLocation();
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const path = location.pathname.replace(/^\/scripts\//, "");

    const { data: script, isLoading, error } = useScript(path);
    const { data: schema } = useScriptSchema();
    const { data: allScripts } = useScripts();
    const saveScript = useSaveScript();

    const [rawContent, setRawContent] = useState("");
    const [parsed, setParsed] = useState<ParsedScript>({ handlers: {} });
    const [parseError, setParseError] = useState<string | null>(null);
    const [dirty, setDirty] = useState(false);
    const [activeTab, setActiveTab] = useState<string>(() => searchParams.get("tab") ?? "structured");
    const [selection, setSelection] = useState<ScriptItemSelection | null>(() => {
        const handler = searchParams.get("handler");
        return handler ? { type: "handler", name: handler } : null;
    });
    const [showDiff, setShowDiff] = useState(false);

    const fileName = path.split("/").pop()?.replace(/\.ya?ml$/, "") ?? path;
    usePageTitle(`${fileName} - Scripts`);
    useUnsavedChanges(dirty);

    const { otherConstKeys, otherCustomActionNames } = useMemo(() => {
        if (!allScripts) return { otherConstKeys: [] as CrossFileEntry[], otherCustomActionNames: [] as CrossFileEntry[] };
        const localConsts = new Set(parsed.consts ? Object.keys(parsed.consts) : []);
        const localActions = new Set(parsed.custom_actions ? Object.keys(parsed.custom_actions) : []);
        const consts: CrossFileEntry[] = [];
        const actions: CrossFileEntry[] = [];
        for (const s of allScripts) {
            if (s.path === path) continue;
            for (const name of s.constNames ?? []) {
                if (!localConsts.has(name)) consts.push({ name, file: s.path });
            }
            for (const name of s.customActionNames ?? []) {
                if (!localActions.has(name)) actions.push({ name, file: s.path });
            }
        }
        return { otherConstKeys: consts, otherCustomActionNames: actions };
    }, [allScripts, path, parsed.consts, parsed.custom_actions]);

    useEffect(() => {
        if (script) {
            setRawContent(script.rawYaml);
            try {
                const p = parseScript(script.rawYaml);
                setParsed(p);
                setParseError(null);
                const urlHandler = searchParams.get("handler");
                if (urlHandler && p.handlers[urlHandler]) {
                    setSelection({ type: "handler", name: urlHandler });
                } else {
                    const firstHandler = Object.keys(p.handlers)[0];
                    setSelection(firstHandler ? { type: "handler", name: firstHandler } : null);
                }
            } catch (e) {
                setParseError(String(e));
            }
            setDirty(false);
        }
    }, [script]);

    const handleSave = useCallback(() => {
        const content = activeTab === "structured" ? stringifyScript(parsed) : rawContent;
        saveScript.mutate(
            { path, content },
            { onSuccess: () => {
                setDirty(false);
                if (activeTab === "structured") {
                    setRawContent(content);
                }
            }},
        );
    }, [activeTab, parsed, rawContent, path, saveScript]);

    const handleKeyDown = useCallback((e: KeyboardEvent) => {
        if ((e.metaKey || e.ctrlKey) && e.key === "s") {
            e.preventDefault();
            if (dirty) handleSave();
        }
    }, [dirty, handleSave]);

    useEffect(() => {
        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [handleKeyDown]);

    useEffect(() => {
        const params = new URLSearchParams(searchParams);
        let changed = false;
        const handlerName = selection?.type === "handler" ? selection.name : null;
        if (handlerName) {
            if (params.get("handler") !== handlerName) { params.set("handler", handlerName); changed = true; }
        } else {
            if (params.has("handler")) { params.delete("handler"); changed = true; }
        }
        if (activeTab !== "structured") {
            if (params.get("tab") !== activeTab) { params.set("tab", activeTab); changed = true; }
        } else {
            if (params.has("tab")) { params.delete("tab"); changed = true; }
        }
        if (changed) setSearchParams(params, { replace: true });
    }, [selection, activeTab]);

    const handleTabChange = (tab: string) => {
        if (tab === "structured" && activeTab === "raw") {
            try {
                const p = parseScript(rawContent);
                setParsed(p);
                setParseError(null);
                if (selection?.type === "handler" && !p.handlers[selection.name]) {
                    const firstHandler = Object.keys(p.handlers)[0];
                    setSelection(firstHandler ? { type: "handler", name: firstHandler } : null);
                }
            } catch (e) {
                setParseError(String(e));
                return;
            }
        } else if (tab === "raw" && activeTab === "structured") {
            setRawContent(stringifyScript(parsed));
        }
        setActiveTab(tab);
    };

    const handleStructuredChange = (updated: ParsedScript) => {
        setParsed(updated);
        setDirty(true);
    };

    const diffLines = useMemo(() => {
        if (!showDiff || !script) return [];
        const currentContent = activeTab === "structured" ? stringifyScript(parsed) : rawContent;
        return computeUnifiedDiff(normalizeYaml(script.rawYaml), normalizeYaml(currentContent));
    }, [showDiff, script, activeTab, parsed, rawContent]);

    if (isLoading) return <div className="p-4 text-muted-foreground">Loading...</div>;
    if (error) return <div className="p-4 text-destructive">Error: {error.message}</div>;

    return (
        <div className="flex h-full flex-col">
            <div className="flex items-center gap-2 border-b border-border px-3 py-2">
                <Button variant="ghost" size="sm" onClick={() => navigate("/scripts")}>
                    <ArrowLeft className="mr-1 h-3 w-3" />
                    Back
                </Button>
                <div className="flex-1">
                    <span className="text-sm font-medium">{path}</span>
                    {script && (
                        <span className="ml-2 text-xs text-muted-foreground">
                            {script.handlerCount} handler{script.handlerCount !== 1 ? "s" : ""}
                        </span>
                    )}
                </div>
                {dirty && (
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setShowDiff(!showDiff)}
                    >
                        <GitCompare className="mr-1 h-3 w-3" />
                        {showDiff ? "Hide diff" : "Show diff"}
                    </Button>
                )}
                <Button
                    size="sm"
                    disabled={!dirty || saveScript.isPending}
                    onClick={handleSave}
                >
                    <Save className="mr-1 h-3 w-3" />
                    {saveScript.isPending ? "Saving..." : "Save"}
                </Button>
            </div>

            {showDiff && (
                <div className="border-b border-border bg-muted/30 max-h-[40vh] overflow-auto">
                    <div className="flex items-center justify-between px-3 py-1.5 border-b border-border/50 sticky top-0 bg-muted/50 backdrop-blur-sm">
                        <span className="text-xs font-medium text-muted-foreground">Unsaved changes</span>
                        <button className="text-muted-foreground hover:text-foreground" onClick={() => setShowDiff(false)}>
                            <X className="h-3.5 w-3.5" />
                        </button>
                    </div>
                    {diffLines.length === 0 ? (
                        <div className="px-3 py-4 text-xs text-muted-foreground text-center">No changes</div>
                    ) : (
                        <pre className="text-xs font-mono leading-relaxed">
                            {diffLines.map((line, i) => (
                                <div
                                    key={i}
                                    className={
                                        line.type === "add" ? "bg-accent-green-tint text-accent-green" :
                                        line.type === "remove" ? "bg-accent-red-tint text-accent-red" :
                                        line.text === "···" ? "text-muted-foreground/40 text-center" :
                                        "text-muted-foreground"
                                    }
                                >
                                    <span className="inline-block w-8 text-right text-muted-foreground/40 select-none pr-1">
                                        {line.oldNum ?? ""}
                                    </span>
                                    <span className="inline-block w-8 text-right text-muted-foreground/40 select-none pr-2">
                                        {line.newNum ?? ""}
                                    </span>
                                    <span className="select-none">{line.type === "add" ? "+" : line.type === "remove" ? "-" : " "}</span>
                                    {line.text}
                                </div>
                            ))}
                        </pre>
                    )}
                </div>
            )}

            <Tabs value={activeTab} onValueChange={handleTabChange} className="flex flex-1 flex-col overflow-hidden min-h-0">
                <div className="border-b border-border px-3">
                    <TabsList className="h-8">
                        <TabsTrigger value="structured" className="text-xs px-3 py-1">Structured</TabsTrigger>
                        <TabsTrigger value="raw" className="text-xs px-3 py-1">Raw YAML</TabsTrigger>
                    </TabsList>
                    {parseError && activeTab === "structured" && (
                        <span className="ml-2 text-xs text-destructive">Parse error: {parseError}</span>
                    )}
                </div>

                <TabsContent value="structured" className="flex-1 overflow-hidden m-0 p-0 min-h-0">
                    {schema ? (
                        <ScriptClipboardProvider>
                        <ExprContextProvider
                            handlerVarKeys={selection?.type === "handler" && parsed.handlers[selection.name]?.var ? Object.keys(parsed.handlers[selection.name].var!) : []}
                            constKeys={parsed.consts ? Object.keys(parsed.consts) : []}
                            otherConstKeys={otherConstKeys}
                            otherCustomActionNames={otherCustomActionNames}
                            customActions={parsed.custom_actions}
                        >
                            <div className="flex h-full min-h-0 overflow-hidden">
                                <div className="w-64 shrink-0 border-r border-border overflow-hidden">
                                    <HandlerList
                                        script={parsed}
                                        selection={selection}
                                        onSelect={setSelection}
                                        onChange={handleStructuredChange}
                                    />
                                </div>
                                <div className="flex-1 overflow-hidden">
                                    {renderDetail(parsed, selection, schema, handleStructuredChange)}
                                </div>
                            </div>
                        </ExprContextProvider>
                        </ScriptClipboardProvider>
                    ) : (
                        <div className="p-4 text-muted-foreground">Loading schema...</div>
                    )}
                </TabsContent>

                <TabsContent value="raw" className="flex-1 overflow-hidden m-0 p-0 min-h-0">
                    <YamlEditor
                        value={rawContent}
                        onChange={(v) => {
                            setRawContent(v);
                            setDirty(true);
                        }}
                    />
                </TabsContent>
            </Tabs>
        </div>
    );
}
