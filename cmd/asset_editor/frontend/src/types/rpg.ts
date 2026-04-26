export interface Skill {
    id: string
    name: string
    description: string
    ticks: SkillTick[]
}

export interface SkillTick {
    stance?: string
    effects?: TickEffect[]
    animation?: TickAnimation
}

export interface TickEffect {
    damage?: TickDamage
    status?: TickStatus
}

export interface TickDamage {
    amount: number
    variance?: number
    miss_rate?: number
    scaled_by?: DamageScaling
}

export interface DamageScaling {
    target_status?: Record<string, Record<string, number>>
    source_status?: Record<string, Record<string, number>>
}

export interface TickStatus {
    type: string
    stacks: number
    target: string
    require?: string
}

export interface TickAnimation {
    source?: Transformation[]
    target?: Transformation[]
}

export interface Transformation {
    type: string
    speed?: number
    repetitions?: number
}

export interface Primortal {
    type: string
    name: string
    description: string
    xeno_log_index?: number
    base_sync: number
    unlockable_skills?: Record<string, UnlockableSkill>
    combat_archetypes?: Record<string, CombatArchetype>
}

export interface UnlockableSkill {
    cost: number
    prerequisites?: string[]
}

export interface CombatArchetype {
    additional_sync?: number
    additional_sync_variance?: number
    skill_pool?: SkillPool
}

export interface SkillPool {
    random?: RandomPool
}

export interface RandomPool {
    initial_ordered_skills?: string[]
    weighted_skills?: Record<string, number>
}

export interface StatusEntry {
    type: string
    name: string
    description: string
}

export interface StanceEntry {
    name: string
    description: string
}

export interface CombatInfo {
    statuses: StatusEntry[]
    stances: StanceEntry[]
}
