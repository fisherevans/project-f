import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { spriteKeys } from "@/api/sprites";

interface WebSocketState {
    connected: boolean;
}

const WebSocketContext = createContext<WebSocketState>({ connected: false });

export function useWebSocket() {
    return useContext(WebSocketContext);
}

export function WebSocketProvider({ children }: { children: ReactNode }) {
    const [connected, setConnected] = useState(false);
    const queryClient = useQueryClient();
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

    useEffect(() => {
        function connect() {
            const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
            const ws = new WebSocket(`${protocol}//${window.location.host}/api/v1/ws`);
            wsRef.current = ws;

            ws.onopen = () => {
                setConnected(true);
            };

            ws.onmessage = (event) => {
                try {
                    const msg = JSON.parse(event.data);
                    if (msg.type === "change" || msg.type === "create" || msg.type === "remove") {
                        queryClient.invalidateQueries({ queryKey: spriteKeys.all });
                    }
                } catch {
                    // ignore malformed messages
                }
            };

            ws.onclose = () => {
                setConnected(false);
                wsRef.current = null;
                reconnectTimer.current = setTimeout(connect, 3000);
            };

            ws.onerror = () => {
                ws.close();
            };
        }

        connect();

        return () => {
            clearTimeout(reconnectTimer.current);
            wsRef.current?.close();
        };
    }, [queryClient]);

    return (
        <WebSocketContext.Provider value={{ connected }}>
            {children}
        </WebSocketContext.Provider>
    );
}
