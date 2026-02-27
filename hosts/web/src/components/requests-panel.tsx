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

const stateLightClass: Record<RequestState, string> = {
  in_progress: "bg-[#f59e0b]",
  succeeded: "bg-[#21b35c]",
  failed: "bg-[#dd3e4d]",
  unknown: "bg-[#c4c4c4]",
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
    class: "min-w-[220px]",
    render: (row) => row.taskName,
  },
  {
    key: "status",
    header: "Status",
    class: "w-[88px]",
    render: (row) => (
      <span class="inline-flex items-center justify-center">
        <span
          class={`h-[11px] w-[11px] rounded-full shadow-[0_0_0_1px_rgba(0,0,0,0.1)] ${stateLightClass[row.state]}`}
          aria-label={stateLabel[row.state]}
          title={stateLabel[row.state]}
        />
      </span>
    ),
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
        tableClass="min-w-[520px]"
      />
    </Surface>
  );
}
