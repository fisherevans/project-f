import { Outlet, NavLink } from "react-router-dom";
import { Image, Volume2, ScrollText, Swords, Shield, Crosshair, HardDrive, BookOpen, Bug, Map } from "lucide-react";

const NAV_ITEMS = [
    { to: "/sprites", label: "Sprites", icon: Image },
    { to: "/audio", label: "Audio", icon: Volume2 },
    { to: "/scripts", label: "Scripts", icon: ScrollText },
    { to: "/skills", label: "Skills", icon: Swords },
    { to: "/primortals", label: "Primortals", icon: Shield },
    { to: "/combat", label: "Combat", icon: Crosshair },
    { to: "/saves", label: "Saves", icon: HardDrive },
    { to: "/tiled", label: "Tiled", icon: Map },
    { to: "/reference", label: "Reference", icon: BookOpen },
    { to: "/debug", label: "Debug", icon: Bug },
];

export function AppLayout() {
    return (
        <div className="flex h-screen overflow-hidden bg-background text-foreground">
            <nav className="flex w-36 flex-col gap-0.5 border-r border-border bg-muted/40 px-2 py-3">
                {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
                    <NavLink
                        key={to}
                        to={to}
                        title={label}
                        className={({ isActive }) =>
                            `flex items-center gap-2 rounded-md px-2 py-1.5 text-xs font-medium transition-colors ${
                                isActive
                                    ? "bg-primary text-primary-foreground"
                                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                            }`
                        }
                    >
                        <Icon className="h-3.5 w-3.5 shrink-0" />
                        {label}
                    </NavLink>
                ))}
            </nav>
            <main className="flex-1 overflow-hidden">
                <Outlet />
            </main>
        </div>
    );
}
