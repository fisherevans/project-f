import { Outlet, NavLink } from "react-router-dom";
import { Image, Volume2, ScrollText, Swords, Bug, HardDrive, BookOpen } from "lucide-react";

export function AppLayout() {
    return (
        <div className="flex h-screen overflow-hidden bg-background text-foreground">
            <nav className="flex w-12 flex-col items-center gap-2 border-r border-border bg-muted/40 py-3">
                <NavLink
                    to="/sprites"
                    className={({ isActive }) =>
                        `flex h-8 w-8 items-center justify-center rounded-md transition-colors ${
                            isActive
                                ? "bg-primary text-primary-foreground"
                                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        }`
                    }
                    title="Sprites"
                >
                    <Image className="h-4 w-4" />
                </NavLink>
                <NavLink
                    to="/audio"
                    className={({ isActive }) =>
                        `flex h-8 w-8 items-center justify-center rounded-md transition-colors ${
                            isActive
                                ? "bg-primary text-primary-foreground"
                                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        }`
                    }
                    title="Audio"
                >
                    <Volume2 className="h-4 w-4" />
                </NavLink>
                <NavLink
                    to="/scripts"
                    className={({ isActive }) =>
                        `flex h-8 w-8 items-center justify-center rounded-md transition-colors ${
                            isActive
                                ? "bg-primary text-primary-foreground"
                                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        }`
                    }
                    title="Scripts"
                >
                    <ScrollText className="h-4 w-4" />
                </NavLink>
                <NavLink
                    to="/rpg"
                    className={({ isActive }) =>
                        `flex h-8 w-8 items-center justify-center rounded-md transition-colors ${
                            isActive
                                ? "bg-primary text-primary-foreground"
                                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        }`
                    }
                    title="RPG"
                >
                    <Swords className="h-4 w-4" />
                </NavLink>
                <NavLink
                    to="/saves"
                    className={({ isActive }) =>
                        `flex h-8 w-8 items-center justify-center rounded-md transition-colors ${
                            isActive
                                ? "bg-primary text-primary-foreground"
                                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        }`
                    }
                    title="Saves"
                >
                    <HardDrive className="h-4 w-4" />
                </NavLink>
                <NavLink
                    to="/reference"
                    className={({ isActive }) =>
                        `flex h-8 w-8 items-center justify-center rounded-md transition-colors ${
                            isActive
                                ? "bg-primary text-primary-foreground"
                                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        }`
                    }
                    title="Reference"
                >
                    <BookOpen className="h-4 w-4" />
                </NavLink>
                <div className="mt-auto" />
                <NavLink
                    to="/debug"
                    className={({ isActive }) =>
                        `flex h-8 w-8 items-center justify-center rounded-md transition-colors ${
                            isActive
                                ? "bg-primary text-primary-foreground"
                                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        }`
                    }
                    title="Debug"
                >
                    <Bug className="h-4 w-4" />
                </NavLink>
            </nav>
            <main className="flex-1 overflow-hidden">
                <Outlet />
            </main>
        </div>
    );
}
