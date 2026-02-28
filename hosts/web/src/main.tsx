import { For, Show, createMemo, createSignal, onCleanup, onMount } from "solid-js";
import { render } from "solid-js/web";
import type { HostMessage } from "./generated/HostMessage";
import type {
  HostMessageAuthorizationsList,
  HostMessageComponentsList,
  HostMessageCreateAuthorization,
  HostMessageLogEvent,
  HostMessagePendingPermissionList,
  HostMessageRequestHistory,
  HostMessageRequestPermission,
  HostMessageRespondPermission,
  HostMessageTasksList,
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

function toDate(value: unknown): Date {
  if (value instanceof Date) {
    return value;
  }
  if (typeof value === "string" || typeof value === "number") {
    const parsed = new Date(value);
    if (!Number.isNaN(parsed.getTime())) {
      return parsed;
    }
  }
  return new Date();
}

function App() {
  const [connected, setConnected] = createSignal(false);
  const [status, setStatus] = createSignal("Connecting");
  const [logs, setLogs] = createSignal<LogRecord[]>([]);
  const [tasks, setTasks] = createSignal<RegisteredTask[]>([]);
  const [components, setComponents] = createSignal<RegisteredComponent[]>([]);
  const [nickname, setNickname] = createSignal("");
  const [tokens, setTokens] = createSignal<TokenRecord[]>([]);
  const [requests, setRequests] = createSignal<RequestRecord[]>([]);
  const [activeTab, setActiveTab] = createSignal<TabKey>("overview");
  const [permissionPrompts, setPermissionPrompts] = createSignal<PermissionPrompt[]>([]);
  const createRequestIds = new Set<string>();

  let socket: WebSocket | null = null;
  let reconnectTimer: number | null = null;
  let disposed = false;

  const lineCount = createMemo(() => logs().length);
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

  const hydrateTasks = (payload: HostMessageTasksList) => {
    setTasks(sortByName(payload.tasks.map((task) => ({ name: task.name, url: task.url }))));
  };

  const hydrateComponents = (payload: HostMessageComponentsList) => {
    setComponents(
      sortByName(
        payload.components.map((component) => ({
          name: component.name,
          version: component.version || "unknown",
          isHealthy: component.is_healthy,
        })),
      ),
    );
  };

  const hydrateAuthorizations = (payload: HostMessageAuthorizationsList, requestId?: string) => {
    setTokens(
      [...payload.authorizations].sort(
        (a, b) =>
          toDate((b.created_date as unknown as string) || Date.now()).getTime() -
          toDate((a.created_date as unknown as string) || Date.now()).getTime(),
      ),
    );

    if (requestId && createRequestIds.has(requestId)) {
      createRequestIds.delete(requestId);
      setActiveTab("authorization");
    }
  };

  const hydrateRequestHistory = (payload: HostMessageRequestHistory) => {
    const nextRequests = payload.requests.map((entry) => {
      const startedAt = toDate(entry.task_start_date);
      const finishedAt = toDate(entry.task_end_date);
      const durationSeconds = Math.max(0, (finishedAt.getTime() - startedAt.getTime()) / 1000);

      const row: RequestRecord = {
        id: String(entry.id),
        taskName: `#${entry.id}`,
        startedAt,
        finishedAt,
        durationSeconds,
        state: entry.succeeded ? "succeeded" : "failed",
        sortTs: startedAt.getTime(),
      };
      return row;
    });

    nextRequests.sort((a, b) => b.sortTs - a.sortTs);
    setRequests(nextRequests);
  };

  const parsePermissionPromptPayload = (
    id: string,
    payload: HostMessageRequestPermission,
  ): PermissionPrompt | null => {
    if (!payload || !payload.permission) {
      return null;
    }

    const permission = payload.permission;
    const container = permission.container.trim().length > 0 ? permission.container.trim() : "unknown-container";
    const action = String(permission.action).trim().length > 0 ? String(permission.action).trim() : "unknown-action";
    const scope = typeof permission.scope === "string" && permission.scope.trim().length > 0 ? permission.scope.trim() : undefined;

    return {
      id,
      container,
      action,
      scope,
    };
  };

  const hydratePendingPermissionList = (payload: HostMessagePendingPermissionList) => {
    const next = payload.pending_permissions
      .map((pending) => parsePermissionPromptPayload(pending.id, { permission: pending.permission }))
      .filter((prompt): prompt is PermissionPrompt => prompt !== null)
      .slice(0, MAX_PERMISSION_POPUPS);
    setPermissionPrompts(next);
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

  const send = <TPayload,>(type: HostMessageType, payload: TPayload, forceId?: string): string | null => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return null;
    }

    const id = forceId ?? nextMessageId();

    socket.send(
      JSON.stringify({
        id,
        type,
        payload,
      } satisfies HostMessage<TPayload>),
    );

    return id;
  };

  const requestInitialSnapshots = () => {
    send("getTasksList", {});
    send("getComponentsList", {});
    send("getRequestHistory", {});
    send("getAuthorizationsList", {});
    send("getPendingPermissionList", {});
  };

  const respondToPermissionPrompt = (promptId: string, approve: boolean) => {
    const sent = send<HostMessageRespondPermission>(
      "respondPermission",
      {
        approve,
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
      requestInitialSnapshots();
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

      if (message.type === "tasksList") {
        hydrateTasks(message.payload as HostMessageTasksList);
        return;
      }

      if (message.type === "componentsList") {
        hydrateComponents(message.payload as HostMessageComponentsList);
        return;
      }

      if (message.type === "authorizationsList") {
        hydrateAuthorizations(message.payload as HostMessageAuthorizationsList, message.id);
        return;
      }

      if (message.type === "requestHistory") {
        hydrateRequestHistory(message.payload as HostMessageRequestHistory);
        return;
      }

      if (message.type === "pendingPermissionList") {
        hydratePendingPermissionList(message.payload as HostMessagePendingPermissionList);
        return;
      }

      if (message.type === "permissionRequest") {
        const prompt = parsePermissionPromptPayload(message.id, message.payload as HostMessageRequestPermission);
        if (!prompt) {
          return;
        }
        enqueuePermissionPrompt(prompt);
        return;
      }

      if (message.type === "logEvent") {
        const incoming = message.payload as HostMessageLogEvent;
        const timestamp = toDate(incoming.date);

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

        if (incoming.channel === "event" && incoming.kind === "taskFinished") {
          send("getRequestHistory", {});
        }
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

    const requestId = send<HostMessageCreateAuthorization>("createAuthorization", {
      nickname: nextNickname,
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
