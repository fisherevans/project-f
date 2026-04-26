import { Outlet } from "react-router-dom"
import { RpgNav } from "./RpgNav"

export function RpgLayout() {
    return (
        <div className="flex h-full flex-col overflow-hidden">
            <RpgNav />
            <div className="flex-1 overflow-hidden">
                <Outlet />
            </div>
        </div>
    )
}
