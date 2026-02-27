import { For, Show } from "solid-js";
import type { RegisteredComponent, RegisteredTask } from "../app-types";
import { Surface, emptyStateClass, monoTextClass } from "./ui";

type OverviewPanelProps = {
  tasks: readonly RegisteredTask[];
  components: readonly RegisteredComponent[];
};

const joinClasses = (...parts: (string | false)[]) => parts.filter(Boolean).join(" ");

export function OverviewPanel(props: OverviewPanelProps) {
  return (
    <Surface title="Overview" subtitle="Registered tasks and components">
      <div class="grid grid-cols-2 gap-3 max-[900px]:grid-cols-1">
        <article class="overflow-hidden rounded-[10px] border border-[#e5e5e5] bg-[#fcfcfc]">
          <header class="flex items-center justify-between gap-2 border-b border-[#e5e5e5] bg-[#fafafa] px-3 py-2.5">
            <h3 class="m-0 text-xs uppercase tracking-[0.06em] text-[#5f5f5f]">Registered Tasks</h3>
            <span class={joinClasses("rounded-full border border-[#e5e5e5] px-2 py-0.5 text-[#5f5f5f]", monoTextClass)}>
              {props.tasks.length}
            </span>
          </header>

          <Show when={props.tasks.length > 0} fallback={<div class={emptyStateClass}>No registered tasks yet.</div>}>
            <ul class="m-0 list-none p-0">
              <For each={props.tasks}>
                {(task) => (
                  <li class="grid grid-cols-[auto_1fr_auto] items-center gap-2.5 border-b border-[#e5e5e5] px-3 py-2.5 last:border-b-0">
                    <span class="min-w-0 [overflow-wrap:anywhere] text-[13px]">{task.name}</span>
                    <code class={joinClasses("whitespace-pre-wrap text-right text-[11px] text-[#5f5f5f] [overflow-wrap:anywhere]", monoTextClass)}>
                      {task.url}
                    </code>
                  </li>
                )}
              </For>
            </ul>
          </Show>
        </article>

        <article class="overflow-hidden rounded-[10px] border border-[#e5e5e5] bg-[#fcfcfc]">
          <header class="flex items-center justify-between gap-2 border-b border-[#e5e5e5] bg-[#fafafa] px-3 py-2.5">
            <h3 class="m-0 text-xs uppercase tracking-[0.06em] text-[#5f5f5f]">Registered Components</h3>
            <span class={joinClasses("rounded-full border border-[#e5e5e5] px-2 py-0.5 text-[#5f5f5f]", monoTextClass)}>
              {props.components.length}
            </span>
          </header>

          <Show
            when={props.components.length > 0}
            fallback={<div class={emptyStateClass}>No registered components yet.</div>}
          >
            <ul class="m-0 list-none p-0">
              <For each={props.components}>
                {(component) => (
                  <li class="grid grid-cols-[auto_1fr_auto] items-center gap-2.5 border-b border-[#e5e5e5] px-3 py-2.5 last:border-b-0">
                    <span
                      class={joinClasses(
                        "h-[11px] w-[11px] rounded-full shadow-[0_0_0_1px_rgba(0,0,0,0.1)]",
                        component.isHealthy ? "bg-[#21b35c]" : "bg-[#dd3e4d]",
                      )}
                      aria-hidden="true"
                    />
                    <span class="min-w-0 [overflow-wrap:anywhere] text-[13px]">{component.name}</span>
                    <code class={joinClasses("whitespace-pre-wrap text-right text-[11px] text-[#5f5f5f] [overflow-wrap:anywhere]", monoTextClass)}>
                      {component.version}
                    </code>
                  </li>
                )}
              </For>
            </ul>
          </Show>
        </article>
      </div>
    </Surface>
  );
}
