import { For } from "solid-js";
import { state } from "../store";

export default function TaskList() {
  return (
    <div class="panel">
      <div class="section-header">Registered Tasks</div>

      <div style="padding-top: 4px;">
        <For
          each={state.tasks}
          fallback={<p class="empty-state">no tasks registered</p>}
        >
          {(task, i) => (
            <div class={`card ${i() === 0 ? "card-first" : ""}`}>
              <div class="card-title">{task.name}</div>
              <div class="card-sub">{task.url}</div>
            </div>
          )}
        </For>
      </div>
    </div>
  );
}
