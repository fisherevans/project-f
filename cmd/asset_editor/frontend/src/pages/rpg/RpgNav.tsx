import { NavLink } from "react-router-dom"
import { cn } from "@/lib/utils"

const links = [
    { to: "/rpg", label: "Skills", end: true },
    { to: "/rpg/primortals", label: "Primortals" },
    { to: "/rpg/combat", label: "Combat" },
]

export function RpgNav() {
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
