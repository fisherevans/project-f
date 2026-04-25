import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { StepEditor } from "./StepEditor";
import { StepKindPicker } from "./StepKindPicker";
import { createEmptyStep, moveItem } from "@/lib/scriptUtils";
import type { StepNode, ScriptSchema } from "@/types/scripts";

interface StepListProps {
    steps: StepNode[];
    schema: ScriptSchema;
    onChange: (steps: StepNode[]) => void;
}

export function StepList({ steps, schema, onChange }: StepListProps) {
    const [showPicker, setShowPicker] = useState(false);

    return (
        <div className="space-y-0.5">
            {steps.map((step, i) => (
                <StepEditor
                    key={i}
                    step={step}
                    schema={schema}
                    onChange={(updated) => {
                        const next = [...steps];
                        next[i] = updated;
                        onChange(next);
                    }}
                    onRemove={() => onChange(steps.filter((_, j) => j !== i))}
                    onMoveUp={i > 0 ? () => onChange(moveItem(steps, i, i - 1)) : undefined}
                    onMoveDown={i < steps.length - 1 ? () => onChange(moveItem(steps, i, i + 1)) : undefined}
                />
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
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 text-xs text-muted-foreground"
                    onClick={() => setShowPicker(true)}
                >
                    <Plus className="mr-1 h-3 w-3" />
                    Add step
                </Button>
            )}
        </div>
    );
}
