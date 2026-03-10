import { createSignal, Show, onMount, createEffect } from "solid-js";
import { state } from "./store";
import { connect, setToken, getToken } from "./ws";
import LogPanel from "./components/LogPanel";
import TaskList from "./components/TaskList";
import ComponentList from "./components/ComponentList";
import AuthPanel from "./components/AuthPanel";
import PermissionDialog from "./components/PermissionDialog";
import SecretDialog from "./components/SecretDialog";
import OauthDialog from "./components/OauthDialog";

type Tab = "logs" | "tasks" | "components" | "auth";

const tabs: { id: Tab; label: string }[] = [
  { id: "logs",       label: "Logs" },
  { id: "tasks",      label: "Tasks" },
  { id: "components", label: "Components" },
  { id: "auth",       label: "Auth" },
];

export default function App() {
  const [tab, setTab] = createSignal<Tab>("logs");
  const [tokenInput, setTokenInput] = createSignal("");
  const [authed, setAuthed] = createSignal(false);

  onMount(() => {
    const params = new URLSearchParams(location.search);
    const t = params.get("token") ?? localStorage.getItem("strata_token");
    if (t) {
      setToken(t);
      setAuthed(true);
      connect();
    }
  });

  createEffect(() => {
    if (state.authFailed) {
      localStorage.removeItem("strata_token");
      setAuthed(false);
    }
  });

  function login() {
    const t = tokenInput().trim();
    if (!t) return;
    setToken(t);
    setAuthed(true);
    connect(true);
  }

  function reconnect() {
    connect(true);
  }

  return (
    <Show
      when={authed()}
      fallback={
        <div class="login-screen">
          <div class="login-box animate-in">
            <div class="login-header">
              <div class="login-title">Strata</div>
              <div class="login-subtitle">// authorization required</div>
            </div>
            <div class="login-body">
              <input
                type="password"
                placeholder="enter token..."
                value={tokenInput()}
                onInput={(e) => setTokenInput(e.currentTarget.value)}
                onKeyDown={(e) => e.key === "Enter" && login()}
                class="input"
                style="flex: unset; width: 100%;"
              />
              <button onClick={login} class="btn btn-primary" style="width: 100%;">
                Connect
              </button>
            </div>
          </div>
        </div>
      }
    >
      <div class="layout-root">
        {/* Sidebar */}
        <aside class="sidebar">
          <div class="sidebar-brand">
            <div class="brand-name">Strata</div>
            <div class="brand-status">
              <div class={`status-dot ${state.connected ? "live" : ""}`} />
              <span>{state.connected ? "live" : "offline"}</span>
            </div>
          </div>
          <nav class="sidebar-nav">
            {tabs.map((t) => (
              <button
                onClick={() => setTab(t.id)}
                class={`nav-item ${tab() === t.id ? "active" : ""}`}
              >
                <span class="nav-marker">{tab() === t.id ? "›" : " "}</span>
                {t.label}
              </button>
            ))}
          </nav>
        </aside>

        {/* Main content */}
        <main class="main">
          <Show when={state.kicked}>
            <div class="kicked-banner">
              <span>⚠ session displaced — another client connected</span>
              <button onClick={reconnect} class="btn btn-ghost btn-sm">
                Reconnect
              </button>
            </div>
          </Show>

          <div style="flex: 1; overflow: hidden;">
            <Show when={tab() === "logs"}>
              <LogPanel />
            </Show>
            <Show when={tab() === "tasks"}>
              <TaskList />
            </Show>
            <Show when={tab() === "components"}>
              <ComponentList />
            </Show>
            <Show when={tab() === "auth"}>
              <AuthPanel />
            </Show>
          </div>
        </main>

        {/* Modals */}
        <Show when={state.permissionRequest}>
          <PermissionDialog />
        </Show>
        <Show when={state.secretRequest}>
          <SecretDialog />
        </Show>
        <Show when={state.oauthRequest}>
          <OauthDialog />
        </Show>
      </div>
    </Show>
  );
}
