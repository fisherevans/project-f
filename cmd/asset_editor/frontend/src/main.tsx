import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createBrowserRouter, Navigate, RouterProvider } from "react-router-dom";
import { WebSocketProvider } from "@/api/websocket";
import { AppLayout } from "@/layouts/AppLayout";
import { SpriteBrowser } from "@/pages/sprites/SpriteBrowser";
import { SpriteEditor } from "@/pages/sprites/SpriteEditor";
import { AudioBrowser } from "@/pages/audio/AudioBrowser";
import { AudioEditor } from "@/pages/audio/AudioEditor";
import { ScriptBrowser } from "@/pages/scripts/ScriptBrowser";
import { ScriptEditor } from "@/pages/scripts/ScriptEditor";
import { SkillBrowser } from "@/pages/rpg/SkillBrowser";
import { PrimortalBrowser } from "@/pages/rpg/PrimortalBrowser";
import { CombatBrowser } from "@/pages/rpg/CombatBrowser";
import { SaveBrowser } from "@/pages/saves/SaveBrowser";
import { SaveEditor } from "@/pages/saves/SaveEditor";
import { ScriptReference } from "@/pages/reference/ScriptReference";
import { DebugLayout } from "@/pages/debug/DebugLayout";
import { DebugOverview } from "@/pages/debug/DebugOverview";
import { DebugGlobals } from "@/pages/debug/DebugGlobals";
import { DebugCommands } from "@/pages/debug/DebugCommands";
import { DebugEntities } from "@/pages/debug/DebugEntities";
import "./index.css";

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: 30_000,
            retry: 1,
        },
    },
});

const router = createBrowserRouter([
    {
        element: <AppLayout />,
        children: [
            { path: "/sprites", element: <SpriteBrowser /> },
            { path: "/sprites/*", element: <SpriteEditor /> },
            { path: "/audio", element: <AudioBrowser /> },
            { path: "/audio/*", element: <AudioEditor /> },
            { path: "/scripts", element: <ScriptBrowser /> },
            { path: "/scripts/*", element: <ScriptEditor /> },
            { path: "/skills", element: <SkillBrowser /> },
            { path: "/primortals", element: <PrimortalBrowser /> },
            { path: "/combat", element: <CombatBrowser /> },
            { path: "/rpg", element: <Navigate to="/skills" replace /> },
            { path: "/rpg/primortals", element: <Navigate to="/primortals" replace /> },
            { path: "/rpg/combat", element: <Navigate to="/combat" replace /> },
            { path: "/saves", element: <SaveBrowser /> },
            { path: "/saves/:id", element: <SaveEditor /> },
            { path: "/reference", element: <ScriptReference /> },
            {
                path: "/debug",
                element: <DebugLayout />,
                children: [
                    { index: true, element: <DebugOverview /> },
                    { path: "globals", element: <DebugGlobals /> },
                    { path: "commands", element: <DebugCommands /> },
                    { path: "entities", element: <DebugEntities /> },
                ],
            },
            { path: "/", element: <Navigate to="/sprites" replace /> },
        ],
    },
]);

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <QueryClientProvider client={queryClient}>
            <WebSocketProvider>
                <RouterProvider router={router} />
            </WebSocketProvider>
        </QueryClientProvider>
    </StrictMode>,
);
