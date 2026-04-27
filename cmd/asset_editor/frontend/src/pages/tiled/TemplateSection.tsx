import { useState, useEffect, useCallback, useMemo } from "react";
import { Link } from "react-router-dom";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
    ChevronDown,
    ChevronRight,
    ExternalLink,
    FileCode,
    Pencil,
    RotateCcw,
} from "lucide-react";
import { usePropertyTemplates } from "@/api/scripts";
import { sendTiledCommand } from "@/api/tiled-bridge";
import type { TiledSelectedObject } from "@/api/tiled-bridge";
import type { PropertyTemplateEntry } from "@/types/scripts";
import { getKnownProperty, KNOWN_ENTITY_PROPERTIES } from "@/lib/entityProperties";

interface TemplateSectionProps {
    object: TiledSelectedObject;
}

export interface ResolvedTemplateData {
    templateName: string;
    templateEntry: PropertyTemplateEntry;
}

export function useResolvedTemplate(object: TiledSelectedObject): ResolvedTemplateData | null {
    const { data: templates } = usePropertyTemplates();
    const templateName = object.properties.template;
    return useMemo(() => {
        if (!templateName || !templates) return null;
        const entry = templates.find((t) => t.name === templateName);
        if (!entry) return null;
        return { templateName, templateEntry: entry };
    }, [templateName, templates]);
}

export function resolveProperty(
    object: TiledSelectedObject,
    templateData: ResolvedTemplateData | null,
    key: string,
): { value: string; source: "entity" | "template" } | null {
    const entityValue = object.properties[key];
    if (entityValue !== undefined && entityValue !== "") {
        return { value: entityValue, source: "entity" };
    }
    if (templateData) {
        const tmplValue = templateData.templateEntry.properties[key];
        if (tmplValue !== undefined && tmplValue !== null && tmplValue !== "") {
            return { value: String(tmplValue), source: "template" };
        }
    }
    return null;
}

export function TemplateSection({ object }: TemplateSectionProps) {
    const { data: templates } = usePropertyTemplates();
    const templateName = object.properties.template ?? "";
    const [mode, setMode] = useState<"view" | "pick">("view");
    const [expanded, setExpanded] = useState(true);

    useEffect(() => {
        setMode("view");
    }, [templateName]);

    const templateEntry = useMemo(() => {
        if (!templateName || !templates) return null;
        return templates.find((t) => t.name === templateName) ?? null;
    }, [templateName, templates]);

    const templateNotFound = templateName && !templateEntry;

    return (
        <div className="space-y-2">
            <button
                className="flex items-center gap-1 text-sm font-semibold border-b border-border pb-1 w-full text-left"
                onClick={() => setExpanded(!expanded)}
            >
                {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                Template
                {templateName && (
                    <span className="text-[10px] font-mono font-normal text-accent-violet ml-1">
                        {templateName}
                    </span>
                )}
            </button>
            {expanded && (
                <div className="space-y-2">
                    {mode === "view" && (
                        <TemplateDisplay
                            object={object}
                            templateName={templateName}
                            templateEntry={templateEntry}
                            templateNotFound={templateNotFound}
                            onPickMode={() => setMode("pick")}
                        />
                    )}
                    {mode === "pick" && (
                        <TemplatePicker
                            object={object}
                            currentTemplate={templateName}
                            templates={templates ?? []}
                            onCancel={() => setMode("view")}
                        />
                    )}
                    {templateEntry && mode === "view" && (
                        <TemplatePropertiesView object={object} templateEntry={templateEntry} />
                    )}
                </div>
            )}
        </div>
    );
}

function TemplateDisplay({
    object,
    templateName,
    templateEntry,
    templateNotFound,
    onPickMode,
}: {
    object: TiledSelectedObject;
    templateName: string;
    templateEntry: PropertyTemplateEntry | null;
    templateNotFound: boolean;
    onPickMode: () => void;
}) {
    const handleRemove = useCallback(() => {
        sendTiledCommand({
            objectId: object.id,
            action: "removeProperty",
            name: "template",
        }, "Remove template");
    }, [object.id]);

    if (!templateName) {
        return (
            <div className="flex items-center gap-2">
                <span className="text-xs text-muted-foreground italic">No template assigned</span>
                <Button
                    variant="outline"
                    size="sm"
                    className="h-6 text-[10px] px-2 ml-auto"
                    onClick={onPickMode}
                >
                    Assign
                </Button>
            </div>
        );
    }

    return (
        <div className="space-y-1">
            <div className="flex items-center gap-2">
                <span className="text-[10px] px-1.5 py-0.5 rounded bg-accent-violet-tint text-accent-violet font-mono font-medium">
                    {templateName}
                </span>
                {templateNotFound && (
                    <span className="text-[10px] text-accent-red">not found</span>
                )}
                <div className="ml-auto flex items-center gap-1">
                    <Button variant="outline" size="sm" className="h-6 text-[10px] px-2" onClick={onPickMode}>
                        Change
                    </Button>
                    <Button variant="outline" size="sm" className="h-6 text-[10px] px-2" onClick={handleRemove}>
                        Remove
                    </Button>
                </div>
            </div>
            {templateEntry && (
                <Link
                    to={`/scripts/${templateEntry.scriptFile}`}
                    className="text-[10px] text-muted-foreground hover:underline flex items-center gap-0.5"
                >
                    <FileCode className="h-2.5 w-2.5" />
                    {templateEntry.scriptFile}
                    <ExternalLink className="h-2 w-2" />
                </Link>
            )}
        </div>
    );
}

function TemplatePicker({
    object,
    currentTemplate,
    templates,
    onCancel,
}: {
    object: TiledSelectedObject;
    currentTemplate: string;
    templates: PropertyTemplateEntry[];
    onCancel: () => void;
}) {
    const [search, setSearch] = useState("");

    const filtered = useMemo(() => {
        const q = search.toLowerCase();
        return templates.filter((t) => {
            if (!q) return true;
            return t.name.toLowerCase().includes(q) || t.scriptFile.toLowerCase().includes(q);
        }).slice(0, 30);
    }, [templates, search]);

    const handleSelect = useCallback((name: string) => {
        sendTiledCommand({
            objectId: object.id,
            action: "setProperty",
            name: "template",
            value: name,
        }, `Assign template ${name}`);
    }, [object.id]);

    return (
        <div className="space-y-1">
            <Input
                className="h-7 text-xs font-mono"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search templates..."
                autoFocus
                onKeyDown={(e) => {
                    if (e.key === "Escape") onCancel();
                }}
            />
            <div className="border border-border rounded-md max-h-48 overflow-y-auto">
                {filtered.map((t) => {
                    const propCount = Object.keys(t.properties).length;
                    const hints = Object.entries(t.properties)
                        .slice(0, 3)
                        .map(([k, v]) => `${k}=${v}`)
                        .join(", ");
                    return (
                        <button
                            key={t.name}
                            className="flex w-full items-start gap-2 px-2 py-1.5 text-left text-xs hover:bg-accent border-b border-border last:border-b-0"
                            onClick={() => handleSelect(t.name)}
                        >
                            <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-1.5">
                                    <span className="font-mono font-medium truncate">{t.name}</span>
                                    {t.name === currentTemplate && (
                                        <span className="text-[9px] px-1 rounded bg-accent-green-tint text-accent-green">current</span>
                                    )}
                                    <span className="text-[10px] text-muted-foreground/50">{propCount} props</span>
                                </div>
                                {hints && (
                                    <p className="text-[10px] text-muted-foreground truncate">{hints}</p>
                                )}
                            </div>
                            <span className="text-[10px] text-muted-foreground shrink-0 truncate max-w-[120px]">
                                {t.scriptFile}
                            </span>
                        </button>
                    );
                })}
                {filtered.length === 0 && (
                    <p className="px-2 py-2 text-[10px] text-muted-foreground">No templates match</p>
                )}
            </div>
            <div className="flex justify-end">
                <Button variant="ghost" size="sm" className="h-6 text-[10px]" onClick={onCancel}>
                    Cancel
                </Button>
            </div>
        </div>
    );
}

function TemplatePropertiesView({
    object,
    templateEntry,
}: {
    object: TiledSelectedObject;
    templateEntry: PropertyTemplateEntry;
}) {
    const templateProps = Object.entries(templateEntry.properties);
    if (templateProps.length === 0) return null;

    return (
        <div className="space-y-1.5">
            <span className="text-[10px] font-medium text-muted-foreground uppercase tracking-wider">
                Template Properties
            </span>
            {templateProps.map(([key, defaultValue]) => (
                <TemplatePropertyRow
                    key={key}
                    propKey={key}
                    defaultValue={defaultValue}
                    object={object}
                />
            ))}
        </div>
    );
}

function TemplatePropertyRow({
    propKey,
    defaultValue,
    object,
}: {
    propKey: string;
    defaultValue: unknown;
    object: TiledSelectedObject;
}) {
    const entityValue = object.properties[propKey];
    const hasOverride = entityValue !== undefined && entityValue !== "" && entityValue !== String(defaultValue);
    const displayValue = hasOverride ? entityValue : String(defaultValue ?? "");
    const known = getKnownProperty(propKey);
    const [editing, setEditing] = useState(false);
    const [editValue, setEditValue] = useState(displayValue);

    useEffect(() => {
        setEditValue(hasOverride ? entityValue : String(defaultValue ?? ""));
        setEditing(false);
    }, [entityValue, defaultValue, hasOverride]);

    const handleCommit = useCallback((val: string) => {
        const trimmed = val.trim();
        if (trimmed === String(defaultValue ?? "")) {
            // Value matches default - remove the override
            sendTiledCommand({
                objectId: object.id,
                action: "removeProperty",
                name: propKey,
            }, `Reset ${propKey} to template default`);
        } else if (trimmed !== (entityValue ?? "")) {
            sendTiledCommand({
                objectId: object.id,
                action: "setProperty",
                name: propKey,
                value: trimmed,
            }, `Set ${propKey}=${trimmed}`);
        }
        setEditing(false);
    }, [object.id, propKey, entityValue, defaultValue]);

    const handleReset = useCallback(() => {
        sendTiledCommand({
            objectId: object.id,
            action: "removeProperty",
            name: propKey,
        }, `Reset ${propKey} to template default`);
    }, [object.id, propKey]);

    const accentClass = hasOverride
        ? "border-l-accent-amber-edge"
        : "border-l-accent-violet-edge";

    return (
        <div className={`border-l-2 ${accentClass} pl-2 space-y-0.5`}>
            <div className="flex items-center gap-1.5">
                <span className="text-[11px] font-mono font-medium">{propKey}</span>
                {known && (
                    <span className="text-[10px] text-muted-foreground/50">{known.label}</span>
                )}
                {hasOverride && (
                    <span className="text-[9px] px-1 rounded bg-accent-amber-tint text-accent-amber">overridden</span>
                )}
                {!hasOverride && (
                    <span className="text-[9px] px-1 rounded bg-accent-violet-tint text-accent-violet">default</span>
                )}
            </div>
            {!editing ? (
                <div className="flex items-center gap-1.5">
                    <TemplatePropertyDisplay propKey={propKey} value={displayValue} known={known} isDefault={!hasOverride} />
                    <div className="ml-auto flex items-center gap-1 shrink-0">
                        <button
                            className="text-muted-foreground hover:text-foreground"
                            onClick={() => { setEditValue(displayValue); setEditing(true); }}
                        >
                            <Pencil className="h-2.5 w-2.5" />
                        </button>
                        {hasOverride && (
                            <button
                                className="text-muted-foreground hover:text-accent-violet"
                                onClick={handleReset}
                                title="Reset to template default"
                            >
                                <RotateCcw className="h-2.5 w-2.5" />
                            </button>
                        )}
                    </div>
                </div>
            ) : (
                <TemplatePropertyEditor
                    propKey={propKey}
                    known={known}
                    value={editValue}
                    onCommit={handleCommit}
                    onCancel={() => setEditing(false)}
                />
            )}
            {hasOverride && (
                <p className="text-[10px] text-muted-foreground/40 italic">
                    template default: {String(defaultValue ?? "")}
                </p>
            )}
        </div>
    );
}

function TemplatePropertyDisplay({
    propKey,
    value,
    known,
    isDefault,
}: {
    propKey: string;
    value: string;
    known: ReturnType<typeof getKnownProperty>;
    isDefault: boolean;
}) {
    const textClass = isDefault
        ? "text-xs font-mono text-muted-foreground italic"
        : "text-xs font-mono";

    if (known?.editorKind === "enum" && known.enumValues) {
        const label = value || "(default)";
        return <span className={textClass}>{label}</span>;
    }

    if (known?.editorKind === "yaml-textarea" && value) {
        const lines = value.split("\n");
        const preview = lines.length > 1 ? `${lines[0]}... (${lines.length} lines)` : value;
        return <span className={`${textClass} truncate`}>{preview}</span>;
    }

    return <span className={`${textClass} truncate`}>{value || "(empty)"}</span>;
}

function TemplatePropertyEditor({
    propKey,
    known,
    value,
    onCommit,
    onCancel,
}: {
    propKey: string;
    known: ReturnType<typeof getKnownProperty>;
    value: string;
    onCommit: (value: string) => void;
    onCancel: () => void;
}) {
    const [val, setVal] = useState(value);

    if (known?.editorKind === "enum" && known.enumValues) {
        return (
            <select
                className="h-6 w-full text-xs font-mono rounded border border-border bg-background px-1"
                value={val}
                autoFocus
                onChange={(e) => { setVal(e.target.value); onCommit(e.target.value); }}
                onBlur={onCancel}
                onKeyDown={(e) => { if (e.key === "Escape") onCancel(); }}
            >
                {known.enumValues.map((v) => (
                    <option key={v} value={v}>{v || "(default)"}</option>
                ))}
            </select>
        );
    }

    if (known?.editorKind === "yaml-textarea") {
        return (
            <div className="space-y-1">
                <textarea
                    className="w-full min-h-[60px] text-xs font-mono rounded border border-border bg-background px-2 py-1 resize-y"
                    value={val}
                    onChange={(e) => setVal(e.target.value)}
                    autoFocus
                    rows={4}
                    onKeyDown={(e) => {
                        if (e.key === "Escape") onCancel();
                    }}
                />
                <div className="flex gap-1 justify-end">
                    <Button variant="ghost" size="sm" className="h-5 text-[10px] px-1.5" onClick={onCancel}>
                        Cancel
                    </Button>
                    <Button size="sm" className="h-5 text-[10px] px-1.5" onClick={() => onCommit(val)}>
                        Apply
                    </Button>
                </div>
            </div>
        );
    }

    return (
        <Input
            className="h-6 text-xs font-mono"
            value={val}
            onChange={(e) => setVal(e.target.value)}
            autoFocus
            onBlur={() => onCommit(val)}
            onKeyDown={(e) => {
                if (e.key === "Enter") onCommit(val);
                if (e.key === "Escape") onCancel();
            }}
        />
    );
}
