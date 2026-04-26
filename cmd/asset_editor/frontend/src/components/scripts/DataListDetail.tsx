import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Plus, Trash2, GripVertical } from "lucide-react";

interface DataListDetailProps {
    name: string;
    items: string[];
    onChange: (items: string[]) => void;
}

export function DataListDetail({ name, items, onChange }: DataListDetailProps) {
    const addItem = () => {
        onChange([...items, ""]);
    };

    const removeItem = (idx: number) => {
        onChange(items.filter((_, i) => i !== idx));
    };

    const updateItem = (idx: number, value: string) => {
        const next = [...items];
        next[idx] = value;
        onChange(next);
    };

    const moveItem = (from: number, to: number) => {
        if (to < 0 || to >= items.length) return;
        const next = [...items];
        const [item] = next.splice(from, 1);
        next.splice(to, 0, item);
        onChange(next);
    };

    return (
        <div className="flex h-full flex-col overflow-hidden">
            <div className="flex items-center gap-2 border-b border-border px-3 py-1.5">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-accent-orange px-1.5 py-0.5 rounded bg-accent-orange-tint">data list</span>
                <code className="text-sm font-mono font-semibold">{name}</code>
                <span className="text-[10px] text-muted-foreground/50">
                    {items.length} item{items.length !== 1 ? "s" : ""}
                </span>
                <div className="flex-1" />
                <Button variant="ghost" size="sm" className="h-6 text-xs text-muted-foreground" onClick={addItem}>
                    <Plus className="mr-1 h-3 w-3" />
                    Add
                </Button>
            </div>
            <ScrollArea className="flex-1 overflow-hidden">
                <div className="p-3 space-y-1">
                    <div className="text-[10px] text-muted-foreground/60 mb-2">
                        Legacy string lists for pick_dialogue, pick_self_dialogue, and pick_chatter steps. Prefer using consts for new lists - they work the same way with pick_* steps.
                    </div>
                    {items.map((item, i) => (
                        <div key={i} className="flex items-start gap-1 group/item">
                            <div className="flex flex-col shrink-0 mt-1.5">
                                <button
                                    className="text-muted-foreground/30 hover:text-muted-foreground cursor-grab"
                                    onMouseDown={(e) => {
                                        e.preventDefault();
                                        const startY = e.clientY;
                                        const itemHeight = 32;
                                        let currentFrom = i;
                                        const handleMove = (ev: MouseEvent) => {
                                            const delta = Math.round((ev.clientY - startY) / itemHeight);
                                            const target = i + delta;
                                            if (target !== currentFrom && target >= 0 && target < items.length) {
                                                moveItem(currentFrom, target);
                                                currentFrom = target;
                                            }
                                        };
                                        const handleUp = () => {
                                            window.removeEventListener("mousemove", handleMove);
                                            window.removeEventListener("mouseup", handleUp);
                                        };
                                        window.addEventListener("mousemove", handleMove);
                                        window.addEventListener("mouseup", handleUp);
                                    }}
                                >
                                    <GripVertical className="h-3 w-3" />
                                </button>
                            </div>
                            <span className="text-[10px] text-muted-foreground/40 w-5 text-right shrink-0 mt-1.5">{i}</span>
                            <Input
                                className="h-7 flex-1 text-xs"
                                value={item}
                                onChange={(e) => updateItem(i, e.target.value)}
                                placeholder="message text"
                            />
                            <button
                                className="text-muted-foreground hover:text-destructive opacity-0 group-hover/item:opacity-100 shrink-0 mt-1.5"
                                onClick={() => removeItem(i)}
                            >
                                <Trash2 className="h-3 w-3" />
                            </button>
                        </div>
                    ))}
                    {items.length === 0 && (
                        <div className="text-xs text-muted-foreground text-center py-4">
                            Empty list. Click "Add" to add entries.
                        </div>
                    )}
                </div>
            </ScrollArea>
        </div>
    );
}
