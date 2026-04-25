import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus } from "lucide-react";
import { HookSection } from "./HookSection";
import { getHandlerHookKeys, getAvailableHooks } from "@/lib/scriptUtils";
import type { HandlerDef, RuleDef, ScriptSchema } from "@/types/scripts";

interface HandlerDetailProps {
    handlerName: string;
    handler: HandlerDef;
    schema: ScriptSchema;
    onChange: (handler: HandlerDef) => void;
}

export function HandlerDetail({ handlerName, handler, schema, onChange }: HandlerDetailProps) {
    const hookKeys = getHandlerHookKeys(handler);
    const availableHooks = getAvailableHooks(handler, schema);

    const handleAddHook = (hookKey: string) => {
        onChange({ ...handler, [hookKey]: [] });
    };

    const handleRemoveHook = (hookKey: string) => {
        const next = { ...handler };
        delete next[hookKey];
        onChange(next);
    };

    const handleUpdateHook = (hookKey: string, rules: RuleDef[]) => {
        onChange({ ...handler, [hookKey]: rules });
    };

    return (
        <div className="flex h-full flex-col">
            <div className="flex items-center gap-2 border-b border-border px-3 py-1.5">
                <code className="text-sm font-mono font-semibold">{handlerName}</code>
                <div className="flex-1" />
                {availableHooks.length > 0 && (
                    <Select onValueChange={handleAddHook}>
                        <SelectTrigger className="h-7 w-auto text-xs gap-1">
                            <Plus className="h-3 w-3" />
                            <SelectValue placeholder="Add hook" />
                        </SelectTrigger>
                        <SelectContent>
                            {availableHooks.map((hook) => {
                                const hookDef = Object.values(schema.eventHooks).find((h) => h.yamlKey === hook);
                                return (
                                    <SelectItem key={hook} value={hook}>
                                        <span className="font-mono">{hook}</span>
                                        {hookDef && <span className="ml-2 text-muted-foreground">{hookDef.name}</span>}
                                    </SelectItem>
                                );
                            })}
                        </SelectContent>
                    </Select>
                )}
            </div>
            <ScrollArea className="flex-1">
                <div className="space-y-3 p-3">
                    {hookKeys.length === 0 && (
                        <div className="text-center text-sm text-muted-foreground py-8">
                            No event hooks defined. Add one to get started.
                        </div>
                    )}
                    {hookKeys.map((hookKey) => {
                        const hookDef = Object.values(schema.eventHooks).find((h) => h.yamlKey === hookKey);
                        return (
                            <HookSection
                                key={hookKey}
                                hookKey={hookKey}
                                hookDef={hookDef}
                                rules={handler[hookKey]}
                                schema={schema}
                                onChange={(rules) => handleUpdateHook(hookKey, rules)}
                                onRemove={() => handleRemoveHook(hookKey)}
                            />
                        );
                    })}
                </div>
            </ScrollArea>
        </div>
    );
}
