import type { RequestRecord, RequestState } from "../app-types";
import { DataTable, Surface, monoTextClass, type TableColumn } from "./ui";

type RequestsPanelProps = {
  requests: readonly RequestRecord[];
};

const stateLabel: Record<RequestState, string> = {
  in_progress: "In Progress",
  succeeded: "Succeeded",
  failed: "Failed",
  unknown: "Unknown",
};

const stateClass: Record<RequestState, string> = {
  in_progress: "border-[#e5e5e5] bg-white text-[#111111]",
  succeeded: "border-[#21b35c] bg-[#effcf3] text-[#10753a]",
  failed: "border-[#dd3e4d] bg-[#fff1f2] text-[#a51f2d]",
  unknown: "border-[#e5e5e5] bg-[#fafafa] text-[#5f5f5f]",
};

const rowClassByState: Partial<Record<RequestState, string>> = {
  failed: "[&>td:first-child]:border-l-[3px] [&>td:first-child]:border-l-[#dd3e4d]",
  succeeded: "[&>td:first-child]:border-l-[3px] [&>td:first-child]:border-l-[#21b35c]",
};

const renderTimestamp = (row: RequestRecord) => {
  const date = row.startedAt ?? row.finishedAt;
  return date ? date.toLocaleString() : "-";
};

const renderDuration = (value: number | undefined) => {
  if (typeof value !== "number" || Number.isNaN(value)) {
    return "-";
  }
  if (value < 1) {
    return `${Math.round(value * 1000)}ms`;
  }
  return `${value.toFixed(2)}s`;
};

const requestColumns: readonly TableColumn<RequestRecord>[] = [
  {
    key: "time",
    header: "Time",
    class: `${monoTextClass} w-[190px]`,
    render: (row) => renderTimestamp(row),
  },
  {
    key: "task",
    header: "Task",
    render: (row) => row.taskName,
  },
  {
    key: "method",
    header: "Method",
    class: `${monoTextClass} w-[90px]`,
    render: (row) => row.method,
  },
  {
    key: "path",
    header: "Path",
    class: monoTextClass,
    render: (row) => row.path,
  },
  {
    key: "status",
    header: "Status",
    class: "w-[130px]",
    render: (row) => (
      <span class={`inline-flex rounded-full border px-2 py-1 text-[11px] font-semibold ${stateClass[row.state]}`}>
        {stateLabel[row.state]}
      </span>
    ),
  },
  {
    key: "code",
    header: "Code",
    class: `${monoTextClass} w-[90px]`,
    render: (row) => (typeof row.statusCode === "number" ? row.statusCode : "-"),
  },
  {
    key: "duration",
    header: "Duration",
    class: `${monoTextClass} w-[110px]`,
    render: (row) => renderDuration(row.durationSeconds),
  },
];

export function RequestsPanel(props: RequestsPanelProps) {
  return (
    <Surface title="Requests" subtitle="Task request history.">
      <DataTable
        columns={requestColumns}
        rows={props.requests}
        emptyLabel="No requests yet. Waiting for task events..."
        getRowId={(row) => row.id}
        rowClass={(row) => rowClassByState[row.state]}
      />
    </Surface>
  );
}
