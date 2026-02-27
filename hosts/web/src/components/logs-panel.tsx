import type { LogRecord } from "../app-types";
import { DataTable, Surface, monoTextClass, type TableColumn } from "./ui";

type LogsPanelProps = {
  logs: readonly LogRecord[];
};

const rowClassByKind: Record<string, string> = {
  error: "[&>td:first-child]:font-semibold [&>td:first-child]:underline",
  container: "[&>td:first-child]:border-l-[3px] [&>td:first-child]:border-l-[#111111]",
};

const logColumns: readonly TableColumn<LogRecord>[] = [
  {
    key: "date",
    header: "Time",
    class: `${monoTextClass} w-[110px]`,
    render: (row) => row.date,
  },
  {
    key: "channel",
    header: "Channel",
    class: monoTextClass,
    render: (row) => row.channel,
  },
  {
    key: "namespace",
    header: "Namespace",
    class: monoTextClass,
    render: (row) => row.namespace ?? "-",
  },
  {
    key: "kind",
    header: "Kind",
    class: monoTextClass,
    render: (row) => row.kind,
  },
  {
    key: "message",
    header: "Message",
    render: (row) => row.message,
  },
];

export function LogsPanel(props: LogsPanelProps) {
  return (
    <Surface title="Logs" subtitle="Realtime event stream from the host bus">
      <DataTable
        columns={logColumns}
        rows={props.logs}
        emptyLabel="No logs yet. Waiting for events..."
        getRowId={(row) => row.id}
        rowClass={(row) => rowClassByKind[row.kind]}
      />
    </Surface>
  );
}
