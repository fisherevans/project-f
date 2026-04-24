import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { WebSocketProvider } from "@/api/websocket";
import { AppLayout } from "@/layouts/AppLayout";
import { SpriteBrowser } from "@/pages/sprites/SpriteBrowser";
import { SpriteEditor } from "@/pages/sprites/SpriteEditor";
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
                            <Route path="/" element={<Navigate to="/sprites" replace />} />
                        </Route>
                    </Routes>
                </BrowserRouter>
            </WebSocketProvider>
        </QueryClientProvider>
    </StrictMode>,
);
