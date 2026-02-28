import { For, Show, createMemo, createSignal, onCleanup, onMount } from "solid-js";
import { render } from "solid-js/web";
import type {
  HostMessage,
  HostMessageEventReceived,
  HostMessagePayload,
  HostMessageType,
} from "./generated";
import type {
  LogRecord,
  RequestRecord,
  RegisteredComponent,
  RegisteredTask,
  TabKey,
  TokenRecord,
} from "./app-types";
import { HostHeader } from "./components/app-header";
import { AuthorizationPanel } from "./components/authorization-panel";
import { LogsPanel } from "./components/logs-panel";
import { OverviewPanel } from "./components/overview-panel";
import { RequestsPanel } from "./components/requests-panel";
import { Shell, TabBar } from "./components/ui";
import "./styles.css";

const MAX_LOG_LINES = 1000;
const MAX_PERMISSION_POPUPS = 4;
const RECONNECT_DELAY_MS = 1000;

const tabOptions = [
  { key: "overview", label: "Overview" },
  { key: "authorization", label: "Authorization" },
  { key: "logs", label: "Logs" },
  { key: "requests", label: "Requests" },
] as const satisfies readonly { key: TabKey; label: string }[];

const emptyPayload = (): HostMessagePayload => ({
  hello: undefined,
  hello_ack: undefined,
  subscribe_logs: undefined,
  subscribe_logs_ack: undefined,
  authorization_create: undefined,
  authorization_created: undefined,
  request_permission: undefined,
  permission_response: undefined,
  event_received: undefined,
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

type PermissionPrompt = {
  id: string;
  container: string;
  action: string;
  scope?: string;
};

function App() {
  const [connected, setConnected] = createSignal(false);
  const [status, setStatus] = createSignal("Connecting");
  const [logs, setLogs] = createSignal<LogRecord[]>([]);
  const [tasks, setTasks] = createSignal<RegisteredTask[]>([]);
  const [components, setComponents] = createSignal<RegisteredComponent[]>([]);
  const [nickname, setNickname] = createSignal("");
  const [tokens, setTokens] = createSignal<TokenRecord[]>([]);
  const [activeTab, setActiveTab] = createSignal<TabKey>("overview");
  const [permissionPrompts, setPermissionPrompts] = createSignal<PermissionPrompt[]>([]);
  const createRequestIds = new Set<string>();

  let socket: WebSocket | null = null;
  let reconnectTimer: number | null = null;
  let disposed = false;

  function parsePayloadObject(rawPayload: string | undefined): Record<string, unknown> | null {
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
  }

  const lineCount = createMemo(() => logs().length);
  const requests = createMemo<RequestRecord[]>(() => {
    const rows = logs();
    const map = new Map<string, RequestRecord>();

    const parseDate = (value: unknown): Date | null => {
      if (value instanceof Date && !Number.isNaN(value.getTime())) {
        return value;
      }
      if (typeof value === "string" || typeof value === "number") {
        const parsed = new Date(value);
        if (!Number.isNaN(parsed.getTime())) {
          return parsed;
        }
      }
      return null;
    };

    const parseString = (value: unknown): string | null => {
      if (typeof value !== "string") {
        return null;
      }
      const trimmed = value.trim();
      return trimmed.length > 0 ? trimmed : null;
    };

    const parseNumber = (value: unknown): number | undefined => {
      if (typeof value !== "number" || Number.isNaN(value)) {
        return undefined;
      }
      return value;
    };

    for (const row of rows) {
      if (row.channel !== "event") {
        continue;
      }
      if (row.kind !== "taskStarted" && row.kind !== "taskFinished") {
        continue;
      }

      const payload = parsePayloadObject(row.payload);
      if (!payload) {
        continue;
      }

      const id = parseString(payload.id);
      if (!id) {
        continue;
      }

      const fallbackDate = new Date(row.timestamp);
      const eventDate = parseDate(payload.date) ?? fallbackDate;
      const current = map.get(id) ?? {
        id,
        taskName: "unknown",
        startedAt: null,
        finishedAt: null,
        state: "unknown",
        sortTs: eventDate.getTime(),
      };

      const payloadName = parseString(payload.name);

      if (payloadName) {
        current.taskName = payloadName;
      }
      current.sortTs = Math.max(current.sortTs, eventDate.getTime());

      if (row.kind === "taskStarted") {
        if (!current.startedAt || eventDate.getTime() < current.startedAt.getTime()) {
          current.startedAt = eventDate;
        }
        if (current.state === "unknown") {
          current.state = "in_progress";
        }
      } else {
        if (!current.finishedAt || eventDate.getTime() > current.finishedAt.getTime()) {
          current.finishedAt = eventDate;
        }
        if (!current.startedAt) {
          current.startedAt = eventDate;
        }

        const duration = parseNumber(payload.duration);
        if (typeof duration === "number") {
          current.durationSeconds = duration;
        }

        if (typeof payload.succeeded === "boolean") {
          current.state = payload.succeeded ? "succeeded" : "failed";
        } else {
          current.state = "unknown";
        }
      }

      map.set(id, current);
    }

    return Array.from(map.values()).sort((a, b) => b.sortTs - a.sortTs);
  });
  const requestCount = createMemo(() => requests().length);

  const pushLog = (entry: LogRecord) => {
    setLogs((previous) => {
      const next = [...previous, entry];
      if (next.length <= MAX_LOG_LINES) {
        return next;
      }
      return next.slice(next.length - MAX_LOG_LINES);
    });
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

  const processHostEvent = (incoming: HostMessageEventReceived) => {
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

  const parsePermissionPrompt = (message: HostMessage): PermissionPrompt | null => {
    const raw = message.payload.request_permission as unknown;
    if (!raw || typeof raw !== "object") {
      return null;
    }

    const requestRecord = raw as Record<string, unknown>;
    const maybePermission = requestRecord.permission;
    const permissionRecord =
      maybePermission && typeof maybePermission === "object"
        ? (maybePermission as Record<string, unknown>)
        : requestRecord;

    const rawContainer = permissionRecord.container;
    const rawAction = permissionRecord.action;
    const rawScope = permissionRecord.scope;

    const container =
      typeof rawContainer === "string" && rawContainer.trim().length > 0
        ? rawContainer.trim()
        : "unknown-container";
    const action =
      typeof rawAction === "string" && rawAction.trim().length > 0
        ? rawAction.trim()
        : "unknown-action";
    const scope = typeof rawScope === "string" && rawScope.trim().length > 0 ? rawScope.trim() : undefined;

    return {
      id: message.id,
      container,
      action,
      scope,
    };
  };

  const enqueuePermissionPrompt = (prompt: PermissionPrompt) => {
    setPermissionPrompts((previous) => {
      if (previous.some((entry) => entry.id === prompt.id)) {
        return previous;
      }
      const next = [prompt, ...previous];
      if (next.length <= MAX_PERMISSION_POPUPS) {
        return next;
      }
      return next.slice(0, MAX_PERMISSION_POPUPS);
    });
  };

  const send = (type: HostMessageType, payload: HostMessagePayload, forceId?: string): string | null => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return null;
    }

    const id = forceId ?? nextMessageId();

    socket.send(
      JSON.stringify({
        id,
        type,
        payload,
      } satisfies HostMessage),
    );

    return id;
  };

  const respondToPermissionPrompt = (promptId: string, approve: boolean) => {
    const sent = send(
      "permissionReplied",
      {
        ...emptyPayload(),
        permission_response: {
          approve,
        },
      },
      promptId,
    );
    if (!sent) {
      setStatus("Not connected");
      return;
    }

    setPermissionPrompts((previous) => previous.filter((prompt) => prompt.id !== promptId));
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
          client_name: "strata-web-host",
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
      setPermissionPrompts([]);

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

      if (message.type === "eventReceived" && message.payload.event_received) {
        const incoming = message.payload.event_received as HostMessageEventReceived;
        const timestamp = new Date((incoming.date as unknown as string) || Date.now());

        pushLog({
          id: message.id,
          timestamp: timestamp.getTime(),
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

      if (message.type === "permissionRequest") {
        const prompt = parsePermissionPrompt(message);
        if (!prompt) {
          return;
        }
        enqueuePermissionPrompt(prompt);
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
          setActiveTab("authorization");
        }
        return;
      }

      if (message.type === "error" && message.payload.error) {
        const now = new Date();
        pushLog({
          id: message.id || nextMessageId(),
          timestamp: now.getTime(),
          date: now.toLocaleTimeString(),
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
    <>
      <Shell>
        <HostHeader connected={connected()} status={status()} lineCount={lineCount()} />

        <TabBar
          value={activeTab()}
          onChange={setActiveTab}
          options={tabOptions.map((tab) => {
            if (tab.key === "logs") {
              return { ...tab, count: lineCount() };
            }
            if (tab.key === "requests") {
              return { ...tab, count: requestCount() };
            }
            return tab;
          })}
        />

        <Show when={activeTab() === "overview"}>
          <OverviewPanel tasks={tasks()} components={components()} />
        </Show>

        <Show when={activeTab() === "authorization"}>
          <AuthorizationPanel
            nickname={nickname()}
            onNicknameInput={setNickname}
            onSubmit={onCreateToken}
            tokens={tokens()}
          />
        </Show>

        <Show when={activeTab() === "logs"}>
          <LogsPanel logs={logs()} />
        </Show>

        <Show when={activeTab() === "requests"}>
          <RequestsPanel requests={requests()} />
        </Show>
      </Shell>

      <div class="permission-popups" aria-live="assertive">
        <For each={permissionPrompts()}>
          {(prompt) => (
            <article class="permission-popup">
              <div class="permission-popup__badge">Permission Request</div>
              <p class="permission-popup__summary">
                <span class="permission-popup__container">{prompt.container}</span>
                <span> requests </span>
                <code class="permission-popup__action">{prompt.action}</code>
              </p>
              <Show when={prompt.scope}>
                <p class="permission-popup__scope">
                  Scope: <code>{prompt.scope}</code>
                </p>
              </Show>
              <div class="permission-popup__actions">
                <button type="button" class="permission-popup__button permission-popup__button--deny" onClick={() => respondToPermissionPrompt(prompt.id, false)}>
                  Deny
                </button>
                <button type="button" class="permission-popup__button permission-popup__button--approve" onClick={() => respondToPermissionPrompt(prompt.id, true)}>
                  Approve
                </button>
              </div>
            </article>
          )}
        </For>
      </div>
    </>
  );
}

render(() => <App />, document.getElementById("root")!);
