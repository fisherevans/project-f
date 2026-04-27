import { createContext, useContext, useState, useCallback, type ReactNode } from "react";
import type { RuleDef, StepNode, ConditionNode } from "@/types/scripts";

type ClipboardItem =
    | { type: "rule"; data: RuleDef }
    | { type: "step"; data: StepNode }
    | { type: "steps"; data: StepNode[] }
    | { type: "condition"; data: ConditionNode };

interface ClipboardState {
    item: ClipboardItem | null;
    copyRule: (rule: RuleDef) => void;
    copyStep: (step: StepNode) => void;
    copySteps: (steps: StepNode[]) => void;
    copyCondition: (condition: ConditionNode) => void;
    clear: () => void;
}

const ClipboardContext = createContext<ClipboardState>({
    item: null,
    copyRule: () => {},
    copyStep: () => {},
    copySteps: () => {},
    copyCondition: () => {},
    clear: () => {},
});

export function useScriptClipboard() {
    return useContext(ClipboardContext);
}

function deepClone<T>(obj: T): T {
    return JSON.parse(JSON.stringify(obj));
}

export function ScriptClipboardProvider({ children }: { children: ReactNode }) {
    const [item, setItem] = useState<ClipboardItem | null>(null);

    const copyRule = useCallback((rule: RuleDef) => {
        setItem({ type: "rule", data: deepClone(rule) });
    }, []);

    const copyStep = useCallback((step: StepNode) => {
        setItem({ type: "step", data: deepClone(step) });
    }, []);

    const copySteps = useCallback((steps: StepNode[]) => {
        setItem({ type: "steps", data: deepClone(steps) });
    }, []);

    const copyCondition = useCallback((condition: ConditionNode) => {
        setItem({ type: "condition", data: deepClone(condition) });
    }, []);

    const clear = useCallback(() => setItem(null), []);

    return (
        <ClipboardContext.Provider value={{ item, copyRule, copyStep, copySteps, copyCondition, clear }}>
            {children}
        </ClipboardContext.Provider>
    );
}

export function pasteItem<T>(item: T): T {
    return deepClone(item);
}
