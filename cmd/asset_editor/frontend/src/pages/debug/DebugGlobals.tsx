import { useState } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { useGlobals, useSetGlobal, useDeleteGlobal } from "@/api/debug"
import { Trash2 } from "lucide-react"
import { usePageTitle } from "@/hooks/usePageTitle"

export function DebugGlobals() {
    usePageTitle("Globals - Debug")

    const [prefix, setPrefix] = useState("")
    const [editKey, setEditKey] = useState("")
    const [editType, setEditType] = useState("string")
    const [editValue, setEditValue] = useState("")

    const { data: globals, isLoading, error } = useGlobals(prefix || undefined)
    const setGlobal = useSetGlobal()
    const deleteGlobal = useDeleteGlobal()

    function handleSet() {
        if (!editKey) return
        let val: unknown = editValue
        if (editType === "int") val = parseInt(editValue)
        else if (editType === "float") val = parseFloat(editValue)
        else if (editType === "bool") val = editValue === "true"
        setGlobal.mutate({ key: editKey, type: editType, value: val }, {
            onSuccess: () => { setEditKey(""); setEditValue("") },
        })
    }

    return (
        <div className="flex h-full flex-col gap-4 overflow-hidden p-6">
            <div className="flex items-center gap-3">
                <Input
                    placeholder="Filter by prefix..."
                    value={prefix}
                    onChange={e => setPrefix(e.target.value)}
                    className="w-60"
                />
                <span className="text-xs text-muted-foreground">
                    {globals?.length ?? 0} entries
                </span>
            </div>

            <div className="flex items-center gap-2 rounded-md border bg-card p-3">
                <Input
                    placeholder="key"
                    value={editKey}
                    onChange={e => setEditKey(e.target.value)}
                    className="w-48"
                />
                <select
                    value={editType}
                    onChange={e => setEditType(e.target.value)}
                    className="h-9 rounded-md border bg-background px-3 text-sm"
                >
                    <option value="string">string</option>
                    <option value="int">int</option>
                    <option value="float">float</option>
                    <option value="bool">bool</option>
                </select>
                <Input
                    placeholder="value"
                    value={editValue}
                    onChange={e => setEditValue(e.target.value)}
                    className="w-48"
                    onKeyDown={e => { if (e.key === "Enter") handleSet() }}
                />
                <Button size="sm" onClick={handleSet} disabled={setGlobal.isPending || !editKey}>
                    Set
                </Button>
            </div>

            {isLoading && <div className="text-sm text-muted-foreground">Loading...</div>}
            {error && (
                <div className="text-sm text-destructive">
                    Cannot reach debug API. Is the game running?
                </div>
            )}

            {globals && (
                <div className="flex-1 overflow-auto rounded-md border">
                    <table className="w-full text-sm">
                        <thead className="sticky top-0 bg-muted">
                            <tr>
                                <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Key</th>
                                <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Value</th>
                                <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase w-20">Type</th>
                                <th className="px-3 py-2 w-12"></th>
                            </tr>
                        </thead>
                        <tbody>
                            {globals.map(g => (
                                <tr key={g.key} className="border-t hover:bg-muted/50">
                                    <td className="px-3 py-1.5 font-mono text-xs">{g.key}</td>
                                    <td className="px-3 py-1.5 font-mono text-xs">{JSON.stringify(g.value)}</td>
                                    <td className="px-3 py-1.5">
                                        <Badge variant="outline" className="text-[10px]">
                                            {typeof g.value}
                                        </Badge>
                                    </td>
                                    <td className="px-3 py-1.5">
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            className="h-6 w-6 p-0 text-muted-foreground hover:text-destructive"
                                            onClick={() => deleteGlobal.mutate(g.key)}
                                        >
                                            <Trash2 className="h-3 w-3" />
                                        </Button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    )
}
