import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Plus, Trash2, Check, X } from "lucide-react";
import { getHandlerHookKeys } from "@/lib/scriptUtils";
import type { ParsedScript, HandlerDef } from "@/types/scripts";

interface HandlerListProps {
    script: ParsedScript;
    selectedHandler: string | null;
    onSelect: (name: string) => void;
    onChange: (script: ParsedScript) => void;
}

export function HandlerList({ script, selectedHandler, onSelect, onChange }: HandlerListProps) {
    const [adding, setAdding] = useState(false);
    const [newName, setNewName] = useState("");
    const [renamingKey, setRenamingKey] = useState<string | null>(null);
    const [renameValue, setRenameValue] = useState("");

    const handlerNames = Object.keys(script.handlers);
    const sequenceNames = script.sequences ? Object.keys(script.sequences) : [];

    const handleAdd = () => {
        if (!newName || script.handlers[newName]) return;
        onChange({
            ...script,
            handlers: { ...script.handlers, [newName]: {} },
        });
        onSelect(newName);
        setAdding(false);
        setNewName("");
    };

    const handleDelete = (name: string, e: React.MouseEvent) => {
        e.stopPropagation();
        const next = { ...script.handlers };
        delete next[name];
        onChange({ ...script, handlers: next });
        if (selectedHandler === name) onSelect(handlerNames.find((n) => n !== name) ?? "");
    };

    const handleRename = (oldName: string) => {
        if (!renameValue || renameValue === oldName || script.handlers[renameValue]) {
            setRenamingKey(null);
            return;
        }
        const next: Record<string, HandlerDef> = {};
        for (const [k, v] of Object.entries(script.handlers)) {
            next[k === oldName ? renameValue : k] = v;
        }
        onChange({ ...script, handlers: next });
        if (selectedHandler === oldName) onSelect(renameValue);
        setRenamingKey(null);
    };

    return (
        <div className="flex h-full flex-col">
            <div className="flex items-center justify-between border-b border-border px-2 py-1.5">
                <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Handlers</span>
                <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={() => setAdding(true)}>
                    <Plus className="h-3 w-3" />
                </Button>
            </div>
            <ScrollArea className="flex-1">
                <div className="p-1 space-y-px">
                    {handlerNames.map((name) => {
                        const hookCount = getHandlerHookKeys(script.handlers[name]).length;
                        const isSelected = name === selectedHandler;

                        if (renamingKey === name) {
                            return (
                                <div key={name} className="flex items-center gap-1 px-1 py-0.5">
                                    <Input
                                        className="h-6 flex-1 text-xs font-mono"
                                        value={renameValue}
                                        onChange={(e) => setRenameValue(e.target.value)}
                                        onKeyDown={(e) => {
                                            if (e.key === "Enter") handleRename(name);
                                            if (e.key === "Escape") setRenamingKey(null);
                                        }}
                                        autoFocus
                                    />
                                    <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={() => handleRename(name)}>
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
                                key={name}
                                role="button"
                                tabIndex={0}
                                className={`group flex w-full cursor-pointer items-center gap-1.5 rounded px-2 py-1 text-left text-xs ${
                                    isSelected ? "bg-accent text-accent-foreground" : "hover:bg-accent/50"
                                }`}
                                onClick={() => onSelect(name)}
                                onDoubleClick={() => {
                                    setRenamingKey(name);
                                    setRenameValue(name);
                                }}
                                onKeyDown={(e) => { if (e.key === "Enter") onSelect(name); }}
                            >
                                <span className="flex-1 truncate font-mono">{name}</span>
                                {hookCount > 0 && (
                                    <Badge variant="outline" className="text-[9px] px-1 py-0 shrink-0">{hookCount}</Badge>
                                )}
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-4 w-4 p-0 opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive"
                                    onClick={(e) => handleDelete(name, e)}
                                >
                                    <Trash2 className="h-2.5 w-2.5" />
                                </Button>
                            </div>
                        );
                    })}
                    {adding && (
                        <div className="flex items-center gap-1 px-1 py-0.5">
                            <Input
                                className="h-6 flex-1 text-xs font-mono"
                                placeholder="handler.name"
                                value={newName}
                                onChange={(e) => setNewName(e.target.value)}
                                onKeyDown={(e) => {
                                    if (e.key === "Enter") handleAdd();
                                    if (e.key === "Escape") { setAdding(false); setNewName(""); }
                                }}
                                autoFocus
                            />
                            <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={handleAdd}>
                                <Check className="h-3 w-3" />
                            </Button>
                            <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={() => { setAdding(false); setNewName(""); }}>
                                <X className="h-3 w-3" />
                            </Button>
                        </div>
                    )}
                </div>

                {sequenceNames.length > 0 && (
                    <>
                        <div className="border-t border-border mt-2 pt-1.5 px-2">
                            <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60">Sequences</span>
                        </div>
                        <div className="p-1 space-y-px">
                            {sequenceNames.map((name) => (
                                <div key={name} className="flex items-center gap-1.5 px-2 py-1 text-xs text-muted-foreground">
                                    <span className="font-mono truncate">{name}</span>
                                </div>
                            ))}
                        </div>
                    </>
                )}
            </ScrollArea>
        </div>
    );
}
