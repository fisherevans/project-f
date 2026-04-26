import { useState, useMemo, useRef, useEffect, useCallback } from "react";
import { useLocation } from "react-router-dom";
import { useScriptSchema } from "@/api/scripts";
import { usePageTitle } from "@/hooks/usePageTitle";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Search, ChevronDown, ChevronRight, BookOpen, FileCode } from "lucide-react";
import { getCategoryColor, CATEGORY_ORDER } from "@/components/scripts/StepKindPicker";
import { generateStepExample, generateActionExample, generateConditionExample, generateHookExample } from "@/lib/yamlExamples";
import { ENV_VARS, FUNCTIONS, OPERATORS, EXAMPLES } from "@/lib/exprReferenceData";
import type { ScriptSchema, StepKindDef, CallableDef, ConditionDef, EventHookDef, ParamDef, TemplateVarDef } from "@/types/scripts";

type SectionKey = "steps" | "actions" | "conditions" | "hooks" | "templateVars" | "expressions";

function anchorId(section: string, name: string): string {
    return `${section}-${name}`;
}

export function ScriptReference() {
    const { data: schema, isLoading, error } = useScriptSchema();
    const [search, setSearch] = useState("");
    const [collapsedSections, setCollapsedSections] = useState<Set<SectionKey>>(new Set());
    const contentRef = useRef<HTMLDivElement>(null);
    const location = useLocation();
    usePageTitle("Script Reference");

    const toggleSection = useCallback((key: SectionKey) => {
        setCollapsedSections(prev => {
            const next = new Set(prev);
            if (next.has(key)) next.delete(key);
            else next.add(key);
            return next;
        });
    }, []);

    useEffect(() => {
        if (!schema || !location.hash) return;
        const id = location.hash.slice(1);
        const timeout = setTimeout(() => {
            const el = document.getElementById(id);
            if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
        }, 100);
        return () => clearTimeout(timeout);
    }, [schema, location.hash]);

    if (isLoading) return <div className="p-4 text-muted-foreground">Loading schema...</div>;
    if (error) return <div className="p-4 text-destructive">Error: {error.message}</div>;
    if (!schema) return null;

    return (
        <div className="flex h-full overflow-hidden">
            <SidebarNav schema={schema} search={search} setSearch={setSearch} collapsedSections={collapsedSections} toggleSection={toggleSection} />
            <div ref={contentRef} className="flex-1 overflow-y-auto">
                <ReferenceContent schema={schema} search={search} />
            </div>
        </div>
    );
}

function SidebarNav({ schema, search, setSearch, collapsedSections, toggleSection }: {
    schema: ScriptSchema;
    search: string;
    setSearch: (s: string) => void;
    collapsedSections: Set<SectionKey>;
    toggleSection: (key: SectionKey) => void;
}) {
    const q = search.toLowerCase();

    const stepsByCategory = useMemo(() => {
        const map = new Map<string, StepKindDef[]>();
        for (const def of Object.values(schema.stepKinds)) {
            const cat = def.category;
            if (!map.has(cat)) map.set(cat, []);
            map.get(cat)!.push(def);
        }
        for (const defs of map.values()) defs.sort((a, b) => a.name.localeCompare(b.name));
        return map;
    }, [schema]);

    const filteredSteps = useMemo(() => {
        if (!q) return stepsByCategory;
        const result = new Map<string, StepKindDef[]>();
        for (const [cat, defs] of stepsByCategory) {
            const matching = defs.filter(d => d.name.includes(q) || d.description.toLowerCase().includes(q));
            if (matching.length > 0) result.set(cat, matching);
        }
        return result;
    }, [stepsByCategory, q]);

    const filteredActions = useMemo(() => {
        const all = Object.values(schema.actions).sort((a, b) => a.name.localeCompare(b.name));
        if (!q) return all;
        return all.filter(d => d.name.includes(q) || d.description.toLowerCase().includes(q));
    }, [schema, q]);

    const filteredConditions = useMemo(() => {
        const all = [
            ...Object.values(schema.builtinConditions),
            ...Object.values(schema.conditions),
        ].sort((a, b) => a.name.localeCompare(b.name));
        if (!q) return all;
        return all.filter(d => d.name.includes(q) || d.description.toLowerCase().includes(q));
    }, [schema, q]);

    const filteredHooks = useMemo(() => {
        const all = Object.values(schema.eventHooks).sort((a, b) => a.yamlKey.localeCompare(b.yamlKey));
        if (!q) return all;
        return all.filter(d => d.yamlKey.includes(q) || d.name.toLowerCase().includes(q) || d.description.toLowerCase().includes(q));
    }, [schema, q]);

    const scrollTo = (id: string) => {
        const el = document.getElementById(id);
        if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
    };

    return (
        <div className="w-64 shrink-0 border-r border-border flex flex-col h-full overflow-hidden bg-muted/20">
            <div className="p-3 border-b border-border space-y-2">
                <div className="flex items-center gap-2">
                    <BookOpen className="h-4 w-4 text-accent-violet" />
                    <span className="text-sm font-semibold">Script Reference</span>
                </div>
                <div className="relative">
                    <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-muted-foreground" />
                    <Input
                        className="h-7 text-xs pl-7"
                        placeholder="Search all entries..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                    />
                </div>
            </div>
            <div className="flex-1 overflow-y-auto p-1.5 space-y-1">
                {/* Step Kinds */}
                <NavSection
                    label="Step Kinds"
                    count={Object.keys(schema.stepKinds).length}
                    collapsed={collapsedSections.has("steps")}
                    onToggle={() => toggleSection("steps")}
                    onLabelClick={() => scrollTo("section-steps")}
                >
                    {[...CATEGORY_ORDER.filter(c => filteredSteps.has(c)), ...[...filteredSteps.keys()].filter(c => !CATEGORY_ORDER.includes(c))].map(cat => (
                        <div key={cat}>
                            <div className="px-2 py-0.5 text-[9px] font-semibold uppercase tracking-wider text-muted-foreground/50">{cat}</div>
                            {filteredSteps.get(cat)!.map(def => (
                                <NavEntry key={def.name} name={def.name} onClick={() => scrollTo(anchorId("step", def.name))} />
                            ))}
                        </div>
                    ))}
                </NavSection>

                {/* Actions */}
                <NavSection
                    label="Named Actions"
                    count={Object.keys(schema.actions).length}
                    collapsed={collapsedSections.has("actions")}
                    onToggle={() => toggleSection("actions")}
                    onLabelClick={() => scrollTo("section-actions")}
                >
                    {filteredActions.map(def => (
                        <NavEntry key={def.name} name={def.name} onClick={() => scrollTo(anchorId("action", def.name))} />
                    ))}
                </NavSection>

                {/* Conditions */}
                <NavSection
                    label="Conditions"
                    count={Object.keys(schema.conditions).length + Object.keys(schema.builtinConditions).length}
                    collapsed={collapsedSections.has("conditions")}
                    onToggle={() => toggleSection("conditions")}
                    onLabelClick={() => scrollTo("section-conditions")}
                >
                    {filteredConditions.map(def => (
                        <NavEntry key={def.name} name={def.name} onClick={() => scrollTo(anchorId("condition", def.name))} />
                    ))}
                </NavSection>

                {/* Event Hooks */}
                <NavSection
                    label="Event Hooks"
                    count={Object.keys(schema.eventHooks).length}
                    collapsed={collapsedSections.has("hooks")}
                    onToggle={() => toggleSection("hooks")}
                    onLabelClick={() => scrollTo("section-hooks")}
                >
                    {filteredHooks.map(def => (
                        <NavEntry key={def.yamlKey} name={def.yamlKey} onClick={() => scrollTo(anchorId("hook", def.yamlKey))} />
                    ))}
                </NavSection>

                {/* Template Vars */}
                <NavSection
                    label="Template Variables"
                    count={schema.templateVars.length}
                    collapsed={collapsedSections.has("templateVars")}
                    onToggle={() => toggleSection("templateVars")}
                    onLabelClick={() => scrollTo("section-templateVars")}
                >
                    {schema.templateVars.map(v => (
                        <NavEntry key={v.pattern} name={v.pattern} onClick={() => scrollTo(anchorId("tvar", v.pattern))} />
                    ))}
                </NavSection>

                {/* Expressions */}
                <button
                    className="flex items-center gap-1.5 w-full px-2 py-1 text-xs text-muted-foreground hover:text-foreground rounded hover:bg-muted/50"
                    onClick={() => scrollTo("section-expressions")}
                >
                    <FileCode className="h-3 w-3" />
                    Expressions
                </button>
            </div>
        </div>
    );
}

function NavSection({ label, count, collapsed, onToggle, onLabelClick, children }: {
    label: string;
    count: number;
    collapsed: boolean;
    onToggle: () => void;
    onLabelClick: () => void;
    children: React.ReactNode;
}) {
    return (
        <div>
            <div className="flex items-center gap-1">
                <button className="shrink-0 text-muted-foreground hover:text-foreground p-0.5" onClick={onToggle}>
                    {collapsed ? <ChevronRight className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                </button>
                <button className="flex-1 text-left text-xs font-medium hover:text-foreground text-muted-foreground truncate" onClick={onLabelClick}>
                    {label}
                </button>
                <span className="text-[10px] text-muted-foreground/50 tabular-nums pr-1">{count}</span>
            </div>
            {!collapsed && <div className="ml-3 mt-0.5 space-y-px">{children}</div>}
        </div>
    );
}

function NavEntry({ name, onClick }: { name: string; onClick: () => void }) {
    return (
        <button
            className="block w-full text-left px-2 py-0.5 text-[11px] font-mono text-muted-foreground hover:text-foreground hover:bg-muted/50 rounded truncate"
            onClick={onClick}
        >
            {name}
        </button>
    );
}

function ReferenceContent({ schema, search }: {
    schema: ScriptSchema;
    search: string;
}) {
    const q = search.toLowerCase();

    const stepsByCategory = useMemo(() => {
        const map = new Map<string, StepKindDef[]>();
        for (const def of Object.values(schema.stepKinds)) {
            if (q && !def.name.includes(q) && !def.description.toLowerCase().includes(q)) continue;
            const cat = def.category;
            if (!map.has(cat)) map.set(cat, []);
            map.get(cat)!.push(def);
        }
        for (const defs of map.values()) defs.sort((a, b) => a.name.localeCompare(b.name));
        return map;
    }, [schema, q]);

    const actions = useMemo(() => {
        return Object.values(schema.actions)
            .filter(d => !q || d.name.includes(q) || d.description.toLowerCase().includes(q))
            .sort((a, b) => a.name.localeCompare(b.name));
    }, [schema, q]);

    const conditions = useMemo(() => {
        const all: (CallableDef | ConditionDef)[] = [
            ...Object.values(schema.builtinConditions),
            ...Object.values(schema.conditions),
        ];
        return all
            .filter(d => !q || d.name.includes(q) || d.description.toLowerCase().includes(q))
            .sort((a, b) => a.name.localeCompare(b.name));
    }, [schema, q]);

    const hooks = useMemo(() => {
        return Object.values(schema.eventHooks)
            .filter(d => !q || d.yamlKey.includes(q) || d.name.toLowerCase().includes(q) || d.description.toLowerCase().includes(q))
            .sort((a, b) => a.yamlKey.localeCompare(b.yamlKey));
    }, [schema, q]);

    const templateVars = useMemo(() => {
        return schema.templateVars.filter(v => !q || v.pattern.toLowerCase().includes(q) || v.description.toLowerCase().includes(q));
    }, [schema, q]);

    const categories = [
        ...CATEGORY_ORDER.filter(c => stepsByCategory.has(c)),
        ...[...stepsByCategory.keys()].filter(c => !CATEGORY_ORDER.includes(c)).sort(),
    ];

    return (
        <div className="max-w-3xl mx-auto p-6 space-y-8">
            {/* Header */}
            <div>
                <h1 className="text-lg font-semibold flex items-center gap-2">
                    <BookOpen className="h-5 w-5 text-accent-violet" />
                    Script Reference
                </h1>
                <p className="text-xs text-muted-foreground mt-1">
                    Complete reference for the adventure script system. Step kinds, named actions, conditions, event hooks, template variables, and the expression language.
                </p>
            </div>

            {/* Step Kinds */}
            <section id="section-steps">
                <SectionHeader label="Step Kinds" count={Object.keys(schema.stepKinds).length} description="Each step in a handler rule is a single-key YAML map: { kind: params }." />
                {categories.map(cat => (
                    <div key={cat} className="mt-4">
                        <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground/60 mb-2 flex items-center gap-1.5">
                            <span className={`inline-block w-2 h-2 rounded-sm ${getCategoryColor(cat).split(" ")[0]}`} />
                            {cat}
                            <span className="text-muted-foreground/40">({stepsByCategory.get(cat)!.length})</span>
                        </h3>
                        <div className="space-y-3">
                            {stepsByCategory.get(cat)!.map(def => (
                                <StepKindCard key={def.name} def={def} />
                            ))}
                        </div>
                    </div>
                ))}
                {categories.length === 0 && <EmptySearch />}
            </section>

            {/* Named Actions */}
            <section id="section-actions">
                <SectionHeader label="Named Actions" count={Object.keys(schema.actions).length} description="Invoked via the action step: action: name or action: { name: ..., param: value }." />
                <div className="space-y-3 mt-3">
                    {actions.map(def => (
                        <ActionCard key={def.name} def={def} />
                    ))}
                </div>
                {actions.length === 0 && <EmptySearch />}
            </section>

            {/* Conditions */}
            <section id="section-conditions">
                <SectionHeader label="Conditions" count={Object.keys(schema.conditions).length + Object.keys(schema.builtinConditions).length} description="Used in when clauses and wait_for steps. Built-in conditions (all, any, not, expr) handle logic; named conditions check game state." />
                <div className="space-y-3 mt-3">
                    {conditions.map(def => (
                        <ConditionCard key={def.name} def={def} />
                    ))}
                </div>
                {conditions.length === 0 && <EmptySearch />}
            </section>

            {/* Event Hooks */}
            <section id="section-hooks">
                <SectionHeader label="Event Hooks" count={Object.keys(schema.eventHooks).length} description="Handlers define rules under these YAML keys. Rules fire when the corresponding event occurs." />
                <div className="space-y-3 mt-3">
                    {hooks.map(def => (
                        <HookCard key={def.yamlKey} def={def} />
                    ))}
                </div>
                {hooks.length === 0 && <EmptySearch />}
            </section>

            {/* Template Variables */}
            <section id="section-templateVars">
                <SectionHeader label="Template Variables" count={schema.templateVars.length} description="Available in all string values via {{variable}} syntax." />
                <div className="space-y-2 mt-3">
                    {templateVars.map(v => (
                        <TemplateVarCard key={v.pattern} def={v} />
                    ))}
                </div>
                {templateVars.length === 0 && <EmptySearch />}
            </section>

            {/* Expression Reference */}
            <section id="section-expressions">
                <SectionHeader label="Expression Language" description="Expressions use the expr-lang/expr engine. Used in if.when, switch.on, set_var.value, {{...}} template interpolation, and the expr condition type." />
                <ExpressionReferenceSection />
            </section>

            <div className="pb-8" />
        </div>
    );
}

function SectionHeader({ label, count, description }: { label: string; count?: number; description: string }) {
    return (
        <div className="border-b border-border pb-2">
            <h2 className="text-sm font-semibold flex items-center gap-2">
                {label}
                {count !== undefined && <span className="text-xs font-normal text-muted-foreground/50">({count})</span>}
            </h2>
            <p className="text-[11px] text-muted-foreground mt-0.5">{description}</p>
        </div>
    );
}

function EmptySearch() {
    return <p className="text-xs text-muted-foreground py-4 text-center">No matching entries.</p>;
}

function ParamTable({ params }: { params: ParamDef[] }) {
    if (!params || params.length === 0) return null;
    return (
        <div className="mt-2">
            <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/50 mb-1">Parameters</div>
            <div className="border border-border/50 rounded overflow-hidden">
                <table className="w-full text-xs">
                    <thead>
                        <tr className="bg-muted/30 border-b border-border/30">
                            <th className="text-left px-2 py-1 font-medium text-muted-foreground">Name</th>
                            <th className="text-left px-2 py-1 font-medium text-muted-foreground">Type</th>
                            <th className="text-left px-2 py-1 font-medium text-muted-foreground">Description</th>
                        </tr>
                    </thead>
                    <tbody>
                        {params.map(p => (
                            <tr key={p.name} className="border-b border-border/20 last:border-0">
                                <td className="px-2 py-1 font-mono text-[11px] whitespace-nowrap">
                                    {p.name}
                                    {p.required && <span className="text-accent-red ml-0.5">*</span>}
                                </td>
                                <td className="px-2 py-1 text-muted-foreground whitespace-nowrap">
                                    {p.type}
                                    {p.default !== undefined && <span className="text-muted-foreground/50 ml-1">= {String(p.default)}</span>}
                                    {p.enum && p.enum.length > 0 && (
                                        <div className="flex flex-wrap gap-0.5 mt-0.5">
                                            {p.enum.map(v => (
                                                <span key={v} className="bg-muted/50 rounded px-1 py-px text-[10px] font-mono">{v}</span>
                                            ))}
                                        </div>
                                    )}
                                </td>
                                <td className="px-2 py-1 text-muted-foreground">{p.description}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
}

function YamlBlock({ code }: { code: string }) {
    return (
        <div className="mt-2">
            <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/50 mb-1">Example</div>
            <pre className="bg-canvas text-[11px] font-mono leading-relaxed px-3 py-2 rounded border border-border/30 overflow-x-auto text-foreground/80">
                {code}
            </pre>
        </div>
    );
}

function StepKindCard({ def }: { def: StepKindDef }) {
    const categoryColor = getCategoryColor(def.category);
    const nonStepParams = def.params?.filter(p => p.type !== "steps") ?? [];
    const subStepParams = def.params?.filter(p => p.type === "steps") ?? [];

    const crossLinks: { label: string; anchor: string }[] = [];
    if (def.name === "action") {
        crossLinks.push({ label: "See Named Actions", anchor: "section-actions" });
    } else if (def.name === "custom_action") {
        crossLinks.push({ label: "Custom actions are defined in YAML custom_actions: blocks", anchor: "section-actions" });
    } else if (def.name === "wait_for") {
        crossLinks.push({ label: "See Conditions", anchor: "section-conditions" });
    } else if (def.name === "ref") {
        crossLinks.push({ label: "References reusable sequences defined in the same script file", anchor: "section-steps" });
    }

    return (
        <div id={anchorId("step", def.name)} className="border border-border/50 rounded-md scroll-mt-4">
            <div className="flex items-center gap-2 px-3 py-2 border-b border-border/30 bg-muted/20">
                <code className="text-xs font-mono font-semibold">{def.name}</code>
                <Badge variant="outline" className={`text-[10px] px-1.5 py-0 ${categoryColor}`}>{def.category}</Badge>
                {def.acceptsSubSteps && (
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0 text-muted-foreground">sub-steps</Badge>
                )}
                <span className="text-[10px] text-muted-foreground/50 ml-auto">{def.paramStyle}</span>
            </div>
            <div className="px-3 py-2 space-y-1">
                <p className="text-xs text-muted-foreground">{def.description}</p>
                {nonStepParams.length > 0 && <ParamTable params={nonStepParams} />}
                {subStepParams.length > 0 && (
                    <div className="mt-1">
                        <span className="text-[10px] text-muted-foreground/50">Sub-step fields: </span>
                        {subStepParams.map(p => (
                            <code key={p.name} className="text-[10px] font-mono text-accent-violet mx-0.5">{p.name}</code>
                        ))}
                    </div>
                )}
                <YamlBlock code={generateStepExample(def)} />
                {crossLinks.length > 0 && (
                    <div className="mt-1 flex flex-wrap gap-2">
                        {crossLinks.map(link => (
                            <a key={link.anchor} href={`#${link.anchor}`} className="text-[10px] text-accent-violet hover:underline">{link.label}</a>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}

function ActionCard({ def }: { def: CallableDef }) {
    return (
        <div id={anchorId("action", def.name)} className="border border-border/50 rounded-md scroll-mt-4">
            <div className="flex items-center gap-2 px-3 py-2 border-b border-border/30 bg-muted/20">
                <code className="text-xs font-mono font-semibold">{def.name}</code>
                <Badge variant="outline" className="text-[10px] px-1.5 py-0 bg-accent-teal-tint text-accent-teal">action</Badge>
            </div>
            <div className="px-3 py-2 space-y-1">
                <p className="text-xs text-muted-foreground">{def.description}</p>
                {def.params && def.params.length > 0 && <ParamTable params={def.params} />}
                <YamlBlock code={generateActionExample(def)} />
            </div>
        </div>
    );
}

function ConditionCard({ def }: { def: CallableDef | ConditionDef }) {
    const isBuiltin = "isComposite" in def;
    const isComposite = "isComposite" in def && def.isComposite;
    return (
        <div id={anchorId("condition", def.name)} className="border border-border/50 rounded-md scroll-mt-4">
            <div className="flex items-center gap-2 px-3 py-2 border-b border-border/30 bg-muted/20">
                <code className="text-xs font-mono font-semibold">{def.name}</code>
                {isBuiltin && <Badge variant="outline" className="text-[10px] px-1.5 py-0 bg-accent-amber-tint text-accent-amber">builtin</Badge>}
                {isComposite && <Badge variant="outline" className="text-[10px] px-1.5 py-0 bg-accent-blue-tint text-accent-blue">composite</Badge>}
                {!isBuiltin && <Badge variant="outline" className="text-[10px] px-1.5 py-0 bg-accent-teal-tint text-accent-teal">named</Badge>}
            </div>
            <div className="px-3 py-2 space-y-1">
                <p className="text-xs text-muted-foreground">{def.description}</p>
                {def.params && def.params.length > 0 && <ParamTable params={def.params} />}
                <YamlBlock code={generateConditionExample(def)} />
            </div>
        </div>
    );
}

function HookCard({ def }: { def: EventHookDef }) {
    return (
        <div id={anchorId("hook", def.yamlKey)} className="border border-border/50 rounded-md scroll-mt-4">
            <div className="flex items-center gap-2 px-3 py-2 border-b border-border/30 bg-muted/20">
                <code className="text-xs font-mono font-semibold">{def.yamlKey}</code>
                <span className="text-[11px] text-muted-foreground">{def.name}</span>
                <span className="text-[10px] text-muted-foreground/50 ml-auto font-mono">{def.eventType}</span>
            </div>
            <div className="px-3 py-2 space-y-1">
                <p className="text-xs text-muted-foreground">{def.description}</p>
                {def.filterFields && def.filterFields.length > 0 && <ParamTable params={def.filterFields} />}
                <YamlBlock code={generateHookExample(def)} />
            </div>
        </div>
    );
}

function TemplateVarCard({ def }: { def: TemplateVarDef }) {
    return (
        <div id={anchorId("tvar", def.pattern)} className="flex items-start gap-3 py-2 border-b border-border/20 last:border-0 scroll-mt-4">
            <code className="text-xs font-mono font-semibold text-accent-violet shrink-0">{def.pattern}</code>
            <span className="text-xs text-muted-foreground">{def.description}</span>
        </div>
    );
}

function ExpressionReferenceSection() {
    return (
        <div className="space-y-6 mt-3">
            {/* Environment */}
            <div>
                <h3 className="text-xs font-semibold mb-2">Environment Variables</h3>
                <p className="text-[11px] text-muted-foreground mb-2">Variables available in expression fields. Map types support dot access and bracket indexing.</p>
                <div className="border border-border/50 rounded overflow-hidden">
                    <table className="w-full text-xs">
                        <thead>
                            <tr className="bg-muted/30 border-b border-border/30">
                                <th className="text-left px-2 py-1 font-medium text-muted-foreground">Name</th>
                                <th className="text-left px-2 py-1 font-medium text-muted-foreground">Type</th>
                                <th className="text-left px-2 py-1 font-medium text-muted-foreground">Description</th>
                            </tr>
                        </thead>
                        <tbody>
                            {ENV_VARS.map(v => (
                                <tr key={v.name} className="border-b border-border/20 last:border-0">
                                    <td className="px-2 py-1 font-mono text-[11px] text-accent-violet font-semibold">{v.name}</td>
                                    <td className="px-2 py-1 text-muted-foreground/60">{v.type}</td>
                                    <td className="px-2 py-1 text-muted-foreground">{v.desc}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Functions */}
            <div>
                <h3 className="text-xs font-semibold mb-2">Built-in Functions</h3>
                <div className="space-y-1">
                    {FUNCTIONS.map(f => (
                        <div key={f.name} className="py-1.5 border-b border-border/20 last:border-0">
                            <div className="flex items-start gap-2">
                                <code className="text-[11px] font-mono font-semibold text-accent-teal shrink-0">{f.name}</code>
                                <span className="text-xs text-muted-foreground">{f.desc}</span>
                            </div>
                            <code className="text-[10px] font-mono text-muted-foreground/50 ml-2">{f.example}</code>
                        </div>
                    ))}
                </div>
            </div>

            {/* Operators */}
            <div>
                <h3 className="text-xs font-semibold mb-2">Operators</h3>
                <div className="space-y-1">
                    {OPERATORS.map(o => (
                        <div key={o.op} className="flex items-start gap-3 py-1 border-b border-border/20 last:border-0">
                            <code className="text-[11px] font-mono font-semibold text-accent-amber shrink-0 w-40">{o.op}</code>
                            <span className="text-xs text-muted-foreground">{o.desc}</span>
                        </div>
                    ))}
                </div>
            </div>

            {/* Examples */}
            <div>
                <h3 className="text-xs font-semibold mb-2">Common Patterns</h3>
                <div className="space-y-1">
                    {EXAMPLES.map(e => (
                        <div key={e.expr} className="py-1.5 border-b border-border/20 last:border-0">
                            <code className="text-[11px] font-mono text-foreground/80">{e.expr}</code>
                            <div className="text-[10px] text-muted-foreground mt-0.5">{e.desc}</div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}
