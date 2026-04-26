import { useState, useEffect, useCallback, useMemo } from "react";
import { Link } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import YAML from "yaml";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useScripts, useScript, useSaveScript, useScriptSchema, useTiledUsages } from "@/api/scripts";
import { useTiledBridgeStatus, sendTiledCommand, tiledBridgeKeys } from "@/api/tiled-bridge";
import type { TiledSelection, TiledSelectedObject } from "@/api/tiled-bridge";
import type { HandlerPropDef, ScriptFileEntry, ScriptFileDetail, ParsedScript, HandlerDef, ScriptSchema } from "@/types/scripts";
import { parseScript, stringifyScript } from "@/lib/scriptUtils";
import { apiFetch } from "@/api/client";
import { HandlerDetail } from "@/components/scripts/HandlerDetail";
import { ExprContextProvider } from "@/components/scripts/ExprContext";
import type { CrossFileEntry } from "@/components/scripts/ExprContext";
import {
    Circle,
    ExternalLink,
    MapPin,
    AlertTriangle,
    Check,
    Plug,
    PlugZap,
    ChevronDown,
    ChevronRight,
    Save,
    Users,
    Crosshair,
    Plus,
    Pencil,
} from "lucide-react";

export function TiledCompanion() {
    const [selection, setSelection] = useState<TiledSelection | null>(null);
    const { data: status } = useTiledBridgeStatus();
    const queryClient = useQueryClient();

    // WebSocket listener for real-time selection updates
    useEffect(() => {
        const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        const ws = new WebSocket(`${protocol}//${window.location.host}/api/v1/ws`);

        ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                if (msg.type === "tiled-selection") {
                    setSelection(msg.data);
                    queryClient.invalidateQueries({ queryKey: tiledBridgeKeys.status() });
                }
            } catch {
                // ignore
            }
        };

        ws.onclose = () => {
            setTimeout(() => {
                // Reconnect handled by parent WebSocketProvider
            }, 3000);
        };

        return () => ws.close();
    }, [queryClient]);

    const connected = status?.connected ?? false;
    const objects = selection?.objects ?? [];
    const mapFile = selection?.mapFile ?? "";

    return (
        <div className="flex h-full flex-col overflow-hidden">
            <Header connected={connected} mapFile={mapFile} objectCount={objects.length} />
            <div className="flex-1 overflow-hidden">
                {objects.length === 0 ? (
                    <EmptyState connected={connected} />
                ) : objects.length === 1 ? (
                    <SingleObjectView object={objects[0]} />
                ) : (
                    <MultiObjectView objects={objects} />
                )}
            </div>
        </div>
    );
}

function Header({ connected, mapFile, objectCount }: { connected: boolean; mapFile: string; objectCount: number }) {
    return (
        <div className="flex items-center gap-3 border-b border-border px-4 py-2">
            <div className="flex items-center gap-1.5">
                {connected ? (
                    <PlugZap className="h-3.5 w-3.5 text-accent-green" />
                ) : (
                    <Plug className="h-3.5 w-3.5 text-muted-foreground" />
                )}
                <span className="text-xs font-medium">
                    {connected ? "Connected" : "Waiting for Tiled"}
                </span>
            </div>
            {mapFile && (
                <span className="text-[11px] font-mono text-muted-foreground truncate">
                    {mapFile}
                </span>
            )}
            <span className="ml-auto text-[10px] text-muted-foreground">
                {objectCount > 0 && `${objectCount} selected`}
            </span>
        </div>
    );
}

function EmptyState({ connected }: { connected: boolean }) {
    return (
        <div className="flex h-full items-center justify-center">
            <div className="text-center text-sm text-muted-foreground space-y-2">
                {connected ? (
                    <>
                        <MapPin className="h-8 w-8 mx-auto opacity-40" />
                        <p>Select an entity in Tiled to inspect it here.</p>
                    </>
                ) : (
                    <>
                        <Plug className="h-8 w-8 mx-auto opacity-40" />
                        <p>Tiled companion bridge not connected.</p>
                        <p className="text-xs">
                            Make sure the <code className="font-mono bg-muted px-1 rounded">companion_bridge.js</code> plugin
                            is installed and enabled in Tiled.
                        </p>
                    </>
                )}
            </div>
        </div>
    );
}

function MultiObjectView({ objects }: { objects: TiledSelectedObject[] }) {
    return (
        <ScrollArea className="h-full">
            <div className="p-4 space-y-2">
                <p className="text-xs text-muted-foreground">{objects.length} objects selected</p>
                {objects.map((obj) => (
                    <div key={obj.id} className="border border-border rounded-md p-2 space-y-1">
                        <div className="flex items-center gap-2">
                            <span className="text-xs font-mono font-medium">#{obj.id}</span>
                            {obj.className && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-accent-teal-tint text-accent-teal">
                                    {obj.className}
                                </span>
                            )}
                            {obj.properties.script_ref && (
                                <span className="text-[10px] font-mono text-accent-violet">
                                    {obj.properties.script_ref}
                                </span>
                            )}
                        </div>
                        {obj.properties.entity_id && (
                            <span className="text-[10px] font-mono text-muted-foreground">
                                {obj.properties.entity_id}
                            </span>
                        )}
                    </div>
                ))}
            </div>
        </ScrollArea>
    );
}

function SingleObjectView({ object }: { object: TiledSelectedObject }) {
    const scriptRef = object.properties.script_ref || "";
    const zoneId = object.properties.zone_id || "";
    const { data: scripts } = useScripts();

    // Find which script file contains this handler
    const scriptPath = useMemo(() => {
        if (!scriptRef || !scripts) return null;
        for (const entry of scripts) {
            if (entry.handlerNames.includes(scriptRef)) {
                return entry.path;
            }
        }
        return null;
    }, [scriptRef, scripts]);

    const isZone = object.className === "Zone" || !!zoneId;

    return (
        <ScrollArea className="h-full">
            <div className="p-4 space-y-4">
                <ObjectIdentity object={object} />
                {isZone && <ZoneSection object={object} zoneId={zoneId} scripts={scripts ?? []} />}
                {!isZone && (
                    <HandlerSection
                        object={object}
                        scriptRef={scriptRef}
                        scriptPath={scriptPath}
                        scripts={scripts ?? []}
                    />
                )}
                {!isZone && scriptPath && <HandlerPropsSection object={object} scriptPath={scriptPath} scriptRef={scriptRef} />}
                {!isZone && scriptPath && <HandlerDefinitionSection scriptPath={scriptPath} scriptRef={scriptRef} scripts={scripts ?? []} />}
                <StructuredConfigSection object={object} />
                <ClassPropertiesSection object={object} />
                <RelatedEntitiesSection object={object} scriptRef={scriptRef} />
                <AllPropertiesSection object={object} />
            </div>
        </ScrollArea>
    );
}

const ENTITY_CLASSES = ["ModeBasedEntity", "NPC", "DirectInteraction", "ShadowMob", "Zone"];

function ObjectIdentity({ object }: { object: TiledSelectedObject }) {
    const { data: usages } = useTiledUsages();
    const [editingId, setEditingId] = useState(false);
    const [editingClass, setEditingClass] = useState(false);
    const [idValue, setIdValue] = useState(object.properties.entity_id ?? "");
    const [classValue, setClassValue] = useState(object.className);

    useEffect(() => {
        setIdValue(object.properties.entity_id ?? "");
        setEditingId(false);
    }, [object.properties.entity_id]);

    useEffect(() => {
        setClassValue(object.className);
        setEditingClass(false);
    }, [object.className]);

    const idConflict = useMemo(() => {
        const currentId = idValue.trim();
        if (!currentId || !usages) return null;
        for (const usage of usages) {
            for (const ent of usage.entities) {
                if (ent.entityId === currentId && ent.objectId !== object.id) {
                    return { mapFile: ent.mapFile, objectId: ent.objectId };
                }
            }
        }
        return null;
    }, [usages, idValue, object.id]);

    const commitEntityId = useCallback((val: string) => {
        const trimmed = val.trim();
        if (trimmed === (object.properties.entity_id ?? "")) {
            setEditingId(false);
            return;
        }
        if (trimmed) {
            sendTiledCommand({ objectId: object.id, action: "setProperty", name: "entity_id", value: trimmed });
        } else {
            sendTiledCommand({ objectId: object.id, action: "removeProperty", name: "entity_id" });
        }
        setEditingId(false);
    }, [object.id, object.properties.entity_id]);

    const commitClass = useCallback((val: string) => {
        if (val === object.className) {
            setEditingClass(false);
            return;
        }
        sendTiledCommand({ objectId: object.id, action: "setProperty", name: "class", value: val });
        setEditingClass(false);
    }, [object.id, object.className]);

    return (
        <div className="space-y-1.5">
            <div className="flex items-center gap-2">
                <span className="text-sm font-semibold">Object #{object.id}</span>
                <p className="text-[10px] text-muted-foreground">
                    ({Math.round(object.x)}, {Math.round(object.y)})
                    {object.width > 0 && ` ${Math.round(object.width)}x${Math.round(object.height)}`}
                </p>
            </div>

            {/* Entity class */}
            <div className="flex items-center gap-1.5">
                <span className="text-[10px] text-muted-foreground w-10 shrink-0">class</span>
                {!editingClass ? (
                    <div className="flex items-center gap-1.5 flex-1 min-w-0">
                        {object.className ? (
                            <span className="text-[10px] px-1.5 py-0.5 rounded bg-accent-teal-tint text-accent-teal font-medium">
                                {object.className}
                            </span>
                        ) : (
                            <span className="text-[10px] text-muted-foreground italic">none</span>
                        )}
                        <button className="text-muted-foreground hover:text-foreground" onClick={() => setEditingClass(true)}>
                            <Pencil className="h-2.5 w-2.5" />
                        </button>
                    </div>
                ) : (
                    <div className="flex items-center gap-1 flex-1">
                        <select
                            className="h-6 flex-1 text-xs font-mono rounded border border-border bg-background px-1"
                            value={classValue}
                            autoFocus
                            onChange={(e) => { setClassValue(e.target.value); commitClass(e.target.value); }}
                            onBlur={() => setEditingClass(false)}
                            onKeyDown={(e) => { if (e.key === "Escape") setEditingClass(false); }}
                        >
                            <option value="">-- none --</option>
                            {ENTITY_CLASSES.map((c) => <option key={c} value={c}>{c}</option>)}
                        </select>
                    </div>
                )}
            </div>

            {/* Entity ID */}
            <div className="flex items-center gap-1.5">
                <span className="text-[10px] text-muted-foreground w-10 shrink-0">id</span>
                {!editingId ? (
                    <div className="flex items-center gap-1.5 flex-1 min-w-0">
                        {object.properties.entity_id ? (
                            <span className="text-xs font-mono truncate">{object.properties.entity_id}</span>
                        ) : (
                            <span className="text-[10px] text-muted-foreground italic">none</span>
                        )}
                        <button className="text-muted-foreground hover:text-foreground" onClick={() => setEditingId(true)}>
                            <Pencil className="h-2.5 w-2.5" />
                        </button>
                    </div>
                ) : (
                    <div className="flex-1 space-y-0.5">
                        <Input
                            className="h-6 text-xs font-mono"
                            value={idValue}
                            onChange={(e) => setIdValue(e.target.value)}
                            autoFocus
                            placeholder="map.entity_name"
                            onBlur={() => commitEntityId(idValue)}
                            onKeyDown={(e) => {
                                if (e.key === "Enter") commitEntityId(idValue);
                                if (e.key === "Escape") { setEditingId(false); setIdValue(object.properties.entity_id ?? ""); }
                            }}
                        />
                        {idConflict && (
                            <p className="text-[10px] text-accent-red flex items-center gap-0.5">
                                <AlertTriangle className="h-2.5 w-2.5" />
                                Conflict: obj#{idConflict.objectId} in {idConflict.mapFile}
                            </p>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}

function HandlerSection({
    object,
    scriptRef,
    scriptPath,
    scripts,
}: {
    object: TiledSelectedObject;
    scriptRef: string;
    scriptPath: string | null;
    scripts: ScriptFileEntry[];
}) {
    const [mode, setMode] = useState<"view" | "assign" | "create">("view");
    const [search, setSearch] = useState(scriptRef);
    const [newName, setNewName] = useState("");
    const [targetFile, setTargetFile] = useState("");
    const [fileSearch, setFileSearch] = useState("");
    const [creating, setCreating] = useState(false);
    const [createError, setCreateError] = useState("");
    const queryClient = useQueryClient();

    useEffect(() => {
        setSearch(scriptRef);
        setMode("view");
    }, [scriptRef]);

    const allHandlers = useMemo(() => {
        const handlers: { name: string; file: string }[] = [];
        for (const script of scripts) {
            for (const name of script.handlerNames) {
                handlers.push({ name, file: script.path });
            }
        }
        handlers.sort((a, b) => a.name.localeCompare(b.name));
        return handlers;
    }, [scripts]);

    const filtered = useMemo(() => {
        const q = search.toLowerCase();
        if (!q) return allHandlers.slice(0, 30);
        return allHandlers.filter((h) => h.name.toLowerCase().includes(q) || h.file.toLowerCase().includes(q)).slice(0, 30);
    }, [allHandlers, search]);

    const scriptFiles = useMemo(() => {
        return scripts.map((s) => s.path).sort();
    }, [scripts]);

    const filteredFiles = useMemo(() => {
        const q = fileSearch.toLowerCase();
        if (!q) return scriptFiles.slice(0, 20);
        return scriptFiles.filter((f) => f.toLowerCase().includes(q)).slice(0, 20);
    }, [scriptFiles, fileSearch]);

    const handleAssign = useCallback(
        (handlerName: string) => {
            sendTiledCommand({
                objectId: object.id,
                action: "setProperty",
                name: "script_ref",
                value: handlerName,
            });
            setMode("view");
        },
        [object.id],
    );

    const handleCreate = useCallback(async () => {
        if (!newName.trim() || !targetFile) return;
        const name = newName.trim();
        setCreating(true);
        setCreateError("");
        try {
            const detail = await apiFetch<ScriptFileDetail>(`/scripts/${targetFile}`);
            const parsed = parseScript(detail.rawYaml);
            if (parsed.handlers[name]) {
                setCreateError(`Handler "${name}" already exists in this file`);
                setCreating(false);
                return;
            }
            parsed.handlers[name] = {};
            const content = stringifyScript(parsed);
            await apiFetch(`/scripts/${targetFile}`, {
                method: "PUT",
                body: JSON.stringify({ content }),
            });
            queryClient.invalidateQueries({ queryKey: ["scripts"] });
            handleAssign(name);
        } catch (e) {
            setCreateError(e instanceof Error ? e.message : "Failed to create handler");
        } finally {
            setCreating(false);
        }
    }, [newName, targetFile, queryClient, handleAssign]);

    const resetCreate = useCallback(() => {
        setMode("view");
        setNewName("");
        setTargetFile("");
        setFileSearch("");
        setCreateError("");
    }, []);

    return (
        <div className="space-y-2">
            <div className="flex items-center gap-2 border-b border-border pb-1">
                <span className="text-sm font-semibold">Handler</span>
                {scriptPath && mode === "view" && (
                    <Link
                        to={`/scripts/${scriptPath}`}
                        className="ml-auto text-[10px] text-accent-violet hover:underline flex items-center gap-0.5"
                    >
                        Edit in script editor <ExternalLink className="h-2.5 w-2.5" />
                    </Link>
                )}
            </div>

            {mode === "view" && (
                <div className="flex items-center gap-2">
                    {scriptRef ? (
                        <>
                            <Circle className="h-2.5 w-2.5 fill-accent-green text-accent-green" />
                            <span className="text-xs font-mono font-medium">{scriptRef}</span>
                            {!scriptPath && (
                                <span className="text-[10px] text-accent-red flex items-center gap-0.5">
                                    <AlertTriangle className="h-2.5 w-2.5" /> not found
                                </span>
                            )}
                        </>
                    ) : (
                        <span className="text-xs text-muted-foreground italic">No handler assigned</span>
                    )}
                    <div className="ml-auto flex items-center gap-1">
                        <Button variant="outline" size="sm" className="h-6 text-[10px] px-2" onClick={() => setMode("assign")}>
                            {scriptRef ? "Change" : "Assign"}
                        </Button>
                        <Button variant="outline" size="sm" className="h-6 text-[10px] px-2 gap-0.5" onClick={() => setMode("create")}>
                            <Plus className="h-3 w-3" /> New
                        </Button>
                    </div>
                </div>
            )}

            {mode === "assign" && (
                <div className="space-y-1">
                    <Input
                        className="h-7 text-xs font-mono"
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        placeholder="Search handlers..."
                        autoFocus
                        onKeyDown={(e) => {
                            if (e.key === "Escape") {
                                setMode("view");
                                setSearch(scriptRef);
                            }
                            if (e.key === "Enter" && search) {
                                handleAssign(search);
                            }
                        }}
                    />
                    <div className="border border-border rounded-md max-h-48 overflow-y-auto">
                        {filtered.map((h) => (
                            <button
                                key={h.name}
                                className="flex w-full items-center gap-2 px-2 py-1 text-left text-xs hover:bg-accent border-b border-border last:border-b-0"
                                onClick={() => handleAssign(h.name)}
                            >
                                <span className="font-mono flex-1 truncate">{h.name}</span>
                                {h.name === scriptRef && <Check className="h-3 w-3 text-accent-green" />}
                                <span className="text-[10px] text-muted-foreground truncate max-w-[160px]">{h.file}</span>
                            </button>
                        ))}
                        {filtered.length === 0 && (
                            <p className="px-2 py-2 text-[10px] text-muted-foreground">No handlers match</p>
                        )}
                    </div>
                    <div className="flex justify-end gap-1">
                        <Button variant="ghost" size="sm" className="h-6 text-[10px]" onClick={() => { setMode("view"); setSearch(scriptRef); }}>
                            Cancel
                        </Button>
                    </div>
                </div>
            )}

            {mode === "create" && (
                <div className="space-y-2">
                    <div className="space-y-1">
                        <label className="text-[11px] font-medium">Handler name</label>
                        <Input
                            className="h-7 text-xs font-mono"
                            value={newName}
                            onChange={(e) => setNewName(e.target.value)}
                            placeholder="my_handler_name"
                            autoFocus
                            onKeyDown={(e) => {
                                if (e.key === "Escape") resetCreate();
                            }}
                        />
                    </div>
                    <div className="space-y-1">
                        <label className="text-[11px] font-medium">Script file</label>
                        {targetFile ? (
                            <div className="flex items-center gap-2">
                                <span className="text-xs font-mono truncate flex-1">{targetFile}</span>
                                <Button variant="ghost" size="sm" className="h-5 text-[10px] px-1" onClick={() => setTargetFile("")}>
                                    change
                                </Button>
                            </div>
                        ) : (
                            <>
                                <Input
                                    className="h-7 text-xs font-mono"
                                    value={fileSearch}
                                    onChange={(e) => setFileSearch(e.target.value)}
                                    placeholder="Search script files..."
                                    onKeyDown={(e) => {
                                        if (e.key === "Escape") resetCreate();
                                    }}
                                />
                                <div className="border border-border rounded-md max-h-36 overflow-y-auto">
                                    {filteredFiles.map((f) => (
                                        <button
                                            key={f}
                                            className="flex w-full items-center px-2 py-1 text-left text-xs font-mono hover:bg-accent border-b border-border last:border-b-0"
                                            onClick={() => { setTargetFile(f); setFileSearch(""); }}
                                        >
                                            <span className="truncate">{f}</span>
                                        </button>
                                    ))}
                                    {filteredFiles.length === 0 && (
                                        <p className="px-2 py-2 text-[10px] text-muted-foreground">No files match</p>
                                    )}
                                </div>
                            </>
                        )}
                    </div>
                    {createError && (
                        <p className="text-[10px] text-accent-red">{createError}</p>
                    )}
                    <div className="flex justify-end gap-1">
                        <Button variant="ghost" size="sm" className="h-6 text-[10px]" onClick={resetCreate}>
                            Cancel
                        </Button>
                        <Button
                            size="sm"
                            className="h-6 text-[10px] px-2 gap-0.5"
                            disabled={!newName.trim() || !targetFile || creating}
                            onClick={handleCreate}
                        >
                            <Plus className="h-3 w-3" />
                            {creating ? "Creating..." : "Create & Assign"}
                        </Button>
                    </div>
                </div>
            )}
        </div>
    );
}

function HandlerPropsSection({
    object,
    scriptPath,
    scriptRef,
}: {
    object: TiledSelectedObject;
    scriptPath: string;
    scriptRef: string;
}) {
    const { data: scriptDetail } = useScript(scriptPath);
    const [expanded, setExpanded] = useState(true);

    const handlerProps = useMemo((): HandlerPropDef[] => {
        if (!scriptDetail?.rawYaml || !scriptRef) return [];
        try {
            const parsed = parseScript(scriptDetail.rawYaml);
            const handler = parsed.handlers[scriptRef];
            if (handler?.props) return handler.props;
        } catch {
            // parse error
        }
        return [];
    }, [scriptDetail, scriptRef]);

    if (handlerProps.length === 0) return null;

    return (
        <div className="space-y-2">
            <button
                className="flex items-center gap-1 text-sm font-semibold border-b border-border pb-1 w-full text-left"
                onClick={() => setExpanded(!expanded)}
            >
                {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                Handler Properties
                <span className="text-[10px] font-normal text-muted-foreground ml-1">({handlerProps.length})</span>
            </button>
            {expanded && (
                <div className="space-y-2">
                    {handlerProps.map((prop) => (
                        <PropEditor key={prop.name} prop={prop} object={object} />
                    ))}
                </div>
            )}
        </div>
    );
}

function PropEditor({ prop, object }: { prop: HandlerPropDef; object: TiledSelectedObject }) {
    const currentValue = object.properties[prop.name] ?? "";
    const [value, setValue] = useState(currentValue);

    useEffect(() => {
        setValue(object.properties[prop.name] ?? "");
    }, [object.properties, prop.name]);

    const isMissing = prop.required && !currentValue;
    const hasDefault = prop.default !== undefined && prop.default !== null;

    const handleCommit = useCallback(
        (newValue: string) => {
            if (newValue === currentValue) return;
            sendTiledCommand({
                objectId: object.id,
                action: "setProperty",
                name: prop.name,
                value: newValue,
            });
        },
        [object.id, prop.name, currentValue],
    );

    return (
        <div className="space-y-0.5">
            <div className="flex items-center gap-1.5">
                <span className="text-[11px] font-medium">{prop.name}</span>
                <span className="text-[9px] px-1 py-0 rounded bg-muted text-muted-foreground">{prop.type}</span>
                {prop.required && !currentValue && (
                    <AlertTriangle className="h-2.5 w-2.5 text-accent-red" />
                )}
                {currentValue && !isMissing && (
                    <Check className="h-2.5 w-2.5 text-accent-green" />
                )}
            </div>
            {prop.description && (
                <p className="text-[10px] text-muted-foreground">{prop.description}</p>
            )}
            <div className="flex items-center gap-1">
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    value={value}
                    onChange={(e) => setValue(e.target.value)}
                    placeholder={hasDefault ? `default: ${prop.default}` : prop.required ? "required" : "optional"}
                    onBlur={() => handleCommit(value)}
                    onKeyDown={(e) => {
                        if (e.key === "Enter") handleCommit(value);
                    }}
                />
                {hasDefault && !currentValue && (
                    <Button
                        variant="outline"
                        size="sm"
                        className="h-6 text-[10px] px-1.5 shrink-0"
                        onClick={() => {
                            const def = String(prop.default);
                            setValue(def);
                            handleCommit(def);
                        }}
                    >
                        Use default
                    </Button>
                )}
            </div>
        </div>
    );
}

// --- Handler Definition Section (inline editor) ---

function HandlerDefinitionSection({
    scriptPath,
    scriptRef,
    scripts,
}: {
    scriptPath: string;
    scriptRef: string;
    scripts: ScriptFileEntry[];
}) {
    const { data: scriptDetail } = useScript(scriptPath);
    const { data: schema } = useScriptSchema();
    const { data: allScripts } = useScripts();
    const saveScript = useSaveScript();

    const [expanded, setExpanded] = useState(false);
    const [editedParsed, setEditedParsed] = useState<ParsedScript | null>(null);
    const [dirty, setDirty] = useState(false);

    // Parse the script when data arrives or changes
    const baseParsed = useMemo((): ParsedScript | null => {
        if (!scriptDetail?.rawYaml) return null;
        try {
            return parseScript(scriptDetail.rawYaml);
        } catch {
            return null;
        }
    }, [scriptDetail]);

    // Reset edited state when the base changes (e.g. after save or reload)
    useEffect(() => {
        setEditedParsed(null);
        setDirty(false);
    }, [baseParsed]);

    const currentParsed = editedParsed ?? baseParsed;
    const handler = currentParsed?.handlers[scriptRef] ?? null;

    const handleChange = useCallback(
        (updated: HandlerDef) => {
            if (!currentParsed) return;
            const next: ParsedScript = { ...currentParsed, handlers: { ...currentParsed.handlers, [scriptRef]: updated } };
            setEditedParsed(next);
            setDirty(true);
        },
        [currentParsed, scriptRef],
    );

    const handleSave = useCallback(() => {
        if (!editedParsed) return;
        const content = stringifyScript(editedParsed);
        saveScript.mutate(
            { path: scriptPath, content },
            { onSuccess: () => setDirty(false) },
        );
    }, [editedParsed, scriptPath, saveScript]);

    // ExprContext values
    const handlerVarKeys = useMemo(() => {
        if (!handler?.var) return [];
        return Object.keys(handler.var);
    }, [handler]);

    const constKeys = useMemo(() => {
        if (!currentParsed?.consts) return [];
        return Object.keys(currentParsed.consts);
    }, [currentParsed]);

    const { otherConstKeys, otherCustomActionNames } = useMemo(() => {
        if (!allScripts) return { otherConstKeys: [] as CrossFileEntry[], otherCustomActionNames: [] as CrossFileEntry[] };
        const localConsts = new Set(currentParsed?.consts ? Object.keys(currentParsed.consts) : []);
        const localActions = new Set(currentParsed?.custom_actions ? Object.keys(currentParsed.custom_actions) : []);
        const consts: CrossFileEntry[] = [];
        const actions: CrossFileEntry[] = [];
        for (const s of allScripts) {
            if (s.path === scriptPath) continue;
            for (const name of s.constNames ?? []) {
                if (!localConsts.has(name)) consts.push({ name, file: s.path });
            }
            for (const name of s.customActionNames ?? []) {
                if (!localActions.has(name)) actions.push({ name, file: s.path });
            }
        }
        return { otherConstKeys: consts, otherCustomActionNames: actions };
    }, [allScripts, scriptPath, currentParsed?.consts, currentParsed?.custom_actions]);

    if (!handler || !schema) return null;

    return (
        <div className="space-y-2">
            <button
                className="flex items-center gap-1 text-sm font-semibold border-b border-border pb-1 w-full text-left"
                onClick={() => setExpanded(!expanded)}
            >
                {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                Handler Definition
                {dirty && <span className="text-[10px] font-normal text-accent-amber ml-1">(unsaved)</span>}
            </button>
            {expanded && (
                <div className="space-y-2">
                    {dirty && (
                        <div className="flex items-center gap-2">
                            <Button
                                size="sm"
                                className="h-6 text-[10px] px-2 gap-1"
                                disabled={saveScript.isPending}
                                onClick={handleSave}
                            >
                                <Save className="h-3 w-3" />
                                {saveScript.isPending ? "Saving..." : "Save"}
                            </Button>
                            <span className="text-[10px] text-muted-foreground">Changes to {scriptPath}</span>
                        </div>
                    )}
                    <div className="border border-border rounded-md overflow-hidden">
                        <ExprContextProvider
                            handlerVarKeys={handlerVarKeys}
                            constKeys={constKeys}
                            otherConstKeys={otherConstKeys}
                            otherCustomActionNames={otherCustomActionNames}
                            customActions={currentParsed?.custom_actions}
                        >
                            <HandlerDetail
                                handlerName={scriptRef}
                                handler={handler}
                                schema={schema}
                                onChange={handleChange}
                            />
                        </ExprContextProvider>
                    </div>
                </div>
            )}
        </div>
    );
}

// --- Structured Config Section ---

interface PresenceConfig {
    block_ingress?: boolean;
    is_interactable?: boolean;
    impedance?: number;
}

interface AnimationEntry {
    name: string;
    colorMask?: string;
    offset?: { x: number; y: number };
}

interface LightEntry {
    color: string;
    size: number;
    modifier?: string;
}

interface RenderConfig {
    mode?: string;
    animations?: Record<string, AnimationEntry[]>;
    lights?: Record<string, LightEntry[]>;
}

function StructuredConfigSection({ object }: { object: TiledSelectedObject }) {
    const isModeEntity = object.className === "ModeBasedEntity";
    const hasPresence = "presence_config" in object.properties;
    const hasRender = "render_config" in object.properties;
    const showPresence = hasPresence || isModeEntity || object.className === "NPC";
    const showRender = hasRender || isModeEntity;

    if (!showPresence && !showRender) return null;

    return (
        <div className="space-y-4">
            {showPresence && <PresenceConfigEditor object={object} />}
            {showRender && <RenderConfigEditor object={object} />}
        </div>
    );
}

function PresenceConfigEditor({ object }: { object: TiledSelectedObject }) {
    const [expanded, setExpanded] = useState(true);

    const config = useMemo((): PresenceConfig => {
        try {
            const raw = object.properties.presence_config;
            if (!raw) return {};
            return (YAML.parse(raw) ?? {}) as PresenceConfig;
        } catch {
            return {};
        }
    }, [object.properties.presence_config]);

    const commitConfig = useCallback(
        (updated: PresenceConfig) => {
            const yamlStr = YAML.stringify(updated, { indent: 2, lineWidth: 0 }).trim();
            sendTiledCommand({
                objectId: object.id,
                action: "setProperty",
                name: "presence_config",
                value: yamlStr,
            });
        },
        [object.id],
    );

    return (
        <div className="space-y-2">
            <button
                className="flex items-center gap-1 text-sm font-semibold border-b border-border pb-1 w-full text-left"
                onClick={() => setExpanded(!expanded)}
            >
                {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                Presence Config
            </button>
            {expanded && (
                <div className="space-y-2 pl-1">
                    <label className="flex items-center gap-2 text-xs">
                        <input
                            type="checkbox"
                            className="h-3.5 w-3.5"
                            checked={config.block_ingress ?? true}
                            onChange={(e) => commitConfig({ ...config, block_ingress: e.target.checked })}
                        />
                        <span>block_ingress</span>
                        <span className="text-[10px] text-muted-foreground">(default: true)</span>
                    </label>
                    <label className="flex items-center gap-2 text-xs">
                        <input
                            type="checkbox"
                            className="h-3.5 w-3.5"
                            checked={config.is_interactable ?? false}
                            onChange={(e) => commitConfig({ ...config, is_interactable: e.target.checked })}
                        />
                        <span>is_interactable</span>
                    </label>
                    <div className="flex items-center gap-2">
                        <label className="text-xs shrink-0">impedance</label>
                        <Input
                            className="h-6 w-20 text-xs font-mono"
                            type="number"
                            step="0.1"
                            value={config.impedance ?? ""}
                            onChange={(e) => {
                                const v = e.target.value;
                                const next = { ...config };
                                if (v === "") {
                                    delete next.impedance;
                                } else {
                                    next.impedance = parseFloat(v);
                                }
                                commitConfig(next);
                            }}
                            placeholder="0"
                        />
                    </div>
                </div>
            )}
        </div>
    );
}

const LIGHT_MODIFIERS = ["", "flicker", "pulse_slow", "pulse_medium", "pulse_fast"];

function RenderConfigEditor({ object }: { object: TiledSelectedObject }) {
    const [expanded, setExpanded] = useState(true);
    const [addingMode, setAddingMode] = useState(false);
    const [newModeName, setNewModeName] = useState("");

    const config = useMemo((): RenderConfig => {
        try {
            const raw = object.properties.render_config;
            if (!raw) return {};
            return (YAML.parse(raw) ?? {}) as RenderConfig;
        } catch {
            return {};
        }
    }, [object.properties.render_config]);

    const commitConfig = useCallback(
        (updated: RenderConfig) => {
            const cleaned: RenderConfig = {};
            if (updated.mode) cleaned.mode = updated.mode;
            if (updated.animations && Object.keys(updated.animations).length > 0) cleaned.animations = updated.animations;
            if (updated.lights && Object.keys(updated.lights).length > 0) cleaned.lights = updated.lights;
            const yamlStr = YAML.stringify(cleaned, { indent: 2, lineWidth: 0 }).trim();
            sendTiledCommand({
                objectId: object.id,
                action: "setProperty",
                name: "render_config",
                value: yamlStr,
            });
        },
        [object.id],
    );

    const allModes = useMemo(() => {
        const modes = new Set<string>();
        if (config.animations) Object.keys(config.animations).forEach((m) => modes.add(m));
        if (config.lights) Object.keys(config.lights).forEach((m) => modes.add(m));
        return Array.from(modes).sort((a, b) => {
            if (a === "") return -1;
            if (b === "") return 1;
            return a.localeCompare(b);
        });
    }, [config]);

    const handleAddMode = useCallback(() => {
        const name = newModeName.trim();
        const anims = { ...config.animations, [name]: [] };
        commitConfig({ ...config, animations: anims });
        setNewModeName("");
        setAddingMode(false);
    }, [newModeName, config, commitConfig]);

    const handleRemoveMode = useCallback((mode: string) => {
        const anims = { ...config.animations };
        const lights = { ...config.lights };
        delete anims[mode];
        delete lights[mode];
        commitConfig({ ...config, animations: anims, lights: lights });
    }, [config, commitConfig]);

    const updateAnimation = useCallback((mode: string, index: number, entry: AnimationEntry) => {
        const entries = [...(config.animations?.[mode] ?? [])];
        entries[index] = entry;
        commitConfig({ ...config, animations: { ...config.animations, [mode]: entries } });
    }, [config, commitConfig]);

    const addAnimation = useCallback((mode: string) => {
        const entries = [...(config.animations?.[mode] ?? []), { name: "" }];
        commitConfig({ ...config, animations: { ...config.animations, [mode]: entries } });
    }, [config, commitConfig]);

    const removeAnimation = useCallback((mode: string, index: number) => {
        const entries = [...(config.animations?.[mode] ?? [])];
        entries.splice(index, 1);
        commitConfig({ ...config, animations: { ...config.animations, [mode]: entries } });
    }, [config, commitConfig]);

    const updateLight = useCallback((mode: string, index: number, entry: LightEntry) => {
        const entries = [...(config.lights?.[mode] ?? [])];
        entries[index] = entry;
        commitConfig({ ...config, lights: { ...config.lights, [mode]: entries } });
    }, [config, commitConfig]);

    const addLight = useCallback((mode: string) => {
        const entries = [...(config.lights?.[mode] ?? []), { color: "#fff", size: 1 }];
        commitConfig({ ...config, lights: { ...config.lights, [mode]: entries } });
    }, [config, commitConfig]);

    const removeLight = useCallback((mode: string, index: number) => {
        const entries = [...(config.lights?.[mode] ?? [])];
        entries.splice(index, 1);
        commitConfig({ ...config, lights: { ...config.lights, [mode]: entries } });
    }, [config, commitConfig]);

    return (
        <div className="space-y-2">
            <button
                className="flex items-center gap-1 text-sm font-semibold border-b border-border pb-1 w-full text-left"
                onClick={() => setExpanded(!expanded)}
            >
                {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                Render Config
                {allModes.length > 0 && (
                    <span className="text-[10px] font-normal text-muted-foreground ml-1">({allModes.length} modes)</span>
                )}
            </button>
            {expanded && (
                <div className="space-y-3 pl-1">
                    {/* Default mode */}
                    <div className="flex items-center gap-2">
                        <label className="text-xs shrink-0">default mode</label>
                        <Input
                            className="h-6 flex-1 text-xs font-mono"
                            value={config.mode ?? ""}
                            onChange={(e) => commitConfig({ ...config, mode: e.target.value || undefined })}
                            placeholder="(empty = first)"
                        />
                    </div>

                    {/* Mode entries */}
                    {allModes.map((mode) => (
                        <RenderModeSection
                            key={mode}
                            mode={mode}
                            animations={config.animations?.[mode] ?? []}
                            lights={config.lights?.[mode] ?? []}
                            onUpdateAnimation={(i, e) => updateAnimation(mode, i, e)}
                            onAddAnimation={() => addAnimation(mode)}
                            onRemoveAnimation={(i) => removeAnimation(mode, i)}
                            onUpdateLight={(i, e) => updateLight(mode, i, e)}
                            onAddLight={() => addLight(mode)}
                            onRemoveLight={(i) => removeLight(mode, i)}
                            onRemoveMode={() => handleRemoveMode(mode)}
                        />
                    ))}

                    {/* Add mode */}
                    {!addingMode ? (
                        <Button variant="outline" size="sm" className="h-6 text-[10px] px-2 gap-0.5" onClick={() => setAddingMode(true)}>
                            <Plus className="h-3 w-3" /> Add mode
                        </Button>
                    ) : (
                        <div className="flex items-center gap-1">
                            <Input
                                className="h-6 flex-1 text-xs font-mono"
                                value={newModeName}
                                onChange={(e) => setNewModeName(e.target.value)}
                                placeholder="mode name (empty = default)"
                                autoFocus
                                onKeyDown={(e) => {
                                    if (e.key === "Enter") handleAddMode();
                                    if (e.key === "Escape") { setAddingMode(false); setNewModeName(""); }
                                }}
                            />
                            <Button size="sm" className="h-6 text-[10px] px-2" onClick={handleAddMode}>Add</Button>
                            <Button variant="ghost" size="sm" className="h-6 text-[10px] px-1" onClick={() => { setAddingMode(false); setNewModeName(""); }}>
                                Cancel
                            </Button>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

function RenderModeSection({
    mode,
    animations,
    lights,
    onUpdateAnimation,
    onAddAnimation,
    onRemoveAnimation,
    onUpdateLight,
    onAddLight,
    onRemoveLight,
    onRemoveMode,
}: {
    mode: string;
    animations: AnimationEntry[];
    lights: LightEntry[];
    onUpdateAnimation: (index: number, entry: AnimationEntry) => void;
    onAddAnimation: () => void;
    onRemoveAnimation: (index: number) => void;
    onUpdateLight: (index: number, entry: LightEntry) => void;
    onAddLight: () => void;
    onRemoveLight: (index: number) => void;
    onRemoveMode: () => void;
}) {
    const [expanded, setExpanded] = useState(true);

    return (
        <div className="border border-border rounded-md overflow-hidden">
            <div className="flex items-center gap-2 bg-muted/40 px-2 py-1">
                <button className="flex items-center gap-1 flex-1 text-left" onClick={() => setExpanded(!expanded)}>
                    {expanded ? <ChevronDown className="h-2.5 w-2.5" /> : <ChevronRight className="h-2.5 w-2.5" />}
                    <span className="text-[11px] font-mono font-medium text-accent-teal">
                        {mode === "" ? "(default)" : mode}
                    </span>
                    <span className="text-[10px] text-muted-foreground ml-1">
                        {animations.length}a {lights.length}l
                    </span>
                </button>
                <button className="text-[10px] text-muted-foreground hover:text-accent-red" onClick={onRemoveMode}>
                    remove
                </button>
            </div>
            {expanded && (
                <div className="p-2 space-y-2">
                    {/* Animations */}
                    <div className="space-y-1">
                        <span className="text-[10px] font-medium text-muted-foreground uppercase tracking-wider">Animations</span>
                        {animations.map((entry, i) => (
                            <div key={i} className="flex items-start gap-1 group">
                                <div className="flex-1 space-y-0.5">
                                    <Input
                                        className="h-6 text-xs font-mono"
                                        value={entry.name}
                                        onChange={(e) => onUpdateAnimation(i, { ...entry, name: e.target.value })}
                                        placeholder="sprites/path:animation"
                                    />
                                    <div className="flex items-center gap-1">
                                        <Input
                                            className="h-5 w-20 text-[10px] font-mono"
                                            value={entry.colorMask ?? ""}
                                            onChange={(e) => onUpdateAnimation(i, { ...entry, colorMask: e.target.value || undefined })}
                                            placeholder="colorMask"
                                        />
                                        <Input
                                            className="h-5 w-12 text-[10px] font-mono"
                                            type="number"
                                            step="0.125"
                                            value={entry.offset?.x ?? ""}
                                            onChange={(e) => {
                                                const x = e.target.value ? parseFloat(e.target.value) : 0;
                                                const y = entry.offset?.y ?? 0;
                                                onUpdateAnimation(i, { ...entry, offset: (x || y) ? { x, y } : undefined });
                                            }}
                                            placeholder="oX"
                                        />
                                        <Input
                                            className="h-5 w-12 text-[10px] font-mono"
                                            type="number"
                                            step="0.125"
                                            value={entry.offset?.y ?? ""}
                                            onChange={(e) => {
                                                const y = e.target.value ? parseFloat(e.target.value) : 0;
                                                const x = entry.offset?.x ?? 0;
                                                onUpdateAnimation(i, { ...entry, offset: (x || y) ? { x, y } : undefined });
                                            }}
                                            placeholder="oY"
                                        />
                                    </div>
                                </div>
                                <button
                                    className="mt-1 text-muted-foreground hover:text-accent-red opacity-0 group-hover:opacity-100"
                                    onClick={() => onRemoveAnimation(i)}
                                >
                                    <Plus className="h-3 w-3 rotate-45" />
                                </button>
                            </div>
                        ))}
                        <Button variant="ghost" size="sm" className="h-5 text-[10px] px-1.5 gap-0.5" onClick={onAddAnimation}>
                            <Plus className="h-2.5 w-2.5" /> animation
                        </Button>
                    </div>

                    {/* Lights */}
                    <div className="space-y-1">
                        <span className="text-[10px] font-medium text-muted-foreground uppercase tracking-wider">Lights</span>
                        {lights.map((entry, i) => (
                            <div key={i} className="flex items-center gap-1 group">
                                <Input
                                    className="h-6 w-16 text-xs font-mono"
                                    value={entry.color}
                                    onChange={(e) => onUpdateLight(i, { ...entry, color: e.target.value })}
                                    placeholder="#fff"
                                />
                                <div
                                    className="h-5 w-5 rounded border border-border shrink-0"
                                    style={{ backgroundColor: entry.color }}
                                />
                                <Input
                                    className="h-6 w-14 text-xs font-mono"
                                    type="number"
                                    step="0.25"
                                    min="0"
                                    value={entry.size}
                                    onChange={(e) => onUpdateLight(i, { ...entry, size: parseFloat(e.target.value) || 0 })}
                                    placeholder="size"
                                />
                                <select
                                    className="h-6 text-[10px] font-mono rounded border border-border bg-background px-1 flex-1"
                                    value={entry.modifier ?? ""}
                                    onChange={(e) => onUpdateLight(i, { ...entry, modifier: e.target.value || undefined })}
                                >
                                    <option value="">no modifier</option>
                                    {LIGHT_MODIFIERS.filter(Boolean).map((m) => <option key={m} value={m}>{m}</option>)}
                                </select>
                                <button
                                    className="text-muted-foreground hover:text-accent-red opacity-0 group-hover:opacity-100"
                                    onClick={() => onRemoveLight(i)}
                                >
                                    <Plus className="h-3 w-3 rotate-45" />
                                </button>
                            </div>
                        ))}
                        <Button variant="ghost" size="sm" className="h-5 text-[10px] px-1.5 gap-0.5" onClick={onAddLight}>
                            <Plus className="h-2.5 w-2.5" /> light
                        </Button>
                    </div>
                </div>
            )}
        </div>
    );
}

// --- Class-Specific Properties Section ---

const NPC_MOVEMENTS = ["", "static", "horiz"];
const NPC_SPEEDS = ["", "fast"];
const FACING_DIRECTIONS = ["", "up", "down", "left", "right"];

function ClassPropertiesSection({ object }: { object: TiledSelectedObject }) {
    const cls = object.className;
    if (cls === "NPC") return <NPCPropertiesEditor object={object} />;
    if (cls === "DirectInteraction") return <DirectInteractionEditor object={object} />;
    if (cls === "ShadowMob") return <ShadowMobEditor object={object} />;
    return <MetadataEditor object={object} />;
}

function NPCPropertiesEditor({ object }: { object: TiledSelectedObject }) {
    const [expanded, setExpanded] = useState(true);

    const commitProp = useCallback((name: string, value: string) => {
        if (value) {
            sendTiledCommand({ objectId: object.id, action: "setProperty", name, value });
        } else {
            sendTiledCommand({ objectId: object.id, action: "removeProperty", name });
        }
    }, [object.id]);

    return (
        <div className="space-y-2">
            <button
                className="flex items-center gap-1 text-sm font-semibold border-b border-border pb-1 w-full text-left"
                onClick={() => setExpanded(!expanded)}
            >
                {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                NPC Behavior
            </button>
            {expanded && (
                <div className="space-y-2 pl-1">
                    <div className="flex items-center gap-2">
                        <label className="text-xs shrink-0 w-28">movement</label>
                        <select
                            className="h-6 flex-1 text-xs font-mono rounded border border-border bg-background px-1"
                            value={object.properties.movement ?? ""}
                            onChange={(e) => commitProp("movement", e.target.value)}
                        >
                            {NPC_MOVEMENTS.map((m) => <option key={m} value={m}>{m || "(default - wanders)"}</option>)}
                        </select>
                    </div>
                    <div className="flex items-center gap-2">
                        <label className="text-xs shrink-0 w-28">speed</label>
                        <select
                            className="h-6 flex-1 text-xs font-mono rounded border border-border bg-background px-1"
                            value={object.properties.speed ?? ""}
                            onChange={(e) => commitProp("speed", e.target.value)}
                        >
                            {NPC_SPEEDS.map((s) => <option key={s} value={s}>{s || "(default - normal)"}</option>)}
                        </select>
                    </div>
                    <div className="flex items-center gap-2">
                        <label className="text-xs shrink-0 w-28">idle facing</label>
                        <select
                            className="h-6 flex-1 text-xs font-mono rounded border border-border bg-background px-1"
                            value={object.properties.idle_facing_direction ?? ""}
                            onChange={(e) => commitProp("idle_facing_direction", e.target.value)}
                        >
                            {FACING_DIRECTIONS.map((d) => <option key={d} value={d}>{d || "(none)"}</option>)}
                        </select>
                    </div>
                    <div className="flex items-center gap-2">
                        <label className="text-xs shrink-0 w-28">active zone</label>
                        <Input
                            className="h-6 flex-1 text-xs font-mono"
                            key={object.properties.active_player_zone ?? ""}
                            defaultValue={object.properties.active_player_zone ?? ""}
                            onBlur={(e) => commitProp("active_player_zone", e.target.value)}
                            onKeyDown={(e) => { if (e.key === "Enter") (e.target as HTMLInputElement).blur(); }}
                            placeholder="only active when player in zone"
                        />
                    </div>
                    <MetadataEditor object={object} />
                </div>
            )}
        </div>
    );
}

function DirectInteractionEditor({ object }: { object: TiledSelectedObject }) {
    const [expanded, setExpanded] = useState(true);

    return (
        <div className="space-y-2">
            <button
                className="flex items-center gap-1 text-sm font-semibold border-b border-border pb-1 w-full text-left"
                onClick={() => setExpanded(!expanded)}
            >
                {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                Direct Interaction
            </button>
            {expanded && (
                <div className="space-y-1 pl-1">
                    <p className="text-[10px] text-muted-foreground">Redirects interactions to another entity</p>
                    <div className="flex items-center gap-2">
                        <label className="text-xs shrink-0">target</label>
                        <Input
                            className="h-6 flex-1 text-xs font-mono"
                            key={object.properties.target ?? ""}
                            defaultValue={object.properties.target ?? ""}
                            onBlur={(e) => {
                                sendTiledCommand({ objectId: object.id, action: "setProperty", name: "target", value: e.target.value });
                            }}
                            onKeyDown={(e) => { if (e.key === "Enter") (e.target as HTMLInputElement).blur(); }}
                            placeholder="entity_id to redirect to"
                        />
                    </div>
                </div>
            )}
        </div>
    );
}

function ShadowMobEditor({ object }: { object: TiledSelectedObject }) {
    const [expanded, setExpanded] = useState(true);

    const config = useMemo(() => {
        try {
            const raw = object.properties.shadow_mob;
            if (!raw) return {} as Record<string, string>;
            const parsed = YAML.parse(raw);
            if (!parsed || typeof parsed !== "object") return {};
            const flat: Record<string, string> = {};
            for (const [k, v] of Object.entries(parsed)) {
                flat[k] = v != null ? String(v) : "";
            }
            return flat;
        } catch {
            return {};
        }
    }, [object.properties.shadow_mob]);

    const commitConfig = useCallback((key: string, value: string) => {
        const updated = { ...config };
        if (value) {
            updated[key] = value;
        } else {
            delete updated[key];
        }
        const yamlStr = YAML.stringify(updated, { indent: 2, lineWidth: 0 }).trim();
        sendTiledCommand({ objectId: object.id, action: "setProperty", name: "shadow_mob", value: yamlStr });
    }, [object.id, config]);

    const fields = [
        { key: "combat_id", label: "combat_id", placeholder: "combat encounter id" },
        { key: "broadcast_id", label: "broadcast_id", placeholder: "broadcast on trigger" },
        { key: "opponent_pool", label: "opponent_pool", placeholder: "random opponent pool" },
        { key: "opponent", label: "opponent", placeholder: "primortal type" },
        { key: "opponent_archetype", label: "archetype", placeholder: "opponent archetype" },
        { key: "combat_background", label: "background", placeholder: "combat bg" },
        { key: "leash_radius", label: "leash_radius", placeholder: "5" },
        { key: "respawn_delay", label: "respawn_delay", placeholder: "10" },
    ];

    return (
        <div className="space-y-2">
            <button
                className="flex items-center gap-1 text-sm font-semibold border-b border-border pb-1 w-full text-left"
                onClick={() => setExpanded(!expanded)}
            >
                {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                Shadow Mob Config
            </button>
            {expanded && (
                <div className="space-y-1.5 pl-1">
                    {fields.map((f) => (
                        <div key={f.key} className="flex items-center gap-2">
                            <label className="text-xs shrink-0 w-24 truncate">{f.label}</label>
                            <Input
                                className="h-6 flex-1 text-xs font-mono"
                                key={`${f.key}-${config[f.key] ?? ""}`}
                                defaultValue={config[f.key] ?? ""}
                                onBlur={(e) => commitConfig(f.key, e.target.value)}
                                onKeyDown={(e) => { if (e.key === "Enter") (e.target as HTMLInputElement).blur(); }}
                                placeholder={f.placeholder}
                            />
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

interface TalkerConfig {
    preset?: string;
    voice_pitch?: number;
    energy?: number;
}

function MetadataEditor({ object }: { object: TiledSelectedObject }) {
    const rawMeta = object.properties.metadata;
    const talkerConfig = useMemo((): TalkerConfig => {
        if (!rawMeta) return {};
        try {
            const parsed = YAML.parse(rawMeta);
            return parsed?.talker_config ?? {};
        } catch {
            return {};
        }
    }, [rawMeta]);

    const hasTalkerConfig = rawMeta?.includes("talker_config");
    if (!hasTalkerConfig && !rawMeta) return null;

    const commitTalkerConfig = (updated: TalkerConfig) => {
        let meta: Record<string, unknown> = {};
        if (rawMeta) {
            try { meta = YAML.parse(rawMeta) ?? {}; } catch { meta = {}; }
        }
        const cleaned: TalkerConfig = {};
        if (updated.preset) cleaned.preset = updated.preset;
        if (updated.voice_pitch != null && updated.voice_pitch !== 0) cleaned.voice_pitch = updated.voice_pitch;
        if (updated.energy != null && updated.energy !== 0) cleaned.energy = updated.energy;
        if (Object.keys(cleaned).length > 0) {
            meta.talker_config = cleaned;
        } else {
            delete meta.talker_config;
        }
        if (Object.keys(meta).length === 0) {
            sendTiledCommand({ objectId: object.id, action: "removeProperty", name: "metadata" });
        } else {
            const yamlStr = YAML.stringify(meta, { indent: 2, lineWidth: 0 }).trim();
            sendTiledCommand({ objectId: object.id, action: "setProperty", name: "metadata", value: yamlStr });
        }
    };

    return (
        <div className="space-y-1">
            <span className="text-[10px] font-medium text-muted-foreground uppercase tracking-wider">Talker Config</span>
            <div className="flex items-center gap-2">
                <label className="text-xs shrink-0 w-20">preset</label>
                <select
                    className="h-6 flex-1 text-xs font-mono rounded border border-border bg-background px-1"
                    value={talkerConfig.preset ?? ""}
                    onChange={(e) => commitTalkerConfig({ ...talkerConfig, preset: e.target.value || undefined })}
                >
                    <option value="">normal</option>
                    <option value="slow">slow</option>
                    <option value="kid">kid</option>
                </select>
            </div>
            <div className="flex items-center gap-2">
                <label className="text-xs shrink-0 w-20">voice_pitch</label>
                <Input
                    className="h-6 w-20 text-xs font-mono"
                    type="number"
                    step="0.1"
                    min="-1"
                    max="1"
                    value={talkerConfig.voice_pitch ?? ""}
                    onChange={(e) => {
                        const v = e.target.value === "" ? undefined : parseFloat(e.target.value);
                        commitTalkerConfig({ ...talkerConfig, voice_pitch: v });
                    }}
                    placeholder="0"
                />
                <span className="text-[10px] text-muted-foreground">-1 (deep) to 1 (high)</span>
            </div>
            <div className="flex items-center gap-2">
                <label className="text-xs shrink-0 w-20">energy</label>
                <Input
                    className="h-6 w-20 text-xs font-mono"
                    type="number"
                    step="0.1"
                    min="-1"
                    max="1"
                    value={talkerConfig.energy ?? ""}
                    onChange={(e) => {
                        const v = e.target.value === "" ? undefined : parseFloat(e.target.value);
                        commitTalkerConfig({ ...talkerConfig, energy: v });
                    }}
                    placeholder="0"
                />
                <span className="text-[10px] text-muted-foreground">-1 (quiet) to 1 (loud)</span>
            </div>
        </div>
    );
}

// --- Related Entities Section ---

function RelatedEntitiesSection({ object, scriptRef }: { object: TiledSelectedObject; scriptRef: string }) {
    const { data: usages } = useTiledUsages();
    const [expanded, setExpanded] = useState(false);

    const { sameHandler, nearby } = useMemo(() => {
        const sameHandler: { mapFile: string; objectId: number; entityId?: string; objectType?: string; x: number; y: number }[] = [];
        const nearby: { mapFile: string; objectId: number; entityId?: string; objectType?: string; x: number; y: number; distance: number; handlerName?: string }[] = [];

        if (!usages) return { sameHandler, nearby };

        // Collect entities with the same handler
        if (scriptRef) {
            const usage = usages.find((u) => u.handlerName === scriptRef);
            if (usage) {
                for (const ent of usage.entities) {
                    if (ent.objectId === object.id) continue;
                    sameHandler.push({
                        mapFile: ent.mapFile,
                        objectId: ent.objectId,
                        entityId: ent.entityId,
                        objectType: ent.objectType,
                        x: ent.x,
                        y: ent.y,
                    });
                }
            }
        }

        // Collect nearby entities (within 100px)
        const threshold = 100;
        for (const usage of usages) {
            for (const ent of usage.entities) {
                if (ent.objectId === object.id) continue;
                const dx = ent.x - object.x;
                const dy = ent.y - object.y;
                const dist = Math.sqrt(dx * dx + dy * dy);
                if (dist <= threshold) {
                    // Avoid duplicates if already in sameHandler
                    const alreadyListed = sameHandler.some((s) => s.objectId === ent.objectId && s.mapFile === ent.mapFile);
                    if (!alreadyListed) {
                        nearby.push({
                            mapFile: ent.mapFile,
                            objectId: ent.objectId,
                            entityId: ent.entityId,
                            objectType: ent.objectType,
                            x: ent.x,
                            y: ent.y,
                            distance: Math.round(dist),
                            handlerName: usage.handlerName,
                        });
                    }
                }
            }
        }

        nearby.sort((a, b) => a.distance - b.distance);
        return { sameHandler, nearby };
    }, [usages, scriptRef, object.id, object.x, object.y]);

    const total = sameHandler.length + nearby.length;
    if (total === 0) return null;

    return (
        <div className="space-y-2">
            <button
                className="flex items-center gap-1 text-sm font-semibold border-b border-border pb-1 w-full text-left"
                onClick={() => setExpanded(!expanded)}
            >
                {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                Related Entities
                <span className="text-[10px] font-normal text-muted-foreground ml-1">({total})</span>
            </button>
            {expanded && (
                <div className="space-y-3">
                    {sameHandler.length > 0 && (
                        <div className="space-y-1">
                            <div className="flex items-center gap-1 text-[11px] font-medium text-accent-violet">
                                <Users className="h-3 w-3" />
                                Same handler ({sameHandler.length})
                            </div>
                            {sameHandler.map((ent, i) => (
                                <div key={i} className="flex items-center gap-2 text-[10px] pl-2">
                                    <span className="font-mono text-muted-foreground">obj#{ent.objectId}</span>
                                    {ent.objectType && (
                                        <span className="px-1 rounded bg-accent-teal-tint text-accent-teal">{ent.objectType}</span>
                                    )}
                                    {ent.entityId && (
                                        <span className="font-mono text-foreground">{ent.entityId}</span>
                                    )}
                                    <span className="text-muted-foreground ml-auto">({Math.round(ent.x)}, {Math.round(ent.y)})</span>
                                </div>
                            ))}
                        </div>
                    )}
                    {nearby.length > 0 && (
                        <div className="space-y-1">
                            <div className="flex items-center gap-1 text-[11px] font-medium text-accent-orange">
                                <Crosshair className="h-3 w-3" />
                                Nearby ({nearby.length})
                            </div>
                            {nearby.map((ent, i) => (
                                <div key={i} className="flex items-center gap-2 text-[10px] pl-2">
                                    <span className="font-mono text-muted-foreground">obj#{ent.objectId}</span>
                                    {ent.objectType && (
                                        <span className="px-1 rounded bg-accent-teal-tint text-accent-teal">{ent.objectType}</span>
                                    )}
                                    {ent.entityId && (
                                        <span className="font-mono text-foreground">{ent.entityId}</span>
                                    )}
                                    {ent.handlerName && (
                                        <span className="font-mono text-accent-violet">{ent.handlerName}</span>
                                    )}
                                    <span className="text-muted-foreground ml-auto">{ent.distance}px</span>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

// --- Zone Section ---

function ZoneSection({
    object,
    zoneId,
    scripts,
}: {
    object: TiledSelectedObject;
    zoneId: string;
    scripts: ScriptFileEntry[];
}) {
    const [editingId, setEditingId] = useState(false);
    const [idValue, setIdValue] = useState(zoneId);

    useEffect(() => {
        setIdValue(zoneId);
        setEditingId(false);
    }, [zoneId]);

    const handleCommitId = useCallback(
        (newValue: string) => {
            if (newValue === zoneId) { setEditingId(false); return; }
            sendTiledCommand({
                objectId: object.id,
                action: "setProperty",
                name: "zone_id",
                value: newValue,
            });
            setEditingId(false);
        },
        [object.id, zoneId],
    );

    // Find handlers that reference this zone in their filters
    // (This is a best-effort search through handler names - full parsing
    // would require loading every script file. For now, show the zone ID
    // and provide a link to search in scripts.)
    return (
        <div className="space-y-2">
            <div className="flex items-center gap-2 border-b border-border pb-1">
                <span className="text-sm font-semibold">Zone</span>
                <span className="text-[10px] px-1.5 py-0.5 rounded bg-accent-orange-tint text-accent-orange font-medium">
                    Zone
                </span>
            </div>
            <div className="space-y-1">
                <label className="text-[11px] font-medium">zone_id</label>
                {!editingId ? (
                    <div className="flex items-center gap-2">
                        {zoneId ? (
                            <span className="text-xs font-mono">{zoneId}</span>
                        ) : (
                            <span className="text-xs text-muted-foreground italic">No zone_id set</span>
                        )}
                        <Button variant="outline" size="sm" className="ml-auto h-6 text-[10px] px-2" onClick={() => setEditingId(true)}>
                            Edit
                        </Button>
                    </div>
                ) : (
                    <div className="flex items-center gap-1">
                        <Input
                            className="h-6 flex-1 text-xs font-mono"
                            value={idValue}
                            onChange={(e) => setIdValue(e.target.value)}
                            autoFocus
                            onKeyDown={(e) => {
                                if (e.key === "Enter") handleCommitId(idValue);
                                if (e.key === "Escape") { setEditingId(false); setIdValue(zoneId); }
                            }}
                            onBlur={() => handleCommitId(idValue)}
                        />
                    </div>
                )}
            </div>
            {zoneId && (
                <p className="text-[10px] text-muted-foreground">
                    Handlers referencing this zone will appear in <code className="font-mono">on_zone_activity</code> hooks
                    with a filter matching <code className="font-mono">{zoneId}</code>.
                </p>
            )}
        </div>
    );
}

// --- All Properties Section ---

function AllPropertiesSection({ object }: { object: TiledSelectedObject }) {
    const [expanded, setExpanded] = useState(false);
    const entries = Object.entries(object.properties).filter(([k]) => k !== "script_ref");

    if (entries.length === 0) return null;

    return (
        <div className="space-y-2">
            <button
                className="flex items-center gap-1 text-sm font-semibold border-b border-border pb-1 w-full text-left"
                onClick={() => setExpanded(!expanded)}
            >
                {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                All Properties
                <span className="text-[10px] font-normal text-muted-foreground ml-1">({entries.length})</span>
            </button>
            {expanded && (
                <div className="space-y-1">
                    {entries.map(([key, val]) => (
                        <div key={key} className="flex items-center gap-2 text-xs">
                            <span className="font-mono text-muted-foreground w-32 truncate shrink-0">{key}</span>
                            <span className="font-mono truncate">{val}</span>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
