import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { apiFetch } from "@/api/client"
import type { AudioEntry, AudioDetail, AudioMetadata } from "@/types/audio"

export const audioKeys = {
    all: ["audio"] as const,
    list: () => [...audioKeys.all, "list"] as const,
    detail: (path: string) => [...audioKeys.all, "detail", path] as const,
}

export function useAudioFiles() {
    return useQuery({
        queryKey: audioKeys.list(),
        queryFn: () => apiFetch<AudioEntry[]>("/audio"),
    })
}

export function useAudioFile(path: string) {
    return useQuery({
        queryKey: audioKeys.detail(path),
        queryFn: () => apiFetch<AudioDetail>(`/audio/${path}`),
        enabled: !!path,
    })
}

export function useSaveAudio() {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: ({ path, metadata }: { path: string; metadata: AudioMetadata }) =>
            apiFetch<AudioDetail>(`/audio/${path}`, {
                method: "PUT",
                body: JSON.stringify(metadata),
            }),
        onSuccess: (_data, variables) => {
            queryClient.invalidateQueries({ queryKey: audioKeys.detail(variables.path) })
            queryClient.invalidateQueries({ queryKey: audioKeys.list() })
        },
    })
}
