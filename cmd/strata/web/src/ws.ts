import { setState, type LogEntry } from "./store";

let socket: WebSocket | null = null;
let token: string = "";
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let wasKicked = false;
let didOpen = false;
let failedBeforeOpenCount = 0;
const MAX_FAILED_BEFORE_OPEN = 5;

export function getToken(): string {
  return token;
}

export function setToken(t: string) {
  token = t;
  localStorage.setItem("strata_token", t);
}

export function connect(explicit = false) {
  if (!token) return;
  if (socket && socket.readyState <= WebSocket.OPEN) return;

  if (explicit) failedBeforeOpenCount = 0;
  wasKicked = false;
  didOpen = false;

  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  const url = `${proto}//${location.host}/ws?token=${encodeURIComponent(token)}`;
  socket = new WebSocket(url);

  socket.onopen = () => {
    didOpen = true;
    failedBeforeOpenCount = 0;
    setState("connected", true);
    setState("kicked", false);
    setState("authFailed", false);
  };

  socket.onclose = () => {
    setState("connected", false);
    socket = null;
    if (!didOpen) {
      // Connection closed before opening — could be startup race or auth rejection
      failedBeforeOpenCount++;
      if (failedBeforeOpenCount < MAX_FAILED_BEFORE_OPEN) {
        scheduleReconnect();
        return;
      }
      // Auth genuinely rejected after multiple attempts
      failedBeforeOpenCount = 0;
      setState("authFailed", true);
      return;
    }
    // Auto-reconnect unless we were kicked
    if (!wasKicked) {
      scheduleReconnect();
    }
  };

  socket.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data);
      handleMessage(msg);
    } catch {}
  };
}

function scheduleReconnect() {
  if (reconnectTimer) return;
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null;
    connect();
  }, 2000);
}

function handleMessage(msg: { type: string; id?: string; payload: any }) {
  switch (msg.type) {
    case "sync":
      setState("logs", msg.payload.logs ?? []);
      setState("tasks", msg.payload.tasks ?? []);
      setState("components", msg.payload.components ?? []);
      setState("triggers", msg.payload.triggers ?? []);
      setState("authorizations", msg.payload.authorizations ?? []);
      break;
    case "log":
      setState("logs", (prev) => [...prev, msg.payload as LogEntry]);
      break;
    case "taskRegistered":
      setState("tasks", (prev) => [...prev, msg.payload]);
      break;
    case "componentRegistered":
      setState("components", (prev) => [...prev, msg.payload]);
      break;
    case "taskTriggered":
      setState("triggers", (prev) => [...prev, msg.payload]);
      break;
    case "authorizationsUpdated":
      setState("authorizations", msg.payload.authorizations ?? []);
      break;
    case "permissionRequest":
      setState("permissionRequest", {
        id: msg.id!,
        permission: msg.payload,
      });
      break;
    case "secretRequest":
      setState("secretRequest", {
        id: msg.id!,
        ...msg.payload,
      });
      break;
    case "oauthRequest":
      setState("oauthRequest", {
        id: msg.id!,
        ...msg.payload,
      });
      break;
    case "kicked":
      wasKicked = true;
      setState("kicked", true);
      setState("connected", false);
      socket?.close();
      socket = null;
      break;
  }
}

export function send(type: string, id: string, payload: any) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type, id, payload }));
  }
}
