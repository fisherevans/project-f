import { NavLink } from "react-router-dom"
import { cn } from "@/lib/utils"

const links = [
    { to: "/debug", label: "Overview", end: true },
    { to: "/debug/globals", label: "Globals" },
    { to: "/debug/commands", label: "Commands" },
    { to: "/debug/entities", label: "Entities" },
]

export function DebugNav() {
    return (
        <div className="flex items-center gap-1 border-b px-6 py-1">
            {links.map(link => (
                <NavLink
                    key={link.to}
                    to={link.to}
                    end={link.end}
                    className={({ isActive }) =>
                        cn(
                            "px-3 py-1.5 text-xs font-medium rounded-md transition-colors",
                            isActive
                                ? "bg-accent text-accent-foreground"
                                : "text-muted-foreground hover:text-foreground hover:bg-muted"
                        )
                    }
                >
                    {link.label}
                </NavLink>
            ))}
        </div>
    )
}
