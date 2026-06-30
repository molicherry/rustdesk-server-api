import { useEffect, useRef, useState } from "react";

export type WSStatus = "connecting" | "connected" | "disconnected";

export interface WSEvent {
  type: string;
  payload: unknown;
}

export function useWebSocket() {
  const [status, setStatus] = useState<WSStatus>("connecting");
  const [lastEvent, setLastEvent] = useState<WSEvent | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const token = localStorage.getItem("api-token");
    if (!token) {
      setStatus("disconnected");
      return;
    }

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setStatus("connected");
      ws.send(JSON.stringify({ type: "auth", token }));
    };

    ws.onmessage = (ev) => {
      try {
        const data = JSON.parse(ev.data);
        setLastEvent({ type: data.type || "message", payload: data });
      } catch {
        setLastEvent({ type: "raw", payload: ev.data });
      }
    };

    ws.onclose = () => {
      setStatus("disconnected");
    };

    ws.onerror = () => {
      setStatus("disconnected");
    };

    return () => {
      ws.close();
    };
  }, []);

  return { status, lastEvent, ws: wsRef.current };
}
