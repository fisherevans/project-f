import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { apiFetch } from "./client"
import type { OverlayFlowEntry, OverlayFlowDetail, NamedRectsDetail, OverlayFlow, OverlayRect } from "@/types/overlays"

export const overlayKeys = {
    all: ["overlays"] as const,
    list: () => [...overlayKeys.all, "list"] as const,
    detail: (name: string) => [...overlayKeys.all, "detail", name] as const,
    rects: () => [...overlayKeys.all, "rects"] as const,
}

export function useOverlays() {
    return useQuery({
        queryKey: overlayKeys.list(),
        queryFn: () => apiFetch<OverlayFlowEntry[]>("/overlays"),
    })
}

export function useOverlay(name: string) {
    return useQuery({
        queryKey: overlayKeys.detail(name),
        queryFn: () => apiFetch<OverlayFlowDetail>(`/overlays/${name}`),
        enabled: !!name,
    })
}

export function useSaveOverlay() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: ({ name, content }: { name: string; content: string }) =>
            apiFetch<void>(`/overlays/${name}`, {
                method: "PUT",
                body: content,
                headers: { "Content-Type": "text/plain" },
            }),
        onSuccess: (_data, { name }) => {
            qc.invalidateQueries({ queryKey: overlayKeys.list() })
            qc.invalidateQueries({ queryKey: overlayKeys.detail(name) })
        },
    })
}

export function useDeleteOverlay() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: (name: string) =>
            apiFetch<void>(`/overlays/${name}`, { method: "DELETE" }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: overlayKeys.list() })
        },
    })
}

export function useNamedRects() {
    return useQuery({
        queryKey: overlayKeys.rects(),
        queryFn: () => apiFetch<NamedRectsDetail>("/overlays/rects"),
    })
}

export function useSaveNamedRects() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: (content: string) =>
            apiFetch<void>("/overlays/rects", {
                method: "PUT",
                body: content,
                headers: { "Content-Type": "text/plain" },
            }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: overlayKeys.rects() })
        },
    })
}

export interface PreviewHighlightRequest {
    flowName?: string
    flow?: OverlayFlow
    rects?: Record<string, OverlayRect>
    duration?: number
}

const DEBUG_API = "http://localhost:8091/api/v1"

export function usePreviewHighlight() {
    return useMutation({
        mutationFn: (req: PreviewHighlightRequest) =>
            fetch(`${DEBUG_API}/debug/highlight`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(req),
            }).then(res => {
                if (!res.ok) throw new Error(`Preview failed: ${res.statusText}`)
            }),
    })
}

export function useDismissHighlight() {
    return useMutation({
        mutationFn: () =>
            fetch(`${DEBUG_API}/debug/highlight`, { method: "DELETE" }).then(res => {
                if (!res.ok) throw new Error(`Dismiss failed: ${res.statusText}`)
            }),
    })
}
