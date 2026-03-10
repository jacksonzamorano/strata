import { createSignal } from "solid-js";
import { state, setState } from "../store";
import { send } from "../ws";

export default function OauthDialog() {
  const [callbackUrl, setCallbackUrl] = createSignal("");
  const req = () => state.oauthRequest!;

  function submit() {
    send("oauthResponse", req().id, { url: callbackUrl() });
    setState("oauthRequest", null);
    setCallbackUrl("");
  }

  return (
    <div class="modal-backdrop">
      <div class="modal animate-in" style="width: 480px;">
        <div class="modal-header">
          <span class="modal-icon">⇌</span>
          <span class="modal-title">OAuth Authentication</span>
        </div>
        <div class="modal-body">
          <div class="modal-ns">{req().namespace}</div>
          <p class="modal-text" style="font-size: 10.5px; color: var(--text-3);">
            Open the link below and complete authentication, then paste the callback URL.
          </p>
          <a
            href={req().url}
            target="_blank"
            rel="noopener noreferrer"
            class="oauth-link"
          >
            {req().url}
          </a>
          <input
            type="text"
            value={callbackUrl()}
            onInput={(e) => setCallbackUrl(e.currentTarget.value)}
            onKeyDown={(e) => e.key === "Enter" && submit()}
            class="input"
            style="width: 100%; flex: unset;"
            placeholder="paste callback url..."
          />
        </div>
        <div class="modal-footer">
          <button onClick={submit} class="btn btn-primary" style="flex: 1;">
            Submit
          </button>
        </div>
      </div>
    </div>
  );
}
