import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { apiFetch } from "./client"
import type { Skill, Primortal, CombatInfo } from "@/types/rpg"

export function useSkills() {
    return useQuery({
        queryKey: ["rpg", "skills"],
        queryFn: () => apiFetch<Skill[]>("/rpg/skills"),
    })
}

export function useSkill(id: string) {
    return useQuery({
        queryKey: ["rpg", "skills", id],
        queryFn: () => apiFetch<Skill>(`/rpg/skills/${id}`),
        enabled: !!id,
    })
}

export function useSaveSkill() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: (skill: Skill) =>
            apiFetch<void>(`/rpg/skills/${skill.id}`, {
                method: "PUT",
                body: JSON.stringify(skill),
            }),
        onSuccess: (_data, skill) => {
            qc.invalidateQueries({ queryKey: ["rpg", "skills"] })
            qc.invalidateQueries({ queryKey: ["rpg", "skills", skill.id] })
        },
    })
}

export function useDeleteSkill() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: (id: string) =>
            apiFetch<void>(`/rpg/skills/${id}`, { method: "DELETE" }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["rpg", "skills"] })
        },
    })
}

export function usePrimortals() {
    return useQuery({
        queryKey: ["rpg", "primortals"],
        queryFn: () => apiFetch<Primortal[]>("/rpg/primortals"),
    })
}

export function usePrimortal(type_: string) {
    return useQuery({
        queryKey: ["rpg", "primortals", type_],
        queryFn: () => apiFetch<Primortal>(`/rpg/primortals/${type_}`),
        enabled: !!type_,
    })
}

export function useSavePrimortal() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: (primortal: Primortal) =>
            apiFetch<void>(`/rpg/primortals/${primortal.type}`, {
                method: "PUT",
                body: JSON.stringify(primortal),
            }),
        onSuccess: (_data, primortal) => {
            qc.invalidateQueries({ queryKey: ["rpg", "primortals"] })
            qc.invalidateQueries({ queryKey: ["rpg", "primortals", primortal.type] })
        },
    })
}

export function useDeletePrimortal() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: (type_: string) =>
            apiFetch<void>(`/rpg/primortals/${type_}`, { method: "DELETE" }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["rpg", "primortals"] })
        },
    })
}

export function useCombatInfo() {
    return useQuery({
        queryKey: ["rpg", "combat"],
        queryFn: () => apiFetch<CombatInfo>("/rpg/combat"),
    })
}
