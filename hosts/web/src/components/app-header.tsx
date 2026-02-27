import { AppHeader, MetaChip, StatusBadge, monoTextClass } from "./ui";

type HostHeaderProps = {
  connected: boolean;
  status: string;
  lineCount: number;
};

export function HostHeader(props: HostHeaderProps) {
  return (
    <AppHeader
      eyebrow="tasklib"
      title="Control Plane"
      meta={
        <>
          <StatusBadge online={props.connected} />
          <MetaChip>{props.status}</MetaChip>
          <MetaChip class={monoTextClass}>{props.lineCount} logs</MetaChip>
        </>
      }
    />
  );
}
