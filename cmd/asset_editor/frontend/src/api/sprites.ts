import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { apiFetch } from "@/api/client"
import type { SpriteEntry, SpriteDetail, SpriteMetadata, ScaffoldRequest } from "@/types/sprites"

export const spriteKeys = {
    all: ["sprites"] as const,
    list: () => [...spriteKeys.all, "list"] as const,
    detail: (path: string) => [...spriteKeys.all, "detail", path] as const,
}

export function useSprites() {
    return useQuery({
        queryKey: spriteKeys.list(),
        queryFn: () => apiFetch<SpriteEntry[]>("/sprites"),
    })
}

export function useSprite(path: string) {
    return useQuery({
        queryKey: spriteKeys.detail(path),
        queryFn: () => apiFetch<SpriteDetail>(`/sprites/${path}`),
        enabled: !!path,
    })
}

export function useSaveSprite() {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: ({ path, metadata }: { path: string; metadata: SpriteMetadata }) =>
            apiFetch<SpriteDetail>(`/sprites/${path}`, {
                method: "PUT",
                body: JSON.stringify(metadata),
            }),
        onSuccess: (_data, variables) => {
            queryClient.invalidateQueries({ queryKey: spriteKeys.detail(variables.path) })
            queryClient.invalidateQueries({ queryKey: spriteKeys.list() })
        },
    })
}

export function useDeleteSpriteSidecar() {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: (path: string) =>
            apiFetch<void>(`/sprites/${path}`, { method: "DELETE" }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: spriteKeys.all })
        },
    })
}

export function useScaffoldSprite() {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: (req: ScaffoldRequest) =>
            apiFetch<{ status: string; output: string }>("/sprites/_scaffold", {
                method: "POST",
                body: JSON.stringify(req),
            }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: spriteKeys.all })
        },
    })
}
