import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Plus, ClipboardPaste } from "lucide-react";
import { StepEditor } from "./StepEditor";
import { StepKindPicker } from "./StepKindPicker";
import { useScriptClipboard, pasteItem } from "./ScriptClipboard";
import { createEmptyStep, moveItem } from "@/lib/scriptUtils";
import type { StepNode, ScriptSchema } from "@/types/scripts";

interface StepListProps {
    steps: StepNode[];
    schema: ScriptSchema;
    onChange: (steps: StepNode[]) => void;
}

export function StepList({ steps, schema, onChange }: StepListProps) {
    const [showPicker, setShowPicker] = useState(false);
    const clipboard = useScriptClipboard();
    const canPasteStep = clipboard.item?.type === "step" || clipboard.item?.type === "steps";

    const pasteStepAt = (index: number) => {
        if (clipboard.item?.type === "step") {
            const pasted = pasteItem(clipboard.item.data);
            const next = [...steps];
            next.splice(index, 0, pasted);
            onChange(next);
        } else if (clipboard.item?.type === "steps") {
            const pasted = pasteItem(clipboard.item.data);
            const next = [...steps];
            next.splice(index, 0, ...pasted);
            onChange(next);
        }
    };

    const cloneStep = (index: number) => {
        const cloned: StepNode = JSON.parse(JSON.stringify(steps[index]));
        const next = [...steps];
        next.splice(index + 1, 0, cloned);
        onChange(next);
    };

    return (
        <div className="space-y-1">
            {steps.map((step, i) => (
                <div key={i} className="group/step">
                    <StepEditor
                        step={step}
                        stepIndex={i}
                        schema={schema}
                        onChange={(updated) => {
                            const next = [...steps];
                            next[i] = updated;
                            onChange(next);
                        }}
                        onRemove={() => onChange(steps.filter((_, j) => j !== i))}
                        onMoveUp={i > 0 ? () => onChange(moveItem(steps, i, i - 1)) : undefined}
                        onMoveDown={i < steps.length - 1 ? () => onChange(moveItem(steps, i, i + 1)) : undefined}
                        onClone={() => cloneStep(i)}
                        onPasteBefore={canPasteStep ? () => pasteStepAt(i) : undefined}
                        onPasteAfter={canPasteStep ? () => pasteStepAt(i + 1) : undefined}
                    />
                </div>
            ))}
            {showPicker ? (
                <StepKindPicker
                    schema={schema}
                    onSelect={(kind) => {
                        const stepDef = schema.stepKinds[kind];
                        onChange([...steps, createEmptyStep(kind, stepDef)]);
                        setShowPicker(false);
                    }}
                    onCancel={() => setShowPicker(false)}
                />
            ) : (
                <div className="flex items-center gap-1">
                    <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 text-xs text-muted-foreground"
                        onClick={() => setShowPicker(true)}
                    >
                        <Plus className="mr-1 h-3 w-3" />
                        Add step
                    </Button>
                    {canPasteStep && steps.length === 0 && (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 text-xs text-accent-violet"
                            onClick={() => pasteStepAt(0)}
                        >
                            <ClipboardPaste className="mr-1 h-3 w-3" />
                            Paste
                        </Button>
                    )}
                </div>
            )}
        </div>
    );
}
