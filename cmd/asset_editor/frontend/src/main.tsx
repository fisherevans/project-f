import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { WebSocketProvider } from "@/api/websocket";
import { AppLayout } from "@/layouts/AppLayout";
import { SpriteBrowser } from "@/pages/sprites/SpriteBrowser";
import { SpriteEditor } from "@/pages/sprites/SpriteEditor";
import { AudioBrowser } from "@/pages/audio/AudioBrowser";
import { AudioEditor } from "@/pages/audio/AudioEditor";
import { ScriptBrowser } from "@/pages/scripts/ScriptBrowser";
import { ScriptEditor } from "@/pages/scripts/ScriptEditor";
import "./index.css";

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: 30_000,
            retry: 1,
        },
    },
});

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <QueryClientProvider client={queryClient}>
            <WebSocketProvider>
                <BrowserRouter>
                    <Routes>
                        <Route element={<AppLayout />}>
                            <Route path="/sprites" element={<SpriteBrowser />} />
                            <Route path="/sprites/*" element={<SpriteEditor />} />
                            <Route path="/audio" element={<AudioBrowser />} />
                            <Route path="/audio/*" element={<AudioEditor />} />
                            <Route path="/scripts" element={<ScriptBrowser />} />
                            <Route path="/scripts/*" element={<ScriptEditor />} />
                            <Route path="/" element={<Navigate to="/sprites" replace />} />
                        </Route>
                    </Routes>
                </BrowserRouter>
            </WebSocketProvider>
        </QueryClientProvider>
    </StrictMode>,
);
