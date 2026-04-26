import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/api/client";

export interface TiledSelectedObject {
    id: number;
    name: string;
    className: string;
    x: number;
    y: number;
    width: number;
    height: number;
    properties: Record<string, string>;
}

export interface TiledSelection {
    mapFile: string;
    objects: TiledSelectedObject[];
    time: string;
}

export interface TiledBridgeStatus {
    connected: boolean;
    lastSeen: TiledSelection | null;
}

export interface TiledCommand {
    id?: string;
    objectId: number;
    action: "setProperty" | "removeProperty";
    name: string;
    value?: string;
}

export interface TiledCommandAck {
    id: string;
    ok: boolean;
    message?: string;
}

export const tiledBridgeKeys = {
    all: ["tiled-bridge"] as const,
    selection: () => [...tiledBridgeKeys.all, "selection"] as const,
    status: () => [...tiledBridgeKeys.all, "status"] as const,
};

export function useTiledBridgeStatus() {
    return useQuery({
        queryKey: tiledBridgeKeys.status(),
        queryFn: () => apiFetch<TiledBridgeStatus>("/tiled-bridge/status"),
        refetchInterval: 5000,
    });
}

let commandCounter = 0;
let commandTracker: ((id: string, description: string) => void) | null = null;

export function setCommandTracker(tracker: ((id: string, description: string) => void) | null) {
    commandTracker = tracker;
}

export async function sendTiledCommand(cmd: TiledCommand, description?: string): Promise<string> {
    const id = `web-${Date.now()}-${++commandCounter}`;
    await apiFetch<void>("/tiled-bridge/commands", {
        method: "POST",
        body: JSON.stringify({ ...cmd, id }),
    });
    if (commandTracker) {
        const desc = description ?? `${cmd.action} ${cmd.name} on obj#${cmd.objectId}`;
        commandTracker(id, desc);
    }
    return id;
}
