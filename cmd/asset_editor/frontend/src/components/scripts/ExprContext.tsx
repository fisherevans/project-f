import { createContext, useContext, useState, type ReactNode } from "react";
import { ExpressionHelpModal } from "./ExpressionHelpModal";
import type { CustomActionDef } from "@/types/scripts";

interface ExprContextValue {
    handlerVarKeys: string[];
    constKeys: string[];
    customActions: Record<string, CustomActionDef>;
    openHelp: () => void;
}

const ExprCtx = createContext<ExprContextValue>({
    handlerVarKeys: [],
    constKeys: [],
    customActions: {},
    openHelp: () => {},
});

export function useExprContext() {
    return useContext(ExprCtx);
}

export function ExprContextProvider({
    handlerVarKeys,
    constKeys,
    customActions,
    children,
}: {
    handlerVarKeys: string[];
    constKeys: string[];
    customActions?: Record<string, CustomActionDef>;
    children: ReactNode;
}) {
    const [helpOpen, setHelpOpen] = useState(false);

    return (
        <ExprCtx.Provider value={{ handlerVarKeys, constKeys, customActions: customActions ?? {}, openHelp: () => setHelpOpen(true) }}>
            {children}
            {helpOpen && <ExpressionHelpModal onClose={() => setHelpOpen(false)} />}
        </ExprCtx.Provider>
    );
}
