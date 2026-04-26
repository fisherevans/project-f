package rpg

import (
    "fmt"
    "io/fs"
    "sort"
    "strings"

    "fisherevans.com/project/f/internal/schema"
    "gopkg.in/yaml.v3"
)

func LoadFromFS(fsys fs.FS) error {
    if err := LoadSkillsFromFS(fsys, "rpg/skills"); err != nil {
        return fmt.Errorf("loading skills: %w", err)
    }
    if err := LoadPrimortalsFromFS(fsys, "rpg/primortals"); err != nil {
        return fmt.Errorf("loading primortals: %w", err)
    }
    return nil
}

func LoadSkillsFromFS(fsys fs.FS, dir string) error {
    entries, err := fs.ReadDir(fsys, dir)
    if err != nil {
        return fmt.Errorf("reading skills dir: %w", err)
    }
    for _, entry := range entries {
        if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
            continue
        }
        data, err := fs.ReadFile(fsys, dir+"/"+entry.Name())
        if err != nil {
            return fmt.Errorf("reading %s: %w", entry.Name(), err)
        }
        var s schema.RPGSkill
        if err := yaml.Unmarshal(data, &s); err != nil {
            return fmt.Errorf("parsing %s: %w", entry.Name(), err)
        }
        sk, err := convertSchemaSkill(s)
        if err != nil {
            return fmt.Errorf("converting %s: %w", entry.Name(), err)
        }
        sk.validate()
        if _, exists := Skills[sk.Id]; exists {
            panic("duplicate skill ID: " + sk.Id)
        }
        Skills[sk.Id] = sk
    }
    return nil
}

func LoadPrimortalsFromFS(fsys fs.FS, dir string) error {
    entries, err := fs.ReadDir(fsys, dir)
    if err != nil {
        return fmt.Errorf("reading primortals dir: %w", err)
    }
    for _, entry := range entries {
        if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
            continue
        }
        data, err := fs.ReadFile(fsys, dir+"/"+entry.Name())
        if err != nil {
            return fmt.Errorf("reading %s: %w", entry.Name(), err)
        }
        var p schema.RPGPrimortal
        if err := yaml.Unmarshal(data, &p); err != nil {
            return fmt.Errorf("parsing %s: %w", entry.Name(), err)
        }
        pr := convertSchemaPrimortal(p)
        pr.register()
    }
    return nil
}

func convertSchemaSkill(s schema.RPGSkill) (Skill, error) {
    ticks := make([]SkillTick, len(s.Ticks))
    for i, t := range s.Ticks {
        tick, err := convertSchemaTick(t)
        if err != nil {
            return Skill{}, fmt.Errorf("tick %d: %w", i, err)
        }
        ticks[i] = tick
    }
    return Skill{
        Id:          SkillId(s.Id),
        Name:        s.Name,
        Description: s.Description,
        Ticks:       ticks,
    }, nil
}

func convertSchemaTick(t schema.RPGSkillTick) (SkillTick, error) {
    stance, err := parseStance(t.Stance)
    if err != nil {
        return SkillTick{}, err
    }
    st := SkillTick{
        StanceType: stance,
    }
    for _, e := range t.Effects {
        st.Effects = append(st.Effects, convertSchemaEffect(e))
    }
    if t.Animation != nil {
        anim, err := convertSchemaAnimation(*t.Animation)
        if err != nil {
            return SkillTick{}, err
        }
        st.AnimationConfig = anim
    }
    return st, nil
}

func parseStance(s string) (CombatStance, error) {
    switch strings.ToLower(s) {
    case "", "none":
        return TickStanceNone, nil
    case "defending":
        return TickStanceDefending, nil
    case "reflecting":
        return TickStanceReflecting, nil
    case "vulnerable":
        return TickStanceVulnerable, nil
    case "exposed":
        return TickStanceExposed, nil
    default:
        return TickStanceNone, fmt.Errorf("unknown stance %q", s)
    }
}

func convertSchemaEffect(e schema.RPGTickEffect) SkillTickEffect {
    out := SkillTickEffect{}
    if e.Damage != nil {
        d := &SkillTickDamage{
            Amount:         e.Damage.Amount,
            RandomVariance: e.Damage.Variance,
            MissRate:       e.Damage.MissRate,
        }
        if e.Damage.ScaledBy != nil {
            d.ScaledBy = SkillTickDamageScalers{
                TargetStatus: convertSchemaStatusScalers(e.Damage.ScaledBy.TargetStatus),
                SourceStatus: convertSchemaStatusScalers(e.Damage.ScaledBy.SourceStatus),
            }
        }
        out.Damage = d
    }
    if e.Status != nil {
        out.Self = e.Status.Target == "self"
        s := &SkillTickStatus{
            Status: StatusType(e.Status.Type),
            Stacks: e.Status.Stacks,
        }
        switch e.Status.Require {
        case "existing_stacks":
            s.RequireExistingStacks = SkillTickStatusRequireExistingStacks
        case "no_stacks":
            s.RequireExistingStacks = SkillTickStatusRequireNoStacks
        }
        out.Status = s
    }
    return out
}

func convertSchemaStatusScalers(m map[string]map[string]float64) map[StatusType]map[StatusLevel]float64 {
    if len(m) == 0 {
        return nil
    }
    out := make(map[StatusType]map[StatusLevel]float64)
    for st, levels := range m {
        lm := make(map[StatusLevel]float64)
        for lvl, mult := range levels {
            switch lvl {
            case "1":
                lm[StatusLevel1] = mult
            case "2":
                lm[StatusLevel2] = mult
            case "3":
                lm[StatusLevel3] = mult
            }
        }
        out[StatusType(st)] = lm
    }
    return out
}

func parseTransformationType(s string) (SkillTickCombatantTransformationType, error) {
    switch strings.ToLower(s) {
    case "", "none":
        return SkillTickCombatantTransformationTypeNone, nil
    case "pounce":
        return SkillTickCombatantTransformationTypePounce, nil
    case "recoil":
        return SkillTickCombatantTransformationTypeRecoil, nil
    case "wiggle":
        return SkillTickCombatantTransformationTypeWiggle, nil
    case "hop":
        return SkillTickCombatantTransformationTypeHop, nil
    default:
        return SkillTickCombatantTransformationTypeNone, fmt.Errorf("unknown transformation type %q", s)
    }
}

func convertSchemaAnimation(a schema.RPGTickAnimation) (SkillTickAnimationConfig, error) {
    cfg := SkillTickAnimationConfig{}
    for _, t := range a.Source {
        typ, err := parseTransformationType(t.Type)
        if err != nil {
            return SkillTickAnimationConfig{}, fmt.Errorf("source: %w", err)
        }
        speed := t.Speed
        if speed == 0 {
            speed = 1.0
        }
        reps := t.Repetitions
        if reps == 0 {
            reps = 1
        }
        cfg.SourceTransformations = append(cfg.SourceTransformations, SkillTickCombatantTransformation{
            Type:        typ,
            Speed:       speed,
            Repetitions: reps,
        })
    }
    for _, t := range a.Target {
        typ, err := parseTransformationType(t.Type)
        if err != nil {
            return SkillTickAnimationConfig{}, fmt.Errorf("target: %w", err)
        }
        speed := t.Speed
        if speed == 0 {
            speed = 1.0
        }
        reps := t.Repetitions
        if reps == 0 {
            reps = 1
        }
        cfg.TargetTransformations = append(cfg.TargetTransformations, SkillTickCombatantTransformation{
            Type:        typ,
            Speed:       speed,
            Repetitions: reps,
        })
    }
    return cfg, nil
}

func convertSchemaPrimortal(p schema.RPGPrimortal) Primortal {
    pr := Primortal{
        Type:         PrimortalType(p.Type),
        Name:         p.Name,
        Description:  p.Description,
        XenoLogIndex: p.XenoLogIndex,
        BaseSync:     p.BaseSync,
    }
    if len(p.UnlockableSkills) > 0 {
        pr.UnlockableSkills = make(map[SkillId]UnlockableSkill)
        for sid, us := range p.UnlockableSkills {
            prereqs := make([]SkillId, len(us.Prerequisites))
            for i, pre := range us.Prerequisites {
                prereqs[i] = SkillId(pre)
            }
            pr.UnlockableSkills[SkillId(sid)] = UnlockableSkill{
                Cost:          us.Cost,
                Prerequisites: prereqs,
            }
        }
    }
    if len(p.CombatArchetypes) > 0 {
        pr.CombatArchetypes = make(map[string]PrimortalCombatArchetype)
        names := make([]string, 0, len(p.CombatArchetypes))
        for name := range p.CombatArchetypes {
            names = append(names, name)
        }
        sort.Strings(names)
        for _, name := range names {
            arch := p.CombatArchetypes[name]
            a := PrimortalCombatArchetype{
                AdditionalSync:         arch.AdditionalSync,
                AdditionalSyncVariance: arch.AdditionalSyncVariance,
            }
            if arch.SkillPool != nil && arch.SkillPool.Random != nil {
                pool := &CombatSkillPoolRandom{}
                for _, sid := range arch.SkillPool.Random.InitialOrderedSkills {
                    pool.InitialOrderedSkills = append(pool.InitialOrderedSkills, SkillId(sid))
                }
                if len(arch.SkillPool.Random.WeightedSkills) > 0 {
                    pool.WeightedSkills = make(map[SkillId]int)
                    for sid, w := range arch.SkillPool.Random.WeightedSkills {
                        pool.WeightedSkills[SkillId(sid)] = w
                    }
                }
                a.SkillPool = CombatSkillPool{Random: pool}
            }
            pr.CombatArchetypes[name] = a
        }
    }
    return pr
}
