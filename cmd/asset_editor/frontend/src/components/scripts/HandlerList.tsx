import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Plus, Trash2, Check, X } from "lucide-react";
import { getHandlerHookKeys } from "@/lib/scriptUtils";
import type { ParsedScript, HandlerDef, ScriptItemSelection, CustomActionDef, SequenceDef } from "@/types/scripts";

interface HandlerListProps {
    script: ParsedScript;
    selection: ScriptItemSelection | null;
    onSelect: (sel: ScriptItemSelection | null) => void;
    onChange: (script: ParsedScript) => void;
}

type AddingType = "handler" | "custom_action" | "sequence" | "const" | "data" | "property_template" | null;

function isSelected(selection: ScriptItemSelection | null, type: ScriptItemSelection["type"], name: string): boolean {
    return selection?.type === type && selection.name === name;
}

export function HandlerList({ script, selection, onSelect, onChange }: HandlerListProps) {
    const [adding, setAdding] = useState<AddingType>(null);
    const [newName, setNewName] = useState("");
    const [renamingKey, setRenamingKey] = useState<{ type: ScriptItemSelection["type"]; name: string } | null>(null);
    const [renameValue, setRenameValue] = useState("");

    const handlerNames = Object.keys(script.handlers);
    const sequenceNames = script.sequences ? Object.keys(script.sequences) : [];
    const customActionNames = script.custom_actions ? Object.keys(script.custom_actions) : [];
    const constNames = script.consts ? Object.keys(script.consts) : [];
    const dataNames = script.data ? Object.keys(script.data) : [];
    const templateNames = script.property_templates ? Object.keys(script.property_templates) : [];

    const startAdd = (type: AddingType) => {
        setAdding(type);
        setNewName("");
    };

    const cancelAdd = () => {
        setAdding(null);
        setNewName("");
    };

    const handleAdd = () => {
        if (!newName || !adding) return;
        switch (adding) {
            case "handler":
                if (script.handlers[newName]) return;
                onChange({ ...script, handlers: { ...script.handlers, [newName]: {} } });
                onSelect({ type: "handler", name: newName });
                break;
            case "custom_action":
                if (script.custom_actions?.[newName]) return;
                onChange({
                    ...script,
                    custom_actions: { ...(script.custom_actions ?? {}), [newName]: { steps: [] } },
                });
                onSelect({ type: "custom_action", name: newName });
                break;
            case "sequence":
                if (script.sequences?.[newName]) return;
                onChange({
                    ...script,
                    sequences: { ...(script.sequences ?? {}), [newName]: { steps: [] } },
                });
                onSelect({ type: "sequence", name: newName });
                break;
            case "const":
                if (script.consts?.[newName] !== undefined) return;
                onChange({
                    ...script,
                    consts: { ...(script.consts ?? {}), [newName]: "" },
                });
                onSelect({ type: "const", name: newName });
                break;
            case "data":
                if (script.data?.[newName]) return;
                onChange({
                    ...script,
                    data: { ...(script.data ?? {}), [newName]: [] },
                });
                onSelect({ type: "data", name: newName });
                break;
            case "property_template":
                if (script.property_templates?.[newName]) return;
                onChange({
                    ...script,
                    property_templates: { ...(script.property_templates ?? {}), [newName]: {} },
                });
                onSelect({ type: "property_template", name: newName });
                break;
        }
        cancelAdd();
    };

    const handleDelete = (type: ScriptItemSelection["type"], name: string, e: React.MouseEvent) => {
        e.stopPropagation();
        switch (type) {
            case "handler": {
                const next = { ...script.handlers };
                delete next[name];
                onChange({ ...script, handlers: next });
                break;
            }
            case "custom_action": {
                const next = { ...(script.custom_actions ?? {}) };
                delete next[name];
                onChange({ ...script, custom_actions: Object.keys(next).length > 0 ? next : undefined });
                break;
            }
            case "sequence": {
                const next = { ...(script.sequences ?? {}) };
                delete next[name];
                onChange({ ...script, sequences: Object.keys(next).length > 0 ? next : undefined });
                break;
            }
            case "const": {
                const next = { ...(script.consts ?? {}) };
                delete next[name];
                onChange({ ...script, consts: Object.keys(next).length > 0 ? next : undefined });
                break;
            }
            case "data": {
                const next = { ...(script.data ?? {}) };
                delete next[name];
                onChange({ ...script, data: Object.keys(next).length > 0 ? next : undefined });
                break;
            }
            case "property_template": {
                const next = { ...(script.property_templates ?? {}) };
                delete next[name];
                onChange({ ...script, property_templates: Object.keys(next).length > 0 ? next : undefined });
                break;
            }
        }
        if (isSelected(selection, type, name)) {
            onSelect(null);
        }
    };

    const handleRename = (type: ScriptItemSelection["type"], oldName: string) => {
        if (!renameValue || renameValue === oldName) {
            setRenamingKey(null);
            return;
        }
        switch (type) {
            case "handler": {
                if (script.handlers[renameValue]) { setRenamingKey(null); return; }
                const next: Record<string, HandlerDef> = {};
                for (const [k, v] of Object.entries(script.handlers)) {
                    next[k === oldName ? renameValue : k] = v;
                }
                onChange({ ...script, handlers: next });
                break;
            }
            case "custom_action": {
                if (script.custom_actions?.[renameValue]) { setRenamingKey(null); return; }
                const next: Record<string, CustomActionDef> = {};
                for (const [k, v] of Object.entries(script.custom_actions ?? {})) {
                    next[k === oldName ? renameValue : k] = v;
                }
                onChange({ ...script, custom_actions: next });
                break;
            }
            case "sequence": {
                if (script.sequences?.[renameValue]) { setRenamingKey(null); return; }
                const next: Record<string, SequenceDef> = {};
                for (const [k, v] of Object.entries(script.sequences ?? {})) {
                    next[k === oldName ? renameValue : k] = v;
                }
                onChange({ ...script, sequences: next });
                break;
            }
            case "const": {
                if (script.consts?.[renameValue] !== undefined) { setRenamingKey(null); return; }
                const next: Record<string, unknown> = {};
                for (const [k, v] of Object.entries(script.consts ?? {})) {
                    next[k === oldName ? renameValue : k] = v;
                }
                onChange({ ...script, consts: next });
                break;
            }
            case "data": {
                if (script.data?.[renameValue]) { setRenamingKey(null); return; }
                const next: Record<string, string[]> = {};
                for (const [k, v] of Object.entries(script.data ?? {})) {
                    next[k === oldName ? renameValue : k] = v;
                }
                onChange({ ...script, data: next });
                break;
            }
            case "property_template": {
                if (script.property_templates?.[renameValue]) { setRenamingKey(null); return; }
                const next: Record<string, Record<string, unknown>> = {};
                for (const [k, v] of Object.entries(script.property_templates ?? {})) {
                    next[k === oldName ? renameValue : k] = v;
                }
                onChange({ ...script, property_templates: next });
                break;
            }
        }
        if (isSelected(selection, type, oldName)) {
            onSelect({ type, name: renameValue });
        }
        setRenamingKey(null);
    };

    const renderAddRow = (type: AddingType, placeholder: string) => {
        if (adding !== type) return null;
        return (
            <div className="flex items-center gap-1 px-1 py-0.5">
                <Input
                    className="h-6 flex-1 text-xs font-mono"
                    placeholder={placeholder}
                    value={newName}
                    onChange={(e) => setNewName(e.target.value)}
                    onKeyDown={(e) => {
                        if (e.key === "Enter") handleAdd();
                        if (e.key === "Escape") cancelAdd();
                    }}
                    autoFocus
                />
                <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={handleAdd}>
                    <Check className="h-3 w-3" />
                </Button>
                <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={cancelAdd}>
                    <X className="h-3 w-3" />
                </Button>
            </div>
        );
    };

    const renderItem = (
        type: ScriptItemSelection["type"],
        name: string,
        badge?: React.ReactNode,
    ) => {
        const selected = isSelected(selection, type, name);
        const isRenaming = renamingKey?.type === type && renamingKey.name === name;

        if (isRenaming) {
            return (
                <div key={`${type}-${name}`} className="flex items-center gap-1 px-1 py-0.5">
                    <Input
                        className="h-6 flex-1 text-xs font-mono"
                        value={renameValue}
                        onChange={(e) => setRenameValue(e.target.value)}
                        onKeyDown={(e) => {
                            if (e.key === "Enter") handleRename(type, name);
                            if (e.key === "Escape") setRenamingKey(null);
                        }}
                        autoFocus
                    />
                    <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={() => handleRename(type, name)}>
                        <Check className="h-3 w-3" />
                    </Button>
                    <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={() => setRenamingKey(null)}>
                        <X className="h-3 w-3" />
                    </Button>
                </div>
            );
        }

        return (
            <div
                key={`${type}-${name}`}
                role="button"
                tabIndex={0}
                className={`group flex w-full cursor-pointer items-center gap-1.5 rounded px-2 py-1 text-left text-xs ${
                    selected ? "bg-accent text-accent-foreground" : "hover:bg-accent/50"
                }`}
                onClick={() => onSelect({ type, name })}
                onDoubleClick={() => {
                    setRenamingKey({ type, name });
                    setRenameValue(name);
                }}
                onKeyDown={(e) => { if (e.key === "Enter") onSelect({ type, name }); }}
            >
                <span className="flex-1 truncate font-mono">{name}</span>
                {badge}
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-4 w-4 p-0 opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive"
                    onClick={(e) => handleDelete(type, name, e)}
                >
                    <Trash2 className="h-2.5 w-2.5" />
                </Button>
            </div>
        );
    };

    const renderSectionHeader = (label: string, type: AddingType) => (
        <div className="flex items-center justify-between border-t border-border mt-2 pt-1.5 px-2">
            <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60">{label}</span>
            <Button variant="ghost" size="sm" className="h-4 w-4 p-0" onClick={() => startAdd(type)}>
                <Plus className="h-2.5 w-2.5 text-muted-foreground/60" />
            </Button>
        </div>
    );

    return (
        <div className="flex h-full flex-col overflow-hidden">
            <div className="flex items-center justify-between border-b border-border px-2 py-1.5">
                <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Handlers</span>
                <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={() => startAdd("handler")}>
                    <Plus className="h-3 w-3" />
                </Button>
            </div>
            <ScrollArea className="flex-1 overflow-hidden">
                <div className="p-1 space-y-px">
                    {handlerNames.map((name) => {
                        const hookCount = getHandlerHookKeys(script.handlers[name]).length;
                        return renderItem(
                            "handler",
                            name,
                            hookCount > 0 ? (
                                <Badge variant="outline" className="text-[9px] px-1 py-0 shrink-0">{hookCount}</Badge>
                            ) : undefined,
                        );
                    })}
                    {renderAddRow("handler", "handler.name")}
                </div>

                {(customActionNames.length > 0 || adding === "custom_action") && (
                    <>
                        {renderSectionHeader("Custom Actions", "custom_action")}
                        <div className="p-1 space-y-px">
                            {customActionNames.map((name) =>
                                renderItem(
                                    "custom_action",
                                    name,
                                    script.custom_actions?.[name]?.params ? (
                                        <span className="text-[10px] text-muted-foreground/50 shrink-0">
                                            ({script.custom_actions[name].params!.length})
                                        </span>
                                    ) : undefined,
                                ),
                            )}
                            {renderAddRow("custom_action", "action_name")}
                        </div>
                    </>
                )}
                {customActionNames.length === 0 && adding !== "custom_action" && (
                    <>
                        {renderSectionHeader("Custom Actions", "custom_action")}
                    </>
                )}

                {(sequenceNames.length > 0 || adding === "sequence") && (
                    <>
                        {renderSectionHeader("Sequences", "sequence")}
                        <div className="p-1 space-y-px">
                            {sequenceNames.map((name) => renderItem("sequence", name))}
                            {renderAddRow("sequence", "sequence_name")}
                        </div>
                    </>
                )}
                {sequenceNames.length === 0 && adding !== "sequence" && (
                    <>
                        {renderSectionHeader("Sequences", "sequence")}
                    </>
                )}

                {(dataNames.length > 0 || adding === "data") && (
                    <>
                        {renderSectionHeader("Data Lists", "data")}
                        <div className="p-1 space-y-px">
                            {dataNames.map((name) => {
                                const count = script.data?.[name]?.length ?? 0;
                                return renderItem(
                                    "data",
                                    name,
                                    count > 0 ? (
                                        <span className="text-[10px] text-muted-foreground/50 shrink-0">[{count}]</span>
                                    ) : undefined,
                                );
                            })}
                            {renderAddRow("data", "list_name")}
                        </div>
                    </>
                )}
                {dataNames.length === 0 && adding !== "data" && (
                    <>
                        {renderSectionHeader("Data Lists", "data")}
                    </>
                )}

                {(constNames.length > 0 || adding === "const") && (
                    <>
                        {renderSectionHeader("Constants", "const")}
                        <div className="p-1 space-y-px">
                            {constNames.map((name) => {
                                const val = script.consts![name];
                                const typeHint = Array.isArray(val) ? `[${(val as unknown[]).length}]` : typeof val === "object" ? "map" : typeof val;
                                return renderItem(
                                    "const",
                                    name,
                                    <span className="text-[10px] text-muted-foreground/50 shrink-0">{typeHint}</span>,
                                );
                            })}
                            {renderAddRow("const", "const_name")}
                        </div>
                    </>
                )}
                {constNames.length === 0 && adding !== "const" && (
                    <>
                        {renderSectionHeader("Constants", "const")}
                    </>
                )}

                {(templateNames.length > 0 || adding === "property_template") && (
                    <>
                        {renderSectionHeader("Property Templates", "property_template")}
                        <div className="p-1 space-y-px">
                            {templateNames.map((name) => {
                                const propCount = Object.keys(script.property_templates![name]).length;
                                return renderItem(
                                    "property_template",
                                    name,
                                    propCount > 0 ? (
                                        <span className="text-[10px] text-muted-foreground/50 shrink-0">{propCount} props</span>
                                    ) : undefined,
                                );
                            })}
                            {renderAddRow("property_template", "template.name")}
                        </div>
                    </>
                )}
                {templateNames.length === 0 && adding !== "property_template" && (
                    <>
                        {renderSectionHeader("Property Templates", "property_template")}
                    </>
                )}
            </ScrollArea>
        </div>
    );
}
