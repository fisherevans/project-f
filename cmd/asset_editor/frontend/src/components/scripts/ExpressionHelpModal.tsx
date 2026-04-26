import { useState } from "react";
import { X, HelpCircle } from "lucide-react";
import { ENV_VARS, FUNCTIONS, OPERATORS, EXAMPLES } from "@/lib/exprReferenceData";

export function ExpressionHelpModal({ onClose }: { onClose: () => void }) {
    const [tab, setTab] = useState<"env" | "functions" | "operators" | "examples">("env");

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
            <div className="w-[640px] max-h-[80vh] rounded-lg border border-border bg-popover shadow-xl flex flex-col" onClick={(e) => e.stopPropagation()}>
                <div className="flex items-center gap-2 border-b border-border px-4 py-2.5">
                    <HelpCircle className="h-4 w-4 text-accent-violet" />
                    <span className="text-sm font-semibold">Expression Reference</span>
                    <div className="flex-1" />
                    <button className="text-muted-foreground hover:text-foreground" onClick={onClose}>
                        <X className="h-4 w-4" />
                    </button>
                </div>
                <div className="flex gap-1 px-4 py-2 border-b border-border/50">
                    {(["env", "functions", "operators", "examples"] as const).map((t) => (
                        <button
                            key={t}
                            className={`text-xs px-2 py-1 rounded ${tab === t ? "bg-accent-violet-tint text-accent-violet font-medium" : "text-muted-foreground hover:text-foreground"}`}
                            onClick={() => setTab(t)}
                        >
                            {t === "env" ? "Environment" : t.charAt(0).toUpperCase() + t.slice(1)}
                        </button>
                    ))}
                </div>
                <div className="flex-1 overflow-y-auto p-4">
                    {tab === "env" && (
                        <div className="space-y-1">
                            <p className="text-xs text-muted-foreground mb-3">
                                Variables available in expression fields. Map types support dot access and bracket indexing.
                            </p>
                            {ENV_VARS.map((v) => (
                                <div key={v.name} className="flex items-start gap-2 py-1.5 border-b border-border/30 last:border-0">
                                    <code className="text-xs font-mono font-semibold text-accent-violet shrink-0 w-16">{v.name}</code>
                                    <span className="text-[10px] text-muted-foreground/60 shrink-0 w-12">{v.type}</span>
                                    <span className="text-xs text-muted-foreground">{v.desc}</span>
                                </div>
                            ))}
                        </div>
                    )}
                    {tab === "functions" && (
                        <div className="space-y-1">
                            <p className="text-xs text-muted-foreground mb-3">
                                Built-in functions. All are available in any expression field.
                            </p>
                            {FUNCTIONS.map((f) => (
                                <div key={f.name} className="py-1.5 border-b border-border/30 last:border-0">
                                    <div className="flex items-start gap-2">
                                        <code className="text-xs font-mono font-semibold text-accent-teal shrink-0">{f.name}</code>
                                        <span className="text-xs text-muted-foreground">{f.desc}</span>
                                    </div>
                                    <code className="text-[11px] font-mono text-muted-foreground/60 ml-2">{f.example}</code>
                                </div>
                            ))}
                        </div>
                    )}
                    {tab === "operators" && (
                        <div className="space-y-1">
                            <p className="text-xs text-muted-foreground mb-3">
                                Operators follow standard precedence. Parentheses override.
                            </p>
                            {OPERATORS.map((o) => (
                                <div key={o.op} className="flex items-start gap-3 py-1.5 border-b border-border/30 last:border-0">
                                    <code className="text-xs font-mono font-semibold text-accent-amber shrink-0 w-44">{o.op}</code>
                                    <span className="text-xs text-muted-foreground">{o.desc}</span>
                                </div>
                            ))}
                        </div>
                    )}
                    {tab === "examples" && (
                        <div className="space-y-1">
                            <p className="text-xs text-muted-foreground mb-3">
                                Common expression patterns for script steps and conditions.
                            </p>
                            {EXAMPLES.map((e) => (
                                <div key={e.expr} className="py-1.5 border-b border-border/30 last:border-0">
                                    <code className="text-xs font-mono text-foreground">{e.expr}</code>
                                    <div className="text-[11px] text-muted-foreground mt-0.5">{e.desc}</div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
                <div className="border-t border-border px-4 py-2">
                    <p className="text-[10px] text-muted-foreground/50">
                        Powered by expr-lang/expr. String interpolation uses {"{{expr}}"} syntax. Bare expression fields (if.when, switch.on, set_var.value) evaluate without braces.
                    </p>
                </div>
            </div>
        </div>
    );
}

export function ExpressionHelpLink({ onClick }: { onClick: () => void }) {
    return (
        <button
            className="inline-flex items-center gap-0.5 text-[10px] text-accent-violet/60 hover:text-accent-violet"
            onClick={onClick}
        >
            <HelpCircle className="h-2.5 w-2.5" />
            <span>expr reference</span>
        </button>
    );
}
