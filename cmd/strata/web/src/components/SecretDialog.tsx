import { createSignal } from "solid-js";
import { state, setState } from "../store";
import { send } from "../ws";

export default function SecretDialog() {
  const [value, setValue] = createSignal("");
  const req = () => state.secretRequest!;

  function submit() {
    send("secretResponse", req().id, { secret: value() });
    setState("secretRequest", null);
    setValue("");
  }

  return (
    <div class="modal-backdrop">
      <div class="modal animate-in">
        <div class="modal-header">
          <span class="modal-icon">⚿</span>
          <span class="modal-title">Secret Requested</span>
        </div>
        <div class="modal-body">
          <div class="modal-ns">{req().namespace}</div>
          <p class="modal-text">{req().prompt}</p>
          <input
            type="password"
            value={value()}
            onInput={(e) => setValue(e.currentTarget.value)}
            onKeyDown={(e) => e.key === "Enter" && submit()}
            class="input"
            style="width: 100%; flex: unset;"
            placeholder="enter secret..."
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
