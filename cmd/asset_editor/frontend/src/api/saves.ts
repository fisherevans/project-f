import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { apiFetch } from "./client"
import type { SaveSummary, SaveDetail } from "@/types/saves"

export function useSaves() {
    return useQuery({
        queryKey: ["saves"],
        queryFn: () => apiFetch<SaveSummary[]>("/saves"),
    })
}

export function useSave(id: string) {
    return useQuery({
        queryKey: ["saves", id],
        queryFn: () => apiFetch<SaveDetail>(`/saves/${id}`),
        enabled: !!id,
    })
}

export function useSaveSave() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: ({ id, rawYaml }: { id: string; rawYaml: string }) =>
            apiFetch<void>(`/saves/${id}`, {
                method: "PUT",
                body: JSON.stringify({ raw_yaml: rawYaml }),
            }),
        onSuccess: (_data, { id }) => {
            qc.invalidateQueries({ queryKey: ["saves"] })
            qc.invalidateQueries({ queryKey: ["saves", id] })
        },
    })
}

export function useDeleteSave() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: (id: string) =>
            apiFetch<void>(`/saves/${id}`, { method: "DELETE" }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["saves"] })
        },
    })
}

export function useCloneSave() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: ({ srcId, newId }: { srcId: string; newId: string }) =>
            apiFetch<void>(`/saves/${srcId}/clone`, {
                method: "POST",
                body: JSON.stringify({ new_id: newId }),
            }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["saves"] })
        },
    })
}
