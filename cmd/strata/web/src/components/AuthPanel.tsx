import { createSignal, For } from "solid-js";
import { state } from "../store";
import { send } from "../ws";

export default function AuthPanel() {
  const [nickname, setNickname] = createSignal("");

  function createAuth() {
    const n = nickname().trim();
    if (!n) return;
    send("createAuthorization", "", { nickname: n });
    setNickname("");
  }

  function deleteAuth(secret: string) {
    send("deleteAuthorization", "", { secret });
  }

  return (
    <div class="panel">
      <div class="section-header">Authorizations</div>

      <div class="auth-form">
        <input
          type="text"
          placeholder="nickname..."
          value={nickname()}
          onInput={(e) => setNickname(e.currentTarget.value)}
          onKeyDown={(e) => e.key === "Enter" && createAuth()}
          class="input"
        />
        <button onClick={createAuth} class="btn btn-primary">
          Create
        </button>
      </div>

      <For
        each={state.authorizations}
        fallback={<p class="empty-state">no authorizations</p>}
      >
        {(auth) => (
          <div class="auth-card">
            <div>
              <div class="auth-card-name">{auth.nickname || "<unnamed>"}</div>
              <div class="auth-card-secret">
                {auth.secret.slice(0, 8)}···{auth.secret.slice(-4)}
              </div>
              <div class="auth-card-meta">
                {auth.source} &nbsp;·&nbsp;{" "}
                {new Date(auth.created_date).toLocaleDateString(undefined, {
                  month: "short",
                  day: "numeric",
                  year: "numeric",
                })}
              </div>
            </div>
            <button
              onClick={() => deleteAuth(auth.secret)}
              class="btn btn-danger btn-sm"
            >
              Revoke
            </button>
          </div>
        )}
      </For>
    </div>
  );
}
