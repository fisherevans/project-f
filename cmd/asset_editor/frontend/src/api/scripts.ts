import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { apiFetch } from "@/api/client"
import type { ScriptFileEntry, ScriptFileDetail, ScriptSchema } from "@/types/scripts"

export const scriptKeys = {
    all: ["scripts"] as const,
    list: () => [...scriptKeys.all, "list"] as const,
    detail: (path: string) => [...scriptKeys.all, "detail", path] as const,
    schema: () => [...scriptKeys.all, "schema"] as const,
}

export function useScriptSchema() {
    return useQuery({
        queryKey: scriptKeys.schema(),
        queryFn: () => apiFetch<ScriptSchema>("/script-schema"),
        staleTime: Infinity,
    })
}

export function useScripts() {
    return useQuery({
        queryKey: scriptKeys.list(),
        queryFn: () => apiFetch<ScriptFileEntry[]>("/scripts"),
    })
}

export function useScript(path: string) {
    return useQuery({
        queryKey: scriptKeys.detail(path),
        queryFn: () => apiFetch<ScriptFileDetail>(`/scripts/${path}`),
        enabled: !!path,
    })
}

export function useSaveScript() {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: ({ path, content }: { path: string; content: string }) =>
            apiFetch<ScriptFileDetail>(`/scripts/${path}`, {
                method: "PUT",
                body: JSON.stringify({ content }),
            }),
        onSuccess: (_data, variables) => {
            queryClient.invalidateQueries({ queryKey: scriptKeys.detail(variables.path) })
            queryClient.invalidateQueries({ queryKey: scriptKeys.list() })
        },
    })
}

export function useDeleteScript() {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: (path: string) =>
            apiFetch<void>(`/scripts/${path}`, { method: "DELETE" }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: scriptKeys.all })
        },
    })
}
