import { For } from "solid-js";
import { state } from "../store";

const nsColorMap: Record<string, string> = {};
const nsClasses = ["ns-0", "ns-1", "ns-2", "ns-3", "ns-4", "ns-5", "ns-6", "ns-7"];

function getNsClass(ns: string): string {
  if (!nsColorMap[ns]) {
    nsColorMap[ns] = nsClasses[Object.keys(nsColorMap).length % nsClasses.length];
  }
  return nsColorMap[ns];
}

function formatTime(ts: string): string {
  try {
    const d = new Date(ts);
    const hh = String(d.getHours()).padStart(2, "0");
    const mm = String(d.getMinutes()).padStart(2, "0");
    const ss = String(d.getSeconds()).padStart(2, "0");
    const ms = String(d.getMilliseconds()).padStart(3, "0");
    return `${hh}:${mm}:${ss}.${ms}`;
  } catch {
    return ts;
  }
}

export default function LogPanel() {
  return (
    <div class="panel">
      <div class="section-header">Logs</div>
      <div class="log-list">
        <For
          each={[...state.logs].reverse()}
          fallback={<p class="empty-state">no log entries yet</p>}
        >
          {(log) => (
            <div class="log-entry">
              <span class="log-ts">{formatTime(log.timestamp)}</span>
              <span class="log-sep" />
              <span class={`log-ns ${getNsClass(log.namespace || "global")}`}>
                {log.namespace || "global"}
              </span>
              <span class="log-sep" />
              <span class="log-msg">{log.message}</span>
            </div>
          )}
        </For>
      </div>
    </div>
  );
}
