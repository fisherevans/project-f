import { useState, useEffect, useCallback } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useScript, useSaveScript, useScriptSchema } from "@/api/scripts";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ArrowLeft, Save } from "lucide-react";
import { HandlerList } from "@/components/scripts/HandlerList";
import { HandlerDetail } from "@/components/scripts/HandlerDetail";
import { parseScript, stringifyScript } from "@/lib/scriptUtils";
import type { ParsedScript } from "@/types/scripts";

export function ScriptEditor() {
    const location = useLocation();
    const navigate = useNavigate();
    const path = location.pathname.replace(/^\/scripts\//, "");

    const { data: script, isLoading, error } = useScript(path);
    const { data: schema } = useScriptSchema();
    const saveScript = useSaveScript();

    const [rawContent, setRawContent] = useState("");
    const [parsed, setParsed] = useState<ParsedScript>({ handlers: {} });
    const [parseError, setParseError] = useState<string | null>(null);
    const [dirty, setDirty] = useState(false);
    const [activeTab, setActiveTab] = useState<string>("structured");
    const [selectedHandler, setSelectedHandler] = useState<string | null>(null);

    useEffect(() => {
        if (script) {
            setRawContent(script.rawYaml);
            try {
                const p = parseScript(script.rawYaml);
                setParsed(p);
                setParseError(null);
                const firstHandler = Object.keys(p.handlers)[0] ?? null;
                setSelectedHandler(firstHandler);
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
            { onSuccess: () => setDirty(false) },
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

    const handleTabChange = (tab: string) => {
        if (tab === "structured" && activeTab === "raw") {
            try {
                const p = parseScript(rawContent);
                setParsed(p);
                setParseError(null);
                if (!selectedHandler || !p.handlers[selectedHandler]) {
                    setSelectedHandler(Object.keys(p.handlers)[0] ?? null);
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
                <Button
                    size="sm"
                    disabled={!dirty || saveScript.isPending}
                    onClick={handleSave}
                >
                    <Save className="mr-1 h-3 w-3" />
                    {saveScript.isPending ? "Saving..." : "Save"}
                </Button>
            </div>

            <Tabs value={activeTab} onValueChange={handleTabChange} className="flex flex-1 flex-col overflow-hidden">
                <div className="border-b border-border px-3">
                    <TabsList className="h-8">
                        <TabsTrigger value="structured" className="text-xs px-3 py-1">Structured</TabsTrigger>
                        <TabsTrigger value="raw" className="text-xs px-3 py-1">Raw YAML</TabsTrigger>
                    </TabsList>
                    {parseError && activeTab === "structured" && (
                        <span className="ml-2 text-xs text-destructive">Parse error: {parseError}</span>
                    )}
                </div>

                <TabsContent value="structured" className="flex-1 overflow-hidden m-0 p-0">
                    {schema ? (
                        <div className="flex h-full">
                            <div className="w-64 shrink-0 border-r border-border overflow-hidden">
                                <HandlerList
                                    script={parsed}
                                    selectedHandler={selectedHandler}
                                    onSelect={setSelectedHandler}
                                    onChange={handleStructuredChange}
                                />
                            </div>
                            <div className="flex-1 overflow-hidden">
                                {selectedHandler && parsed.handlers[selectedHandler] ? (
                                    <HandlerDetail
                                        handlerName={selectedHandler}
                                        handler={parsed.handlers[selectedHandler]}
                                        schema={schema}
                                        onChange={(handler) => {
                                            handleStructuredChange({
                                                ...parsed,
                                                handlers: { ...parsed.handlers, [selectedHandler]: handler },
                                            });
                                        }}
                                    />
                                ) : (
                                    <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
                                        {Object.keys(parsed.handlers).length === 0
                                            ? "No handlers. Add one from the left panel."
                                            : "Select a handler to edit."}
                                    </div>
                                )}
                            </div>
                        </div>
                    ) : (
                        <div className="p-4 text-muted-foreground">Loading schema...</div>
                    )}
                </TabsContent>

                <TabsContent value="raw" className="flex-1 overflow-hidden m-0 p-0">
                    <textarea
                        className="h-full w-full resize-none bg-background p-4 font-mono text-xs leading-relaxed outline-none"
                        value={rawContent}
                        onChange={(e) => {
                            setRawContent(e.target.value);
                            setDirty(true);
                        }}
                        spellCheck={false}
                    />
                </TabsContent>
            </Tabs>
        </div>
    );
}
