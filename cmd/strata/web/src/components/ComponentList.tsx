import { For } from "solid-js";
import { state } from "../store";

export default function ComponentList() {
  return (
    <div class="panel">
      <div class="section-header">Registered Components</div>

      <div style="padding-top: 4px;">
        <For
          each={state.components}
          fallback={<p class="empty-state">no components registered</p>}
        >
          {(comp, i) => (
            <div class={`card ${i() === 0 ? "card-first" : ""}`}>
              <div class="card-row">
                <div style="display: flex; align-items: center; gap: 10px; min-width: 0;">
                  <div class={`status-mark ${comp.suceeded ? "ok" : "err"}`} />
                  <div>
                    <div class="card-title">{comp.name}</div>
                    <div class="card-sub">v{comp.version} &nbsp;·&nbsp; {comp.path}</div>
                    {comp.error && (
                      <div class="card-error">{comp.error}</div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}
        </For>
      </div>
    </div>
  );
}
