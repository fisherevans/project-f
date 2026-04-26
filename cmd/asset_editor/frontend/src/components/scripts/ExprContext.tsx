import { createContext, useContext, useState, type ReactNode } from "react";
import { ExpressionHelpModal } from "./ExpressionHelpModal";

interface ExprContextValue {
    handlerVarKeys: string[];
    constKeys: string[];
    openHelp: () => void;
}

const ExprCtx = createContext<ExprContextValue>({
    handlerVarKeys: [],
    constKeys: [],
    openHelp: () => {},
});

export function useExprContext() {
    return useContext(ExprCtx);
}

export function ExprContextProvider({
    handlerVarKeys,
    constKeys,
    children,
}: {
    handlerVarKeys: string[];
    constKeys: string[];
    children: ReactNode;
}) {
    const [helpOpen, setHelpOpen] = useState(false);

    return (
        <ExprCtx.Provider value={{ handlerVarKeys, constKeys, openHelp: () => setHelpOpen(true) }}>
            {children}
            {helpOpen && <ExpressionHelpModal onClose={() => setHelpOpen(false)} />}
        </ExprCtx.Provider>
    );
}
