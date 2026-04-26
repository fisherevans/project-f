import { createContext, useContext, useState, type ReactNode } from "react";
import { ExpressionHelpModal } from "./ExpressionHelpModal";
import type { CustomActionDef } from "@/types/scripts";

export interface CrossFileEntry {
    name: string;
    file: string;
}

interface ExprContextValue {
    handlerVarKeys: string[];
    constKeys: string[];
    otherConstKeys: CrossFileEntry[];
    otherCustomActionNames: CrossFileEntry[];
    customActions: Record<string, CustomActionDef>;
    openHelp: () => void;
}

const ExprCtx = createContext<ExprContextValue>({
    handlerVarKeys: [],
    constKeys: [],
    otherConstKeys: [],
    otherCustomActionNames: [],
    customActions: {},
    openHelp: () => {},
});

export function useExprContext() {
    return useContext(ExprCtx);
}

export function ExprContextProvider({
    handlerVarKeys,
    constKeys,
    otherConstKeys,
    otherCustomActionNames,
    customActions,
    children,
}: {
    handlerVarKeys: string[];
    constKeys: string[];
    otherConstKeys?: CrossFileEntry[];
    otherCustomActionNames?: CrossFileEntry[];
    customActions?: Record<string, CustomActionDef>;
    children: ReactNode;
}) {
    const [helpOpen, setHelpOpen] = useState(false);

    return (
        <ExprCtx.Provider value={{
            handlerVarKeys,
            constKeys,
            otherConstKeys: otherConstKeys ?? [],
            otherCustomActionNames: otherCustomActionNames ?? [],
            customActions: customActions ?? {},
            openHelp: () => setHelpOpen(true),
        }}>
            {children}
            {helpOpen && <ExpressionHelpModal onClose={() => setHelpOpen(false)} />}
        </ExprCtx.Provider>
    );
}
