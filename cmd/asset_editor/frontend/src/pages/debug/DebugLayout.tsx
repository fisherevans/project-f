import { Outlet } from "react-router-dom"
import { DebugNav } from "./DebugNav"

export function DebugLayout() {
    return (
        <div className="flex h-full flex-col overflow-hidden">
            <DebugNav />
            <div className="flex-1 overflow-hidden">
                <Outlet />
            </div>
        </div>
    )
}
