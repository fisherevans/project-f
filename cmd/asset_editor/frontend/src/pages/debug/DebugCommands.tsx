import { useState, useRef, useEffect } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { useRunCommand } from "@/api/debug"
import { usePageTitle } from "@/hooks/usePageTitle"

interface HistoryEntry {
    command: string
    output: string[]
    error?: string
}

export function DebugCommands() {
    usePageTitle("Commands - Debug")

    const [input, setInput] = useState("")
    const [history, setHistory] = useState<HistoryEntry[]>([])
    const [cmdHistory, setCmdHistory] = useState<string[]>([])
    const [historyPos, setHistoryPos] = useState(-1)
    const runCommand = useRunCommand()
    const outputRef = useRef<HTMLDivElement>(null)

    useEffect(() => {
        outputRef.current?.scrollTo(0, outputRef.current.scrollHeight)
    }, [history])

    function handleSubmit() {
        const cmd = input.trim()
        if (!cmd) return
        setInput("")
        setCmdHistory(prev => [...prev, cmd])
        setHistoryPos(-1)

        runCommand.mutate(cmd, {
            onSuccess: (data) => {
                setHistory(prev => [...prev, { command: cmd, output: data.output }])
            },
            onError: (err) => {
                setHistory(prev => [...prev, { command: cmd, output: [], error: String(err) }])
            },
        })
    }

    function handleKeyDown(e: React.KeyboardEvent) {
        if (e.key === "Enter") {
            handleSubmit()
        } else if (e.key === "ArrowUp") {
            e.preventDefault()
            if (cmdHistory.length === 0) return
            const next = historyPos === -1 ? cmdHistory.length - 1 : Math.max(0, historyPos - 1)
            setHistoryPos(next)
            setInput(cmdHistory[next] ?? "")
        } else if (e.key === "ArrowDown") {
            e.preventDefault()
            if (historyPos === -1) return
            const next = historyPos + 1
            if (next >= cmdHistory.length) {
                setHistoryPos(-1)
                setInput("")
            } else {
                setHistoryPos(next)
                setInput(cmdHistory[next] ?? "")
            }
        }
    }

    return (
        <div className="flex h-full flex-col gap-3 p-6">
            <div
                ref={outputRef}
                className="flex-1 overflow-auto rounded-md border bg-card p-4 font-mono text-xs leading-relaxed"
            >
                {history.length === 0 && (
                    <span className="text-muted-foreground">
                        Type a console command below. Same commands as the in-game console (tp, map, global, broadcast, etc.)
                    </span>
                )}
                {history.map((entry, i) => (
                    <div key={i} className="mb-3">
                        <div className="text-primary font-medium">$ {entry.command}</div>
                        {entry.output.map((line, j) => (
                            <div key={j} className="pl-2 text-foreground">{line}</div>
                        ))}
                        {entry.error && (
                            <div className="pl-2 text-destructive">{entry.error}</div>
                        )}
                    </div>
                ))}
            </div>

            <div className="flex items-center gap-2">
                <span className="text-primary font-mono text-sm font-bold">$</span>
                <Input
                    value={input}
                    onChange={e => setInput(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder="Enter command..."
                    autoFocus
                    className="flex-1 font-mono"
                />
                <Button size="sm" onClick={handleSubmit} disabled={runCommand.isPending}>
                    Run
                </Button>
            </div>
        </div>
    )
}
