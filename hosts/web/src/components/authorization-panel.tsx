import { Show } from "solid-js";
import type { TokenRecord, TokenState } from "../app-types";
import { Button, DataTable, FieldLabel, Surface, TextArea, TextInput, emptyStateClass, monoTextClass, type TableColumn } from "./ui";

type AuthorizationPanelProps = {
  nickname: string;
  onNicknameInput: (value: string) => void;
  onSubmit: (event: SubmitEvent) => void;
  latestToken: TokenState;
  tokens: TokenRecord[];
};

const tokenColumns: readonly TableColumn<TokenRecord>[] = [
  {
    key: "nickname",
    header: "Nickname",
    render: (token) => token.nickname || "Untitled",
  },
  {
    key: "source",
    header: "Source",
    render: (token) => token.source,
  },
  {
    key: "created",
    header: "Created",
    class: `${monoTextClass} w-[180px]`,
    render: (token) => {
      const createdDate = new Date((token.created_date as unknown as string) || Date.now());
      return createdDate.toLocaleString();
    },
  },
  {
    key: "secret",
    header: "Secret",
    render: (token) => (
      <code class={`${monoTextClass} m-0 block max-w-full whitespace-pre-wrap break-all`}>{token.secret}</code>
    ),
  },
];

export function AuthorizationPanel(props: AuthorizationPanelProps) {
  return (
    <Surface title="Authorization" subtitle="Generate credentials for API and client sessions">
      <form class="grid grid-cols-[1fr_auto] items-center gap-2 max-[900px]:grid-cols-1" onSubmit={props.onSubmit}>
        <FieldLabel for="nickname-input">Nickname</FieldLabel>
        <TextInput
          id="nickname-input"
          value={props.nickname}
          onInput={(event) => props.onNicknameInput(event.currentTarget.value)}
          placeholder="service-worker-01"
          maxLength={64}
          autocomplete="off"
        />
        <Button type="submit" class="max-[900px]:w-full">
          Create Token
        </Button>
      </form>

      <Show when={props.latestToken}>
        {(token) => (
          <div class="mt-4">
            <div class="mb-2 text-xs text-[#5f5f5f]">
              New secret <span class="ml-1.5 font-semibold text-[#111111]">{token().source}</span>
            </div>
            <TextArea readonly value={token().secret} class={`${monoTextClass} min-h-[130px] resize-y`} />
          </div>
        )}
      </Show>

      <div class="mt-4">
        <div class="mb-2 text-xs text-[#5f5f5f]">
          Existing tokens <span class="ml-1.5 font-semibold text-[#111111]">{props.tokens.length}</span>
        </div>
        <Show when={props.tokens.length > 0} fallback={<div class={emptyStateClass}>No tokens found.</div>}>
          <DataTable columns={tokenColumns} rows={props.tokens} getRowId={(token) => token.secret} emptyLabel="No tokens found." />
        </Show>
      </div>
    </Surface>
  );
}
