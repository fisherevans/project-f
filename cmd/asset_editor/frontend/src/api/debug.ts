import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type {
    StateInfo,
    GlobalEntry,
    EntityListResponse,
    EntityDetail,
    TeleportEntry,
    ZoneRect,
    CommandResponse,
} from "../types/debug";

const DEBUG_URL = "/api/v1/debug";

async function debugFetch<T>(path: string, options?: RequestInit): Promise<T> {
    const res = await fetch(`${DEBUG_URL}${path}`, {
        ...options,
        headers: {
            "Content-Type": "application/json",
            ...options?.headers,
        },
    });
    if (!res.ok) {
        const text = await res.text().catch(() => res.statusText);
        throw new Error(`${res.status}: ${text}`);
    }
    if (res.status === 204) return undefined as T;
    return res.json() as Promise<T>;
}

export function useGameState() {
    return useQuery({
        queryKey: ["debug", "state"],
        queryFn: () => debugFetch<StateInfo>("/state"),
        refetchInterval: 2000,
        retry: false,
    });
}

export function useGameSave() {
    return useQuery({
        queryKey: ["debug", "save"],
        queryFn: () => debugFetch<unknown>("/save"),
        retry: false,
    });
}

export function useGlobals(prefix?: string) {
    return useQuery({
        queryKey: ["debug", "globals", prefix],
        queryFn: () => {
            const params = prefix ? `?prefix=${encodeURIComponent(prefix)}` : "";
            return debugFetch<GlobalEntry[]>(`/globals${params}`);
        },
        retry: false,
    });
}

export function useSetGlobal() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ key, type, value }: { key: string; type: string; value: unknown }) =>
            debugFetch<GlobalEntry>(`/globals/${key}`, {
                method: "POST",
                body: JSON.stringify({ type, value }),
            }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["debug", "globals"] });
        },
    });
}

export function useDeleteGlobal() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (key: string) =>
            debugFetch<GlobalEntry>(`/globals/${key}`, { method: "DELETE" }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["debug", "globals"] });
        },
    });
}

export function useRunCommand() {
    return useMutation({
        mutationFn: (command: string) =>
            debugFetch<CommandResponse>("/command", {
                method: "POST",
                body: JSON.stringify({ command }),
            }),
    });
}

export function useTeleport() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (req: { target?: string; x?: number; y?: number }) =>
            debugFetch<CommandResponse>("/teleport", {
                method: "POST",
                body: JSON.stringify(req),
            }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["debug", "state"] });
            qc.invalidateQueries({ queryKey: ["debug", "entities"] });
        },
    });
}

export function useLoadMap() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (req: { name: string; waypoint?: string }) =>
            debugFetch<CommandResponse>("/map", {
                method: "POST",
                body: JSON.stringify(req),
            }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["debug"] });
        },
    });
}

export function useEntities() {
    return useQuery({
        queryKey: ["debug", "entities"],
        queryFn: () => debugFetch<EntityListResponse>("/entities"),
        refetchInterval: 3000,
        retry: false,
    });
}

export function useEntity(id: string) {
    return useQuery({
        queryKey: ["debug", "entities", id],
        queryFn: () => debugFetch<EntityDetail>(`/entities/${id}`),
        enabled: !!id,
        retry: false,
    });
}

export function useTeleports() {
    return useQuery({
        queryKey: ["debug", "teleports"],
        queryFn: () => debugFetch<TeleportEntry[]>("/teleports"),
        retry: false,
    });
}

export function useZones() {
    return useQuery({
        queryKey: ["debug", "zones"],
        queryFn: () => debugFetch<ZoneRect[]>("/zones"),
        retry: false,
    });
}
