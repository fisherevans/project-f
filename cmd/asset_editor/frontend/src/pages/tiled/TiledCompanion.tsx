import { useState, useEffect, useCallback, useMemo } from "react";
import { Link } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useScripts, useScript } from "@/api/scripts";
import { useTiledBridgeStatus, sendTiledCommand, tiledBridgeKeys } from "@/api/tiled-bridge";
import type { TiledSelection, TiledSelectedObject } from "@/api/tiled-bridge";
import type { HandlerPropDef, ScriptFileEntry } from "@/types/scripts";
import { parseScript } from "@/lib/scriptUtils";
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
                <AllPropertiesSection object={object} />
            </div>
        </ScrollArea>
    );
}

function ObjectIdentity({ object }: { object: TiledSelectedObject }) {
    return (
        <div className="space-y-1">
            <div className="flex items-center gap-2">
                <span className="text-sm font-semibold">Object #{object.id}</span>
                {object.className && (
                    <span className="text-[10px] px-1.5 py-0.5 rounded bg-accent-teal-tint text-accent-teal font-medium">
                        {object.className}
                    </span>
                )}
            </div>
            {object.properties.entity_id && (
                <p className="text-xs font-mono text-muted-foreground">{object.properties.entity_id}</p>
            )}
            <p className="text-[10px] text-muted-foreground">
                Position: ({Math.round(object.x)}, {Math.round(object.y)})
                {object.width > 0 && ` - Size: ${Math.round(object.width)}x${Math.round(object.height)}`}
            </p>
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
    const [editing, setEditing] = useState(false);
    const [search, setSearch] = useState(scriptRef);

    useEffect(() => {
        setSearch(scriptRef);
        setEditing(false);
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

    const handleAssign = useCallback(
        (handlerName: string) => {
            sendTiledCommand({
                objectId: object.id,
                action: "setProperty",
                name: "script_ref",
                value: handlerName,
            });
            setEditing(false);
        },
        [object.id],
    );

    return (
        <div className="space-y-2">
            <div className="flex items-center gap-2 border-b border-border pb-1">
                <span className="text-sm font-semibold">Handler</span>
                {scriptPath && (
                    <Link
                        to={`/scripts/${scriptPath}`}
                        className="ml-auto text-[10px] text-accent-violet hover:underline flex items-center gap-0.5"
                    >
                        Edit in script editor <ExternalLink className="h-2.5 w-2.5" />
                    </Link>
                )}
            </div>

            {!editing ? (
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
                    <Button variant="outline" size="sm" className="ml-auto h-6 text-[10px] px-2" onClick={() => setEditing(true)}>
                        {scriptRef ? "Change" : "Assign"}
                    </Button>
                </div>
            ) : (
                <div className="space-y-1">
                    <Input
                        className="h-7 text-xs font-mono"
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        placeholder="Search handlers..."
                        autoFocus
                        onKeyDown={(e) => {
                            if (e.key === "Escape") {
                                setEditing(false);
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
                        <Button variant="ghost" size="sm" className="h-6 text-[10px]" onClick={() => { setEditing(false); setSearch(scriptRef); }}>
                            Cancel
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
