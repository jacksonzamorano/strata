import { For, Show, createSignal, onCleanup, splitProps, type JSX } from "solid-js";

export type TabOption<T extends string> = {
  key: T;
  label: string;
  count?: number;
};

export type TableColumn<T> = {
  key: string;
  header: string;
  class?: string;
  render: (row: T) => JSX.Element;
};

type SurfaceProps = {
  title: string;
  subtitle?: string;
  class?: string;
  children: JSX.Element;
  actions?: JSX.Element;
};

const joinClasses = (...parts: (string | undefined)[]) => parts.filter(Boolean).join(" ");

const pillClass = "rounded-full border border-[#e5e5e5] px-[10px] py-1.5 text-xs font-semibold";
const inputClass =
  "w-full rounded-[10px] border border-[#e5e5e5] bg-white px-3 py-2.5 text-sm text-[#111111] focus:border-[#111111] focus:outline-none focus:shadow-[0_0_0_1px_#111111]";

export const monoTextClass = "font-mono text-xs";
export const emptyStateClass = "px-[22px] py-[22px] text-center text-[#5f5f5f]";

export function Shell(props: { children: JSX.Element }) {
  return <main class="mx-auto grid max-w-[1080px] gap-4 p-[28px] max-[900px]:p-4">{props.children}</main>;
}

export function AppHeader(props: { eyebrow: string; title: string; meta: JSX.Element }) {
  return (
    <header class="flex flex-wrap items-start justify-between gap-3 rounded-[12px] border border-[#e5e5e5] bg-white p-4 max-[640px]:p-3">
      <div>
        <div class="text-[11px] uppercase tracking-[0.1em] text-[#5f5f5f]">{props.eyebrow}</div>
        <h1 class="m-0 mt-[6px] text-[clamp(1.5rem,3vw,2rem)] leading-[1.1] tracking-[-0.02em]">{props.title}</h1>
      </div>
      <div class="flex flex-wrap items-center gap-2">{props.meta}</div>
    </header>
  );
}

export function Surface(props: SurfaceProps) {
  return (
    <section class={joinClasses("rounded-[12px] border border-[#e5e5e5] bg-white p-4 max-[640px]:p-3", props.class)}>
      <div class="mb-4 flex justify-between gap-3">
        <div>
          <h2 class="m-0 text-[1.05rem] tracking-[-0.01em]">{props.title}</h2>
          <Show when={props.subtitle}>
            <p class="m-0 mt-1 text-xs text-[#5f5f5f]">{props.subtitle}</p>
          </Show>
        </div>
        <Show when={props.actions}>{props.actions}</Show>
      </div>
      {props.children}
    </section>
  );
}

export function StatusBadge(props: { online: boolean }) {
  return (
    <span
      class={joinClasses(
        pillClass,
        "uppercase tracking-[0.04em]",
        props.online ? "border-[#111111] bg-[#111111] text-white" : "bg-white text-[#111111]",
      )}
    >
      {props.online ? "Online" : "Offline"}
    </span>
  );
}

export function MetaChip(props: { children: JSX.Element; class?: string }) {
  return <span class={joinClasses(pillClass, "bg-[#fafafa] text-[#5f5f5f]", props.class)}>{props.children}</span>;
}

export function FieldLabel(props: { for: string; class?: string; children: JSX.Element }) {
  return (
    <label class={joinClasses("col-[1/-1] text-[11px] font-semibold uppercase tracking-[0.08em] text-[#5f5f5f]", props.class)} for={props.for}>
      {props.children}
    </label>
  );
}

export function TextInput(props: JSX.InputHTMLAttributes<HTMLInputElement>) {
  const [local, rest] = splitProps(props, ["class"]);
  return <input {...rest} class={joinClasses(inputClass, local.class)} />;
}

export function TextArea(props: JSX.TextareaHTMLAttributes<HTMLTextAreaElement>) {
  const [local, rest] = splitProps(props, ["class"]);
  return <textarea {...rest} class={joinClasses(inputClass, local.class)} />;
}

type ButtonProps = JSX.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "ghost";
  size?: "default" | "table";
  completeValue?: string;
  completeDurationMs?: number;
};

export function Button(
  props: ButtonProps,
) {
  const [local, rest] = splitProps(props, [
    "children",
    "class",
    "variant",
    "size",
    "onClick",
    "completeValue",
    "completeDurationMs",
  ]);
  const [completed, setCompleted] = createSignal(false);
  let completeTimer: number | undefined;

  const clearCompleteTimer = () => {
    if (completeTimer !== undefined) {
      window.clearTimeout(completeTimer);
      completeTimer = undefined;
    }
  };

  const startCompleteState = () => {
    if (!local.completeValue) {
      return;
    }

    setCompleted(true);
    clearCompleteTimer();
    completeTimer = window.setTimeout(() => {
      setCompleted(false);
      completeTimer = undefined;
    }, local.completeDurationMs ?? 3000);
  };

  onCleanup(clearCompleteTimer);

  const handleClick: JSX.EventHandler<HTMLButtonElement, MouseEvent> = (event) => {
    startCompleteState();
    const clickHandler = local.onClick as JSX.EventHandler<HTMLButtonElement, MouseEvent> | undefined;
    clickHandler?.(event);
  };

  return (
    <button
      {...rest}
      onClick={handleClick}
      class={joinClasses(
        "inline-flex cursor-pointer items-center justify-center font-semibold",
        local.size === "table" ? "h-7 rounded-[8px] px-2.5 text-[11px] leading-[1.2] whitespace-nowrap" : "rounded-[10px] px-[14px] py-2.5 text-[13px]",
        local.variant === "ghost"
          ? joinClasses(
              "border border-[#e5e5e5] bg-white text-[#5f5f5f] transition-colors duration-200",
              local.completeValue && completed() ? "border-[#d8d8d8] bg-[#f7f7f7] text-[#111111]" : undefined,
            )
          : "border border-[#111111] bg-[#111111] text-white hover:bg-[#2a2a2a]",
        local.class,
      )}
    >
      <Show
        when={local.completeValue}
        fallback={local.children}
      >
        <span class={joinClasses("relative inline-block text-center", local.size === "table" ? "min-w-[44px]" : "min-w-[68px]")}>
          <span class={joinClasses("block transition-all duration-200", completed() ? "-translate-y-0.5 opacity-0" : "translate-y-0 opacity-100")}>
            {local.children}
          </span>
          <span
            class={joinClasses(
              "pointer-events-none absolute inset-0 block transition-all duration-200",
              completed() ? "translate-y-0 opacity-100" : "translate-y-0.5 opacity-0",
            )}
          >
            {local.completeValue}
          </span>
        </span>
      </Show>
    </button>
  );
}

export function TabButton(props: {
  selected: boolean;
  label: string;
  count?: number;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      class={joinClasses(
        "inline-flex cursor-pointer items-center justify-center gap-2 rounded-[9px] bg-transparent px-[10px] py-[9px] text-[13px] font-semibold text-[#5f5f5f] hover:bg-[#f2f2f2]",
        props.selected ? "bg-white text-[#111111] shadow-[inset_0_0_0_1px_#e5e5e5]" : undefined,
      )}
      aria-pressed={props.selected}
      onClick={props.onClick}
    >
      <span>{props.label}</span>
      <Show when={typeof props.count === "number"}>
        <span class="min-w-[24px] rounded-full bg-[#111111] px-2 py-0.5 text-[11px] leading-[1.3] text-white">
          {props.count}
        </span>
      </Show>
    </button>
  );
}

export function TabBar<T extends string>(props: {
  value: T;
  options: readonly TabOption<T>[];
  onChange: (key: T) => void;
}) {
  const gridClass =
    props.options.length <= 1
      ? "grid-cols-1"
      : props.options.length === 2
        ? "grid-cols-2"
        : props.options.length === 3
          ? "grid-cols-3"
          : props.options.length === 4
            ? "grid-cols-4"
            : "grid-cols-5";

  return (
    <nav
      class={joinClasses(
        "grid gap-1 rounded-[12px] border border-[#e5e5e5] bg-[#fbfbfb] p-1 max-[640px]:grid-cols-1",
        gridClass,
      )}
      aria-label="Sections"
    >
      <For each={props.options}>
        {(option) => <TabButton selected={props.value === option.key} label={option.label} count={option.count} onClick={() => props.onChange(option.key)} />}
      </For>
    </nav>
  );
}

export function StatTile(props: { label: string; value: JSX.Element; note?: string }) {
  return (
    <article class="rounded-[10px] border border-[#e5e5e5] bg-[#fcfcfc] p-3">
      <div class="text-[11px] uppercase tracking-[0.06em] text-[#5f5f5f]">{props.label}</div>
      <div class="mt-2 text-base font-semibold text-[#111111]">{props.value}</div>
      <Show when={props.note}>
        <div class="mt-1 text-xs text-[#5f5f5f]">{props.note}</div>
      </Show>
    </article>
  );
}

export function DataTable<T>(props: {
  columns: readonly TableColumn<T>[];
  rows: readonly T[];
  getRowId: (row: T, index: number) => string;
  rowClass?: (row: T) => string | undefined;
  emptyLabel?: string;
  tableClass?: string;
  wrapClass?: string;
  cellVerticalAlign?: "center" | "top";
}) {
  const cellAlignClass = props.cellVerticalAlign === "top" ? "align-top" : "align-middle";

  return (
    <div class={joinClasses("max-h-[min(62vh,640px)] overflow-auto rounded-[10px] border border-[#e5e5e5]", props.wrapClass)}>
      <table class={joinClasses("w-full min-w-[780px] border-collapse max-[640px]:min-w-[640px]", props.tableClass)}>
        <thead>
          <tr>
            <For each={props.columns}>
              {(column) => (
                <th
                  class={joinClasses(
                    "sticky top-0 z-[1] border-b border-[#e5e5e5] bg-[#fafafa] px-3 py-2.5 text-left text-[11px] uppercase tracking-[0.06em] text-[#5f5f5f]",
                    column.class,
                  )}
                >
                  {column.header}
                </th>
              )}
            </For>
          </tr>
        </thead>
        <tbody>
          <Show
            when={props.rows.length > 0}
            fallback={
              <tr>
                <td class={emptyStateClass} colSpan={props.columns.length}>
                  {props.emptyLabel ?? "No data"}
                </td>
              </tr>
            }
          >
            <For each={props.rows}>
              {(row, index) => (
                <tr
                  class={joinClasses("hover:bg-[#f7f7f7]", props.rowClass?.(row))}
                  data-rowid={props.getRowId(row, index())}
                >
                  <For each={props.columns}>
                    {(column) => (
                      <td class={joinClasses("border-b border-[#e5e5e5] px-3 py-2.5 text-left text-xs text-[#242424]", cellAlignClass, column.class)}>
                        {column.render(row)}
                      </td>
                    )}
                  </For>
                </tr>
              )}
            </For>
          </Show>
        </tbody>
      </table>
    </div>
  );
}
