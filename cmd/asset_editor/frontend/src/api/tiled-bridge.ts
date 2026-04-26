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
    objectId: number;
    action: "setProperty" | "removeProperty";
    name: string;
    value?: string;
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

export async function sendTiledCommand(cmd: TiledCommand): Promise<void> {
    await apiFetch<void>("/tiled-bridge/commands", {
        method: "POST",
        body: JSON.stringify(cmd),
    });
}
