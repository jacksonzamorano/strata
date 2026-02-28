import { Show } from "solid-js";
import type { TokenRecord } from "../app-types";
import { Button, DataTable, FieldLabel, Surface, TextInput, emptyStateClass, monoTextClass, type TableColumn } from "./ui";

type AuthorizationPanelProps = {
  nickname: string;
  onNicknameInput: (value: string) => void;
  onSubmit: (event: SubmitEvent) => void;
  tokens: TokenRecord[];
};

const visibleSecretChars = 8;

function maskSecret(secret: string) {
  if (secret.length <= visibleSecretChars * 2) {
    return secret;
  }

  return `${secret.slice(0, visibleSecretChars)}...${secret.slice(-visibleSecretChars)}`;
}

async function copySecret(secret: string) {
  if (!secret) {
    return false;
  }

  try {
    await navigator.clipboard.writeText(secret);
    return true;
  } catch {
    const textarea = document.createElement("textarea");
    textarea.value = secret;
    textarea.setAttribute("readonly", "true");
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    document.body.append(textarea);
    textarea.select();
    try {
      return document.execCommand("copy");
    } finally {
      textarea.remove();
    }
  }
}

function SecretDisplay(props: { secret: string; class?: string }) {
  const handleCopy = () => {
    void copySecret(props.secret);
  };

  return (
    <div class={`flex min-w-0 items-center gap-1.5 [&>button]:shrink-0 max-[640px]:flex-col max-[640px]:items-stretch max-[640px]:[&>button]:w-full ${props.class ?? ""}`}>
      <code class={`${monoTextClass} m-0 block min-w-0 flex-1 whitespace-pre-wrap break-all`}>{maskSecret(props.secret)}</code>
      <Button
        type="button"
        variant="ghost"
        size="table"
        completeValue="Copied"
        onClick={handleCopy}
      >
        Copy
      </Button>
    </div>
  );
}

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
    render: (token) => <SecretDisplay secret={token.secret} />,
  },
];

export function AuthorizationPanel(props: AuthorizationPanelProps) {
  return (
    <Surface title="Authorization" subtitle="Manage API keys.">
      <form class="grid grid-cols-[1fr_auto] items-center gap-2 max-[900px]:grid-cols-1 max-[900px]:[&>button]:w-full" onSubmit={props.onSubmit}>
        <FieldLabel for="nickname-input">Nickname</FieldLabel>
        <TextInput
          id="nickname-input"
          value={props.nickname}
          onInput={(event) => props.onNicknameInput(event.currentTarget.value)}
          placeholder="service-worker-01"
          maxLength={64}
          autocomplete="off"
        />
        <Button type="submit">
          Create Token
        </Button>
      </form>

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
