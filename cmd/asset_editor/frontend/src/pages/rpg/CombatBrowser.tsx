import { useCombatInfo } from "@/api/rpg"
import { usePageTitle } from "@/hooks/usePageTitle"

const statusColors: Record<string, string> = {
    warded: "border-accent-blue-edge bg-accent-blue-tint",
    burning: "border-accent-orange-edge bg-accent-orange-tint",
    poisoned: "border-accent-green-edge bg-accent-green-tint",
    ionized: "border-accent-yellow-edge bg-accent-yellow-tint",
    mending: "border-accent-teal-edge bg-accent-teal-tint",
}

const stanceColors: Record<string, string> = {
    Defending: "border-accent-blue-edge bg-accent-blue-tint",
    Reflecting: "border-accent-violet-edge bg-accent-violet-tint",
    Vulnerable: "border-accent-orange-edge bg-accent-orange-tint",
    Exposed: "border-accent-red-edge bg-accent-red-tint",
}

export function CombatBrowser() {
    const { data, isLoading, error } = useCombatInfo()
    usePageTitle("Combat - RPG")

    if (isLoading) return <div className="p-6 text-sm text-muted-foreground">Loading...</div>
    if (error) return <div className="p-6 text-sm text-destructive">Failed to load combat info</div>
    if (!data) return null

    return (
        <div className="h-full overflow-auto p-6">
            <div className="max-w-2xl space-y-8">
                <section>
                    <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-1">Status Effects</h2>
                    <p className="text-xs text-muted-foreground mb-3">
                        Status effects are stacking conditions applied during combat via skill tick effects. Each stack adds to the effect's intensity. Stacks are consumed or decay based on the status type. Used in skill editors as effect targets and in damage scaling conditions.
                    </p>
                    <div className="grid gap-2">
                        {data.statuses.map(s => (
                            <div
                                key={s.type}
                                className={`rounded-md border p-3 ${statusColors[s.type] ?? "border-border bg-muted/30"}`}
                            >
                                <div className="flex items-baseline gap-2 mb-1">
                                    <span className="text-sm font-medium">{s.name}</span>
                                    <span className="font-mono text-[10px] text-muted-foreground">{s.type}</span>
                                </div>
                                <p className="text-xs text-muted-foreground">{s.description}</p>
                            </div>
                        ))}
                    </div>
                </section>

                <section>
                    <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider mb-1">Combat Stances</h2>
                    <p className="text-xs text-muted-foreground mb-3">
                        Stances modify damage taken or dealt for the duration of the stance. Set per-tick in the skill editor. Stances that modify vulnerability must last at least 2 ticks to give the opponent a window to exploit or endure them.
                    </p>
                    <div className="grid gap-2">
                        {data.stances.map(s => (
                            <div
                                key={s.name}
                                className={`rounded-md border p-3 ${stanceColors[s.name] ?? "border-border bg-muted/30"}`}
                            >
                                <span className="text-sm font-medium">{s.name}</span>
                                <p className="text-xs text-muted-foreground mt-1">{s.description}</p>
                            </div>
                        ))}
                    </div>
                </section>
            </div>
        </div>
    )
}
