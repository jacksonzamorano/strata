import { Show, createMemo, createSignal, onCleanup, onMount } from "solid-js";
import { render } from "solid-js/web";
import type {
  HostMessage,
  HostMessageEventRecieved,
  HostMessagePayload,
  HostMessageType,
} from "./generated";
import type {
  LogRecord,
  RegisteredComponent,
  RegisteredTask,
  TabKey,
  TokenRecord,
  TokenState,
} from "./app-types";
import { HostHeader } from "./components/app-header";
import { AuthorizationPanel } from "./components/authorization-panel";
import { LogsPanel } from "./components/logs-panel";
import { OverviewPanel } from "./components/overview-panel";
import { Shell, TabBar } from "./components/ui";
import "./styles.css";

const MAX_LOG_LINES = 1000;
const RECONNECT_DELAY_MS = 1000;

const tabOptions = [
  { key: "overview", label: "Overview" },
  { key: "authorization", label: "Authorization" },
  { key: "logs", label: "Logs" },
] as const satisfies readonly { key: TabKey; label: string }[];

const emptyPayload = (): HostMessagePayload => ({
  hello: undefined,
  hello_ack: undefined,
  subscribe_logs: undefined,
  subscribe_logs_ack: undefined,
  authorization_create: undefined,
  authorization_created: undefined,
  event_recieved: undefined,
  error: undefined,
});

const nextMessageId = (() => {
  let seed = 0;
  return () => {
    seed += 1;
    return `${Date.now()}-${seed}`;
  };
})();

const sortByName = <T extends { name: string }>(rows: T[]): T[] =>
  rows.sort((a, b) => a.name.localeCompare(b.name));

function App() {
  const [connected, setConnected] = createSignal(false);
  const [status, setStatus] = createSignal("Connecting");
  const [logs, setLogs] = createSignal<LogRecord[]>([]);
  const [tasks, setTasks] = createSignal<RegisteredTask[]>([]);
  const [components, setComponents] = createSignal<RegisteredComponent[]>([]);
  const [nickname, setNickname] = createSignal("");
  const [latestToken, setLatestToken] = createSignal<TokenState>(null);
  const [tokens, setTokens] = createSignal<TokenRecord[]>([]);
  const [activeTab, setActiveTab] = createSignal<TabKey>("overview");
  const createRequestIds = new Set<string>();

  let socket: WebSocket | null = null;
  let reconnectTimer: number | null = null;
  let disposed = false;

  const lineCount = createMemo(() => logs().length);

  const pushLog = (entry: LogRecord) => {
    setLogs((previous) => {
      const next = [...previous, entry];
      if (next.length <= MAX_LOG_LINES) {
        return next;
      }
      return next.slice(next.length - MAX_LOG_LINES);
    });
  };

  const parsePayloadObject = (rawPayload: string | undefined): Record<string, unknown> | null => {
    if (!rawPayload) {
      return null;
    }
    try {
      const parsed = JSON.parse(rawPayload) as unknown;
      if (!parsed || typeof parsed !== "object") {
        return null;
      }
      return parsed as Record<string, unknown>;
    } catch {
      return null;
    }
  };

  const upsertTask = (name: string, url: string) => {
    setTasks((previous) => {
      const next = previous.filter((task) => task.name !== name);
      next.push({ name, url });
      return sortByName(next);
    });
  };

  const upsertComponent = (
    name: string,
    update: (existing: RegisteredComponent | undefined) => RegisteredComponent,
  ) => {
    setComponents((previous) => {
      const index = previous.findIndex((component) => component.name === name);
      const existing = index >= 0 ? previous[index] : undefined;
      const next = [...previous];
      const updated = update(existing);
      if (index >= 0) {
        next[index] = updated;
      } else {
        next.push(updated);
      }
      return sortByName(next);
    });
  };

  const processHostEvent = (incoming: HostMessageEventRecieved) => {
    if (incoming.channel !== "event") {
      return;
    }

    const payload = parsePayloadObject(incoming.payload ?? undefined);
    if (!payload) {
      return;
    }

    if (incoming.kind === "taskRegistered") {
      const name = typeof payload.name === "string" ? payload.name.trim() : "";
      const url = typeof payload.url === "string" ? payload.url.trim() : "";
      if (!name || !url) {
        return;
      }
      upsertTask(name, url);
      return;
    }

    if (incoming.kind === "componentRegistered") {
      const name = typeof payload.name === "string" ? payload.name.trim() : "";
      const version = typeof payload.version === "string" ? payload.version.trim() : "";
      const status = payload.suceeded ?? payload.succeeded;
      if (!name || typeof status !== "boolean") {
        return;
      }
      upsertComponent(name, (existing) => ({
        name,
        version: version || existing?.version || "unknown",
        isHealthy: status,
      }));
      return;
    }

    if (incoming.kind === "componentReady") {
      const name = typeof payload.name === "string" ? payload.name.trim() : "";
      const succeeded = payload.succeeded;
      if (!name || typeof succeeded !== "boolean") {
        return;
      }
      upsertComponent(name, (existing) => ({
        name,
        version: existing?.version || "unknown",
        isHealthy: succeeded,
      }));
    }
  };

  const send = (type: HostMessageType, payload: HostMessagePayload): string | null => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return null;
    }

    const id = nextMessageId();

    socket.send(
      JSON.stringify({
        id,
        type,
        payload,
      } satisfies HostMessage),
    );

    return id;
  };

  const connect = () => {
    if (disposed) {
      return;
    }

    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }

    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    socket = new WebSocket(`${protocol}://${window.location.host}/ws`);

    socket.onopen = () => {
      setConnected(true);
      setStatus("Connected");

      send("hello", {
        ...emptyPayload(),
        hello: {
          protocol_version: "1",
          client_name: "tasklib-web-host",
        },
      });

      send("subscribeLogs", {
        ...emptyPayload(),
        subscribe_logs: {
          tail: 200,
        },
      });
    };

    socket.onclose = () => {
      setConnected(false);
      setStatus("Disconnected (retrying)");
      createRequestIds.clear();

      if (!disposed) {
        reconnectTimer = window.setTimeout(connect, RECONNECT_DELAY_MS);
      }
    };

    socket.onerror = () => {
      setStatus("Socket error");
    };

    socket.onmessage = (event) => {
      let message: HostMessage;
      try {
        message = JSON.parse(event.data) as HostMessage;
      } catch {
        return;
      }

      if (message.type === "eventRecieved" && message.payload.event_recieved) {
        const incoming = message.payload.event_recieved as HostMessageEventRecieved;
        const timestamp = new Date((incoming.date as unknown as string) || Date.now());

        pushLog({
          id: message.id,
          date: timestamp.toLocaleTimeString(),
          channel: incoming.channel,
          kind: incoming.kind,
          namespace: incoming.namespace ?? undefined,
          message: incoming.message,
          payload: incoming.payload ?? undefined,
        });

        processHostEvent(incoming);
        return;
      }

      if (message.type === "authorizationCreated" && message.payload.authorization_created) {
        const token = message.payload.authorization_created;

        setTokens((previous) => {
          const index = previous.findIndex((existing) => existing.secret === token.secret);
          const next = [...previous];
          if (index >= 0) {
            next[index] = token;
          } else {
            next.push(token);
          }
          next.sort(
            (a, b) =>
              new Date((b.created_date as unknown as string) || Date.now()).getTime() -
              new Date((a.created_date as unknown as string) || Date.now()).getTime(),
          );
          return next;
        });

        if (createRequestIds.has(message.id)) {
          createRequestIds.delete(message.id);
          setLatestToken(token);
          setActiveTab("authorization");
        }
        return;
      }

      if (message.type === "error" && message.payload.error) {
        pushLog({
          id: message.id || nextMessageId(),
          date: new Date().toLocaleTimeString(),
          channel: "error",
          kind: message.payload.error.code,
          message: message.payload.error.message,
        });
      }
    };
  };

  const onCreateToken = (event: SubmitEvent) => {
    event.preventDefault();
    const nextNickname = nickname().trim();

    if (!nextNickname) {
      setStatus("Nickname is required");
      return;
    }

    const requestId = send("authorizationCreate", {
      ...emptyPayload(),
      authorization_create: {
        nickname: nextNickname,
      },
    });
    if (!requestId) {
      setStatus("Not connected");
      return;
    }
    createRequestIds.add(requestId);

    setNickname("");
  };

  onMount(connect);

  onCleanup(() => {
    disposed = true;

    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }

    if (socket) {
      socket.close();
      socket = null;
    }
  });

  return (
    <Shell>
      <HostHeader connected={connected()} status={status()} lineCount={lineCount()} />

      <TabBar
        value={activeTab()}
        onChange={setActiveTab}
        options={tabOptions.map((tab) =>
          tab.key === "logs" ? { ...tab, count: lineCount() } : tab,
        )}
      />

      <Show when={activeTab() === "overview"}>
        <OverviewPanel tasks={tasks()} components={components()} />
      </Show>

      <Show when={activeTab() === "authorization"}>
        <AuthorizationPanel
          nickname={nickname()}
          onNicknameInput={setNickname}
          onSubmit={onCreateToken}
          latestToken={latestToken()}
          tokens={tokens()}
        />
      </Show>

      <Show when={activeTab() === "logs"}>
        <LogsPanel logs={logs()} />
      </Show>
    </Shell>
  );
}

render(() => <App />, document.getElementById("root")!);
